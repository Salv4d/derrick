package config

const DefaultNixRegistry = "github:NixOS/nixpkgs/nixos-unstable"

type EnvVar struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
	Validation  string `yaml:"validation,omitempty"`
}

type ProjectConfig struct {
	Name         string            `yaml:"name" validate:"required,lowercase"`
	Version      string            `yaml:"version" validate:"required"`
	Dependencies Dependencies      `yaml:"dependencies"`
	Hooks        LifecycleHooks    `yaml:"hooks"`
	Validations  []ValidationCheck `yaml:"validations" validate:"dive"`
	Env          map[string]EnvVar `yaml:"env"`
	Profiles     map[string]Profile `yaml:"profiles,omitempty" validate:"dive"`
}

type Dependencies struct {
	NixPackages          []string `yaml:"nix_packages"`
	NixRegistry          string   `yaml:"nix_registry" validate:"omitempty"`
	DockerCompose        string   `yaml:"docker_compose,omitempty" validate:"omitempty,filepath"`
	DockerComposeProfiles []string `yaml:"docker_compose_profiles,omitempty"`
}

type LifecycleHooks struct {
	PreInit   []string `yaml:"pre_init,omitempty"`
	PostInit  []string `yaml:"post_init,omitempty"`
	PreStart  []string `yaml:"pre_start,omitempty"`
	PostStart []string `yaml:"post_start,omitempty"`
	PreBuild  []string `yaml:"pre_build,omitempty"`
	PostStop  []string `yaml:"post_stop,omitempty"`
}

type ValidationCheck struct {
	Name    string `yaml:"name" validate:"required"`
	Command string `yaml:"command" validate:"required"`
	AutoFix string `yaml:"auto_fix,omitempty"`
}

type Profile struct {
	Extend       string            `yaml:"extend,omitempty"`
	Dependencies *Dependencies     `yaml:"dependencies,omitempty"`
	Hooks        *LifecycleHooks   `yaml:"hooks,omitempty"`
	Validations  []ValidationCheck `yaml:"validations,omitempty" validate:"dive"`
	Env          map[string]EnvVar `yaml:"env,omitempty"`
}
