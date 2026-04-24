package main

import (
	"os"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/state"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// DerrickContext carries initialized project state for subcommands.
type DerrickContext struct {
	Config *config.ProjectConfig
	State  *state.EnvironmentState
	Cwd    string
}

// RunDerrick is a middleware for commands that need a loaded project.
// It handles config parsing, path resolution, and logging setup.
func RunDerrick(fn func(ctx *DerrickContext, cmd *cobra.Command, args []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFastf("Failed to get working directory: %v", err)
		}

		_ = ui.SetLogFile(cwd)

		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			// Special case: 'start <alias>' should work even without a local derrick.yaml
			if cmd.Name() == "start" && len(args) == 1 {
				ctx := &DerrickContext{
					Config: &config.ProjectConfig{},
					State:  &state.EnvironmentState{},
					Cwd:    cwd,
				}
				fn(ctx, cmd, args)
				return
			}
			ui.FailFast(err)
		}

		projectState, _ := state.Load(cwd)

		ctx := &DerrickContext{
			Config: cfg,
			State:  projectState,
			Cwd:    cwd,
		}

		fn(ctx, cmd, args)
	}
}

// Root command for the Derrick CLI.
var rootCmd = &cobra.Command{
	Use:   "derrick",
	Short: "Derrick is a local development environment orchestrator.",
	Long: `Derrick unites the absolute reproducibility of Nix with
Docker Compose containerization, wrapping them in a strict
state validation and hook execution system.`,
	Run: func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetBool("version")
		if v {
			RunVersion()
			return
		}
		cmd.Help()
	},
}

// Execute runs the root command and handles errors.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		ui.FailFast(err)
	}
}

var configFile string
var profileName string
var jsonOutput bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&ui.DebugMode, "debug", false, "Enable verbose debug output and stream raw command logs")
	rootCmd.PersistentFlags().StringVarP(&configFile, "file", "f", "derrick.yaml", "Custom configuration file path")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "", "Derrick profile to execute")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON instead of decorated output (status, doctor, version)")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if jsonOutput {
			ui.Quiet = true
		}
	}
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number and seamlessly check for updates")
}
