package cli_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI verifies the command-line interface behavior, including help output
// and error handling when derrick.yaml is missing.
func TestCLI(t *testing.T) {
	t.Run("Root Command (--help)", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../cmd/derrick", "--help")

		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output

		err := cmd.Run()
		require.NoError(t, err, "Running the CLI with --help should not fail")

		outStr := output.String()
		assert.Contains(t, outStr, "Derrick", "The help menu should print the application name")
		assert.Contains(t, outStr, "Available Commands:", "The help menu should list available commands")
		assert.Contains(t, outStr, "doctor", "The doctor command should be listed")
	})

	t.Run("Doctor Command (Missing YAML)", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../cmd/derrick", "doctor")

		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output

		err := cmd.Run()
		require.Error(t, err, "Doctor should fail if derrick.yaml is missing")

		outStr := output.String()
		assert.Contains(t, outStr, "CRITICAL ERROR", "FailFast should trigger the critical error UI")
		assert.Contains(t, outStr, "no such file or directory", "It should report the missing file")
	})
}
