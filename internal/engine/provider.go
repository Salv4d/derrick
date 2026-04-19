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
type Provider interface {
	// Name returns the human-readable backend identifier ("docker", "nix", …).
	Name() string

	// IsAvailable checks that the required tooling exists on the host.
	IsAvailable() error

	// Start activates the environment (starts containers, enters nix develop, etc.).
	Start(cfg *config.ProjectConfig, flags Flags) error

	// Stop tears down the environment gracefully.
	Stop(cfg *config.ProjectConfig) error

	// Shell opens an interactive shell inside the managed environment.
	Shell(cfg *config.ProjectConfig) error

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
	default:
		return &NixProvider{}
	}
}
