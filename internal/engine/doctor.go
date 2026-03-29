package engine

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

func RunAudit(cfg *config.ProjectConfig) {
	ui.Section("Environment Audit")
	issues := 0

	useNix := len(cfg.Dependencies.NixPackages) > 0
	if useNix {
		ui.Task("Checking Nix package manager")
		if IsNixInstalled() {
			ui.Success("OK")
		} else {
			ui.Error("MISSING")
			ui.Warning("  -> Run: curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install")
			issues++
		}
	}

	if cfg.Dependencies.Dockerfile != "" {
		ui.Task("Checking Docker daemon")
		if IsDockerInstalled() {
			ui.Success("OK")
		} else {
			ui.Error("MISSING OR PERMISSION DENIED")
			ui.Warning("  -> Ensure Docker is running and your user is in the 'docker' group.")
			issues++
		}
	}

	canUseNixBubble := useNix && IsNixInstalled()

	if len(cfg.Validations) > 0 {
		ui.Section("State Validations")
		for _, check := range cfg.Validations {
			ui.SubTaskf("Checking %s", check.Name)
			err := executeCommand(check.Command, canUseNixBubble)
			if err == nil {
				ui.Success("OK")
			} else {
				ui.Error("FAILED")
				ui.Errorf("     Error: %v", err)

				if check.AutoFix != "" {
					ui.Warningf("     Fix available: The 'start' command will run '%s' to attempt recovery.", check.AutoFix)
				}
				issues++
			}
		}
	}

	fmt.Println()
	if issues == 0 {
		ui.Success("Your environment is perfectly healthy! You are ready to run 'derrick start'.")
	} else {
		ui.Warningf("Found %d issue(s) in your environment. Please fix them before starting.", issues)
	}
}
