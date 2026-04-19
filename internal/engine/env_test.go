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

// TestValidateAndLoadEnv_AllSet verifies that environment validation and loading
// succeeds when all required environment variables are already set.
func TestValidateAndLoadEnv_AllSet(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("TEST_VAR_1", "value1")
	defer os.Unsetenv("TEST_VAR_1")

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

	_, err := ValidateAndLoadEnv(tempDir, cfg, false)
	assert.NoError(t, err)
}

// TestValidateAndLoadEnv_Validation verifies behavior when environment variables
// fail validation, including abort, skip, and update user choices.
func TestValidateAndLoadEnv_Validation(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("VALID_VAR", "ok")
	defer os.Unsetenv("VALID_VAR")

	os.Setenv("INVALID_VAR", "fail")
	defer os.Unsetenv("INVALID_VAR")

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
					Validation: "true",
				},
			},
		}
		_, err := ValidateAndLoadEnv(tempDir, cfg, false)
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
					Validation: "false",
				},
			},
		}

		_, err := ValidateAndLoadEnv(tempDir, cfgFail, false)
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
					Validation: "false",
				},
			},
		}

		_, err := ValidateAndLoadEnv(tempDir, cfgFail, false)
		assert.NoError(t, err)
	})

	t.Run("Validation Failure - Update", func(t *testing.T) {
		selectCount := 0
		promptSelect = func(title string, options []huh.Option[string]) (string, error) {
			selectCount++
			if selectCount == 1 {
				return "update", nil
			}
			return "abort", nil
		}

		promptInput = func(title, description string) (string, error) {
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

		_, err := ValidateAndLoadEnv(tempDir, cfgFail, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, selectCount)
	})
}

// TestValidateAndLoadEnv_Default verifies that default values are written to
// the .env file when environment variables are not set.
func TestValidateAndLoadEnv_Default(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	cfg := &config.ProjectConfig{
		Env: map[string]config.EnvVar{
			"DEFAULT_VAR": {
				Default: "my-default-value",
			},
		},
	}

	envVars, err := ValidateAndLoadEnv(tempDir, cfg, false)
	assert.NoError(t, err)

	content, _ := os.ReadFile(envPath)
	assert.Contains(t, string(content), `DEFAULT_VAR="my-default-value"`)
	assert.Contains(t, envVars, `DEFAULT_VAR=my-default-value`)
	assert.Empty(t, os.Getenv("DEFAULT_VAR"), "ValidateAndLoadEnv must not mutate os.Environ")
}

// TestUpdateEnvFile verifies that the .env file is correctly updated with new
// variable values while preserving existing content and comments.
func TestUpdateEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	initialContent := "EXISTING_KEY=\"old_value\"\n# Comment\nSOME_OTHER_KEY=\"keep_this\""
	err := os.WriteFile(envPath, []byte(initialContent), 0o600)
	require.NoError(t, err)

	vars := map[string]string{
		"EXISTING_KEY": "new_value",
		"KEY2":         "VALUE2 \"WITH QUOTES\"",
	}

	err = updateEnvFile(envPath, vars)
	require.NoError(t, err)

	content, err := os.ReadFile(envPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, `EXISTING_KEY="new_value"`)
	assert.NotContains(t, contentStr, `EXISTING_KEY="old_value"`)
	assert.Contains(t, contentStr, `SOME_OTHER_KEY="keep_this"`)
	assert.Contains(t, contentStr, `# Comment`)
	assert.Contains(t, contentStr, `KEY2="VALUE2 \"WITH QUOTES\""`)
}

// TestValidateAndLoadEnv_AutoDiscover verifies that environment variables are
// discovered from a base file (e.g., .env.example) and prompts for missing values.
func TestValidateAndLoadEnv_AutoDiscover(t *testing.T) {
	tempDir := t.TempDir()

	baseFileContent := "DISCOVERED_VAR=\nEXISTING_VAR=exists"
	baseFilePath := filepath.Join(tempDir, ".env.example")
	err := os.WriteFile(baseFilePath, []byte(baseFileContent), 0o644)
	require.NoError(t, err)

	cfg := &config.ProjectConfig{
		EnvManagement: config.EnvManagement{
			BaseFile:      ".env.example",
			PromptMissing: true,
		},
		Env: map[string]config.EnvVar{
			"DISCOVERED_VAR": {
				Description: "this came from yaml",
			},
		},
	}

	origInput := promptInput
	defer func() { promptInput = origInput }()
	
	promptCount := 0
	promptInput = func(title, description string) (string, error) {
		promptCount++
		assert.Equal(t, "Enter value for DISCOVERED_VAR", title)
		assert.Equal(t, "this came from yaml", description)
		return "magic-value", nil
	}

	_, err = ValidateAndLoadEnv(tempDir, cfg, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, promptCount, "Should prompt once for DISCOVERED_VAR")
	
	envPath := filepath.Join(tempDir, ".env")
	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	contentStr := string(content)
	
	assert.Contains(t, contentStr, `EXISTING_VAR=exists`)
	assert.Contains(t, contentStr, `DISCOVERED_VAR="magic-value"`)
}
