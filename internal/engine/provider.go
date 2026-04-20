package engine

import "github.com/Salv4d/derrick/internal/config"

// EnvironmentStatus describes the live state of a managed environment.
type EnvironmentStatus struct {
	Running bool
	Details string
}

// Flags carries the resolved set of custom and built-in flags for a command.
type Flags struct {
	// Active maps custom flag names (from derrick.yaml flags:) to whether they were set.
	Active map[string]bool
	// Reset indicates --reset was passed: providers should rebuild from scratch.
	Reset bool
	// Env holds resolved KEY=VALUE pairs injected into the provider's child
	// processes (compose up, nix develop, etc.).
	Env []string
}

// Provider abstracts an environment backend.
//
// Each backend (Docker, Nix) implements this interface so the CLI layer can remain
// completely agnostic of the underlying isolation technology.
//
// Lifecycle split: Provision materializes the environment so that setup hooks
// can run inside it (writes flake.nix, compose override, resolves deps). Start
// boots the running services. Splitting these lets `setup` hooks (npm install,
// go mod download) run inside a ready sandbox before services come up.
type Provider interface {
	// Name returns the human-readable backend identifier ("docker", "nix", …).
	Name() string

	// IsAvailable checks that the required tooling exists on the host.
	IsAvailable() error

	// Provision materializes the environment: writes generated config files
	// (.derrick/flake.nix, compose override) and resolves dependencies so the
	// sandbox is ready for setup hooks. Does NOT boot long-running services.
	Provision(cfg *config.ProjectConfig) error

	// Start boots long-running services (docker compose up). Pure file/resolve
	// work belongs in Provision.
	Start(cfg *config.ProjectConfig, flags Flags) error

	// Stop tears down the environment gracefully.
	Stop(cfg *config.ProjectConfig) error

	// Shell opens an interactive shell inside the managed environment,
	// or runs args as a single command when args is non-empty.
	Shell(cfg *config.ProjectConfig, args []string) error

	// Status returns the live operational state.
	Status(cfg *config.ProjectConfig) (EnvironmentStatus, error)
}

// ResolveProvider returns the correct Provider for the given project config.
// It respects the Provider field and falls back to auto-detection.
func ResolveProvider(cfg *config.ProjectConfig) Provider {
	switch cfg.ActiveProvider() {
	case "docker":
		return &DockerProvider{}
	case "nix":
		return &NixProvider{}
	case "hybrid":
		return NewHybridProvider()
	default:
		return &NixProvider{}
	}
}
