package engine

import (
	"fmt"
	"os/exec"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
)

// DockerProvider implements Provider using Docker Compose.
type DockerProvider struct{}

func (d *DockerProvider) Name() string { return "docker" }

// IsAvailable checks that the docker binary is reachable.
func (d *DockerProvider) IsAvailable() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return &DerrickError{
			Message: "Docker is not installed or not in PATH.",
			Fix:     "Install Docker Desktop from https://www.docker.com/products/docker-desktop or the Docker Engine for Linux.",
		}
	}
	// Probe the daemon with a lightweight command.
	if err := RunSilent("docker info"); err != nil {
		return &DerrickError{
			Message: "Docker daemon is not running.",
			Fix:     "Start Docker Desktop, or run: sudo systemctl start docker",
		}
	}
	return nil
}

// Start brings up the Docker Compose project, creating the shared network first.
func (d *DockerProvider) Start(cfg *config.ProjectConfig, _ Flags) error {
	if cfg.Docker.Compose == "" {
		return fmt.Errorf("no docker.compose file specified in derrick.yaml")
	}

	overridePath, overrideErr := GenerateNetworkOverride(cfg.Docker.Compose, ".derrick")

	args := []string{"compose", "-f", cfg.Docker.Compose}
	if overrideErr == nil && overridePath != "" {
		args = append(args, "-f", overridePath)
	} else {
		ui.Warningf("Network overlay skipped: %v", overrideErr)
	}
	for _, p := range cfg.Docker.Profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d", "--remove-orphans")

	ui.Taskf("Starting containers from [%s]", cfg.Docker.Compose)
	cmd := exec.Command("docker", args...)
	return RunCommand(cmd)
}

// Stop tears down the Docker Compose project.
func (d *DockerProvider) Stop(cfg *config.ProjectConfig) error {
	if cfg.Docker.Compose == "" {
		return nil
	}

	args := []string{"compose", "-f", cfg.Docker.Compose}
	for _, p := range cfg.Docker.Profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down")

	ui.Task("Stopping containers")
	cmd := exec.Command("docker", args...)
	return RunCommand(cmd)
}

// Shell opens a shell inside the first running service of the Compose project.
func (d *DockerProvider) Shell(cfg *config.ProjectConfig) error {
	if cfg.Docker.Compose == "" {
		return fmt.Errorf("no docker.compose file specified in derrick.yaml")
	}

	args := []string{"compose", "-f", cfg.Docker.Compose, "exec"}
	if len(cfg.Docker.Profiles) > 0 {
		args = append(args, "--profile", cfg.Docker.Profiles[0])
	}
	// Open a shell in the first service by convention — users can override via flags.
	args = append(args, "app", "sh", "-c", "bash || sh")

	cmd := exec.Command("docker", args...)
	return RunCommand(cmd)
}

// Status inspects running containers for the Compose project.
func (d *DockerProvider) Status(cfg *config.ProjectConfig) (EnvironmentStatus, error) {
	if cfg.Docker.Compose == "" {
		return EnvironmentStatus{}, nil
	}

	args := []string{"compose", "-f", cfg.Docker.Compose, "ps", "--services", "--filter", "status=running"}
	cmd := exec.Command("docker", args...)

	out, err := cmd.Output()
	if err != nil {
		return EnvironmentStatus{Running: false, Details: "compose project not running"}, nil
	}

	running := len(out) > 0
	return EnvironmentStatus{Running: running, Details: string(out)}, nil
}
