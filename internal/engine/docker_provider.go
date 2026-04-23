package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

// Provision writes .derrick/docker-compose.override.yml so every service has
// the com.derrick.managed label and a host.docker.internal host entry. When
// docker.networks is declared, it also ensures those external networks exist
// and attaches every service to them. No containers are booted here.
func (d *DockerProvider) Provision(cfg *config.ProjectConfig) error {
	if cfg.Docker.Compose == "" {
		return fmt.Errorf("no docker.compose file specified in derrick.yaml")
	}
	if len(cfg.Docker.Networks) > 0 {
		if err := EnsureNetworks(cfg.Docker.Networks); err != nil {
			return err
		}
	}
	if _, err := GenerateNetworkOverride(cfg.Docker.Compose, ".derrick", cfg.Docker.Networks); err != nil {
		return fmt.Errorf("failed to generate docker network overlay: %w", err)
	}
	return nil
}

// Start brings up the Docker Compose project using the override Provision
// produced. The override path is rebuilt here to avoid storing state on the
// receiver; Provision is idempotent so this is cheap.
func (d *DockerProvider) Start(cfg *config.ProjectConfig, flags Flags) error {
	if cfg.Docker.Compose == "" {
		return fmt.Errorf("no docker.compose file specified in derrick.yaml")
	}

	overridePath := filepath.Join(".derrick", "docker-compose.override.yml")

	args := []string{"compose", "-p", cfg.Name, "-f", cfg.Docker.Compose, "-f", overridePath}
	for _, p := range cfg.Docker.Profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d", "--remove-orphans")

	ui.Taskf("Starting containers from [%s]", cfg.Docker.Compose)
	cmd := exec.Command("docker", args...)
	// Inject resolved env so compose interpolation (${VAR}) resolves without
	// polluting the parent process with os.Setenv.
	cmd.Env = append(os.Environ(), flags.Env...)
	return RunCommand(cmd)
}

// Stop tears down the Docker Compose project.
func (d *DockerProvider) Stop(cfg *config.ProjectConfig) error {
	if cfg.Docker.Compose == "" {
		return nil
	}

	overridePath := filepath.Join(".derrick", "docker-compose.override.yml")

	args := []string{"compose", "-p", cfg.Name, "-f", cfg.Docker.Compose}
	if _, err := os.Stat(overridePath); err == nil {
		args = append(args, "-f", overridePath)
	}

	for _, p := range cfg.Docker.Profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down")

	ui.Task("Stopping containers")
	cmd := exec.Command("docker", args...)
	return RunCommand(cmd)
}

// Shell opens a shell inside the target service of the Compose project,
// or runs cmdArgs as a single command when cmdArgs is non-empty.
// The service is resolved in priority order: docker.shell in derrick.yaml,
// then the first service declared in the compose file.
func (d *DockerProvider) Shell(cfg *config.ProjectConfig, cmdArgs []string) error {
	if cfg.Docker.Compose == "" {
		return fmt.Errorf("no docker.compose file specified in derrick.yaml")
	}

	service := cfg.Docker.Shell
	if service == "" {
		var err error
		service, err = FirstService(cfg.Docker.Compose)
		if err != nil {
			return fmt.Errorf("could not determine shell target service: %w", err)
		}
	}

	args := []string{"compose", "-p", cfg.Name, "-f", cfg.Docker.Compose, "exec"}
	for _, p := range cfg.Docker.Profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, service)
	if len(cmdArgs) > 0 {
		args = append(args, cmdArgs...)
	} else {
		args = append(args, "sh", "-c", "bash || sh")
	}

	cmd := exec.Command("docker", args...)
	return RunCommand(cmd)
}

// Status inspects running containers for the Compose project.
func (d *DockerProvider) Status(cfg *config.ProjectConfig) (EnvironmentStatus, error) {
	if cfg.Docker.Compose == "" {
		return EnvironmentStatus{}, nil
	}

	args := []string{"compose", "-p", cfg.Name, "-f", cfg.Docker.Compose, "ps", "--services", "--filter", "status=running"}
	cmd := exec.Command("docker", args...)

	out, err := cmd.Output()
	if err != nil {
		return EnvironmentStatus{Running: false, Details: "compose project not running"}, nil
	}

	running := len(out) > 0
	return EnvironmentStatus{Running: running, Details: string(out)}, nil
}
