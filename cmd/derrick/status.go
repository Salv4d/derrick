package main

import (
	"encoding/json"
	"fmt"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

type statusReport struct {
	Project   string `json:"project"`
	Version   string `json:"version"`
	Provider  string `json:"provider"`
	Running   bool   `json:"running"`
	Details   string `json:"details,omitempty"`
	LastKnown string `json:"last_known,omitempty"`
}

// statusCmd reports whether the project's environment is currently running.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Report whether the project's environment is running",
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config
		projectState := ctx.State

		provider := engine.ResolveProvider(cfg)
		status, err := provider.Status(cfg)
		if err != nil {
			ui.FailFast(err)
		}

		report := statusReport{
			Project:  cfg.Name,
			Version:  cfg.Version,
			Provider: provider.Name(),
			Running:  status.Running,
			Details:  status.Details,
		}
		report.LastKnown = string(projectState.Status)

		if jsonOutput {
			out, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				ui.FailFast(err)
			}
			fmt.Println(string(out))
			return
		}

		fmt.Printf("project:  %s (v%s)\n", report.Project, report.Version)
		fmt.Printf("provider: %s\n", report.Provider)
		if report.Running {
			ui.Successf("running")
		} else {
			ui.Warning("not running")
		}
		if report.Details != "" {
			fmt.Printf("details:  %s\n", report.Details)
		}
		if report.LastKnown != "" {
			fmt.Printf("last known: %s\n", report.LastKnown)
		}
	}),
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
