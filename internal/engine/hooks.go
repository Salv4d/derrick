package engine

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Salv4d/derrick/internal/ui"
)

func ExecuteHook(stage string, command string) {
	if command == "" {
		return
	}

	ui.Info(fmt.Sprintf("Executing hook: [%s]", stage))

	cmd := exec.Command("bash", "-c", command)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		ui.FailFast(fmt.Errorf("Lifecycle hook [%s] failed with exit code: %v", stage, err))
	}

		fmt.Println(ui.SuccessInline(fmt.Sprintf("[%s] completed successfully.\n", stage)))
}