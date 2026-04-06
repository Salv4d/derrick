package config

import (
	"gopkg.in/yaml.v3"
)

// DefaultNixRegistry is the default Nix registry URL.
const DefaultNixRegistry = "github:NixOS/nixpkgs/nixos-unstable"

// NixPackage represents a Nix package dependency.
type NixPackage struct {
	Name     string `yaml:"package"`
	Registry string `yaml:"registry,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling for NixPackage.
func (n *NixPackage) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		n.Name = value.Value
		return nil
	}

	type alias NixPackage
	var tmp alias
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	n.Name = tmp.Name
	n.Registry = tmp.Registry
	return nil
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
	Validation  string `yaml:"validation,omitempty"`
}

// ProjectConfig is the root configuration structure for a Derrick project.
type ProjectConfig struct {
	Name          string             `yaml:"name" validate:"required,lowercase"`
	Version       string             `yaml:"version" validate:"required"`
	Requires      []string           `yaml:"requires,omitempty"`
	Dependencies Dependencies      `yaml:"dependencies"`
	Hooks        LifecycleHooks    `yaml:"hooks"`
	Validations   []ValidationCheck  `yaml:"validations" validate:"dive"`
	EnvManagement EnvManagement      `yaml:"env_management,omitempty"`
	Env           map[string]EnvVar  `yaml:"env"`
	Profiles      map[string]Profile `yaml:"profiles,omitempty" validate:"dive"`
}

// EnvManagement defines environment variable management settings.
type EnvManagement struct {
	BaseFile      string `yaml:"base_file,omitempty" validate:"omitempty,filepath"`
	PromptMissing bool   `yaml:"prompt_missing,omitempty"`
}

// Dependencies groups Nix and Docker Compose dependencies.
type Dependencies struct {
	NixPackages          []NixPackage `yaml:"nix_packages"`
	NixRegistry          string   `yaml:"nix_registry" validate:"omitempty"`
	DockerCompose        string   `yaml:"docker_compose,omitempty" validate:"omitempty,filepath"`
	DockerComposeProfiles []string `yaml:"docker_compose_profiles,omitempty"`
}

// LifecycleHooks contains commands to run at different lifecycle stages.
type LifecycleHooks struct {
	PreInit   []string `yaml:"pre_init,omitempty"`
	PostInit  []string `yaml:"post_init,omitempty"`
	PreStart  []string `yaml:"pre_start,omitempty"`
	PostStart []string `yaml:"post_start,omitempty"`
	PreBuild  []string `yaml:"pre_build,omitempty"`
	PostStop  []string `yaml:"post_stop,omitempty"`
}

// ValidationCheck defines a validation command and optional auto-fix.
type ValidationCheck struct {
	Name    string `yaml:"name" validate:"required"`
	Command string `yaml:"command" validate:"required"`
	AutoFix string `yaml:"auto_fix,omitempty"`
}

// Profile defines a named configuration profile that extends the base.
type Profile struct {
	Extend       string            `yaml:"extend,omitempty"`
	Dependencies *Dependencies     `yaml:"dependencies,omitempty"`
	Hooks         *LifecycleHooks    `yaml:"hooks,omitempty"`
	Validations   []ValidationCheck  `yaml:"validations,omitempty" validate:"dive"`
	EnvManagement *EnvManagement     `yaml:"env_management,omitempty"`
	Env           map[string]EnvVar  `yaml:"env,omitempty"`
}
