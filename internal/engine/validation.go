package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

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

func executeCommand(command string, useNix bool) error {
	var cmd *exec.Cmd
	if useNix {
		nixArgs := WrapWithNix(command, "")
		ui.Debugf("Executing Nix command: %v", nixArgs)
		cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
	} else {
		ui.Debugf("Executing Bash command: bash -c %q", command)
		cmd = exec.Command("bash", "-c", command)
	}

	var stderr bytes.Buffer

	if useNix {
		cmd.Env = NixEnv()
	}

	if ui.DebugMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

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
