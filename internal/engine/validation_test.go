package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRunner_Run verifies that shell commands are executed correctly via the Runner,
// including handling of valid commands, non-existent commands, and failing commands.
func TestRunner_Run(t *testing.T) {
	runner := &Runner{}

	t.Run("Valid Bash Command", func(t *testing.T) {
		err := runner.Run("true")

		assert.NoError(t, err, "Executing a valid command should not return an error")
	})

	t.Run("Invalid Bash Command", func(t *testing.T) {
		err := runner.Run("this_command_does_not_exist_123")

		assert.Error(t, err, "Executing a non-existent command should return an error")
		assert.Contains(t, err.Error(), "not found", "The error message should contain the shell stderr output")
	})

	t.Run("Command that fails intentionally", func(t *testing.T) {
		err := runner.Run("ls /directory_that_will_never_exist")

		assert.Error(t, err, "Executing a command that fails should return an error")
		assert.Contains(t, err.Error(), "No such file or directory", "The error message should contain the exact OS failure reason")
	})
}
