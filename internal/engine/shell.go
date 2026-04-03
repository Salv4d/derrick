package engine

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type ShellEngine struct{}

func NewShellEngine() *ShellEngine {
	return &ShellEngine{}
}

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
		cmdArgs := []string{"develop", flakePath, "--command"}
		cmdArgs = append(cmdArgs, args...)
		cmd = exec.Command(nixPath, cmdArgs...)
	} else {
		customPrompt := "export PS1='\\e[34m(derrick-sandbox)\\e[0m \\w > '; bash --norc"
		cmd = exec.Command(nixPath, "develop", flakePath, "-c", "sh", "-c", customPrompt)
	}

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
