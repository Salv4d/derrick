package engine

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// ShellEngine manages interactive shell sessions inside the Nix sandbox.
type ShellEngine struct{}

// NewShellEngine creates a new shell engine instance.
func NewShellEngine() *ShellEngine {
	return &ShellEngine{}
}

// EnterSandbox starts an interactive shell or runs a command inside the Nix sandbox.
func (e *ShellEngine) EnterSandbox(flakeDir string, args []string) error {
	nixPath, err := exec.LookPath("nix")
	if err != nil {
		return fmt.Errorf("nix is not installed or not in PATH.\nResolution: Install Nix via 'curl -L https://nixos.org/nix/install | sh'")
	}

	if _, err := os.Stat(filepath.Join(flakeDir, "flake.nix")); os.IsNotExist(err) {
		return fmt.Errorf("sandbox not found at %s.\nResolution: Run 'derrick start' to initialize the environment first", flakeDir)
	}

	flakePath := fmt.Sprintf("path:%s#default", flakeDir)

	var cmd *exec.Cmd

	if len(args) > 0 {
		cmdArgs := []string{"develop", "--impure", flakePath, "--command"}
		cmdArgs = append(cmdArgs, args...)
		cmd = exec.Command(nixPath, cmdArgs...)
	} else {
		cmd = exec.Command(nixPath, "develop", "--impure", flakePath)
	}

	cmd.Env = NixEnv()

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	go func() {
		for sig := range sigChan {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("shell exited with status: %d", exitError.ExitCode())
		}
		return fmt.Errorf("failed to start sandbox shell: %w", err)
	}

	return nil
}
