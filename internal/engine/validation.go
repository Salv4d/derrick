package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

func RunValidations(checks []config.ValidationCheck) {
	if len(checks) == 0 {
		return
	}

	ui.Info("Running state validations...")

	for _, check := range checks {
		fmt.Printf("  Checking %s... ", check.Name)

		err := executeCommand(check.Command)
		if err == nil {
			fmt.Println(ui.SuccessInline("OK"))
			continue
		}

		if check.AutoFix == "" {
			fmt.Println(ui.ErrorInline("FAILED"))
			ui.FailFast(fmt.Errorf("Validation '%s' failed.\nCommand: %s\nError: %v", check.Name, check.Command, err))
		}

		fmt.Println(ui.WarningInline("FAILED. Attempting auto-fix..."))

		fixErr := executeCommand(check.AutoFix)
		if fixErr != nil {
			ui.FailFast(fmt.Errorf("Auto-fix for '%s' failed.\nCommand: %s, Error: %v", check.Name, check.AutoFix, fixErr))
		}

		fmt.Printf("  Re-checking %s... ", check.Name)
		recheckErr := executeCommand(check.Command)
		if recheckErr != nil {
			fmt.Println(ui.ErrorInline("FAILED"))
			ui.FailFast(fmt.Errorf("Validation '%s' still failing after auto-fix.\nError: %v", check.Name, recheckErr))
		}

		fmt.Println(ui.SuccessInline("OK (Fixed)"))
	}
}

func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("%s", errMsg)
		}
		return err
	}
	return nil
}