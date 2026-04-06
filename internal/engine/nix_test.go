package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Salv4d/derrick/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrapWithNix verifies that shell commands are correctly wrapped with
// nix develop arguments for different command patterns.
func TestWrapWithNix(t *testing.T) {
	absPath, err := filepath.Abs(".derrick")
	require.NoError(t, err, "Failed to resolve absolute path")

	expectedFlakePath := "path:" + absPath + "#default"

	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "Simple single word command",
			command:  "node -v",
			expected: []string{"nix", "develop", "--impure", expectedFlakePath, "-c", "bash", "-c", "node -v"},
		},
		{
			name:     "Complex command with single quotes",
			command:  "echo 'Hello World'",
			expected: []string{"nix", "develop", "--impure", expectedFlakePath, "-c", "bash", "-c", "echo 'Hello World'"},
		},
		{
			name:     "Command with pipes and logical operators",
			command:  "test -f .env || echo 'Missing'",
			expected: []string{"nix", "develop", "--impure", expectedFlakePath, "-c", "bash", "-c", "test -f .env || echo 'Missing'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithNix(tt.command, "")

			assert.Equal(t, tt.expected, result, "The generated Nix command should match the expected output")
		})
	}
}

// TestEnsureNixEnvironment verifies that a Nix flake is generated with the
// correct packages and custom registry in the .derrick directory.
func TestEnsureNixEnvironment(t *testing.T) {
	tempDir := t.TempDir()

	originalWD, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer os.Chdir(originalWD)

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change working directory")

	mockPackages := []config.NixPackage{{Name: "golang"}, {Name: "python3"}}
	customRegistry := "github:NixOS/nixpkgs/nixos-22.11"

	err = EnsureNixEnvironment("derrick.yaml", mockPackages, customRegistry, "")
	assert.NoError(t, err, "EnsureNixEnvironment should not return an error")

	derrickDir := ".derrick"
	assert.DirExists(t, derrickDir, "The .derrick hidden directory should be created")

	flakePath := filepath.Join(derrickDir, "flake.nix")
	assert.FileExists(t, flakePath, "The flake.nix file should be created")

	contentBytes, err := os.ReadFile(flakePath)
	require.NoError(t, err, "Should be able to read the generated flake.nix")

	contentStr := string(contentBytes)

	assert.Contains(t, contentStr, customRegistry, "The generated flake should contain the custom registry URL")
	for _, pkg := range mockPackages {
		assert.Contains(t, contentStr, pkg.Name, "The generated flake should contain the injected package")
	}
}
