package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ParseConfig reads and validates a derrick.yaml file, optionally applying a named profile.
func ParseConfig(filename string, profileName string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var cfg ProjectConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err = decoder.Decode(&cfg); err != nil {
		return nil, enhanceYAMLError(data, err)
	}

	if err = checkSchema(&cfg); err != nil {
		return nil, err
	}

	if err = validate.Struct(cfg); err != nil {
		return nil, formatValidationError(err)
	}

	if profileName != "" {
		if err = applyProfile(&cfg, profileName); err != nil {
			return nil, err
		}
	}

	// Apply defaults.
	if cfg.Nix.Registry == "" {
		cfg.Nix.Registry = DefaultNixRegistry
	}

	return &cfg, nil
}

// checkSchema validates that the file's declared schema is one this binary
// can handle. A missing/zero schema is accepted as legacy and upgraded in
// memory; a future schema fails fast with a hint to update derrick.
func checkSchema(cfg *ProjectConfig) error {
	switch {
	case cfg.Schema == 0:
		cfg.Schema = CurrentSchema
		return nil
	case cfg.Schema == CurrentSchema:
		return nil
	case cfg.Schema > CurrentSchema:
		return fmt.Errorf("derrick.yaml schema v%d is newer than this binary (v%d). Run 'derrick update'", cfg.Schema, CurrentSchema)
	default:
		// No older schemas yet. When one is retired, run its migration here.
		return fmt.Errorf("derrick.yaml schema v%d is no longer supported (current: v%d)", cfg.Schema, CurrentSchema)
	}
}

func applyProfile(cfg *ProjectConfig, profileName string) error {
	return applyProfileChain(cfg, profileName, make(map[string]bool))
}

// applyProfileChain walks the extend: chain while tracking visited profiles so
// A extends B extends A (or any longer cycle) aborts with a clear error instead
// of stack-overflowing.
func applyProfileChain(cfg *ProjectConfig, profileName string, visited map[string]bool) error {
	if visited[profileName] {
		return fmt.Errorf("circular profile extension detected at '%s'", profileName)
	}
	visited[profileName] = true

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' is not defined in derrick.yaml", profileName)
	}

	if profile.Extend != "" {
		if err := applyProfileChain(cfg, profile.Extend, visited); err != nil {
			return err
		}
	}

	mergeProfileToConfig(cfg, profile)
	return nil
}

func mergeProfileToConfig(cfg *ProjectConfig, p Profile) {
	if p.Docker != nil {
		if p.Docker.Compose != "" {
			cfg.Docker.Compose = p.Docker.Compose
		}
		if len(p.Docker.Profiles) > 0 {
			cfg.Docker.Profiles = append(cfg.Docker.Profiles, p.Docker.Profiles...)
		}
		if len(p.Docker.Networks) > 0 {
			cfg.Docker.Networks = append(cfg.Docker.Networks, p.Docker.Networks...)
		}
	}

	if p.Nix != nil {
		if len(p.Nix.Packages) > 0 {
			cfg.Nix.Packages = append(cfg.Nix.Packages, p.Nix.Packages...)
		}
		if p.Nix.Registry != "" {
			cfg.Nix.Registry = p.Nix.Registry
		}
	}

	if p.Hooks != nil {
		cfg.Hooks.BeforeStart = append(cfg.Hooks.BeforeStart, p.Hooks.BeforeStart...)
		cfg.Hooks.Setup = append(cfg.Hooks.Setup, p.Hooks.Setup...)
		cfg.Hooks.AfterStart = append(cfg.Hooks.AfterStart, p.Hooks.AfterStart...)
		cfg.Hooks.BeforeStop = append(cfg.Hooks.BeforeStop, p.Hooks.BeforeStop...)
		cfg.Hooks.AfterStop = append(cfg.Hooks.AfterStop, p.Hooks.AfterStop...)
	}

	if len(p.Validations) > 0 {
		cfg.Validations = append(cfg.Validations, p.Validations...)
	}

	if p.EnvManagement != nil {
		if p.EnvManagement.BaseFile != "" {
			cfg.EnvManagement.BaseFile = p.EnvManagement.BaseFile
		}
		if p.EnvManagement.PromptMissing {
			cfg.EnvManagement.PromptMissing = true
		}
	}

	if len(p.Env) > 0 {
		if cfg.Env == nil {
			cfg.Env = make(map[string]EnvVar)
		}
		for k, v := range p.Env {
			cfg.Env[k] = v
		}
	}
}

func formatValidationError(err error) error {
	var sb strings.Builder
	sb.WriteString("Invalid configuration in derrick.yaml:\n\n")

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			yamlPath := strings.TrimPrefix(e.Namespace(), "ProjectConfig.")
			fmt.Fprintf(&sb, "  ✖ Field '%s' failed validation: must satisfy '%s'\n", yamlPath, e.Tag())
		}
		return errors.New(sb.String())
	}

	return err
}

func enhanceYAMLError(fileData []byte, originalErr error) error {
	errMsg := originalErr.Error()

	re := regexp.MustCompile(`line (\d+)`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) < 2 {
		return originalErr
	}

	lineNum, err := strconv.Atoi(matches[1])
	if err != nil {
		return originalErr
	}

	lines := strings.Split(string(fileData), "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return originalErr
	}

	errorLine := lines[lineNum-1]
	trimmed := strings.TrimSpace(errorLine)
	indent := len(errorLine) - len(trimmed)
	indicator := strings.Repeat(" ", indent) + strings.Repeat("^", len(trimmed))

	var sb strings.Builder
	fmt.Fprintf(&sb, "Syntax error in derrick.yaml at line %d:\n\n", lineNum)
	if lineNum > 1 {
		fmt.Fprintf(&sb, "  %3d | %s\n", lineNum-1, lines[lineNum-2])
	}
	fmt.Fprintf(&sb, "  %3d | %s\n", lineNum, errorLine)
	fmt.Fprintf(&sb, "       %s\n\n", indicator)
	fmt.Fprintf(&sb, "Detail: %s", errMsg)

	return errors.New(sb.String())
}
