package engine

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// RunAudit performs a comprehensive, read-only audit of the local environment.
func RunAudit(cfg *config.ProjectConfig) {
	ui.Section("Environment Audit")
	issues := 0

	useNix := cfg.ActiveProvider() == "nix"
	if useNix {
		ui.SubTask("Checking Nix package manager")
		if IsNixInstalled() {
			ui.Success("OK")
		} else {
			ui.Error("MISSING")
			ui.Warning("  -> Run: curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install")
			issues++
		}
	}

	if cfg.Docker.Compose != "" {
		ui.SubTask("Checking Docker daemon")
		if IsDockerInstalled() {
			ui.Success("OK")
		} else {
			ui.Error("MISSING OR PERMISSION DENIED")
			ui.Warning("  -> Ensure Docker is running and your user is in the 'docker' group.")
			issues++
		}
	}

	canUseNixBubble := useNix && IsNixInstalled()

	if canUseNixBubble {
		ui.SubTask("Bootstrapping dry-run Nix sandbox for validations")
		err := BootEnvironment("derrick.yaml", cfg.Nix.Packages, cfg.Nix.Registry, "")
		if err != nil {
			ui.Warningf("  -> Failed to bootstrap Nix evaluation sandbox: %v", err)
			canUseNixBubble = false
		} else {
			ui.Success("OK")
		}
	}

	if len(cfg.Validations) > 0 {
		ui.Section("State Validations")
		for _, check := range cfg.Validations {
			ui.SubTaskf("Checking %s", check.Name)
			err := executeCommand(check.Command, canUseNixBubble, nil)
			if err == nil {
				ui.Success("OK")
			} else {
				ui.Error("FAILED")
				ui.Errorf("     Error: %v", err)
				if check.AutoFix != "" {
					ui.Warningf("     Fix available: 'start' will run '%s' to attempt recovery.", check.AutoFix)
				}
				issues++
			}
		}
	}

	fmt.Println()
	if issues == 0 {
		ui.Success("Environment is healthy. Ready to run 'derrick start'.")
	} else {
		ui.Warningf("Found %d issue(s). Please fix them before starting.", issues)
	}
}
