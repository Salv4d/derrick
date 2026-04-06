package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// RunUpdate checks for a newer release and prints the download command.
func RunUpdate() {
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
				ui.Warningf("A new version is available: %s!", release.TagName)
				fmt.Println("\nTo flawlessly update your binary instantly, run:")
				fmt.Printf("curl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64\n")
				fmt.Printf("chmod +x derrick && sudo mv derrick /usr/local/bin/\n\n")
			} else {
				ui.Successf("You are already on the latest version (%s). No update required!", Version)
			}
		}
	} else {
		ui.Successf("No standard release tags found on remote repository.")
	}
}

// UpdateCmd is the cobra command for "derrick update".
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and elegantly retrieve the upgrade script",
	Run: func(cmd *cobra.Command, args []string) {
		RunUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
