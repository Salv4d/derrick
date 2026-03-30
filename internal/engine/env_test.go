package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndLoadEnv_AllSet(t *testing.T) {
	tempDir := t.TempDir()
	
	// Set environment variables
	os.Setenv("TEST_VAR_1", "value1")
	defer os.Unsetenv("TEST_VAR_1")
	
	cfg := &config.ProjectConfig{
		Env: map[string]config.EnvVar{
			"TEST_VAR_1": {
				Description: "A test variable",
				Required:    true,
			},
			"TEST_VAR_2": {
				Description: "Optional variable",
				Required:    false,
			},
		},
	}

	err := ValidateAndLoadEnv(tempDir, cfg, false)
	assert.NoError(t, err)
}

func TestValidateAndLoadEnv_Validation(t *testing.T) {
	tempDir := t.TempDir()
	
	os.Setenv("VALID_VAR", "ok")
	defer os.Unsetenv("VALID_VAR")
	
	os.Setenv("INVALID_VAR", "fail")
	defer os.Unsetenv("INVALID_VAR")
	
	cfg := &config.ProjectConfig{
		Env: map[string]config.EnvVar{
			"VALID_VAR": {
				Required:   true,
				Validation: "true", // success
			},
		},
	}

	err := ValidateAndLoadEnv(tempDir, cfg, false)
	assert.NoError(t, err)

	cfgFail := &config.ProjectConfig{
		Env: map[string]config.EnvVar{
			"INVALID_VAR": {
				Required:   true,
				Validation: "false", // fail
			},
		},
	}

	err = ValidateAndLoadEnv(tempDir, cfgFail, false)
	assert.Error(t, err)
}

func TestAppendToEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	
	vars := map[string]string{
		"KEY1": "VALUE1",
		"KEY2": "VALUE2 \"WITH QUOTES\"",
	}
	
	err := appendToEnvFile(envPath, vars)
	require.NoError(t, err)
	
	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	
	assert.Contains(t, string(content), `KEY1="VALUE1"`)
	assert.Contains(t, string(content), `KEY2="VALUE2 \"WITH QUOTES\""`)
}
