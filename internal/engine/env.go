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

// envSlice converts a KEY→VALUE map into the KEY=VALUE slice expected by exec.Cmd.Env.
func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

// ValidateAndLoadEnv resolves every variable declared in cfg.Env, prompting
// or applying defaults as needed. It returns the resolved KEY=VALUE pairs so
// callers can inject them explicitly into subprocesses via cmd.Env.
//
// The function does NOT mutate os.Environ: values from .env on disk are read
// into a local map, and new values are propagated via the returned slice.
func ValidateAndLoadEnv(projectDir string, cfg *config.ProjectConfig, useNix bool) ([]string, error) {
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

	// Read .env into a local map instead of loading it into the process env.
	resolved := make(map[string]string)
	if fileEnv, err := godotenv.Read(envPath); err == nil {
		for k, v := range fileEnv {
			resolved[k] = v
		}
	}

	newEnvValues := make(map[string]string)

	for key, rules := range cfg.Env {
		if cfg.EnvManagement.PromptMissing {
			rules.Required = true
		}

		val, exists := resolved[key]
		if !exists || val == "" {
			// Fall back to the caller's shell env, but never persist that
			// fallback back to disk — it's just a runtime value.
			if shellVal := os.Getenv(key); shellVal != "" {
				val = shellVal
				exists = true
				resolved[key] = val
			}
		}

		if !exists || val == "" {
			if rules.Default != "" {
				val = rules.Default
				ui.Infof("Using default value for %s: %s", key, val)
				newEnvValues[key] = val
				resolved[key] = val
			} else if rules.Required {
				ui.Section("Environment Configuration")
				ui.Warningf("Required variable %s is missing.", key)

				input, err := promptInput(fmt.Sprintf("Enter value for %s", key), rules.Description)
				if err != nil {
					return nil, fmt.Errorf("configuration aborted by user")
				}

				input = strings.TrimSpace(input)
				if input == "" {
					ui.Warningf("%s is required but was left empty. Proceeding with caution.", key)
				} else {
					val = input
					newEnvValues[key] = val
					resolved[key] = val
				}
			}
		}

		for rules.Validation != "" && val != "" {
			ui.Taskf("Validating %s value", key)

			runner := &Runner{
				UseNix: useNix,
				Env:    envSlice(resolved),
			}
			err := runner.Run(rules.Validation)
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
				return nil, fmt.Errorf("configuration aborted for %s", key)
			}

			if choice == "skip" {
				ui.Warningf("Skipping validation for %s. Proceeding with caution.", key)
				break
			}

			if choice == "update" {
				input, err := promptInput(fmt.Sprintf("Enter new value for %s", key), rules.Description)
				if err != nil {
					return nil, fmt.Errorf("configuration aborted")
				}

				val = strings.TrimSpace(input)
				resolved[key] = val
				newEnvValues[key] = val
			}
		}
	}

	if len(newEnvValues) > 0 {
		if err := updateEnvFile(envPath, newEnvValues); err != nil {
			return nil, err
		}
	}

	return envSlice(resolved), nil
}

func updateEnvFile(path string, vars map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file with these variables
			var sb strings.Builder
			for k, v := range vars {
				fmt.Fprintf(&sb, "%s=%q\n", k, v)
			}
			return os.WriteFile(path, []byte(sb.String()), 0600)
		}
		return fmt.Errorf("failed to read env file %s: %w", path, err)
	}
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

	// Atomic write: write to a temp file in the same directory, then rename.
	// This prevents a partially-written .env if the process is interrupted.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(output), 0o600); err != nil {
		return fmt.Errorf("failed to write temp env file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to replace %s: %w", path, err)
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
