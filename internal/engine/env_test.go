package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndLoadEnv_AllSet(t *testing.T) {
	tempDir := t.TempDir()
	
	// Set environment variables
	os.Setenv("TEST_VAR_1", "value1")
	defer os.Unsetenv("TEST_VAR_1")

	// Mock prompters just in case
	origSelect := promptSelect
	origInput := promptInput
	defer func() {
		promptSelect = origSelect
		promptInput = origInput
	}()
	promptSelect = func(title string, options []huh.Option[string]) (string, error) { return "abort", nil }
	promptInput = func(title, description string) (string, error) { return "", nil }
	
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

	// Save original prompters
	origSelect := promptSelect
	origInput := promptInput
	defer func() {
		promptSelect = origSelect
		promptInput = origInput
	}()

	t.Run("Validation Success", func(t *testing.T) {
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
	})

	t.Run("Validation Failure - Abort", func(t *testing.T) {
		promptSelect = func(title string, options []huh.Option[string]) (string, error) {
			return "abort", nil
		}

		cfgFail := &config.ProjectConfig{
			Env: map[string]config.EnvVar{
				"INVALID_VAR": {
					Required:   true,
					Validation: "false", // fail
				},
			},
		}

		err := ValidateAndLoadEnv(tempDir, cfgFail, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration aborted")
	})

	t.Run("Validation Failure - Skip", func(t *testing.T) {
		promptSelect = func(title string, options []huh.Option[string]) (string, error) {
			return "skip", nil
		}

		cfgFail := &config.ProjectConfig{
			Env: map[string]config.EnvVar{
				"INVALID_VAR": {
					Required:   true,
					Validation: "false", // fail
				},
			},
		}

		err := ValidateAndLoadEnv(tempDir, cfgFail, false)
		assert.NoError(t, err)
	})

	t.Run("Validation Failure - Update", func(t *testing.T) {
		selectCount := 0
		promptSelect = func(title string, options []huh.Option[string]) (string, error) {
			selectCount++
			if selectCount == 1 {
				return "update", nil
			}
			return "abort", nil // Should not reach here if update works and re-validates (unless second validation also fails)
		}

		promptInput = func(title, description string) (string, error) {
			// Change env var to something that passes validation
			os.Setenv("INVALID_VAR", "new-value")
			return "new-value", nil
		}

		cfgFail := &config.ProjectConfig{
			Env: map[string]config.EnvVar{
				"INVALID_VAR": {
					Required:   true,
					Validation: "test \"$INVALID_VAR\" = \"new-value\"",
				},
			},
		}

		// Initially INVALID_VAR is "fail", so validation "test ... = new-value" will fail.
		// Then we mock "update" and return "new-value", and validation should pass.
		err := ValidateAndLoadEnv(tempDir, cfgFail, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, selectCount)
	})
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
