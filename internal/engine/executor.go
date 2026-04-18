package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Salv4d/derrick/internal/ui"
)

// DerrickError is a structured error with an actionable fix message.
type DerrickError struct {
	Message string
	Fix     string
}

func (e *DerrickError) Error() string {
	if e.Fix != "" {
		return fmt.Sprintf("%s\n\n  Fix: %s", e.Message, e.Fix)
	}
	return e.Message
}

// known translates raw stderr patterns from wrapped tools into DerrickErrors.
var known = []struct {
	pattern *regexp.Regexp
	message string
	fix     string
}{
	{
		pattern: regexp.MustCompile(`(?i)permission denied.*docker\.sock`),
		message: "Docker socket permission denied.\nYour user does not have access to the Docker daemon.",
		fix:     "sudo usermod -aG docker $USER && newgrp docker",
	},
	{
		pattern: regexp.MustCompile(`(?i)cannot connect to the docker daemon`),
		message: "Docker daemon is not running.",
		fix:     "Start Docker Desktop, or run: sudo systemctl start docker",
	},
	{
		pattern: regexp.MustCompile(`(?i)bind: address already in use|port is already allocated`),
		message: "A required port is already in use by another process.",
		fix:     "Stop the conflicting service, or adjust the ports in your docker-compose file.",
	},
	{
		pattern: regexp.MustCompile(`(?i)pull access denied|repository does not exist`),
		message: "Docker image not found or access denied.",
		fix:     "Check the image name in your docker-compose file and ensure you are logged in: docker login",
	},
	{
		pattern: regexp.MustCompile(`(?i)error: attribute '.*' missing|error: flake .* does not provide`),
		message: "Nix package not found in the registry.",
		fix:     "Check the package name at https://search.nixos.org/packages and update your derrick.yaml",
	},
	{
		pattern: regexp.MustCompile(`(?i)nix: command not found|nix is not installed`),
		message: "Nix is not installed on this system.",
		fix:     `curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install`,
	},
}

// translateError checks a raw stderr string against known patterns and returns a
// DerrickError when a match is found, or falls back to a plain error.
func translateError(stderr string, original error) error {
	for _, k := range known {
		if k.pattern.MatchString(stderr) {
			return &DerrickError{Message: k.message, Fix: k.fix}
		}
	}
	if stderr != "" {
		return fmt.Errorf("%s", strings.TrimSpace(stderr))
	}
	return original
}

// Run executes a shell command, capturing stderr for error translation.
// stdout is streamed directly to the terminal. Use RunSilent when output must
// be suppressed.
func Run(command string) error {
	return runCmd(exec.Command("bash", "-c", command), false)
}

// RunInEnv executes a command with a custom environment appended to the
// current process environment.
func RunInEnv(command string, env []string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(), env...)
	return runCmd(cmd, false)
}

// RunSilent executes a command, discarding stdout but still translating errors.
func RunSilent(command string) error {
	return runCmd(exec.Command("bash", "-c", command), true)
}

// RunCommand executes a pre-built *exec.Cmd with error translation.
func RunCommand(cmd *exec.Cmd) error {
	return runCmd(cmd, false)
}

func runCmd(cmd *exec.Cmd, silent bool) error {
	var stderr bytes.Buffer

	if silent {
		cmd.Stderr = &stderr
	} else {
		if ui.DebugMode {
			cmd.Stdout = os.Stdout
			cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = &stderr
		}
	}

	if err := cmd.Run(); err != nil {
		return translateError(stderr.String(), err)
	}
	return nil
}

// executeCommand is the internal helper for hooks and validations.
// It wraps either bash or nix-develop depending on useNix.
func executeCommand(command string, useNix bool) error {
	var cmd *exec.Cmd
	if useNix {
		nixArgs := WrapWithNix(command, "")
		ui.Debugf("Executing via Nix: %v", nixArgs)
		cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
		cmd.Env = NixEnv()
	} else {
		ui.Debugf("Executing: bash -c %q", command)
		cmd = exec.Command("bash", "-c", command)
	}
	return runCmd(cmd, false)
}
