package main

import (
	"fmt"
	"os"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/state"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// statusCmd reports whether the project's environment is currently running.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Report whether the project's environment is running",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			ui.FailFast(err)
		}

		cwd, _ := os.Getwd()
		projectState, _ := state.Load(cwd)

		provider := engine.ResolveProvider(cfg)
		status, err := provider.Status(cfg)
		if err != nil {
			ui.FailFast(err)
		}

		fmt.Printf("project:  %s (v%s)\n", cfg.Name, cfg.Version)
		fmt.Printf("provider: %s\n", provider.Name())

		if status.Running {
			ui.Successf("running")
		} else {
			ui.Warning("not running")
		}
		if status.Details != "" {
			fmt.Printf("details:  %s\n", status.Details)
		}
		if projectState != nil && projectState.Status != "" {
			fmt.Printf("last known: %s\n", projectState.Status)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
