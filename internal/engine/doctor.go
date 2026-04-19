package engine

import (
	"fmt"
	"os"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// AuditCheck is the result of a single audit step.
type AuditCheck struct {
	Name  string `json:"name"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// AuditReport summarizes the doctor run for machine consumers.
type AuditReport struct {
	Project     string       `json:"project"`
	Provider    string       `json:"provider"`
	Tools       []AuditCheck `json:"tools"`
	Validations []AuditCheck `json:"validations"`
	Issues      int          `json:"issues"`
}

// RunAudit performs a comprehensive, read-only audit of the local environment
// and returns a structured report alongside any UI output.
func RunAudit(cfg *config.ProjectConfig) AuditReport {
	report := AuditReport{
		Project:  cfg.Name,
		Provider: cfg.ActiveProvider(),
	}

	ui.Section("Environment Audit")

	useNix := cfg.ActiveProvider() == "nix"
	if useNix {
		ui.SubTask("Checking Nix package manager")
		check := AuditCheck{Name: "nix"}
		if IsNixInstalled() {
			ui.Success("OK")
			check.OK = true
		} else {
			ui.Error("MISSING")
			ui.Warning("  -> Run: curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install")
			check.Error = "nix not found in PATH"
			report.Issues++
		}
		report.Tools = append(report.Tools, check)
	}

	if cfg.Docker.Compose != "" {
		ui.SubTask("Checking Docker daemon")
		check := AuditCheck{Name: "docker"}
		if IsDockerInstalled() {
			ui.Success("OK")
			check.OK = true
		} else {
			ui.Error("MISSING OR PERMISSION DENIED")
			ui.Warning("  -> Ensure Docker is running and your user is in the 'docker' group.")
			check.Error = "docker not available"
			report.Issues++
		}
		report.Tools = append(report.Tools, check)
	}

	canUseNixBubble := useNix && IsNixInstalled()

	var auditSandbox string
	if canUseNixBubble {
		ui.SubTask("Bootstrapping dry-run Nix sandbox for validations")
		// Doctor is documented read-only: generate the flake in a temp
		// directory and clean up, so .derrick/ is never mutated.
		tmp, err := os.MkdirTemp("", "derrick-doctor-*")
		if err != nil {
			ui.Warningf("  -> Failed to create audit sandbox: %v", err)
			canUseNixBubble = false
		} else {
			defer os.RemoveAll(tmp)
			if err := BootEnvironment("derrick.yaml", cfg.Nix.Packages, cfg.Nix.Registry, tmp); err != nil {
				ui.Warningf("  -> Failed to bootstrap Nix evaluation sandbox: %v", err)
				canUseNixBubble = false
			} else {
				auditSandbox = tmp
				ui.Success("OK")
			}
		}
	}

	if len(cfg.Validations) > 0 {
		ui.Section("State Validations")
		for _, check := range cfg.Validations {
			ui.SubTaskf("Checking %s", check.Name)
			result := AuditCheck{Name: check.Name}
			err := executeCommandIn(check.Command, canUseNixBubble, auditSandbox, nil)
			if err == nil {
				ui.Success("OK")
				result.OK = true
			} else {
				ui.Error("FAILED")
				ui.Errorf("     Error: %v", err)
				if check.AutoFix != "" {
					ui.Warningf("     Fix available: 'start' will run '%s' to attempt recovery.", check.AutoFix)
				}
				result.Error = err.Error()
				report.Issues++
			}
			report.Validations = append(report.Validations, result)
		}
	}

	if !ui.Quiet {
		fmt.Println()
		if report.Issues == 0 {
			ui.Success("Environment is healthy. Ready to run 'derrick start'.")
		} else {
			ui.Warningf("Found %d issue(s). Please fix them before starting.", report.Issues)
		}
	}

	return report
}
