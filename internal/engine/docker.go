package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
	"gopkg.in/yaml.v3"
)

// ComposeMap represents the structure of a docker-compose.yml for parsing.
type ComposeMap struct {
	Services map[string]interface{} `yaml:"services"`
}

// OverrideMap represents the docker-compose.override.yml structure.
type OverrideMap struct {
	Services map[string]ServiceOverride `yaml:"services"`
}

// ServiceOverride defines service-specific overrides for docker-compose.
type ServiceOverride struct {
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
}

// GenerateNetworkOverride creates a docker-compose.override.yml that injects
// host.docker.internal into every service's extra_hosts. The project's default
// network is left to Docker Compose so it remains scoped to this project and
// never conflicts with other Derrick projects.
func GenerateNetworkOverride(composeFile string, outDir string) (string, error) {
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return "", err
	}

	var base ComposeMap
	if err := yaml.Unmarshal(data, &base); err != nil {
		return "", fmt.Errorf("failed to parse %s: %v", composeFile, err)
	}

	override := OverrideMap{
		Services: make(map[string]ServiceOverride),
	}

	for svcName := range base.Services {
		override.Services[svcName] = ServiceOverride{
			ExtraHosts: []string{"host.docker.internal:host-gateway"},
		}
	}

	overrideData, err := yaml.Marshal(&override)
	if err != nil {
		return "", err
	}

	if outDir == "" {
		outDir = ".derrick"
	}

	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		return "", err
	}

	overridePath := filepath.Join(outDir, "docker-compose.override.yml")
	err = os.WriteFile(overridePath, overrideData, 0o644)
	if err != nil {
		return "", err
	}

	return overridePath, nil
}

// IsDockerInstalled checks if the docker binary is available in PATH.
func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// StartContainers brings up the docker-compose project.
func StartContainers(composeFile string, profiles []string) error {
	if composeFile == "" {
		return nil
	}

	ui.Taskf("Starting Docker containers from [%s]", composeFile)

	overridePath, err := GenerateNetworkOverride(composeFile, ".derrick")
	if err != nil {
		ui.Warningf("Failed to generate network overrides (clustering disabled): %v", err)
	}

	args := []string{"compose", "-f", composeFile}
	if err == nil && overridePath != "" {
		args = append(args, "-f", overridePath)
	}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d")

	cmd := exec.Command("docker", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
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

// StopContainers stops the docker-compose project.
func StopContainers(composeFile string, profiles []string) error {
	if composeFile == "" {
		return nil
	}

	ui.Taskf("Stopping Docker containers from [%s]", composeFile)

	overridePath, err := GenerateNetworkOverride(composeFile, ".derrick")

	args := []string{"compose", "-f", composeFile}
	if err == nil && overridePath != "" {
		args = append(args, "-f", overridePath)
	}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down")

	cmd := exec.Command("docker", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
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
