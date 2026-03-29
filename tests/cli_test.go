package cli_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	t.Run("Root Command (--help)", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../cmd/derrick/main.go", "--help")

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		require.NoError(t, err, "Running the CLI with --help should not fail")

		output := stdout.String()
		assert.Contains(t, output, "Derrick", "The help menu should print the application name")
		assert.Contains(t, output, "Available Commands:", "The help menu should list available commands")
		assert.Contains(t, output, "doctor", "The doctor command should be listed")
	})

	t.Run("Doctor Command (Missing YAML)", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../cmd/derrick/main.go", "doctor")
		
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		require.Error(t, err, "Doctor should fail if derrick.yaml is missing")

		output := stdout.String()
		assert.Contains(t, output, "CRITICAL ERROR", "FailFast should trigger the critical error UI")
		assert.Contains(t, output, "no such file or directory", "It should report the missing file")
	})
}