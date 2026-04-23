package main

import (
	"fmt"
	"os"
	"os/exec"
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

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFast(err)
		}
		_ = ui.SetLogFile(cwd)

		// Load persisted state first so stop hooks can see the flags the user
		// passed to `derrick start`. Without this, hooks with when: flag:<name>
		// never fire on stop.
		projectState, _ := state.Load(cwd)
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
			childChain := appendStopChain(os.Getenv(stopChainEnv), cfg.Name)

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
					chain := parseStopChain(os.Getenv(stopChainEnv))
					if chain[dep.Name] {
						return nil // Already stopping in this chain
					}

					ui.Infof("Stopping dependency: %s", dep.Name)
					cmdArgs := []string{"stop"}
					if profileName != "" {
						cmdArgs = append(cmdArgs, "--profile", profileName)
					}

					depCmd := exec.Command(os.Args[0], cmdArgs...)
					depCmd.Dir = depPath
					depCmd.Stdout = os.Stdout
					depCmd.Stderr = os.Stderr
					depCmd.Stdin = os.Stdin
					depCmd.Env = append(os.Environ(), stopChainEnv+"="+childChain)

					if err := depCmd.Run(); err != nil {
						return fmt.Errorf("dependency '%s' failed to stop: %w", dep.Name, err)
					}
					return nil
				})
			}

			if err := g.Wait(); err != nil {
				ui.FailFast(err)
			}
		}
	},
}

// parseStopChain deserializes DERRICK_STOP_CHAIN into a set for membership checks.
func parseStopChain(raw string) map[string]bool {
	chain := make(map[string]bool)
	if raw == "" {
		return chain
	}
	for _, name := range strings.Split(raw, ",") {
		if name = strings.TrimSpace(name); name != "" {
			chain[name] = true
		}
	}
	return chain
}

// appendStopChain returns a new chain string with the given project name appended.
func appendStopChain(raw, name string) string {
	if raw == "" {
		return name
	}
	return raw + "," + name
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
