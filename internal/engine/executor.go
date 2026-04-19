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
	"github.com/google/shlex"
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

// shellMetaRe detects characters that require bash interpretation.
// Piped commands, redirects, subshells, logical operators, and variable expansions
// all need a real shell; plain argument lists do not.
var shellMetaRe = regexp.MustCompile(`[|><;&$` + "`" + `]|\|\||&&`)

// buildCmd converts a command string into an *exec.Cmd, choosing the safest
// possible execution path:
//   - No shell metacharacters → shlex split + direct exec (no injection surface)
//   - Shell metacharacters detected  → bash -c (required for pipes, redirects, etc.)
func buildCmd(command string) (*exec.Cmd, error) {
	if shellMetaRe.MatchString(command) {
		return exec.Command("/bin/sh", "-c", command), nil
	}
	args, err := shlex.Split(command)
	if err != nil || len(args) == 0 {
		return nil, fmt.Errorf("invalid command string: %q: %w", command, err)
	}
	return exec.Command(args[0], args[1:]...), nil
}

// translateError checks a raw stderr string against known patterns and returns a
// DerrickError when a match is found, or falls back to a plain error.
// It always preserves stderr content so callers never face a blind failure.
func translateError(stderr string, original error) error {
	for _, k := range known {
		if k.pattern.MatchString(stderr) {
			return &DerrickError{Message: k.message, Fix: k.fix}
		}
	}
	if stderr != "" {
		return fmt.Errorf("%s", strings.TrimSpace(stderr))
	}
	// Wrap so callers see the exit code rather than a raw exec.ExitError.
	return fmt.Errorf("command failed: %w", original)
}

// Run executes a shell command, capturing stderr for error translation.
// stdout is streamed directly to the terminal.
func Run(command string) error {
	cmd, err := buildCmd(command)
	if err != nil {
		return err
	}
	return runCmd(cmd, false)
}

// RunInEnv executes a command with a custom environment appended to the
// current process environment.
func RunInEnv(command string, env []string) error {
	cmd, err := buildCmd(command)
	if err != nil {
		return err
	}
	cmd.Env = append(os.Environ(), env...)
	return runCmd(cmd, false)
}

// RunSilent executes a command, discarding stdout but still translating errors.
func RunSilent(command string) error {
	cmd, err := buildCmd(command)
	if err != nil {
		return err
	}
	return runCmd(cmd, true)
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
// Hook commands come from user YAML and may contain arbitrary shell syntax,
// so we always use a POSIX shell. When useNix is true the command runs inside
// the project's nix develop environment instead.
//
// extraEnv is appended to the process environment as KEY=VALUE pairs. When
// duplicate keys are present Go's exec package keeps the last value, so extras
// deliberately override matching entries in os.Environ()/NixEnv().
func executeCommand(command string, useNix bool, extraEnv []string) error {
	return executeCommandIn(command, useNix, "", extraEnv)
}

// executeCommandIn runs a command optionally wrapped with `nix develop` from a
// specific flake directory. An empty flakeDir defaults to ".derrick".
func executeCommandIn(command string, useNix bool, flakeDir string, extraEnv []string) error {
	var cmd *exec.Cmd
	if useNix {
		nixArgs := WrapWithNix(command, flakeDir)
		ui.Debugf("Executing via Nix: %v", nixArgs)
		cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
		cmd.Env = append(NixEnv(), extraEnv...)
	} else {
		ui.Debugf("Executing: /bin/sh -c %q", command)
		cmd = exec.Command("/bin/sh", "-c", command)
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	return runCmd(cmd, false)
}
