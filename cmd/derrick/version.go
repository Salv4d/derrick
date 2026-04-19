// Derrick CLI — local development orchestrator.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// Version is the current version of Derrick CLI.
var Version = "0.1.0-alpha"

// GithubRelease represents a GitHub API release object.
type GithubRelease struct {
	TagName string `json:"tag_name"`
}

// RunVersion prints the current version and checks for updates.
func RunVersion() {
	ui.PrintHeader()
	fmt.Printf("Derrick CLI version %s\n\n", Version)

	ui.Taskf("Checking for latest releases on GitHub...")

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/Salv4d/derrick/releases/latest")
	if err != nil {
		ui.Warningf("Could not reach GitHub to check for updates: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var release GithubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
			if release.TagName != Version && release.TagName != "v"+Version {
				ui.Warningf("A new version is available: %s! Run 'derrick update' to efficiently upgrade.", release.TagName)
			} else {
				ui.Successf("Your version is exactly up to date!")
			}
		}
	} else {
		ui.Successf("No standard release tags found on remote repository.")
	}
}

// versionCmd prints the current version and gracefully checks for updates.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version and gracefully check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		RunVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
