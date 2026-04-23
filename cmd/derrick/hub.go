package main

import (
	"fmt"
	"sort"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Manage the local project hub",
	Long: `The Derrick Hub stores project aliases and their Git URLs in ~/.derrick/config.yaml.
This allows you to quickly clone and start projects from anywhere using 'derrick start <alias>'.`,
}

var hubAddCmd = &cobra.Command{
	Use:   "add [alias] [git-url]",
	Short: "Add a project alias to the hub",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		url := args[1]

		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		hub.Projects[alias] = url
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}

		ui.Successf("Project '%s' added to hub: %s", alias, url)
	},
}

var hubListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects in the hub",
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		if len(hub.Projects) == 0 {
			ui.Info("Hub is empty.")
			return
		}

		ui.Section("Derrick Hub Projects")
		
		// Sort aliases for consistent output
		aliases := make([]string, 0, len(hub.Projects))
		for a := range hub.Projects {
			aliases = append(aliases, a)
		}
		sort.Strings(aliases)

		for _, a := range aliases {
			fmt.Printf("  %-20s %s\n", a, hub.Projects[a])
		}
	},
}

var hubRemoveCmd = &cobra.Command{
	Use:   "remove [alias]",
	Short: "Remove a project alias from the hub",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		if _, exists := hub.Projects[alias]; !exists {
			ui.FailFastf("Alias '%s' not found in hub.", alias)
		}

		delete(hub.Projects, alias)
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}

		ui.Successf("Project '%s' removed from hub.", alias)
	},
}

func init() {
	hubCmd.AddCommand(hubAddCmd)
	hubCmd.AddCommand(hubListCmd)
	hubCmd.AddCommand(hubRemoveCmd)
	rootCmd.AddCommand(hubCmd)
}
