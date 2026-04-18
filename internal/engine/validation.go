package engine

import (
	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// RunValidations runs each validation check, attempting auto-fixes when defined.
func RunValidations(checks []config.ValidationCheck, useNix bool) {
	if len(checks) == 0 {
		return
	}

	ui.Section("Environment Validation")

	for _, check := range checks {
		ui.SubTask("Checking " + check.Name)

		err := executeCommand(check.Command, useNix)
		if err == nil {
			ui.Success("OK")
			continue
		}

		if check.AutoFix == "" {
			ui.Error("FAILED")
			ui.FailFastf("Validation '%s' failed.\nCommand: %s\nError: %v", check.Name, check.Command, err)
		}

		ui.Warning("FAILED. Attempting auto-fix...")

		fixErr := executeCommand(check.AutoFix, useNix)
		if fixErr != nil {
			ui.FailFastf("Auto-fix for '%s' failed.\nCommand: %s, Error: %v", check.Name, check.AutoFix, fixErr)
		}

		ui.Task("Re-checking " + check.Name)
		recheckErr := executeCommand(check.Command, useNix)
		if recheckErr != nil {
			ui.Error("FAILED")
			ui.FailFastf("Validation '%s' still failing after auto-fix.\nError: %v", check.Name, recheckErr)
		}

		ui.Success("OK (Fixed)")
	}
}
