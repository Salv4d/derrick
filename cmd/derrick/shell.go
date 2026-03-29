package main

import (
	"os"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Drop into the isoled Nix development sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		ui.Infof("Opening sandbox at %s", cwd)

		eng := engine.NewShellEngine()
		if err := eng.EnterSandbox(cwd); err != nil {
			ui.FailFast(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
