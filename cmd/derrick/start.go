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
)

// startChainEnv carries the list of project names already booting in the
// current recursive `derrick start` chain. Each child inherits it via its
// process environment, so cycles abort instead of fork-bombing.
const startChainEnv = "DERRICK_START_CHAIN"

var (
	startReset       bool
	startCustomFlags []string
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
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFastf("Failed to get working directory: %v", err)
		}

		// If an alias was given, resolve it via the Hub and delegate to a
		// subprocess running inside the target directory. This replaces the
		// old os.Chdir call and keeps the parent process's working directory
		// untouched — which matters because relative paths in derrick.yaml
		// (compose files, env base files, etc.) are resolved against the
		// process cwd.
		if len(args) == 1 {
			alias := args[0]
			targetPath := resolveAlias(alias, cwd)
			if targetPath != cwd {
				childArgs := []string{"start"}
				if profileName != "" {
					childArgs = append(childArgs, "--profile", profileName)
				}
				for _, f := range startCustomFlags {
					childArgs = append(childArgs, "--flag", f)
				}
				if startReset {
					childArgs = append(childArgs, "--reset")
				}
				child := exec.Command(os.Args[0], childArgs...)
				child.Dir = targetPath
				child.Stdout = os.Stdout
				child.Stderr = os.Stderr
				child.Stdin = os.Stdin
				if err := child.Run(); err != nil {
					ui.FailFast(fmt.Errorf("start failed for alias '%s': %w", alias, err))
				}
				return
			}
		}

		// ── Configuration ──────────────────────────────────────────────────────
		ui.Section("Configuration")

		if profileName != "" {
			ui.Taskf("Loading %s (profile: %s)", configFile, profileName)
		} else {
			ui.Taskf("Loading %s", configFile)
		}

		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			ui.FailFast(err)
		}
		ui.Successf("Project: %s  v%s  [%s]", cfg.Name, cfg.Version, cfg.ActiveProvider())

		// ── Cycle detection (requires:) ────────────────────────────────────────
		chain := parseStartChain(os.Getenv(startChainEnv))
		if chain[cfg.Name] {
			ui.FailFast(fmt.Errorf("circular dependency detected: '%s' is already booting in this chain [%s]", cfg.Name, os.Getenv(startChainEnv)))
		}

		// ── Custom flags ───────────────────────────────────────────────────────
		activeFlags := resolveCustomFlags(cfg, cmd, startCustomFlags)
		if startReset {
			activeFlags["reset"] = true
		}

		flags := engine.Flags{Active: activeFlags, Reset: startReset}

		// ── Dependency resolution ──────────────────────────────────────────────
		if len(cfg.Requires) > 0 {
			ui.Section("Dependency Resolution")
			resolver, err := engine.NewDependencyResolver()
			if err != nil {
				ui.Warningf("Hub unavailable: %v", err)
			} else {
				if err := resolver.ResolveAndClone(cwd, cfg.Requires); err != nil {
					ui.FailFast(err)
				}
			}

			parentDir := filepath.Dir(cwd)
			childChain := appendStartChain(os.Getenv(startChainEnv), cfg.Name)
			for _, dep := range cfg.Requires {
				depPath := filepath.Join(parentDir, dep)
				ui.Infof("Booting dependency: %s", dep)
				cmdArgs := []string{"start"}
				if profileName != "" {
					cmdArgs = append(cmdArgs, "--profile", profileName)
				}
				depCmd := exec.Command(os.Args[0], cmdArgs...)
				depCmd.Dir = depPath
				depCmd.Stdout = os.Stdout
				depCmd.Stderr = os.Stderr
				depCmd.Stdin = os.Stdin
				depCmd.Env = append(os.Environ(), startChainEnv+"="+childChain)
				if err := depCmd.Run(); err != nil {
					ui.FailFast(fmt.Errorf("dependency '%s' failed to start: %w", dep, err))
				}
			}
		}

		// ── Provider selection ─────────────────────────────────────────────────
		provider := engine.ResolveProvider(cfg)
		ui.Taskf("Checking %s availability", provider.Name())
		if err := provider.IsAvailable(); err != nil {
			ui.FailFast(err)
		}
		ui.Successf("%s is ready", provider.Name())

		// ── State ─────────────────────────────────────────────────────────────
		projectState, err := state.Load(cwd)
		if err != nil {
			ui.Warningf("Could not read state file: %v", err)
			projectState = &state.EnvironmentState{Status: state.StatusUnknown}
		}
		firstSetup := !projectState.FirstSetupCompleted

		hookOpts := engine.HookOpts{
			SetupCompleted: !firstSetup,
			ActiveFlags:    activeFlags,
			UseNix:         cfg.ActiveProvider() == "nix",
		}

		// ── Environment variables ──────────────────────────────────────────────
		ui.Section("Environment")
		ui.Task("Validating environment variables")
		resolvedEnv, err := engine.ValidateAndLoadEnv(cwd, cfg, hookOpts.UseNix)
		if err != nil {
			ui.FailFast(err)
		}
		hookOpts.Env = resolvedEnv
		flags.Env = resolvedEnv
		ui.Success("Environment variables loaded")

		// ── Pre-start hooks ───────────────────────────────────────────────────
		if err := engine.ExecuteHooks("start (pre)", cfg.Hooks.Start, hookOpts); err != nil {
			ui.FailFast(err)
		}

		// ── Validations ───────────────────────────────────────────────────────
		if err := engine.RunValidations(cfg.Validations, hookOpts.UseNix, resolvedEnv); err != nil {
			ui.FailFast(err)
		}

		// ── Provider start ────────────────────────────────────────────────────
		ui.Sectionf("%s Orchestration", provider.Name())
		if err := provider.Start(cfg, flags); err != nil {
			ui.FailFast(err)
		}
		ui.Success("Environment is running")

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
	},
}

// resolveAlias looks up an alias in the Derrick Hub and returns the local path
// for the project, cloning it if necessary.
func resolveAlias(alias string, cwd string) string {
	ui.Section("Hub Resolution")
	ui.Taskf("Looking up alias: %s", alias)

	resolver, err := engine.NewDependencyResolver()
	if err != nil {
		ui.FailFastf("Failed to load Derrick Hub: %v", err)
	}

	parentDir := filepath.Dir(cwd)
	targetPath := filepath.Join(parentDir, alias)

	if _, err := os.Stat(targetPath); err == nil {
		ui.Infof("Project '%s' already exists at %s", alias, targetPath)
		return targetPath
	}

	gitURL, err := resolver.Hub.ResolveAlias(alias)
	if err != nil {
		ui.FailFast(fmt.Errorf("alias '%s' not found in Hub.\nAdd it to ~/.derrick/config.yaml:\n\nprojects:\n  %s: <git-url>", alias, alias))
	}

	ui.Taskf("Cloning %s from %s", alias, gitURL)
	cloneCmd := exec.Command("git", "clone", gitURL, targetPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		ui.FailFastf("Clone failed: %v", err)
	}

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

// parseStartChain deserializes DERRICK_START_CHAIN into a set for membership checks.
func parseStartChain(raw string) map[string]bool {
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

// appendStartChain returns a new chain string with the given project name appended.
func appendStartChain(raw, name string) string {
	if raw == "" {
		return name
	}
	return raw + "," + name
}

func init() {
	startCmd.Flags().BoolVar(&startReset, "reset", false, "Rebuild the environment from scratch")
	startCmd.Flags().StringSliceVar(&startCustomFlags, "flag", nil, "Custom project flags (e.g. --flag seed-db)")
	rootCmd.AddCommand(startCmd)
}
