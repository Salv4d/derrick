package main

import (
	"os"
	"time"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/state"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// stopCmd tears down the local development environment.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the local development environment",
	Long: `Stops all running containers or closes the Nix dev shell, then executes
any defined stop lifecycle hooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			ui.FailFast(err)
		}

		cwd, _ := os.Getwd()

		provider := engine.ResolveProvider(cfg)
		ui.Infof("Stopping %s environment: %s", provider.Name(), cfg.Name)

		if err := provider.Stop(cfg); err != nil {
			ui.FailFast(err)
		}

		hookOpts := engine.HookOpts{
			SetupCompleted: true,
			UseNix:         cfg.ActiveProvider() == "nix",
		}
		if err := engine.ExecuteHooks("stop", cfg.Hooks.Stop, hookOpts); err != nil {
			ui.FailFast(err)
		}

		// Update persisted state.
		projectState, _ := state.Load(cwd)
		projectState.Status = state.StatusStopped
		projectState.StoppedAt = time.Now()
		_ = state.Save(cwd, projectState)

		ui.Success("Environment stopped.")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
