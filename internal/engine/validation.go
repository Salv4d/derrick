package engine

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// RunValidations runs each validation check, attempting auto-fixes when defined.
// Returns an error on the first unrecoverable failure so callers in cmd/ can
// decide how to abort; the engine package never calls os.Exit directly.
func RunValidations(checks []config.ValidationCheck, useNix bool, extraEnv []string) error {
	if len(checks) == 0 {
		return nil
	}

	ui.Section("Environment Validation")
	runner := &Runner{
		UseNix: useNix,
		Env:    extraEnv,
	}

	for _, check := range checks {
		ui.SubTask("Checking " + check.Name)

		err := runner.Run(check.Command)
		if err == nil {
			ui.Success("OK")
			continue
		}

		if check.AutoFix == "" {
			ui.Error("FAILED")
			return fmt.Errorf("validation '%s' failed\n  command: %s\n  error: %w", check.Name, check.Command, err)
		}

		ui.Warning("FAILED. Attempting auto-fix...")

		if fixErr := runner.Run(check.AutoFix); fixErr != nil {
			return fmt.Errorf("auto-fix for '%s' failed\n  command: %s\n  error: %w", check.Name, check.AutoFix, fixErr)
		}

		ui.Task("Re-checking " + check.Name)
		if recheckErr := runner.Run(check.Command); recheckErr != nil {
			ui.Error("FAILED")
			return fmt.Errorf("validation '%s' still failing after auto-fix\n  error: %w", check.Name, recheckErr)
		}

		ui.Success("OK (Fixed)")
	}
	return nil
}
