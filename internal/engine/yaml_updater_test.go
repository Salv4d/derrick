package engine

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateYAMLPackage(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()
	defer os.Chdir(originalWD)
	os.Chdir(tempDir)

	initialYAML := `name: test
nix:
  packages:
    - "python3"
    - "nodejs"
`
	err := os.WriteFile("derrick.yaml", []byte(initialYAML), 0o644)
	require.NoError(t, err)

	t.Run("replaces quoted package name", func(t *testing.T) {
		err := UpdateYAMLPackage("derrick.yaml", "python3", "python310")
		require.NoError(t, err)

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), `"python310"`)
		assert.NotContains(t, string(content), `"python3"`)
		assert.Contains(t, string(content), `"nodejs"`, "unrelated package must be untouched")
	})
}

func TestUpdateYAMLRegistry(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()
	defer os.Chdir(originalWD)
	os.Chdir(tempDir)

	t.Run("inserts registry when none exists", func(t *testing.T) {
		yaml := `name: test
nix:
  packages:
    - "go"
`
		require.NoError(t, os.WriteFile("derrick.yaml", []byte(yaml), 0o644))

		require.NoError(t, UpdateYAMLRegistry("derrick.yaml", "github:NixOS/nixpkgs/nixos-22.11"))

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), `registry: "github:NixOS/nixpkgs/nixos-22.11"`)
	})

	t.Run("replaces existing registry", func(t *testing.T) {
		yaml := `name: test
nix:
  registry: "old-registry"
  packages:
    - "go"
`
		require.NoError(t, os.WriteFile("derrick.yaml", []byte(yaml), 0o644))

		require.NoError(t, UpdateYAMLRegistry("derrick.yaml", "new-registry"))

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), `registry: "new-registry"`)
		assert.NotContains(t, string(content), "old-registry")
	})
}
