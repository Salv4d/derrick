package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Salv4d/derrick/internal/config"
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
	Use:   "stop [alias]",
	Short: "Stop the local development environment",
	Long: `Runs before_stop hooks inside the sandbox (so they can still reach live
services), tears down the provider, then runs after_stop hooks on the host.

Passing an alias as the first argument resolves the project via the global
Derrick Hub (~/.derrick/config.yaml).

Passing --all stops all projects currently registered in the Hub that are
marked as 'running'.`,
	Args: cobra.MaximumNArgs(1),
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config
		cwd := ctx.Cwd
		projectState := ctx.State

		// ── Stop All ──────────────────────────────────────────────────────────
		if stopAll {
			hub, err := config.LoadGlobalHub()
			if err != nil {
				ui.FailFastf("Failed to load Hub: %v", err)
			}

			runningPaths := make(map[string]string)
			for _, proj := range hub.Projects {
				if proj.Path != "" {
					s, _ := state.Load(proj.Path)
					if s.Status == state.StatusRunning {
						runningPaths[s.Project] = proj.Path
					}
				}
			}

			if len(runningPaths) == 0 {
				ui.Info("No running projects found in Hub.")
				return
			}

			ui.Infof("Found %d running projects. Stopping all...", len(runningPaths))
			chain := engine.GetChain(stopChainEnv)
			var g errgroup.Group
			for name, path := range runningPaths {
				name, path := name, path
				g.Go(func() error {
					ui.Infof("Stopping: %s", name)
					return engine.ExecuteRecursive(path, "stop", profileName, chain, "", nil, nil)
				})
			}
			if err := g.Wait(); err != nil {
				ui.FailFast(err)
			}
			return
		}

		// ── Stop by Alias ─────────────────────────────────────────────────────
		if len(args) == 1 {
			alias := args[0]
			targetPath := ResolveAlias(alias, cwd)
			if targetPath != cwd {
				chain := engine.GetChain(stopChainEnv)
				if err := engine.ExecuteRecursive(targetPath, "stop", profileName, chain, cfg.Name, nil, nil); err != nil {
					ui.FailFast(fmt.Errorf("stop failed for alias '%s': %w", alias, err))
				}
				return
			}
		}

		// If we reach here, we are stopping the project in the current directory.
		// If cfg is empty (because RunDerrick bypassed config load for an alias/all),
		// it means the ResolveAlias above should have handled it.
		if cfg.Name == "" {
			ui.FailFast(fmt.Errorf("no project found in this directory. Use 'derrick stop <alias>' or 'derrick stop --all'"))
		}

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
			UseNix: false,
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

var stopAll bool

func init() {
	stopCmd.Flags().BoolVarP(&stopAll, "all", "a", false, "Stop all running projects registered in the Hub")
	rootCmd.AddCommand(stopCmd)
}
