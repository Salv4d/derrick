package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
)

func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func StartContainers(composeFile string) error {
	if composeFile == "" {
		return nil
	}

	ui.Info(fmt.Sprintf("Starting Docker containers from [%s]...", composeFile))

	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("Docker Compose failed: %s", errMsg)
		}
		return fmt.Errorf("Docker Compose failed with error: %v", err)
	}

	fmt.Println(ui.SuccessInline("Containers are up and running"))
	return nil
}

func StopContainers(composeFile string) error {
	if composeFile == "" {
		return nil
	}

	ui.Info(fmt.Sprintf("Stopping Docker containers from [%s]...", composeFile))

	cmd := exec.Command("docker", "compose", "-f", composeFile, "down")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("Docker Compose teardown failed: %s", errMsg)
		}
		return fmt.Errorf("Docker Compose teardown failed: %v", err)
	}

	fmt.Println(ui.SuccessInline("Containers stopped and removed"))
	return nil
}