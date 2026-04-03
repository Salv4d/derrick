package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var Version = "0.1.0-alpha"

type GithubRelease struct {
	TagName string `json:"tag_name"`
}

func RunVersion() {
	ui.PrintHeader()
	fmt.Printf("📦 Derrick CLI version \033[1;36m%s\033[0m\n\n", Version)

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
				fmt.Printf("\033[1;33mcurl -L -o derrick https://github.com/Salv4d/derrick/releases/latest/download/derrick-linux-amd64\033[0m\n")
				fmt.Printf("\033[1;33mchmod +x derrick && sudo mv derrick /usr/local/bin/\033[0m\n\n")
			} else {
				ui.Successf("Your version is exactly up to date!")
			}
		}
	} else {
		ui.Successf("No standard release tags found on remote repository.")
	}
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"update"},
	Short:   "Print the current version and gracefully check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		RunVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
