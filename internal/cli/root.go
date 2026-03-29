package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "derrick",
	Short: "Derrick is a local development environment orchestrator.",
	Long: `Derrick unites the absolute reproducibility of Nix with 
Docker Compose containerization, wrapping them in a strict 
state validation and hook execution system.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	
}