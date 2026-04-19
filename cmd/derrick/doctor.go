package main

import (
	"encoding/json"
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// doctorCmd audits the environment and reports missing dependencies.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Audits the environment and reports missing dependencies",
	Long: `Runs a comprehensive, read-only audit of your local environment.
It checks for necessary system dependencies (like Nix and Docker) and runs
all YAML validations without applying any auto-fixes, providing a complete
health report.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			ui.FailFast(err)
		}

		ui.Infof("Loaded contract for project: %s (v%s)\n", cfg.Name, cfg.Version)

		report := engine.RunAudit(cfg)

		if jsonOutput {
			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
