package main

import (
	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the local development environment",
	Long: `Stops all running containers, closes Nix shells, and executes 
any defined post_stop lifecycle hooks to clean up the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()
		ui.Info("Stopping Derrick orchestration...")

		filename := "derrick.yaml"

		cfg, err := config.ParseConfig(filename)
		if err != nil {
			ui.FailFast(err)
		}

		if cfg.Dependencies.Dockerfile != "" {
			if !engine.IsDockerInstalled() {
				ui.FailFastf(
					"Docker is not running. Cannot stop containers.",
				)
			}

			err := engine.StartContainers(cfg.Dependencies.Dockerfile)
			if err != nil {
				ui.FailFast(err)
			}
		}

		useNix := len(cfg.Dependencies.NixPackages) > 0

		engine.ExecuteHook("post_stop", cfg.Hooks.PostStop, useNix)

		ui.Success("Environment successfully stopped!")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
