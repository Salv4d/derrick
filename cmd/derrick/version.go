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
var Version = "0.6.0-beta.4"

// GithubRelease represents a GitHub API release object.
type GithubRelease struct {
	TagName string `json:"tag_name"`
}

type versionReport struct {
	Version    string `json:"version"`
	Latest     string `json:"latest,omitempty"`
	UpToDate   bool   `json:"up_to_date"`
	CheckError string `json:"check_error,omitempty"`
}

// RunVersion prints the current version and checks for updates.
func RunVersion() {
	report := versionReport{Version: Version, UpToDate: true}

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Salv4d/derrick/releases/latest")
	if err != nil {
		report.CheckError = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			var release GithubRelease
			if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
				report.Latest = release.TagName
				report.UpToDate = release.TagName == Version || release.TagName == "v"+Version
			} else {
				report.CheckError = err.Error()
			}
		} else {
			report.CheckError = fmt.Sprintf("github returned status %d", resp.StatusCode)
		}
	}

	if jsonOutput {
		out, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(out))
		return
	}

	ui.PrintHeader()
	fmt.Printf("Derrick CLI version %s\n\n", Version)
	ui.Taskf("Checking for latest releases on GitHub...")

	if report.CheckError != "" {
		ui.Warningf("Could not check for updates: %v", report.CheckError)
		return
	}
	if report.UpToDate {
		ui.Successf("Your version is exactly up to date!")
	} else {
		ui.Warningf("A new version is available: %s! Run 'derrick update' to efficiently upgrade.", report.Latest)
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
