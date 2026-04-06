package main

import (
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

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

func init() {
	rootCmd.PersistentFlags().BoolVar(&ui.DebugMode, "debug", false, "Enable verbose debug output and stream raw command logs")
	rootCmd.PersistentFlags().StringVarP(&configFile, "file", "f", "derrick.yaml", "Custom configuration file path")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "", "Derrick profile to execute")
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number and seamlessly check for updates")
}
