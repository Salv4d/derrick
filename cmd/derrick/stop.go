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

		// Load persisted state first so stop hooks can see the flags the user
		// passed to `derrick start`. Without this, hooks with when: flag:<name>
		// never fire on stop.
		projectState, _ := state.Load(cwd)
		activeFlags := make(map[string]bool)
		for _, f := range projectState.FlagsUsed {
			activeFlags[f] = true
		}

		provider := engine.ResolveProvider(cfg)
		ui.Infof("Stopping %s environment: %s", provider.Name(), cfg.Name)

		if err := provider.Stop(cfg); err != nil {
			ui.FailFast(err)
		}

		hookOpts := engine.HookOpts{
			SetupCompleted: true,
			ActiveFlags:    activeFlags,
			UseNix:         cfg.ActiveProvider() == "nix",
		}
		if err := engine.ExecuteHooks("stop", cfg.Hooks.Stop, hookOpts); err != nil {
			ui.FailFast(err)
		}

		projectState.Status = state.StatusStopped
		projectState.StoppedAt = time.Now()
		_ = state.Save(cwd, projectState)

		ui.Success("Environment stopped.")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
