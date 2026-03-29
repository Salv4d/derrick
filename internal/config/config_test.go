package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	tempDir := t.TempDir()

	validYAML := []byte(`
name: "test-project"
version: "1.0.0"
dependencies:
  nix_packages:
    - "go"
    - "nodejs"
  docker_file: "docker-compose.yml"
hooks:
  pre_init:
    - "echo 'Starting'"
validations:
  - name: "Check Env"
    command: "test -f .env"
`)
	validFilePath := filepath.Join(tempDir, "valid.yaml")
	err := os.WriteFile(validFilePath, validYAML, 0644)

	require.NoError(t, err, "Failed to write valid YAML test file")

	t.Run("Valid Config", func(t *testing.T) {
		cfg, err := ParseConfig(validFilePath)
		require.NoError(t, err, "Parsing a valid YAML file should not return an error")

		assert.Equal(t, "test-project", cfg.Name, "Project name should match")
		assert.Equal(t, "1.0.0", cfg.Version, "Project version should match")

		assert.Len(t, cfg.Dependencies.NixPackages, 2, "Should parse exactly 2 nix packages")
		assert.Equal(t, "go", cfg.Dependencies.NixPackages[0], "First nix package should be 'go'")
		
		assert.Len(t, cfg.Hooks.PreInit, 1, "Should parse exactly 1 pre_init hook")
		assert.Equal(t, "echo 'Starting'", cfg.Hooks.PreInit[0], "The hook command should match")
		
		assert.Len(t, cfg.Validations, 1, "Should parse exactly 1 validation")
		assert.Equal(t, "Check Env", cfg.Validations[0].Name, "Validation name should match")
	})

	invalidYAML := []byte(`
name: "test-project"
version: "1.0.0"
dependencies:
	nix_packages:
	- "go"
		- "bad-identation"
`)

	invalidFilePath := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFilePath, invalidYAML, 0644)
	require.NoError(t, err, "Failed to write invalid yaml test file")

	t.Run("Malformed Config", func(t *testing.T) {
		_, err := ParseConfig(invalidFilePath)

		assert.Error(t, err, "Parsing a malformed YAML file should return an error")
	})

	t.Run("Missing File", func(t *testing.T) {
		missingFilePath := filepath.Join(tempDir, "does_not_exist.yaml")
		_, err := ParseConfig(missingFilePath)

		assert.Error(t, err, "Attempting to parse a non-existing file should return an error")
	})
}