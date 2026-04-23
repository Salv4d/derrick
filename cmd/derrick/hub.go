package main

import (
	"fmt"
	"path/filepath"
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
	Use:   "add [alias] [git-url] [optional-local-path]",
	Short: "Add a project alias to the hub",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		url := args[1]
		path := ""
		if len(args) == 3 {
			path = args[2]
		}

		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		hub.Projects[alias] = config.HubProject{URL: url, Path: path}
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}

		if path != "" {
			ui.Successf("Project '%s' added to hub: %s (at %s)", alias, url, path)
		} else {
			ui.Successf("Project '%s' added to hub: %s", alias, url)
		}
	},
}

var hubListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects and remotes in the hub",
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		ui.Section("Hub Settings")
		ui.Infof("Workspace: %s", hub.Workspace)

		if len(hub.Projects) > 0 {
			ui.Section("Local Projects")
			aliases := make([]string, 0, len(hub.Projects))
			for a := range hub.Projects {
				aliases = append(aliases, a)
			}
			sort.Strings(aliases)
			for _, a := range aliases {
				proj := hub.Projects[a]
				pathInfo := ""
				if proj.Path != "" {
					pathInfo = fmt.Sprintf(" (at %s)", proj.Path)
				}
				fmt.Printf("  %-20s %s%s\n", a, proj.URL, pathInfo)
			}
		}

		if len(hub.Remotes) > 0 {
			ui.Section("Subscribed Remotes")
			for _, r := range hub.Remotes {
				fmt.Printf("  %s\n", r)
			}
		}

		if len(hub.Projects) == 0 && len(hub.Remotes) == 0 {
			ui.Info("Hub is empty.")
		}
	},
}

var hubSubscribeCmd = &cobra.Command{
	Use:   "subscribe [url]",
	Short: "Subscribe to a remote hub index (YAML URL)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		// Avoid duplicates
		for _, r := range hub.Remotes {
			if r == url {
				ui.Infof("Already subscribed to %s", url)
				return
			}
		}

		hub.Remotes = append(hub.Remotes, url)
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}
		ui.Successf("Subscribed to remote hub: %s", url)
	},
}

var hubUnsubscribeCmd = &cobra.Command{
	Use:   "unsubscribe [url]",
	Short: "Unsubscribe from a remote hub index",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		newRemotes := []string{}
		found := false
		for _, r := range hub.Remotes {
			if r == url {
				found = true
				continue
			}
			newRemotes = append(newRemotes, r)
		}

		if !found {
			ui.FailFastf("Not subscribed to %s", url)
		}

		hub.Remotes = newRemotes
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}
		ui.Successf("Unsubscribed from %s", url)
	},
}

var hubWorkspaceCmd = &cobra.Command{
	Use:   "workspace [optional-new-path]",
	Short: "View or update the global projects workspace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := config.LoadGlobalHub()
		if err != nil {
			ui.FailFast(err)
		}

		if len(args) == 0 {
			ui.Infof("Current workspace: %s", hub.Workspace)
			return
		}

		newPath, _ := filepath.Abs(args[0])
		hub.Workspace = newPath
		if err := hub.Save(); err != nil {
			ui.FailFast(err)
		}
		ui.Successf("Workspace updated to: %s", newPath)
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
	hubCmd.AddCommand(hubSubscribeCmd)
	hubCmd.AddCommand(hubUnsubscribeCmd)
	hubCmd.AddCommand(hubWorkspaceCmd)
	rootCmd.AddCommand(hubCmd)
}
