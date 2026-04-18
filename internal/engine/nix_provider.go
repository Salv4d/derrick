package engine

import (
	"fmt"
	"os/exec"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// NixProvider implements Provider using Nix flake dev shells.
type NixProvider struct{}

func (n *NixProvider) Name() string { return "nix" }

// IsAvailable checks that the nix binary is reachable.
func (n *NixProvider) IsAvailable() error {
	if !IsNixInstalled() {
		return &DerrickError{
			Message: "Nix is not installed on this system.",
			Fix:     `curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install`,
		}
	}
	return nil
}

// Start resolves the Nix flake and validates the environment is ready.
func (n *NixProvider) Start(cfg *config.ProjectConfig, _ Flags) error {
	if len(cfg.Nix.Packages) == 0 {
		return fmt.Errorf("no nix.packages specified in derrick.yaml")
	}

	registry := cfg.Nix.Registry
	if registry == "" {
		registry = config.DefaultNixRegistry
	}

	ui.Taskf("Resolving %d Nix packages", len(cfg.Nix.Packages))
	return BootEnvironment("derrick.yaml", cfg.Nix.Packages, registry, "")
}

// Stop is a no-op for Nix: the dev shell exits when the process ends.
func (n *NixProvider) Stop(_ *config.ProjectConfig) error { return nil }

// Shell opens an interactive Nix dev shell.
func (n *NixProvider) Shell(cfg *config.ProjectConfig) error {
	eng := NewShellEngine()
	return eng.EnterSandbox(".derrick", nil)
}

// Status checks whether the Nix environment has been initialized.
func (n *NixProvider) Status(cfg *config.ProjectConfig) (EnvironmentStatus, error) {
	_, err := exec.LookPath("nix")
	if err != nil {
		return EnvironmentStatus{Running: false, Details: "nix not installed"}, nil
	}
	return EnvironmentStatus{Running: true, Details: "nix environment available"}, nil
}
