package engine

import (
	"fmt"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// HookOpts carries the context that determines which hooks should fire.
type HookOpts struct {
	// SetupCompleted is true when this is not the first `derrick start`.
	SetupCompleted bool
	// ActiveFlags maps custom flag names to whether the user passed them.
	ActiveFlags map[string]bool
	// UseNix controls whether commands run inside a Nix dev shell.
	UseNix bool
	// Env holds resolved KEY=VALUE pairs injected into each hook's process
	// environment. Prefer this over os.Setenv so execution stays isolated.
	Env []string
}

// ExecuteHooks runs the hooks for a lifecycle stage, skipping any whose when:
// condition is not satisfied by the current execution context.
func ExecuteHooks(stage string, hooks []config.Hook, opts HookOpts) error {
	if len(hooks) == 0 {
		return nil
	}

	eligible := make([]config.Hook, 0, len(hooks))
	for _, h := range hooks {
		if shouldRun(h.When, opts) {
			eligible = append(eligible, h)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	ui.Sectionf("Lifecycle: %s", stage)

	runner := &Runner{
		UseNix: opts.UseNix,
		Env:    opts.Env,
	}

	for i, hook := range eligible {
		ui.SubTaskf("Step %d/%d: %s", i+1, len(eligible), hook.Run)

		if err := runner.Run(hook.Run); err != nil {
			ui.Error("FAILED")
			return fmt.Errorf("hook [%s] step %d failed\n  command: %s\n  error: %w", stage, i+1, hook.Run, err)
		}
		ui.Success("DONE")
	}

	ui.Successf("[%s] completed", stage)
	return nil
}

// shouldRun evaluates a hook's when: condition against the current execution context.
// Returns true if ANY of the conditions in the slice match.
func shouldRun(conditions config.Condition, opts HookOpts) bool {
	if len(conditions) == 0 {
		return true
	}
	for _, when := range conditions {
		if matchOne(when, opts) {
			return true
		}
	}
	return false
}

func matchOne(when string, opts HookOpts) bool {
	switch {
	case when == "" || when == "always":
		return true
	case when == "first-setup":
		return !opts.SetupCompleted
	case strings.HasPrefix(when, "flag:"):
		flagName := strings.TrimPrefix(when, "flag:")
		return opts.ActiveFlags[flagName]
	default:
		return true
	}
}
