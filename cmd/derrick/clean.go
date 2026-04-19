package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	cleanAll     bool
	cleanNix     bool
	cleanDocker  bool
	cleanVolumes bool
	cleanImages  bool
	cleanConts   bool
	cleanNets    bool
	cleanDryRun  bool
)

// cleanCmd cleans up unused tools, packages, and docker assets.
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up unused tools, packages, and docker assets",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		if !cleanAll && !cleanNix && !cleanDocker && !cleanVolumes && !cleanImages && !cleanConts && !cleanNets {
			var targets []string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("What would you like to clean?").
						Description("Select components to prune. This frees disk space but removes caches.").
						Options(
							huh.NewOption("Nix (Garbage Collect old generations)", "nix").Selected(true),
							huh.NewOption("Docker System (Stopped Containers, Networks, Dangling Images)", "docker-system"),
							huh.NewOption("Docker All Images (Removes ALL unused images)", "docker-images"),
							huh.NewOption("Docker Volumes (Removes ALL unused volumes)", "docker-volumes"),
						).
						Value(&targets),
				),
			)
			if err := form.Run(); err != nil {
				ui.FailFast(err)
			}

			if len(targets) == 0 {
				ui.Warning("Nothing selected. Exiting.")
				return
			}

			for _, t := range targets {
				switch t {
				case "nix":
					cleanNix = true
				case "docker-system":
					cleanDocker = true
				case "docker-images":
					cleanImages = true
				case "docker-volumes":
					cleanVolumes = true
				}
			}
		}

		if cleanAll || cleanNix {
			ui.Section("Nix Store")
			if cleanDryRun {
				ui.Infof("[dry-run] would run: nix-collect-garbage -d")
			} else {
				ui.Task("Executing 'nix-collect-garbage -d' (Clearing all unreachable packages & old profiles)...")
				c := exec.Command("nix-collect-garbage", "-d")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					ui.Warningf("Nix GC failed: %v", err)
				} else {
					ui.Success("Nix Garbage Collection complete.")
				}
			}
		}

		if cleanAll || cleanDocker || cleanVolumes || cleanImages || cleanConts || cleanNets {
			ui.Section("Docker Engine")
			ui.Infof("Scoped to resources labeled %s", engine.DerrickManagedLabel)

			if cleanAll || cleanDocker || cleanConts {
				runDockerPrune("container")
			}
			if cleanAll || cleanNets {
				runDockerPrune("network")
			}
			if cleanAll || cleanImages {
				runDockerPrune("image", "-a")
			}
			if cleanAll || cleanVolumes {
				runDockerPrune("volume")
			}
		}
	},
}

// runDockerPrune prunes a docker resource restricted to derrick-managed
// assets via the label filter — never touches other projects' resources.
func runDockerPrune(resource string, flags ...string) {
	cmdArgs := append([]string{resource, "prune", "-f", "--filter", "label=" + engine.DerrickManagedLabel}, flags...)
	if cleanDryRun {
		ui.Infof("[dry-run] would run: docker %s", strings.Join(cmdArgs, " "))
		return
	}
	ui.Taskf("Pruning derrick-managed %ss", resource)
	c := exec.Command("docker", cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		ui.Warningf("Docker %s prune failed: %v", resource, err)
	} else {
		ui.Successf("Derrick-managed %ss pruned.", resource)
	}
}

func init() {
	cleanCmd.Flags().BoolVarP(&cleanAll, "all", "a", false, "Clean everything automatically without prompting")
	cleanCmd.Flags().BoolVar(&cleanNix, "nix", false, "Only run Nix Garbage Collection")
	cleanCmd.Flags().BoolVar(&cleanDocker, "docker", false, "Only run Docker System Prune")
	cleanCmd.Flags().BoolVar(&cleanVolumes, "volumes", false, "Include Docker Volumes in pruning")
	cleanCmd.Flags().BoolVar(&cleanImages, "images", false, "Clean ALL unused Docker Images (not just dangling)")
	cleanCmd.Flags().BoolVar(&cleanConts, "containers", false, "Clean unused Docker Containers")
	cleanCmd.Flags().BoolVar(&cleanNets, "networks", false, "Clean unused Docker Networks")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Print actions without executing them")

	rootCmd.AddCommand(cleanCmd)
}
