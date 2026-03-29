package engine

import (
	"os"
	"os/exec"

	"github.com/Salv4d/derrick/internal/ui"
)

func ExecuteHook(stage string, commands []string, useNix bool) {
	if len(commands) == 0 {
		return
	}

	ui.Sectionf("Executing Lifecycle: %s", stage)

	for i, command := range commands {
		if command == "" {
			continue
		}

		ui.SubTaskf("Step %d/%d", i+1, len(commands))

		var cmd *exec.Cmd
		if useNix {
			nixArgs := WrapWithNix(command)
			ui.Debugf("Executing hook via Nix: %v", nixArgs)
			cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
		} else {
			ui.Debugf("Executing hook via Bash: bash -c %q", command)
			cmd = exec.Command("bash", "-c", command)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			ui.Error("FAILED")
			ui.FailFastf("Lifecycle hook [%s] failed at step %d.\nCommand: %s\nError: %v", stage, i+1, command, err)
		}
		ui.Success("DONE")
	}

	ui.Successf("[%s] completed successfully.", stage)
}
