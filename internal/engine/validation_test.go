package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExecuteCommand verifies that shell commands are executed correctly,
// including handling of valid commands, non-existent commands, and failing commands.
func TestExecuteCommand(t *testing.T) {
	t.Run("Valid Bash Command", func(t *testing.T) {
		err := executeCommand("true", false, nil)

		assert.NoError(t, err, "Executing a valid command should not return an error")
	})

	t.Run("Invalid Bash Command", func(t *testing.T) {
		err := executeCommand("this_command_does_not_exist_123", false, nil)

		assert.Error(t, err, "Executing a non-existent command should return an error")
		assert.Contains(t, err.Error(), "not found", "The error message should contain the shell stderr output")
	})

	t.Run("Command that fails intentionally", func(t *testing.T) {
		err := executeCommand("ls /directory_that_will_never_exist", false, nil)

		assert.Error(t, err, "Executing a command that fails should return an error")
		assert.Contains(t, err.Error(), "No such file or directory", "The error message should contain the exact OS failure reason")
	})
}
