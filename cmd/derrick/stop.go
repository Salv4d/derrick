package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/state"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// stopChainEnv carries the list of project names already stopping in the
// current recursive `derrick stop` chain.
const stopChainEnv = "DERRICK_STOP_CHAIN"

// stopCmd tears down the local development environment.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the local development environment",
	Long: `Runs before_stop hooks inside the sandbox (so they can still reach live
services), tears down the provider, then runs after_stop hooks on the host.`,
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config
		cwd := ctx.Cwd
		projectState := ctx.State

		activeFlags := make(map[string]bool)
		for _, f := range projectState.FlagsUsed {
			activeFlags[f] = true
		}

		flagList := ""
		if len(projectState.FlagsUsed) > 0 {
			flagList = fmt.Sprintf(" (flags: %s)", strings.Join(projectState.FlagsUsed, ", "))
		}

		provider := engine.ResolveProvider(cfg)
		ui.Infof("Stopping %s environment: %s%s", provider.Name(), cfg.Name, flagList)

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

		// ── Recursive dependency stopping ─────────────────────────────────────
		if len(cfg.Requires) > 0 {
			ui.Section("Stopping Dependencies")
			parentDir := filepath.Dir(cwd)
			chain := engine.GetChain(stopChainEnv)

			var g errgroup.Group

			for _, dep := range cfg.Requires {
				dep := dep // shadow for goroutine
				g.Go(func() error {
					depPath := filepath.Join(parentDir, dep.Name)

					if _, err := os.Stat(depPath); err != nil {
						ui.Warningf("Dependency '%s' directory not found at %s, skipping stop.", dep.Name, depPath)
						return nil
					}

					// Cycle detection
					if chain.Contains(dep.Name) {
						return nil // Already stopping in this chain
					}

					ui.Infof("Stopping dependency: %s", dep.Name)
					if err := engine.ExecuteRecursive(depPath, "stop", profileName, chain, cfg.Name, nil, nil); err != nil {
						return fmt.Errorf("dependency '%s' failed to stop: %w", dep.Name, err)
					}
					return nil
				})
			}

			if err := g.Wait(); err != nil {
				ui.FailFast(err)
			}
		}
	}),
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
