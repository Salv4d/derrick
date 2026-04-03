package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

	if cfg.EnvManagement.BaseFile != "" {
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			basePath := filepath.Join(projectDir, cfg.EnvManagement.BaseFile)
			if _, err := os.Stat(basePath); err == nil {
				ui.Infof("Auto-generating .env from %s", cfg.EnvManagement.BaseFile)
				_ = copyFile(basePath, envPath)
			}
		}
	}

	_ = godotenv.Load(envPath)

	unifiedEnv := make(map[string]config.EnvVar)
	if cfg.Env != nil {
		for k, v := range cfg.Env {
			unifiedEnv[k] = v
		}
	}

	if cfg.EnvManagement.PromptMissing {
		fileEnv, err := godotenv.Read(envPath)
		if err == nil {
			for k, v := range fileEnv {
				if v == "" {
					if val, exists := unifiedEnv[k]; exists {
						val.Required = true
						unifiedEnv[k] = val
					} else {
						unifiedEnv[k] = config.EnvVar{
							Required: true,
						}
					}
				}
			}
		}
	}

	newEnvValues := make(map[string]string)

	for key, rules := range unifiedEnv {
		val, exists := os.LookupEnv(key)

		if !exists || val == "" {
			if rules.Default != "" {
				val = rules.Default
				ui.Infof("Using default value for %s: %s", key, val)
				newEnvValues[key] = val
				os.Setenv(key, val)
			} else if rules.Required {
				ui.Section("Environment Configuration")
				ui.Warningf("Required variable %s is missing.", key)

				input, err := promptInput(fmt.Sprintf("Enter value for %s", key), rules.Description)
				if err != nil {
					return fmt.Errorf("configuration aborted by user")
				}

				input = strings.TrimSpace(input)
				if input == "" {
					ui.Warningf("%s is required but was left empty. Proceeding with caution.", key)
				} else {
					val = input
					newEnvValues[key] = val
					os.Setenv(key, val)
				}
			}
		}

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

			}
		}
	}

	if len(newEnvValues) > 0 {
		return updateEnvFile(envPath, newEnvValues)
	}

	return nil
}

func updateEnvFile(path string, vars map[string]string) error {
	content, _ := os.ReadFile(path)
	lines := strings.Split(string(content), "\n")

	for k, v := range vars {
		found := false

		escapedVal := strings.ReplaceAll(v, "\"", "\\\"")
		newLine := fmt.Sprintf("%s=\"%s\"", k, escapedVal)

		re := regexp.MustCompile(fmt.Sprintf(`^%s=.*$`, regexp.QuoteMeta(k)))
		for i, line := range lines {
			if re.MatchString(strings.TrimSpace(line)) {
				lines[i] = newLine
				found = true
				break
			}
		}

		if !found {

			if len(lines) > 0 && lines[len(lines)-1] != "" {
				lines = append(lines, "")
			}
			lines = append(lines, newLine)
		}
	}

	output := strings.Join(lines, "\n")

	output = strings.TrimLeft(output, "\n")

	err := os.WriteFile(path, []byte(output), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %v", path, err)
	}

	ui.Successf("Updated %s with %d variables", path, len(vars))
	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
