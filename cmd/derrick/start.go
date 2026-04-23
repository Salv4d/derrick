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

// startChainEnv carries the list of project names already booting in the
// current recursive `derrick start` chain. Each child inherits it via its
// process environment, so cycles abort instead of fork-bombing.
const startChainEnv = "DERRICK_START_CHAIN"

// derrickJoinNetworkEnv is injected by a parent project into a required
// dependency's environment so the dependency joins the shared Docker network
// the parent created for cross-project container DNS.
const derrickJoinNetworkEnv = "DERRICK_JOIN_NETWORK"

var (
	startReset       bool
	startCustomFlags []string
	startDryRun      bool
	startRegister    bool
)

// startCmd boots the local development environment defined in derrick.yaml.
// An optional argument selects a project alias from the global Derrick Hub,
// allowing `derrick start react` to fetch, clone, and boot a remote project.
var startCmd = &cobra.Command{
	Use:   "start [alias]",
	Short: "Boot the local development environment",
	Long: `Reads derrick.yaml, resolves the isolation backend (Docker or Nix),
executes lifecycle hooks, and validates the environment.

Passing an alias as the first argument resolves the project via the global
Derrick Hub (~/.derrick/config.yaml) and clones it if needed.`,
	Args: cobra.MaximumNArgs(1),
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config
		cwd := ctx.Cwd
		projectState := ctx.State

		// If an alias was given, resolve it via the Hub and delegate to a
		// subprocess running inside the target directory.
		if len(args) == 1 {
			alias := args[0]
			targetPath := resolveAlias(alias, cwd)
			if targetPath != cwd {
				childArgs := []string{}
				for _, f := range startCustomFlags {
					childArgs = append(childArgs, "--flag", f)
				}
				if startReset {
					childArgs = append(childArgs, "--reset")
				}

				chain := engine.GetChain(startChainEnv)
				if err := engine.ExecuteRecursive(targetPath, "start", profileName, chain, cfg.Name, childArgs, nil); err != nil {
					ui.FailFast(fmt.Errorf("start failed for alias '%s': %w", alias, err))
				}
				return
			}
		}

		// ── Cycle detection (requires:) ────────────────────────────────────────
		chain := engine.GetChain(startChainEnv)
		if chain.Contains(cfg.Name) {
			ui.FailFast(fmt.Errorf("circular dependency detected: '%s' is already booting in this chain [%s]", cfg.Name, chain.Raw))
		}

		// ── Custom flags ───────────────────────────────────────────────────────
		activeFlags := resolveCustomFlags(cfg, cmd, startCustomFlags)
		if startReset {
			activeFlags["reset"] = true
			ui.Warning("Reset flag active: environment will be rebuilt from scratch")
		}

		flagList := ""
		if len(startCustomFlags) > 0 {
			flagList = fmt.Sprintf(" (flags: %s)", strings.Join(startCustomFlags, ", "))
		}
		ui.Successf("Project: %s  v%s  [%s]%s", cfg.Name, cfg.Version, cfg.ActiveProvider(), flagList)

		flags := engine.Flags{Active: activeFlags, Reset: startReset}

		// ── Dependency resolution ──────────────────────────────────────────────
		if len(cfg.Requires) > 0 {
			ui.Section("Dependency Resolution")

			depNames := make([]string, len(cfg.Requires))
			for i, r := range cfg.Requires {
				depNames[i] = r.Name
			}

			resolver, err := engine.NewDependencyResolver()
			if err != nil {
				ui.Warningf("Hub unavailable: %v", err)
			} else {
				if err := resolver.ResolveAndClone(cwd, depNames); err != nil {
					ui.FailFast(err)
				}
			}

			// When any requirement has connect:true, create a shared network
			// named after this project and wire both sides into it.
			sharedNetwork := ""
			for _, r := range cfg.Requires {
				if r.Connect {
					sharedNetwork = "derrick-" + cfg.Name
					break
				}
			}
			if sharedNetwork != "" {
				if err := engine.EnsureNetworks([]string{sharedNetwork}); err != nil {
					ui.FailFast(err)
				}
				cfg.Docker.Networks = appendUnique(cfg.Docker.Networks, sharedNetwork)
			}

			parentDir := filepath.Dir(cwd)

			var g errgroup.Group

			for _, dep := range cfg.Requires {
				dep := dep // shadow for goroutine
				g.Go(func() error {
					depPath := filepath.Join(parentDir, dep.Name)

					// If missing, try to resolve and clone via Hub
					if _, err := os.Stat(depPath); err != nil {
						ui.Infof("Dependency '%s' is missing locally. Attempting Hub resolution...", dep.Name)
						depPath = resolveAlias(dep.Name, cwd)
					}

					ui.Infof("Booting dependency: %s", dep.Name)
					extraArgs := []string{}
					if startDryRun {
						extraArgs = append(extraArgs, "--dry-run")
					}
					if startReset {
						extraArgs = append(extraArgs, "--reset")
					}

					extraEnv := []string{}
					if dep.Connect && sharedNetwork != "" {
						extraEnv = append(extraEnv, derrickJoinNetworkEnv+"="+sharedNetwork)
					}

					if err := engine.ExecuteRecursive(depPath, "start", profileName, chain, cfg.Name, extraArgs, extraEnv); err != nil {
						return fmt.Errorf("dependency '%s' failed to start: %w", dep.Name, err)
					}
					return nil
				})
			}

			if err := g.Wait(); err != nil {
				ui.FailFast(err)
			}
		}

		// Honour DERRICK_JOIN_NETWORK injected by a requiring parent project.
		if joinNet := os.Getenv(derrickJoinNetworkEnv); joinNet != "" {
			cfg.Docker.Networks = appendUnique(cfg.Docker.Networks, joinNet)
		}

		// ── Provider selection ─────────────────────────────────────────────────
		provider := engine.ResolveProvider(cfg)
		ui.Taskf("Checking %s availability", provider.Name())
		if err := provider.IsAvailable(); err != nil {
			ui.FailFast(err)
		}
		ui.Successf("%s is ready", provider.Name())

		// ── State ─────────────────────────────────────────────────────────────
		firstSetup := !projectState.FirstSetupCompleted

		// useSandbox gates whether hooks that run "inside the sandbox"
		// (setup, after_start, before_stop) get wrapped with `nix develop`.
		// True for nix and hybrid; docker-only has no sandbox — hooks run on
		// the host against the compose network.
		useSandbox := cfg.ActiveProvider() == "nix" || cfg.ActiveProvider() == "hybrid"

		// hostOpts: before_start / after_stop — bare host shell.
		hostOpts := engine.HookOpts{
			SetupCompleted: !firstSetup,
			ActiveFlags:    activeFlags,
			UseNix:         false,
		}
		// sandboxOpts: setup / after_start / before_stop — nix dev shell when active.
		sandboxOpts := engine.HookOpts{
			SetupCompleted: !firstSetup,
			ActiveFlags:    activeFlags,
			UseNix:         useSandbox,
		}

		// ── Environment variables ──────────────────────────────────────────────
		ui.Section("Environment")
		ui.Task("Validating environment variables")
		resolvedEnv, err := engine.ValidateAndLoadEnv(cwd, cfg, useSandbox)
		if err != nil {
			ui.FailFast(err)
		}
		hostOpts.Env = resolvedEnv
		sandboxOpts.Env = resolvedEnv
		flags.Env = resolvedEnv
		ui.Success("Environment variables loaded")

		// ── Dry-run plan ──────────────────────────────────────────────────────
		if startDryRun {
			printStartPlan(cfg, provider.Name(), activeFlags)
			return
		}

		// ── before_start hooks (host) ─────────────────────────────────────────
		if err := engine.ExecuteHooks("before_start", cfg.Hooks.BeforeStart, hostOpts); err != nil {
			ui.FailFast(err)
		}

		// ── Provision (materialize flake / compose override) ──────────────────
		ui.Sectionf("%s Provisioning", provider.Name())
		if err := provider.Provision(cfg); err != nil {
			ui.FailFast(err)
		}

		// ── setup hooks (sandbox, services not yet running) ───────────────────
		if err := engine.ExecuteHooks("setup", cfg.Hooks.Setup, sandboxOpts); err != nil {
			ui.FailFast(err)
		}

		// ── Validations (inside sandbox) ──────────────────────────────────────
		if err := engine.RunValidations(cfg.Validations, useSandbox, resolvedEnv); err != nil {
			ui.FailFast(err)
		}

		// ── Provider start (boot services) ────────────────────────────────────
		ui.Sectionf("%s Orchestration", provider.Name())
		if err := provider.Start(cfg, flags); err != nil {
			ui.FailFast(err)
		}
		ui.Success("Environment is running")

		// ── after_start hooks (sandbox, services up) ──────────────────────────
		if err := engine.ExecuteHooks("after_start", cfg.Hooks.AfterStart, sandboxOpts); err != nil {
			ui.FailFast(err)
		}

		// ── Persist state ─────────────────────────────────────────────────────
		projectState.Project = cfg.Name
		projectState.Provider = provider.Name()
		projectState.Status = state.StatusRunning
		projectState.FirstSetupCompleted = true
		projectState.StartedAt = time.Now()
		projectState.FlagsUsed = startCustomFlags
		if err := state.Save(cwd, projectState); err != nil {
			ui.Warningf("Failed to save state: %v", err)
		}

		fmt.Println()
		if firstSetup {
			ui.Successf("%s is ready! (first setup complete)", cfg.Name)
		} else {
			ui.Successf("%s is ready!", cfg.Name)
		}

		// ── Register project in Hub ───────────────────────────────────────────
		if startRegister {
			ui.Section("Hub Registration")
			ui.Task("Detecting Git remote URL")
			cmd := exec.Command("git", "remote", "get-url", "origin")
			cmd.Dir = cwd
			out, err := cmd.Output()
			if err != nil {
				ui.Warning("Could not detect Git remote 'origin'. Use 'derrick hub add' to register manually.")
			} else {
				url := strings.TrimSpace(string(out))
				hub, err := config.LoadGlobalHub()
				if err != nil {
					ui.Warningf("Could not load global hub: %v", err)
				} else {
					absPath, _ := filepath.Abs(cwd)
					hub.Projects[cfg.Name] = config.HubProject{URL: url, Path: absPath}
					if err := hub.Save(); err != nil {
						ui.Warningf("Could not save hub config: %v", err)
					} else {
						ui.Successf("Project '%s' registered in Hub at %s", cfg.Name, absPath)
					}
				}
			}
		}
	}),
}

