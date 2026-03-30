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

func ValidateAndLoadEnv(projectDir string, cfg *config.ProjectConfig, useNix bool) error {
	envPath := filepath.Join(projectDir, ".env")

	_ = godotenv.Load(envPath)

	newEnvValues := make(map[string]string)

	for key, rules := range cfg.Env {
		val, exists := os.LookupEnv(key)

		// 1. Check if required but missing
		if (!exists || val == "") && rules.Required {
			ui.Section("Environment Configuration")
			ui.Warningf("Required variable %s is missing.", key)

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
				return fmt.Errorf("%s is required but was left empty", key)
			}

			val = input
			newEnvValues[key] = val
			os.Setenv(key, val)
		}

		// 2. Perform optional validation if a command is provided
		if rules.Validation != "" && val != "" {
			ui.Taskf("Validating %s value", key)

			err := executeCommand(rules.Validation, useNix)
			if err != nil {
				ui.Error("FAILED")
				return fmt.Errorf("validation failed for %s: %v\n Fix: Ensure the value is correct and try again.", key, err)
			}
			ui.Success("OK")
		}
	}

	if len(newEnvValues) > 0 {
		return appendToEnvFile(envPath, newEnvValues)
	}

	return nil
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
