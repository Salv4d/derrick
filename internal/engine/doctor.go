package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		ui.SubTask("Checking Docker Compose file")
		check := AuditCheck{Name: "docker-compose-file"}
		if _, err := os.Stat(cfg.Docker.Compose); err == nil {
			ui.Success("OK")
			check.OK = true

			ui.SubTask("Validating Docker Compose syntax")
			checkSyntax := AuditCheck{Name: "docker-compose-syntax"}
			if _, err := FirstService(cfg.Docker.Compose); err == nil {
				ui.Success("OK")
				checkSyntax.OK = true
			} else {
				ui.Error("INVALID")
				ui.Warningf("  -> %v", err)
				checkSyntax.Error = err.Error()
				report.Issues++
			}
			report.Tools = append(report.Tools, checkSyntax)
		} else {
			ui.Error("MISSING")
			ui.Warningf("  -> File not found: %s", cfg.Docker.Compose)
			check.Error = fmt.Sprintf("compose file %q not found", cfg.Docker.Compose)
			report.Issues++
		}
		report.Tools = append(report.Tools, check)

		ui.SubTask("Checking Docker daemon")
		check = AuditCheck{Name: "docker"}
		if !IsDockerInstalled() {
			ui.Error("MISSING")
			ui.Warning("  -> Docker binary not found in PATH.")
			check.Error = "docker binary not found"
			report.Issues++
		} else if err := exec.Command("docker", "info").Run(); err != nil {
			ui.Error("PERMISSION DENIED OR NOT RUNNING")
			ui.Warning("  -> Ensure Docker is running and your user is in the 'docker' group.")
			check.Error = "cannot connect to docker daemon"
			report.Issues++
		} else {
			ui.Success("OK")
			check.OK = true
		}
		report.Tools = append(report.Tools, check)
	}

	if cfg.EnvManagement.BaseFile != "" {
		ui.SubTask("Checking Env Base file")
		check := AuditCheck{Name: "env-base-file"}
		if _, err := os.Stat(cfg.EnvManagement.BaseFile); err == nil {
			ui.Success("OK")
			check.OK = true
		} else {
			ui.Error("MISSING")
			ui.Warningf("  -> File not found: %s", cfg.EnvManagement.BaseFile)
			check.Error = fmt.Sprintf("env base file %q not found", cfg.EnvManagement.BaseFile)
			report.Issues++
		}
		report.Tools = append(report.Tools, check)
	}

	if len(cfg.Requires) > 0 {
		ui.SubTask("Checking required projects")
		check := AuditCheck{Name: "requirements"}
		missing := []string{}
		parentDir := ".." // Requirements are expected to be siblings
		for _, req := range cfg.Requires {
			path := filepath.Join(parentDir, req.Name)
			if _, err := os.Stat(path); err != nil {
				missing = append(missing, req.Name)
			}
		}

		if len(missing) == 0 {
			ui.Success("OK")
			check.OK = true
		} else {
			ui.Error("MISSING")
			ui.Warningf("  -> The following required projects are missing: %s", strings.Join(missing, ", "))
			ui.Warning("  -> Run 'derrick start' to resolve and clone them.")
			check.Error = fmt.Sprintf("missing requirements: %s", strings.Join(missing, ", "))
			report.Issues++
		}
		report.Tools = append(report.Tools, check)
	}

	if len(cfg.Nix.Packages) == 0 && cfg.Docker.Compose == "" {
		ui.SubTask("Checking for active provider")
		ui.Error("NONE")
		ui.Warning("  -> No Docker Compose or Nix packages defined. Project is empty.")
		check := AuditCheck{Name: "provider", Error: "no active provider"}
		report.Tools = append(report.Tools, check)
		report.Issues++
	}

	canUseNixBubble := len(cfg.Nix.Packages) > 0 && IsNixInstalled()

	var auditSandbox string
	if canUseNixBubble {
		ui.SubTask("Bootstrapping dry-run Nix sandbox for validations")
		check := AuditCheck{Name: "nix-sandbox"}
		// Doctor is documented read-only: generate the flake in a temp
		// directory and clean up, so .derrick/ is never mutated.
		tmp, err := os.MkdirTemp("", "derrick-doctor-*")
		if err != nil {
			ui.Error("FAILED")
			ui.Warningf("  -> Failed to create audit sandbox: %v", err)
			check.Error = "failed to create temp dir"
			report.Issues++
			canUseNixBubble = false
		} else {
			defer os.RemoveAll(tmp)
			if err := BootEnvironment("derrick.yaml", cfg.Nix.Packages, cfg.Nix.Registry, tmp); err != nil {
				ui.Error("FAILED")
				ui.Warningf("  -> Failed to bootstrap Nix evaluation sandbox: %v", err)
				check.Error = "nix evaluation failed"
				report.Issues++
				canUseNixBubble = false
			} else {
				auditSandbox = tmp
				ui.Success("OK")
				check.OK = true
			}
		}
		report.Tools = append(report.Tools, check)
	}

	if len(cfg.Validations) > 0 {
		ui.Section("State Validations")
		for _, check := range cfg.Validations {
			ui.SubTaskf("Checking %s", check.Name)
			result := AuditCheck{Name: check.Name}

			runner := &Runner{
				UseNix: canUseNixBubble,
				NixDir: auditSandbox,
				Silent: true,
			}
			err := runner.Run(check.Command)
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