// printStartPlan describes what `derrick start` would do without
// executing hooks, validations, or provider commands.
func printStartPlan(cfg *config.ProjectConfig, providerName string, activeFlags map[string]bool) {
	ui.Section("Dry-run plan")
	ui.Infof("Project:  %s v%s", cfg.Name, cfg.Version)
	ui.Infof("Provider: %s", providerName)

	if len(activeFlags) > 0 {
		flagNames := make([]string, 0, len(activeFlags))
		for name := range activeFlags {
			flagNames = append(flagNames, name)
		}
		ui.Infof("Flags:    %s", strings.Join(flagNames, ", "))
	}

	printHookStage("before_start (host)", cfg.Hooks.BeforeStart)
	printHookStage("setup (sandbox, before services)", cfg.Hooks.Setup)
	printHookStage("after_start (sandbox, services up)", cfg.Hooks.AfterStart)

	if len(cfg.Validations) > 0 {
		ui.Info("Validations:")
		for _, v := range cfg.Validations {
			fmt.Printf("    %-30s  %s\n", v.Name, v.Command)
		}
	}

	if cfg.ActiveProvider() == "docker" && cfg.Docker.Compose != "" {
		ui.Infof("Docker:   would run `docker compose -f %s up -d`", cfg.Docker.Compose)
	}
	if cfg.ActiveProvider() == "nix" && len(cfg.Nix.Packages) > 0 {
		names := make([]string, len(cfg.Nix.Packages))
		for i, p := range cfg.Nix.Packages {
			names[i] = p.Name
		}
		ui.Infof("Nix:      would build flake with packages [%s]", strings.Join(names, ", "))
	}

	fmt.Println()
	ui.Success("Dry-run complete. No side effects.")
}

