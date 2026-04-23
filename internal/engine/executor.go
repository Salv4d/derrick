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

// Runner encapsulates the context for executing commands.
type Runner struct {
	WorkDir string
	UseNix  bool
	NixDir  string
	Env     []string
	Silent  bool
}

// NewRunner creates a new runner with default settings.
func NewRunner() *Runner {
	return &Runner{}
}

// Run executes a shell command within the runner's context.
func (r *Runner) Run(command string) error {
	var cmd *exec.Cmd
	if r.UseNix {
		nixArgs := WrapWithNix(command, r.NixDir)
		ui.Debugf("Executing via Nix: %v", nixArgs)
		cmd = exec.Command(nixArgs[0], nixArgs[1:]...)
		cmd.Env = append(NixEnv(), r.Env...)
	} else {
		// Shell execution for arbitrary strings
		if shellMetaRe.MatchString(command) {
			ui.Debugf("Executing: /bin/sh -c %q", command)
			cmd = exec.Command("/bin/sh", "-c", command)
		} else {
			args, err := shlex.Split(command)
			if err != nil || len(args) == 0 {
				return fmt.Errorf("invalid command string: %q: %w", command, err)
			}
			ui.Debugf("Executing: %s", strings.Join(args, " "))
			cmd = exec.Command(args[0], args[1:]...)
		}
		cmd.Env = append(os.Environ(), r.Env...)
	}

	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}

	return runCmd(cmd, r.Silent)
}

// RunCommand executes a pre-built *exec.Cmd within the runner's context.
func (r *Runner) RunCommand(cmd *exec.Cmd) error {
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	if len(r.Env) > 0 {
		cmd.Env = append(cmd.Env, r.Env...)
	}
	return runCmd(cmd, r.Silent)
}

// Global execution helpers (legacy/simple cases)

func Run(command string) error {
	return NewRunner().Run(command)
}

func RunSilent(command string) error {
	r := NewRunner()
	r.Silent = true
	return r.Run(command)
}

func RunInEnv(command string, env []string) error {
	r := NewRunner()
	r.Env = env
	return r.Run(command)
}

// internal helpers

var shellMetaRe = regexp.MustCompile(`[|><;&$` + "`" + `]|\|\||&&`)

func runCmd(cmd *exec.Cmd, silent bool) error {
	var stderr bytes.Buffer

	switch {
	case silent:
		cmd.Stderr = &stderr
	case ui.Quiet:
		cmd.Stdout = io.Discard
		cmd.Stderr = &stderr
	case ui.DebugMode:
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	default:
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return translateError(stderr.String(), err)
	}
	return nil
}

func translateError(stderr string, original error) error {
	for _, k := range known {
		if k.pattern.MatchString(stderr) {
			return &DerrickError{Message: k.message, Fix: k.fix}
		}
	}
	if stderr != "" {
		return fmt.Errorf("%s", strings.TrimSpace(stderr))
	}
	return fmt.Errorf("command failed: %w", original)
}

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
