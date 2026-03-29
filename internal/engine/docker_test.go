package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDockerInstalled(t *testing.T) {
	result := IsDockerInstalled()

	assert.IsType(t, true, result, "IsDockerInstalled should return a boolean value")
}