// printHookStage renders a dry-run entry for one hook stage. Skipping the
// header entirely when the stage is empty keeps the plan output tight.
func printHookStage(label string, hooks []config.Hook) {
	if len(hooks) == 0 {
		return
	}
	ui.Infof("%s hooks:", label)
	for _, h := range hooks {
		conds := strings.Join(h.When, ", ")
		if conds == "" {
			conds = "always"
		}
		fmt.Printf("    [%s] %s\n", conds, h.Run)
	}
}

// resolveAlias looks up an alias in the Derrick Hub and returns the local path
// for the project, cloning it if necessary.
func resolveAlias(alias string, cwd string) string {
	ui.Section("Hub Resolution")
	ui.Taskf("Looking up alias: %s", alias)

	hub, err := config.LoadGlobalHub()
	if err != nil {
		ui.FailFastf("Failed to load Derrick Hub: %v", err)
	}

	proj, err := hub.ResolveAlias(alias)
	if err != nil {
		ui.FailFast(fmt.Errorf("alias '%s' not found in Hub.\nAdd it to ~/.derrick/config.yaml:\n\nprojects:\n  %s: <git-url>", alias, alias))
	}

	// 1. If we have a tracked local path, prefer it.
	if proj.Path != "" {
		if _, err := os.Stat(proj.Path); err == nil {
			ui.Infof("Project '%s' found at registered path: %s", alias, proj.Path)
			return proj.Path
		}
		ui.Warningf("Project '%s' not found at registered path %s, falling back to workspace.", alias, proj.Path)
	}

	// 2. Check sibling directory (legacy behavior)
	parentDir := filepath.Dir(cwd)
	siblingPath := filepath.Join(parentDir, alias)
	if _, err := os.Stat(siblingPath); err == nil {
		ui.Infof("Project '%s' found as sibling: %s", alias, siblingPath)
		return siblingPath
	}

	// 3. Use workspace
	workspace := hub.Workspace
	if workspace == "" {
		// fallback if somehow empty
		home, _ := os.UserHomeDir()
		workspace = filepath.Join(home, "derrick-projects")
	}

	targetPath := filepath.Join(workspace, alias)

	if _, err := os.Stat(targetPath); err == nil {
		ui.Infof("Project '%s' already exists in workspace at %s", alias, targetPath)
		return targetPath
	}

	// 4. Clone to workspace
	ui.Taskf("Cloning %s from %s into workspace", alias, proj.URL)
	if err := os.MkdirAll(workspace, 0755); err != nil {
		ui.FailFastf("Failed to create workspace directory %s: %v", workspace, err)
	}

	cloneCmd := exec.Command("git", "clone", proj.URL, targetPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		ui.FailFastf("Clone failed: %v", err)
	}

	// Record the local path for next time
	hub.Projects[alias] = config.HubProject{URL: proj.URL, Path: targetPath}
	_ = hub.Save()

	ui.Successf("Cloned '%s' to %s", alias, targetPath)
	return targetPath
}

// resolveCustomFlags builds the active flags map by checking which project-defined
// flags were passed on the command line.
func resolveCustomFlags(cfg *config.ProjectConfig, cmd *cobra.Command, rawFlags []string) map[string]bool {
	active := make(map[string]bool)
	for _, name := range rawFlags {
		active[name] = true
	}
	return active
}

// appendUnique appends s to slice only if it is not already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func init() {
	startCmd.Flags().BoolVar(&startReset, "reset", false, "Rebuild the environment from scratch")
	startCmd.Flags().StringSliceVar(&startCustomFlags, "flag", nil, "Custom project flags (e.g. --flag seed-db)")
	startCmd.Flags().BoolVar(&startDryRun, "dry-run", false, "Print what would happen without executing hooks or starting the provider")
	startCmd.Flags().BoolVar(&startRegister, "register", false, "Register this project in the global Derrick Hub")
	rootCmd.AddCommand(startCmd)
}
