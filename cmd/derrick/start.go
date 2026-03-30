package main

import (
	"fmt"
	"os"

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

		ui.Section("Configuration")

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFastf("Failed to get current working directory: %v", err)
		}

		ui.Taskf("Parsing %s contract", configFile)
		cfg, err := config.ParseConfig(configFile)
		if err != nil {
			ui.FailFast(err)
		}
		ui.Successf("Loaded project: %s (v%s)", cfg.Name, cfg.Version)

		useNix := len(cfg.Dependencies.NixPackages) > 0

		ui.Task("Validating environment variables")
		err = engine.ValidateAndLoadEnv(cwd, cfg, useNix)
		if err != nil {
			ui.FailFast(err)
		}

		ui.Success("Environment state is valid")

		if useNix {
			ui.Section("Nix Sandbox")
			ui.Taskf("Resolving %d Nix packages", len(cfg.Dependencies.NixPackages))

			if !engine.IsNixInstalled() {
				ui.FailFast(fmt.Errorf(
					"This project requires Nix, but it is not installed on your system.\n" +
						" Fix: Run the following command to install it:\n" +
						"curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install",
				))
			}

			err = engine.BootEnvironment(configFile, cfg.Dependencies.NixPackages, cfg.Dependencies.NixRegistry)
			if err != nil {
				ui.FailFast(err)
			}
			ui.Success("Sandbox generated successfully")
		}

		ui.Section("Initialization Lifecycle")

		if len(cfg.Hooks.PreInit) > 0 {
			engine.ExecuteHook("pre_init", cfg.Hooks.PreInit, useNix)
		}

		if len(cfg.Validations) > 0 {
			engine.RunValidations(cfg.Validations, useNix)
		}

		if len(cfg.Hooks.PostInit) > 0 {
			engine.ExecuteHook("post_init", cfg.Hooks.PostInit, useNix)
		}

		if len(cfg.Hooks.PreStart) > 0 {
			engine.ExecuteHook("pre_start", cfg.Hooks.PreStart, useNix)
		}

		if cfg.Dependencies.DockerCompose != "" {
			ui.Section("Docker Orchestration")
			ui.Task("Verifying Docker daemon")

			if !engine.IsDockerInstalled() {
				ui.FailFast(fmt.Errorf(
					"This project requires Docker, but the daemon is not running or not installed.\n" +
						" Fix: Please start Docker Desktop or install the Docker Engine.",
				))
			}

			ui.Task("Starting containers")
			err := engine.StartContainers(cfg.Dependencies.DockerCompose)
			if err != nil {
				ui.FailFast(err)
			}
			ui.Success("Containers are running")
		}

		if len(cfg.Hooks.PostStart) > 0 {
			ui.Section("Post-Flight")
			ui.Task("Executing post_start hook")
			engine.ExecuteHook("post_start", cfg.Hooks.PostStart, useNix)
		}

		fmt.Println()
		ui.Successf("🚀 %s environment is strictly validated and ready!", cfg.Name)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
