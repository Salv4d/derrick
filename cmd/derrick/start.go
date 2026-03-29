package main

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the local development environment",
	Long:  `Reads the derrick.yaml configuration and begins the orchestration process, validating the state and executing the defined hooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()
		ui.Info("Starting Derrick orchestration...")

		filename := "derrick.yaml"
		cfg, err := config.ParseConfig(filename)
		if err != nil {
			ui.FailFast(err)
		}

		ui.Successf("Successfully loaded configuration for project: %s (v%s)\n", cfg.Name, cfg.Version)
		ui.Infof("Found %d Nix packages and %d validation checks to run.\n", len(cfg.Dependencies.NixPackages), len(cfg.Validations))

		useNix := len(cfg.Dependencies.NixPackages) > 0
		if useNix {
			if !engine.IsNixInstalled() {
				ui.FailFast(fmt.Errorf(
					"This project requires Nix, but it is not installed on your system.\n" +
						"To install it on Linux/WSL, run:\n" +
						"curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install",
				))
			}

			err := engine.BootEnvironment(cfg.Dependencies.NixPackages)
			if err != nil {
				ui.FailFast(err)
			}
		}
		engine.ExecuteHook("pre_init", cfg.Hooks.PreInit, useNix)

		engine.RunValidations(cfg.Validations, useNix)

		engine.ExecuteHook("post_init", cfg.Hooks.PostInit, useNix)
		engine.ExecuteHook("pre_start", cfg.Hooks.PreStart, useNix)

		if cfg.Dependencies.Dockerfile != "" {
			if !engine.IsDockerInstalled() {
				ui.FailFast(fmt.Errorf(
					"This project requires Docker, but it is not installed or not running.\n" +
						"Please install Docker Desktop or Docker Engine to continue.",
				))
			}

			err := engine.StartContainers(cfg.Dependencies.Dockerfile)
			if err != nil {
				ui.FailFast(err)
			}
		}

		engine.ExecuteHook("post_start", cfg.Hooks.PostStart, useNix)

		ui.Success("Environment is validated and ready!")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
