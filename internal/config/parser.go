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

func ParseConfig(filename string, profileName string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var config ProjectConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	err = decoder.Decode(&config)
	if err != nil {
		return nil, enhanceYAMLError(data, err)
	}

	err = validate.Struct(config)
	if err != nil {
		return nil, formatValidationError(err)
	}

	if profileName != "" {
		err = applyProfile(&config, profileName)
		if err != nil {
			return nil, err
		}
	}

	if config.Dependencies.NixRegistry == "" {
		config.Dependencies.NixRegistry = DefaultNixRegistry
	}



	return &config, nil
}

func applyProfile(cfg *ProjectConfig, profileName string) error {
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' is not defined in derrick.yaml", profileName)
	}

	if profile.Extend != "" {
		err := applyProfile(cfg, profile.Extend)
		if err != nil {
			return fmt.Errorf("failed to extend profile '%s': %w", profile.Extend, err)
		}
	}

	mergeProfileToConfig(cfg, profile)

	return nil
}

func mergeProfileToConfig(cfg *ProjectConfig, p Profile) {
	if p.Dependencies != nil {
		if len(p.Dependencies.NixPackages) > 0 {
			cfg.Dependencies.NixPackages = append(cfg.Dependencies.NixPackages, p.Dependencies.NixPackages...)
		}
		if p.Dependencies.NixRegistry != "" {
			cfg.Dependencies.NixRegistry = p.Dependencies.NixRegistry
		}
		if p.Dependencies.DockerCompose != "" {
			cfg.Dependencies.DockerCompose = p.Dependencies.DockerCompose
		}
		if len(p.Dependencies.DockerComposeProfiles) > 0 {
			cfg.Dependencies.DockerComposeProfiles = append(cfg.Dependencies.DockerComposeProfiles, p.Dependencies.DockerComposeProfiles...)
		}
	}

	if p.Hooks != nil {
		cfg.Hooks.PreInit = append(cfg.Hooks.PreInit, p.Hooks.PreInit...)
		cfg.Hooks.PostInit = append(cfg.Hooks.PostInit, p.Hooks.PostInit...)
		cfg.Hooks.PreStart = append(cfg.Hooks.PreStart, p.Hooks.PreStart...)
		cfg.Hooks.PostStart = append(cfg.Hooks.PostStart, p.Hooks.PostStart...)
		cfg.Hooks.PreBuild = append(cfg.Hooks.PreBuild, p.Hooks.PreBuild...)
		cfg.Hooks.PostStop = append(cfg.Hooks.PostStop, p.Hooks.PostStop...)
	}

	if len(p.Validations) > 0 {
		cfg.Validations = append(cfg.Validations, p.Validations...)
	}

	if p.EnvManagement != nil {
		if p.EnvManagement.BaseFile != "" {
			cfg.EnvManagement.BaseFile = p.EnvManagement.BaseFile
		}
		// If profile forcefully defines it, override the root. Wait, how do we handle false? 
		// Actually, standard bool merging means true applies over false, but false might not apply over true.
		// Since we want simple behavior, let's just do an override if it's set to true.
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
	var builder strings.Builder
	builder.WriteString("Invalid configuration contract in derrick.yaml:\n\n")

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			yamlPath := strings.TrimPrefix(e.Namespace(), "ProjectConfig.")
			fmt.Fprintf(&builder, "  ✖ Field '%s' failed validation: must be '%s'\n", yamlPath, e.Tag())
		}
		return errors.New(builder.String())
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

	trimmedLine := strings.TrimSpace(errorLine)
	indentLength := len(errorLine) - len(trimmedLine)
	indicator := strings.Repeat(" ", indentLength) + strings.Repeat("^", len(trimmedLine))

	var builder strings.Builder
	fmt.Fprintf(&builder, "Syntax error in derrick.yaml at line %d: \n\n", lineNum)

	if lineNum > 1 {
		fmt.Fprintf(&builder, "  %3d | %s\n", lineNum-1, lines[lineNum-2])
	}

	fmt.Fprintf(&builder, "  %3d | %s\n", lineNum, errorLine)
	fmt.Fprintf(&builder, "       %s\n\n", indicator)

	fmt.Fprintf(&builder, "Detail: %s", errMsg)

	return errors.New(builder.String())
}
