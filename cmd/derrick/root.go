package main

import (
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "derrick",
	Short: "Derrick is a local development environment orchestrator.",
	Long: `Derrick unites the absolute reproducibility of Nix with 
Docker Compose containerization, wrapping them in a strict 
state validation and hook execution system.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		ui.FailFast(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&ui.DebugMode, "debug", false, "Enable verbose debug output and stream raw command logs")
}
