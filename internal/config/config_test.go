package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseConfig verifies that derrick.yaml configuration files are parsed correctly
// including valid configs, custom registries, malformed YAML, missing files, and profile extensions.
func TestParseConfig(t *testing.T) {
	tempDir := t.TempDir()

	validYAML := []byte(`
name: "test-project"
version: "1.0.0"
dependencies:
  nix_packages:
    - "go"
    - "nodejs"
  docker_compose: "docker-compose.yml"
hooks:
  pre_init:
    - "echo 'Starting'"
validations:
  - name: "Check Env"
    command: "test -f .env"
`)
	validFilePath := filepath.Join(tempDir, "valid.yaml")
	err := os.WriteFile(validFilePath, validYAML, 0o644)

	require.NoError(t, err, "Failed to write valid YAML test file")

	t.Run("Valid Config", func(t *testing.T) {
		cfg, err := ParseConfig(validFilePath, "")
		require.NoError(t, err, "Parsing a valid YAML file should not return an error")

		assert.Equal(t, "test-project", cfg.Name, "Project name should match")
		assert.Equal(t, "1.0.0", cfg.Version, "Project version should match")

		assert.Len(t, cfg.Dependencies.NixPackages, 2, "Should parse exactly 2 nix packages")
		assert.Equal(t, "go", cfg.Dependencies.NixPackages[0].Name, "First nix package should be 'go'")

		assert.Equal(t, DefaultNixRegistry, cfg.Dependencies.NixRegistry, "Should use default nix registry when missing")
		assert.Equal(t, "docker-compose.yml", cfg.Dependencies.DockerCompose, "Should use default docker-compose.yml when missing")

		assert.Len(t, cfg.Hooks.PreInit, 1, "Should parse exactly 1 pre_init hook")
		assert.Equal(t, "echo 'Starting'", cfg.Hooks.PreInit[0], "The hook command should match")

		assert.Len(t, cfg.Validations, 1, "Should parse exactly 1 validation")
		assert.Equal(t, "Check Env", cfg.Validations[0].Name, "Validation name should match")
	})

	t.Run("Custom Registry", func(t *testing.T) {
		customYAML := []byte(`
name: "custom-registry"
version: "1.0.0"
dependencies:
  nix_registry: "github:NixOS/nixpkgs/nixos-22.11"
  nix_packages: ["go"]
`)
		customPath := filepath.Join(tempDir, "custom.yaml")
		err := os.WriteFile(customPath, customYAML, 0o644)
		require.NoError(t, err)

		cfg, err := ParseConfig(customPath, "")
		require.NoError(t, err)

		assert.Equal(t, "github:NixOS/nixpkgs/nixos-22.11", cfg.Dependencies.NixRegistry)
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
	err = os.WriteFile(invalidFilePath, invalidYAML, 0o644)
	require.NoError(t, err, "Failed to write invalid yaml test file")

	t.Run("Malformed Config", func(t *testing.T) {
		_, err := ParseConfig(invalidFilePath, "")

		assert.Error(t, err, "Parsing a malformed YAML file should return an error")
	})

	t.Run("Missing File", func(t *testing.T) {
		missingFilePath := filepath.Join(tempDir, "does_not_exist.yaml")
		_, err := ParseConfig(missingFilePath, "")

		assert.Error(t, err, "Attempting to parse a non-existing file should return an error")
	})
	t.Run("Profile Extension", func(t *testing.T) {
		profileYAML := []byte(`
name: "profile-test"
version: "1.0.0"
dependencies:
  nix_packages:
    - "go"
profiles:
  base-worker:
    dependencies:
      docker_compose_profiles: ["cache"]
      nix_packages:
        - "redis"
  advanced-worker:
    extend: "base-worker"
    dependencies:
      docker_compose_profiles: ["worker"]
      nix_packages:
        - "python3"
    hooks:
      pre_start:
        - "echo 'Starting advanced'"
`)
		profilePath := filepath.Join(tempDir, "profile.yaml")
		err := os.WriteFile(profilePath, profileYAML, 0o644)
		require.NoError(t, err)

		cfg, err := ParseConfig(profilePath, "advanced-worker")
		require.NoError(t, err, "Should parse extended profile perfectly")

		assert.Len(t, cfg.Dependencies.NixPackages, 3, "Should merge Root + Base + Advanced")
		expectedPkgs := []NixPackage{{Name: "go"}, {Name: "redis"}, {Name: "python3"}}
		assert.ElementsMatch(t, expectedPkgs, cfg.Dependencies.NixPackages)

		assert.Len(t, cfg.Dependencies.DockerComposeProfiles, 2, "Should accumulate compose profiles")
		assert.ElementsMatch(t, []string{"cache", "worker"}, cfg.Dependencies.DockerComposeProfiles)

		assert.Len(t, cfg.Hooks.PreStart, 1)
		assert.Equal(t, "echo 'Starting advanced'", cfg.Hooks.PreStart[0])
	})
}
