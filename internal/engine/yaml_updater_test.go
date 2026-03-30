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
dependencies:
  nix_packages:
    - "python3"
    - "nodejs"
`
	err := os.WriteFile("derrick.yaml", []byte(initialYAML), 0o644)
	require.NoError(t, err)

	t.Run("Update Package", func(t *testing.T) {
		err := UpdateYAMLPackage("derrick.yaml", "python3", "python310")
		require.NoError(t, err)

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), "- \"python310\"")
		assert.NotContains(t, string(content), "- \"python3\"")
	})
}

func TestUpdateYAMLRegistry(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, _ := os.Getwd()
	defer os.Chdir(originalWD)
	os.Chdir(tempDir)

	t.Run("Inject New Registry", func(t *testing.T) {
		initialYAML := `name: test
dependencies:
  nix_packages:
    - "go"
`
		os.WriteFile("derrick.yaml", []byte(initialYAML), 0o644)

		err := UpdateYAMLRegistry("derrick.yaml", "github:NixOS/nixpkgs/nixos-22.11")
		require.NoError(t, err)

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), "nix_registry: \"github:NixOS/nixpkgs/nixos-22.11\"")
	})

	t.Run("Update Existing Registry", func(t *testing.T) {
		initialYAML := `name: test
dependencies:
  nix_registry: "old-registry"
  nix_packages: ["go"]
`
		os.WriteFile("derrick.yaml", []byte(initialYAML), 0o644)

		err := UpdateYAMLRegistry("derrick.yaml", "new-registry")
		require.NoError(t, err)

		content, _ := os.ReadFile("derrick.yaml")
		assert.Contains(t, string(content), "nix_registry: \"new-registry\"")
		assert.NotContains(t, string(content), "old-registry")
	})
}
