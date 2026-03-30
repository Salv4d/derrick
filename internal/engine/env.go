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

var (
	promptSelect = func(title string, options []huh.Option[string]) (string, error) {
		var choice string
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&choice),
		))
		err := form.Run()
		return choice, err
	}

	promptInput = func(title, description string) (string, error) {
		var input string
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title(title).
				Description(description).
				Value(&input),
		))
		err := form.Run()
		return input, err
	}
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

			input, err := promptInput(fmt.Sprintf("Enter value for %s", key), rules.Description)
			if err != nil {
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
		for {
			if rules.Validation == "" || val == "" {
				break
			}

			ui.Taskf("Validating %s value", key)
			err := executeCommand(rules.Validation, useNix)
			if err == nil {
				ui.Success("OK")
				break
			}

			ui.Error("FAILED")
			ui.Errorf("  Error: %v", err)

			choice, err := promptSelect(fmt.Sprintf("Validation failed for %s. How to proceed?", key), []huh.Option[string]{
				huh.NewOption("Update value", "update"),
				huh.NewOption("Skip validation", "skip"),
				huh.NewOption("Abort", "abort"),
			})

			if err != nil || choice == "abort" {
				return fmt.Errorf("configuration aborted for %s", key)
			}

			if choice == "skip" {
				ui.Warningf("Skipping validation for %s. Proceeding with caution.", key)
				break
			}

			if choice == "update" {
				input, err := promptInput(fmt.Sprintf("Enter new value for %s", key), rules.Description)
				if err != nil {
					return fmt.Errorf("configuration aborted")
				}

				val = strings.TrimSpace(input)
				os.Setenv(key, val)
				newEnvValues[key] = val
				// Continue loop to re-validate new value
			}
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
