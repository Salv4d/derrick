package cli

import (
	"fmt"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Short: "Starts the local development environment",
	Long: `Reads the derrick.yaml configuration and begins the orchestration process, validating the state and executing the defined hooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()
		ui.Info("Starting Derrick orchestration...")

		filename := "derrick.yaml"

		cfg, err := config.ParseConfig(filename)
		if err != nil {
			ui.FailFast(err)
		}

		ui.Success(fmt.Sprintf("Successfully loaded configuration for project: %s (v%s)\n", cfg.Name, cfg.Version))
		ui.Info(fmt.Sprintf("Found %d Nix packages and %d validation checks to run.\n", len(cfg.Dependencies.NixPackages), len(cfg.Validations)))

		engine.ExecuteHook("pre_init", cfg.Hooks.PreInit)

		engine.RunValidations(cfg.Validations)

		engine.ExecuteHook("post_init", cfg.Hooks.PostInit)
		engine.ExecuteHook("pre_start", cfg.Hooks.PreStart)

		ui.Info("⚙️  [Mock] Starting Nix isolated shell and Docker containers...\n")

		ui.Success("Environment is validated and ready!")

		engine.ExecuteHook("post_start", cfg.Hooks.PostStart)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}