package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseConfig(t *testing.T) {
	tempDir := t.TempDir()

	validYAML := []byte(`
name: "test-project"
version: "1.0.0"
provider: docker
docker:
  compose: "docker-compose.yml"
nix:
  packages:
    - "go"
    - "nodejs"
hooks:
  setup:
    - run: "echo 'Starting'"
      when: always
validations:
  - name: "Check Env"
    command: "test -f .env"
`)
	validPath := filepath.Join(tempDir, "valid.yaml")
	require.NoError(t, os.WriteFile(validPath, validYAML, 0o644))

	t.Run("Valid Config", func(t *testing.T) {
		cfg, err := ParseConfig(validPath, "")
		require.NoError(t, err)

		assert.Equal(t, "test-project", cfg.Name)
		assert.Equal(t, "1.0.0", cfg.Version)
		assert.Equal(t, "docker", cfg.Provider)

		assert.Len(t, cfg.Nix.Packages, 2)
		assert.Equal(t, "go", cfg.Nix.Packages[0].Name)
		assert.Equal(t, DefaultNixRegistry, cfg.Nix.Registry)
		assert.Equal(t, "docker-compose.yml", cfg.Docker.Compose)

		assert.Len(t, cfg.Hooks.Setup, 1)
		assert.Equal(t, "echo 'Starting'", cfg.Hooks.Setup[0].Run)
		assert.Equal(t, Condition{"always"}, cfg.Hooks.Setup[0].When)

		assert.Len(t, cfg.Validations, 1)
		assert.Equal(t, "Check Env", cfg.Validations[0].Name)
	})

	t.Run("Plain string hook", func(t *testing.T) {
		yamlData := []byte(`
name: "hook-test"
version: "1.0.0"
hooks:
  setup:
    - "echo hello"
`)
		p := filepath.Join(tempDir, "hook.yaml")
		require.NoError(t, os.WriteFile(p, yamlData, 0o644))

		cfg, err := ParseConfig(p, "")
		require.NoError(t, err)
		assert.Len(t, cfg.Hooks.Setup, 1)
		assert.Equal(t, "echo hello", cfg.Hooks.Setup[0].Run)
		assert.Equal(t, Condition{"always"}, cfg.Hooks.Setup[0].When)
	})

	t.Run("Custom Nix registry", func(t *testing.T) {
		yamlData := []byte(`
name: "custom-registry"
version: "1.0.0"
nix:
  registry: "github:NixOS/nixpkgs/nixos-22.11"
  packages:
    - "go"
`)
		p := filepath.Join(tempDir, "custom.yaml")
		require.NoError(t, os.WriteFile(p, yamlData, 0o644))

		cfg, err := ParseConfig(p, "")
		require.NoError(t, err)
		assert.Equal(t, "github:NixOS/nixpkgs/nixos-22.11", cfg.Nix.Registry)
	})

	t.Run("Malformed Config", func(t *testing.T) {
		badYAML := []byte(`
name: "test-project"
version: "1.0.0"
nix:
	packages:
	- "go"
		- "bad-indentation"
`)
		p := filepath.Join(tempDir, "invalid.yaml")
		require.NoError(t, os.WriteFile(p, badYAML, 0o644))

		_, err := ParseConfig(p, "")
		assert.Error(t, err)
	})

	t.Run("Missing File", func(t *testing.T) {
		_, err := ParseConfig(filepath.Join(tempDir, "does_not_exist.yaml"), "")
		assert.Error(t, err)
	})

	t.Run("Profile Extension", func(t *testing.T) {
		profileYAML := []byte(`
name: "profile-test"
version: "1.0.0"
nix:
  packages:
    - "go"
profiles:
  base-worker:
    docker:
      profiles: ["cache"]
    nix:
      packages:
        - "redis"
  advanced-worker:
    extend: "base-worker"
    docker:
      profiles: ["worker"]
    nix:
      packages:
        - "python3"
    hooks:
      setup:
        - run: "echo 'Starting advanced'"
          when: always
`)
		p := filepath.Join(tempDir, "profile.yaml")
		require.NoError(t, os.WriteFile(p, profileYAML, 0o644))

		cfg, err := ParseConfig(p, "advanced-worker")
		require.NoError(t, err)

		assert.Len(t, cfg.Nix.Packages, 3, "Should merge root + base + advanced")
		expected := []NixPackage{{Name: "go"}, {Name: "redis"}, {Name: "python3"}}
		assert.ElementsMatch(t, expected, cfg.Nix.Packages)

		assert.ElementsMatch(t, []string{"cache", "worker"}, cfg.Docker.Profiles)

		assert.Len(t, cfg.Hooks.Setup, 1)
		assert.Equal(t, "echo 'Starting advanced'", cfg.Hooks.Setup[0].Run)
		assert.Equal(t, Condition{"always"}, cfg.Hooks.Setup[0].When)
	})
}

func TestSchemaVersioning(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("missing schema accepted as legacy", func(t *testing.T) {
		p := filepath.Join(tempDir, "legacy.yaml")
		require.NoError(t, os.WriteFile(p, []byte(`name: "legacy"
version: "1.0.0"
`), 0o644))

		cfg, err := ParseConfig(p, "")
		require.NoError(t, err)
		assert.Equal(t, CurrentSchema, cfg.Schema, "legacy schema should be upgraded in memory")
	})

	t.Run("matching schema accepted", func(t *testing.T) {
		p := filepath.Join(tempDir, "current.yaml")
		require.NoError(t, os.WriteFile(p, []byte(fmt.Sprintf(`schema: %d
name: "current"
version: "1.0.0"
`, CurrentSchema)), 0o644))

		cfg, err := ParseConfig(p, "")
		require.NoError(t, err)
		assert.Equal(t, CurrentSchema, cfg.Schema)
	})

	t.Run("future schema rejected", func(t *testing.T) {
		p := filepath.Join(tempDir, "future.yaml")
		require.NoError(t, os.WriteFile(p, []byte(fmt.Sprintf(`schema: %d
name: "future"
version: "1.0.0"
`, CurrentSchema+1)), 0o644))

		_, err := ParseConfig(p, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "newer than this binary")
	})
}

func TestNixPackageMarshal(t *testing.T) {
	t.Run("plain string when no registry", func(t *testing.T) {
		pkgs := []NixPackage{{Name: "nodejs_20"}}
		out, err := yaml.Marshal(pkgs)
		require.NoError(t, err)
		assert.Contains(t, string(out), "nodejs_20")
		assert.NotContains(t, string(out), "package:")
	})

	t.Run("struct form when registry is set", func(t *testing.T) {
		pkgs := []NixPackage{{Name: "legacy_tool", Registry: "github:NixOS/nixpkgs/nixos-22.11"}}
		out, err := yaml.Marshal(pkgs)
		require.NoError(t, err)
		assert.Contains(t, string(out), "package: legacy_tool")
		assert.Contains(t, string(out), "registry:")
	})
}

func TestActiveProvider(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ProjectConfig
		expected string
	}{
		{"explicit docker", ProjectConfig{Provider: "docker"}, "docker"},
		{"explicit nix", ProjectConfig{Provider: "nix"}, "nix"},
		{"auto with compose", ProjectConfig{Provider: "auto", Docker: DockerConfig{Compose: "docker-compose.yml"}}, "docker"},
		{"auto with nix packages", ProjectConfig{Provider: "auto", Nix: NixConfig{Packages: []NixPackage{{Name: "go"}}}}, "nix"},
		{"empty defaults to nix", ProjectConfig{}, "nix"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cfg.ActiveProvider())
		})
	}
}
