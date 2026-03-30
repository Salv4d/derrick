package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "node -v"},
		},
		{
			name:     "Complex command with single quotes",
			command:  "echo 'Hello World'",
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "echo 'Hello World'"},
		},
		{
			name:     "Command with pipes and logical operators",
			command:  "test -f .env || echo 'Missing'",
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "test -f .env || echo 'Missing'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithNix(tt.command)

			assert.Equal(t, tt.expected, result, "The generated Nix command should match the expected output")
		})
	}
}

func TestEnsureNixEnvironment(t *testing.T) {
	tempDir := t.TempDir()

	originalWD, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer os.Chdir(originalWD)

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change working directory")

	mockPackages := []string{"golang", "python3"}
	customRegistry := "github:NixOS/nixpkgs/nixos-22.11"

	err = EnsureNixEnvironment("derrick.yaml", mockPackages, customRegistry)
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
		assert.Contains(t, contentStr, pkg, "The generated flake should contain the injected package")
	}
}
