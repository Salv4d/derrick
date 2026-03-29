package engine

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Salv4d/derrick/internal/ui"
)

func ExecuteHook(stage string, commands []string, useNix bool) {
	if len(commands) == 0 {
		return
	}

	ui.Infof("Executing hook: [%s] (%d steps)", stage, len(commands))

	for i, command := range commands {
		if command == "" {
			continue
		}

		fmt.Printf("  -> Step %d/%d\n", i+1, len(commands))

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
			ui.FailFastf("Lifecycle hook [%s] failed at step %d.\nCommand: %s\nError: %v", stage, i+1, command, err)
		}
	}

	fmt.Println(ui.SuccessInline(fmt.Sprintf("[%s] completed successfully.\n", stage)))
}