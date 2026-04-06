package main

import (
	"os"
	"path/filepath"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// shellCmd drops into the isolated Nix development sandbox or executes a command inside it.
var shellCmd = &cobra.Command{
	Use:   "shell [command...]",
	Short: "Drop into the isolated Nix development sandbox or execute a command inside it",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()

		if len(args) > 0 {
			ui.Infof("Executing command in sandbox at %s", cwd)
		} else {
			ui.Infof("Opening sandbox at %s", cwd)
		}

		eng := engine.NewShellEngine()
		derrickDir := filepath.Join(cwd, ".derrick")
		if err := eng.EnterSandbox(derrickDir, args); err != nil {
			ui.FailFast(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
