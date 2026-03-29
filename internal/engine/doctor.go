package engine

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

func RunAudit(cfg *config.ProjectConfig) {
	ui.Info("Running Derrick Environment Audit (Doctor)...\n")
	issues := 0

	useNix := len(cfg.Dependencies.NixPackages) > 0
	if useNix {
		fmt.Printf("  Checking Nix package manager...")
		if IsNixInstalled() {
			ui.SuccessInline("OK")
		} else {
			ui.ErrorInline("MISSING")
			ui.WarningInline("-> Run: curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install")
			issues++
		}
	}

	if cfg.Dependencies.Dockerfile != "" {
		fmt.Printf("  Checking Docker daemon... ")
		if IsDockerInstalled() {
			ui.SuccessInline("OK")
		} else {
			ui.ErrorInline("MISSING OR PERMISSION DENIED")
			ui.WarningInline("    -> Ensure Docker is running and your user is in the 'docker' group.")
			issues++
		}
	}

	canUseNixBubble := useNix && IsNixInstalled()

	if len(cfg.Validations) > 0 {
		fmt.Println("\n  State Validations:")
		for _, check := range cfg.Validations {
			fmt.Printf("    Checking %s... ", check.Name)
			err := executeCommand(check.Command, canUseNixBubble)
			if err == nil {
				ui.SuccessInline("OK")
			} else {
				ui.ErrorInline("FAILED")
				fmt.Printf("      -> Error: %v\n", err)
				
				if check.AutoFix != "" {
					fmt.Printf("      -> Fix available: The 'start' command will run '%s' to attempt recovery.\n", check.AutoFix)
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