package engine

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Salv4d/derrick/internal/ui"
)

func ExecuteHook(stage string, command string, useNix bool) {
	if command == "" {
		return
	}

	ui.Info(fmt.Sprintf("Executing hook: [%s]", stage))

	var cmd *exec.Cmd
	if useNix {
		nixArgs := WrapWithNix(command)
		cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
	} else {
		cmd = exec.Command("base", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		ui.FailFast(fmt.Errorf("Lifecycle hook [%s] failed with exit code: %v", stage, err))
	}

		fmt.Println(ui.SuccessInline(fmt.Sprintf("[%s] completed successfully.\n", stage)))
}