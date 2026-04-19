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

	if _, err := os.Stat(flakeDir); os.IsNotExist(err) {
		return fmt.Errorf("sandbox not found at %s.\nResolution: Run 'derrick start' to initialize the environment first", flakeDir)
	}

	flakePath := fmt.Sprintf("path:%s#default", flakeDir)

	var cmd *exec.Cmd

	if len(args) > 0 {
		cmdArgs := []string{"develop", "--impure", flakePath, "--command"}
		cmdArgs = append(cmdArgs, args...)
		cmd = exec.Command(nixPath, cmdArgs...)
	} else {
		histFile := filepath.Join(flakeDir, "shell_history")
		initContent := fmt.Sprintf(
			"export PS1='\\e[34m(derrick-sandbox)\\e[0m \\w > '\n"+
				"export HISTFILE=%q\n"+
				"export HISTSIZE=10000\n"+
				"export HISTFILESIZE=10000\n"+
				"export HISTCONTROL=ignoredups:erasedups\n",
			histFile,
		)
		tmpRC, err := os.CreateTemp("", "derrick-bashrc-*")
		if err != nil {
			return fmt.Errorf("failed to create shell init file: %w", err)
		}
		defer os.Remove(tmpRC.Name())
		if _, err := tmpRC.WriteString(initContent); err != nil {
			tmpRC.Close()
			return fmt.Errorf("failed to write shell init file: %w", err)
		}
		tmpRC.Close()
		cmd = exec.Command(nixPath, "develop", "--impure", flakePath, "-c", "bash", "--init-file", tmpRC.Name())
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
			return fmt.Errorf("shell existed with status: %d", exitError.ExitCode())
		}
		return fmt.Errorf("failed to start sandbox shell: %w", err)
	}

	return nil
}
