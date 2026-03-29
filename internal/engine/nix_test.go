package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWrapWithNix(t *testing.T) {
	absPath, err := filepath.Abs(".derrick")
	if err != nil {
		t.Fatalf("Failed to resolve absolute path: %v", err)
	}
	expectedFlakePath := "path:" + absPath + "#default"

	tests := []struct {
		name string
		command string
		expected []string
	}{
		{
			name: "Simple single word command",
			command: "node -v",
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "node -v"},
		},
		{
			name: "Complex command with single quotes",
			command: "echo 'Hello World'",
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "echo 'Hello World'"},
		},
		{
			name: "Command with pipes and logical operators",
			command: "test -f .env || echo 'Missing'",
			expected: []string{"nix", "develop", expectedFlakePath, "-c", "bash", "-c", "test -f .env || echo 'Missing'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithNix(tt.command)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("\nExpected: %v\nGot:      %v", tt.expected, result)
			}
		})
	}
}

func TestEnsureNixEnvironment(t *testing.T) {
	tempDir := t.TempDir()

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	defer os.Chdir(originalWD)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	mockPackages := []string{"golang", "python3"}
	err = EnsureNixEnvironment(mockPackages)
	if err != nil {
		t.Fatalf("EnsureNixEnvironment failed unexpectedly: %v", err)
	}

	derrickDir := ".derrick"
	if _, err := os.Stat(derrickDir); os.IsNotExist(err) {
		t.Errorf("Expected directory '%s' to be created, but it was not found", derrickDir)
	}

	flakePath := filepath.Join(derrickDir, "flake.nix")
	contentBytes, err := os.ReadFile(flakePath)
	if err != nil {
		t.Fatalf("Failed to read generated flake.nix: %v", err)
	}

	contentStr := string(contentBytes)

	for _, pkg := range mockPackages {
		if !strings.Contains(contentStr, pkg) {
			t.Errorf("Generated flake.nix is missing the required package: %s", pkg)
		}
	}
}