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

func EnsureGlobalNetwork() {
	cmd := exec.Command("docker", "network", "create", "derrick-net")
	_ = cmd.Run() // Ignore errors; network might already exist
}

func StartContainers(composeFile string, profiles []string) error {
	if composeFile == "" {
		return nil
	}

	EnsureGlobalNetwork()

	ui.Taskf("Starting Docker containers from [%s]", composeFile)

	args := []string{"compose", "-f", composeFile}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d")

	cmd := exec.Command("docker", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		ui.Error("FAILED")
		errMsg := strings.TrimSpace(stderr.String())

		if strings.Contains(errMsg, "permission denied") && strings.Contains(errMsg, "docker.sock") {
			return fmt.Errorf(
				"Docker permission denied.\n" +
					"Your current user does not have access to the Docker daemon.\n" +
					"To fix this on Linux/WSL, run the following commands:\n\n" +
					"  sudo usermod -aG docker $USER\n" +
					"  newgrp docker",
			)
		}

		if errMsg != "" {
			return fmt.Errorf("Docker Compose failed: %s", errMsg)
		}
		return fmt.Errorf("Docker Compose failed with error: %v", err)
	}

	ui.Success("DONE")
	return nil
}

func StopContainers(composeFile string, profiles []string) error {
	if composeFile == "" {
		return nil
	}

	ui.Taskf("Stopping Docker containers from [%s]", composeFile)

	args := []string{"compose", "-f", composeFile}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down")

	cmd := exec.Command("docker", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		ui.Error("FAILED")
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("Docker Compose teardown failed: %s", errMsg)
		}
		return fmt.Errorf("Docker Compose teardown failed: %v", err)
	}

	ui.Success("DONE")
	return nil
}
