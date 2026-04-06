package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsDockerInstalled verifies that the function returns a boolean indicating
// Docker availability on the system.
func TestIsDockerInstalled(t *testing.T) {
	result := IsDockerInstalled()

	assert.IsType(t, true, result, "IsDockerInstalled should return a boolean value")
}
