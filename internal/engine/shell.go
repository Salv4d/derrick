package engine

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
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
	if !IsNixInstalled() {
		return fmt.Errorf("nix is not installed or not in PATH.\nResolution: Install Nix via 'curl -L https://nixos.org/nix/install | sh'")
	}

	flakeNix := filepath.Join(flakeDir, "flake.nix")
	if _, err := os.Stat(flakeNix); os.IsNotExist(err) {
		return fmt.Errorf("sandbox not found at %s.\nResolution: Run 'derrick start' to initialize the environment first", flakeDir)
	}

	if len(args) > 0 {
		runner := &Runner{
			UseNix: true,
			NixDir: flakeDir,
		}
		// Run args as a single command string for Nix -c compatibility
		return runner.Run(strings.Join(args, " "))
	}

	// Interactive shell: manual exec to preserve TTY and signal handling
	absFlakeDir, _ := filepath.Abs(flakeDir)
	flakePath := fmt.Sprintf("path:%s#default", absFlakeDir)
	cmd := exec.Command("nix", "develop", "--impure", flakePath)
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
