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
	Long: `Runs before_stop hooks inside the sandbox (so they can still reach live
services), tears down the provider, then runs after_stop hooks on the host.`,
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

		useSandbox := cfg.ActiveProvider() == "nix" || cfg.ActiveProvider() == "hybrid"

		// before_stop runs while services are still reachable (drain, DB
		// dumps, cache flush) — inside the sandbox so tooling is on PATH.
		sandboxOpts := engine.HookOpts{
			SetupCompleted: true,
			ActiveFlags:    activeFlags,
			UseNix:         useSandbox,
		}
		if err := engine.ExecuteHooks("before_stop", cfg.Hooks.BeforeStop, sandboxOpts); err != nil {
			ui.FailFast(err)
		}

		if err := provider.Stop(cfg); err != nil {
			ui.FailFast(err)
		}

		// after_stop runs on the host after teardown — the sandbox may still
		// work but services are gone, so host is the honest contract.
		hostOpts := engine.HookOpts{
			SetupCompleted: true,
			ActiveFlags:    activeFlags,
			UseNix:         false,
		}
		if err := engine.ExecuteHooks("after_stop", cfg.Hooks.AfterStop, hostOpts); err != nil {
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
