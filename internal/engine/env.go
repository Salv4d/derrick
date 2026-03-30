package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

func ValidateAndLoadEnv(projectDir string, cfg *config.ProjectConfig) error {
	envPath := filepath.Join(projectDir, ".env")

	_ = godotenv.Load(envPath)

	var missingVars []string
	newEnvValues := make(map[string]string)

	for key, rules := range cfg.Env {
		val, exists := os.LookupEnv(key)
		if (!exists || val == "") && rules.Required {
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) == 0 {
		ui.Debug("Environment validation passed. No missing variables.")
		return nil
	}

	ui.Section("Environment Validation")
	ui.Warningf("Detected %d missing required environment variables.", len(missingVars))

	for _, key := range missingVars {
		rules := cfg.Env[key]
		var input string

		prompt := huh.NewInput().
			Title(fmt.Sprintf("Enter value for %s", key)).
			Description(rules.Description).
			Value(&input)

		form := huh.NewForm(huh.NewGroup(prompt))

		if err := form.Run(); err != nil {
			return fmt.Errorf("configuration aborted by user")
		}

		input = strings.TrimSpace(input)
		if input == "" {
			return fmt.Errorf("%s is required but was left empty.\n Fix: Rerun `derrick start` and provide a value", key)
		}

		newEnvValues[key] = input
		os.Setenv(key, input)
	}

	return appendToEnvFile(envPath, newEnvValues)
}

func appendToEnvFile(path string, vars map[string]string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v\n Fix: Check file permissions using `ls -la %s` and ensure you have write access", path, err, path)
	}
	defer f.Close()

	for k, v := range vars {
		line := fmt.Sprintf("\n%s=\"%s\"", k, strings.ReplaceAll(v, "\"", "\\\""))
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write secret to %s: %v\n Fix: Ensure your disk is not full and you have write permissions", path, err)
		}
	}

	ui.Successf("Saved %d new variables to %s", len(vars), path)
	return nil
}
