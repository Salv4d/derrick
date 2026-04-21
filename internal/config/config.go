package config

import "gopkg.in/yaml.v3"

// DefaultNixRegistry is the default Nix registry URL.
const DefaultNixRegistry = "github:NixOS/nixpkgs/nixos-unstable"

// CurrentSchema is the derrick.yaml schema version this binary understands.
// Bump when introducing a backwards-incompatible change to ProjectConfig and
// add a migration in parser.go.
const CurrentSchema = 1

// NixPackage represents a Nix package dependency.
type NixPackage struct {
	Name     string `yaml:"package"`
	Registry string `yaml:"registry,omitempty"`
}

// MarshalYAML emits a plain scalar when Registry is empty so that
// yaml.Marshal produces `- nodejs_20` instead of `- package: nodejs_20`.
func (n NixPackage) MarshalYAML() (interface{}, error) {
	if n.Registry == "" {
		return n.Name, nil
	}
	type alias NixPackage
	return alias(n), nil
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

// EnvVar represents an environment variable declaration.
type EnvVar struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
	Validation  string `yaml:"validation,omitempty"`
}

// ValidationCheck defines a command to validate the environment, with an optional auto-fix.
type ValidationCheck struct {
	Name    string `yaml:"name" validate:"required"`
	Command string `yaml:"command" validate:"required"`
	AutoFix string `yaml:"auto_fix,omitempty"`
}

// FlagDef declares a custom project-level flag exposed by `derrick start`.
type FlagDef struct {
	Description string `yaml:"description"`
}

// Hook is a lifecycle command with an optional execution condition.
//
// The When field controls when the hook fires:
//   - "" or "always"     — run every time
//   - "first-setup"      — run only on the first `derrick start` (before state is persisted)
//   - "flag:<name>"      — run only when `derrick start --<name>` is passed
type Hook struct {
	Run  string `yaml:"run"`
	When string `yaml:"when,omitempty"`
}

// UnmarshalYAML lets a Hook be written as a plain string ("echo hi") or a
// full struct ({ run: "echo hi", when: first-setup }).
func (h *Hook) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		h.Run = value.Value
		h.When = "always"
		return nil
	}

	type alias Hook
	var tmp alias
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	h.Run = tmp.Run
	h.When = tmp.When
	return nil
}

// LifecycleHooks contains hooks for each lifecycle stage.
//
// Stages (derrick start):
//   - BeforeStart: host shell, before provider touches anything. Use for
//     precondition checks or inputs that must exist before provisioning.
//   - Setup:       inside the sandbox (nix dev shell when nix is active),
//     after Provision but before services boot. Use for language-dependent
//     setup that doesn't need services — `npm install`, `go mod download`.
//   - AfterStart:  inside the sandbox, after services are up. Use for work
//     that needs both the toolchain and live services — DB seeding, warmup.
//
// Stages (derrick stop):
//   - BeforeStop:  inside the sandbox, while services are still reachable.
//     Use for graceful drain, DB dumps, cache flushes.
//   - AfterStop:   host shell, after teardown. Use for log shipping or local
//     cleanup that can't rely on the sandbox.
type LifecycleHooks struct {
	BeforeStart []Hook `yaml:"before_start,omitempty"`
	Setup       []Hook `yaml:"setup,omitempty"`
	AfterStart  []Hook `yaml:"after_start,omitempty"`
	BeforeStop  []Hook `yaml:"before_stop,omitempty"`
	AfterStop   []Hook `yaml:"after_stop,omitempty"`
}

// DockerConfig holds Docker Compose orchestration settings.
type DockerConfig struct {
	Compose  string   `yaml:"compose,omitempty" validate:"omitempty,filepath"`
	Profiles []string `yaml:"profiles,omitempty"`
	// Shell is the service to exec into when running `derrick shell`.
	// Defaults to the first service defined in the compose file.
	Shell string `yaml:"shell,omitempty"`
	// Networks lists additional Docker networks every service in this project
	// should join. Derrick creates them on start if they don't exist and marks
	// them com.derrick.managed=true. Use this to opt into cross-project
	// container DNS without giving up per-project network isolation by default.
	Networks []string `yaml:"networks,omitempty"`
}

// NixConfig holds Nix sandbox settings.
type NixConfig struct {
	Registry string       `yaml:"registry,omitempty"`
	Packages []NixPackage `yaml:"packages,omitempty"`
}

// EnvManagement defines environment variable management settings.
type EnvManagement struct {
	BaseFile      string `yaml:"base_file,omitempty" validate:"omitempty,filepath"`
	PromptMissing bool   `yaml:"prompt_missing,omitempty"`
}

// Profile defines a named configuration overlay that extends the base config.
type Profile struct {
	Extend        string            `yaml:"extend,omitempty"`
	Docker        *DockerConfig     `yaml:"docker,omitempty"`
	Nix           *NixConfig        `yaml:"nix,omitempty"`
	Hooks         *LifecycleHooks   `yaml:"hooks,omitempty"`
	Validations   []ValidationCheck `yaml:"validations,omitempty" validate:"dive"`
	EnvManagement *EnvManagement    `yaml:"env_management,omitempty"`
	Env           map[string]EnvVar `yaml:"env,omitempty"`
}

// ProjectConfig is the root configuration structure for a Derrick project.
//
// Provider selects the isolation backend:
//   - "docker"     — Docker Compose
//   - "nix"        — Nix flake dev shell
//   - "hybrid"     — Docker (daemon) + Nix (dev shell)
//   - "auto"       — Docker if available, otherwise Nix
type ProjectConfig struct {
	// Schema is the derrick.yaml schema version. Files written before
	// schema versioning was introduced omit it; the parser treats 0 as
	// "legacy, accept and upgrade in memory".
	Schema int `yaml:"schema,omitempty"`

	Name     string `yaml:"name" validate:"required,lowercase"`
	Version  string `yaml:"version" validate:"required"`
	Provider string `yaml:"provider,omitempty"` // docker | nix | auto

	Docker DockerConfig `yaml:"docker,omitempty"`
	Nix    NixConfig    `yaml:"nix,omitempty"`

	Hooks    LifecycleHooks     `yaml:"hooks,omitempty"`
	Flags    map[string]FlagDef `yaml:"flags,omitempty"`
	Requires []string           `yaml:"requires,omitempty"`
	Env      map[string]EnvVar  `yaml:"env,omitempty"`

	// Validation checks run during `derrick start` after the environment boots.
	Validations []ValidationCheck `yaml:"validations,omitempty" validate:"dive"`

	// EnvManagement controls .env file handling.
	EnvManagement EnvManagement `yaml:"env_management,omitempty"`

	// Profiles are named configuration overlays.
	Profiles map[string]Profile `yaml:"profiles,omitempty" validate:"dive"`
}

// ActiveProvider resolves the effective provider, applying "auto" detection.
func (c *ProjectConfig) ActiveProvider() string {
	switch c.Provider {
	case "docker":
		return "docker"
	case "nix":
		return "nix"
	case "hybrid":
		return "hybrid"
	case "auto", "":
		if c.Docker.Compose != "" || len(c.Docker.Profiles) > 0 {
			return "docker"
		}
		if len(c.Nix.Packages) > 0 {
			return "nix"
		}
		return "nix" // default fallback
	}
	return c.Provider
}
