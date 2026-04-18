package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name        string
		stderr      string
		wantMessage string
		wantFix     string
	}{
		{
			name:        "docker socket permission denied",
			stderr:      "Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock",
			wantMessage: "Docker socket permission denied",
			wantFix:     "sudo usermod -aG docker $USER",
		},
		{
			name:        "docker daemon not running",
			stderr:      "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?",
			wantMessage: "Docker daemon is not running",
			wantFix:     "Start Docker Desktop",
		},
		{
			name:        "port already in use",
			stderr:      "Bind for 0.0.0.0:5432 failed: port is already allocated",
			wantMessage: "required port is already in use",
			wantFix:     "Stop the conflicting service",
		},
		{
			name:        "docker image not found",
			stderr:      "pull access denied for myimage, repository does not exist",
			wantMessage: "Docker image not found",
			wantFix:     "docker login",
		},
		{
			name:        "nix package missing",
			stderr:      "error: attribute 'nonexistent-pkg' missing",
			wantMessage: "Nix package not found",
			wantFix:     "search.nixos.org",
		},
		{
			name:        "unknown error falls through as plain string",
			stderr:      "some completely unknown error xyz",
			wantMessage: "some completely unknown error xyz",
			wantFix:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := translateError(tc.stderr, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantMessage)
			if tc.wantFix != "" {
				assert.Contains(t, err.Error(), tc.wantFix)
			}
		})
	}
}

func TestDerrickErrorFormat(t *testing.T) {
	err := &DerrickError{
		Message: "Docker daemon is not running.",
		Fix:     "sudo systemctl start docker",
	}
	msg := err.Error()
	assert.Contains(t, msg, "Docker daemon is not running.")
	assert.Contains(t, msg, "Fix:")
	assert.Contains(t, msg, "sudo systemctl start docker")
}

func TestDerrickErrorNoFix(t *testing.T) {
	err := &DerrickError{Message: "Something went wrong."}
	assert.Equal(t, "Something went wrong.", err.Error())
	assert.NotContains(t, err.Error(), "Fix:")
}

func TestRunSuccess(t *testing.T) {
	err := Run("echo hello")
	assert.NoError(t, err)
}

func TestRunFailure(t *testing.T) {
	err := Run("exit 1")
	assert.Error(t, err)
}
