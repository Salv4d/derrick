package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var updateCheckOnly bool

// RunUpdate fetches the latest GitHub release and atomically replaces the
// running binary. With --check it only reports the available version.
func RunUpdate() {
	ui.Taskf("Checking for latest release on GitHub...")

	release, err := fetchLatestRelease()
	if err != nil {
		ui.Warningf("Could not reach GitHub: %v", err)
		return
	}

	latest := release.TagName
	if latest == Version || latest == "v"+Version {
		ui.Successf("You are already on the latest version (%s).", Version)
		return
	}

	if updateCheckOnly {
		ui.Warningf("A new version is available: %s (current: %s).", latest, Version)
		fmt.Println("Run 'derrick update' to install it.")
		return
	}

	asset := assetName()
	if asset == "" {
		ui.Errorf("No prebuilt binary for %s/%s.", runtime.GOOS, runtime.GOARCH)
		return
	}

	target, err := os.Executable()
	if err != nil {
		ui.Errorf("Could not locate running binary: %v", err)
		return
	}
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		target = resolved
	}

	url := fmt.Sprintf("https://github.com/Salv4d/derrick/releases/download/%s/%s", latest, asset)
	ui.Taskf("Downloading %s...", asset)

	tmp, err := downloadToSibling(url, target)
	if err != nil {
		ui.Errorf("Download failed: %v", err)
		return
	}

	if err := os.Chmod(tmp, 0o755); err != nil {
		os.Remove(tmp)
		ui.Errorf("Could not set executable bit: %v", err)
		return
	}

	if err := os.Rename(tmp, target); err != nil {
		os.Remove(tmp)
		ui.Errorf("Could not replace %s: %v", target, err)
		ui.Warning("If the install path requires elevated permissions, re-run with sudo.")
		return
	}

	ui.Successf("Updated derrick %s → %s", Version, latest)
}

func fetchLatestRelease() (*GithubRelease, error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Salv4d/derrick/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github returned status %d", resp.StatusCode)
	}
	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("no tag in release response")
	}
	return &release, nil
}

func assetName() string {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "derrick-linux-amd64"
	case "linux/arm64":
		return "derrick-linux-arm64"
	case "darwin/amd64":
		return "derrick-darwin-amd64"
	case "darwin/arm64":
		return "derrick-darwin-arm64"
	}
	return ""
}

// downloadToSibling streams url into a temp file in the same directory as
// target, so the final os.Rename is atomic (same filesystem).
func downloadToSibling(url, target string) (string, error) {
	client := http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".derrick-update-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download and install the latest derrick release",
	Run: func(cmd *cobra.Command, args []string) {
		RunUpdate()
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "only report whether an update is available")
	rootCmd.AddCommand(updateCmd)
}
