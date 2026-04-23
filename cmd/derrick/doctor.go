package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// doctorCmd audits the environment and reports missing dependencies.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Audit the environment and report missing dependencies",
	Long: `Runs a comprehensive, read-only audit of your local environment.
It checks for necessary system dependencies (like Nix and Docker) and runs
all YAML validations without applying any auto-fixes, providing a complete
health report.

Exits non-zero when the audit surfaces one or more issues, so CI pipelines
can gate on 'derrick doctor --json' without parsing output.`,
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config

		ui.Infof("Loaded contract for project: %s (v%s)\n", cfg.Name, cfg.Version)

		report := engine.RunAudit(cfg)

		if jsonOutput {
			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))
		}

		if report.Issues > 0 {
			os.Exit(1)
		}
	}),
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
