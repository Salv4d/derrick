package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/discovery"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd initializes a new derrick.yaml configuration.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new derrick.yaml configuration",
	Long:  `Interactively guides you through creating a new 'derrick.yaml' file for your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFastf("Failed to get current working directory: %v", err)
		}

		targetConfig := filepath.Join(cwd, configFile)
		if _, err := os.Stat(targetConfig); err == nil {
			var overwrite bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("A %s file already exists. Overwrite?", configFile)).
				Value(&overwrite).
				Run()
			if err != nil || !overwrite {
				ui.Warningf("Initialization aborted.")
				return
			}
		}

		ui.Section("Discovery")
		ui.Task("Analyzing project structure...")

		metadata := discovery.DiscoverProject(cwd)
		ui.Infof("Detected project: %s (v%s)", metadata.Name, metadata.Version)
		if metadata.Language != "" {
			ui.Infof("Detected language: %s", metadata.Language)
		}

		var (
			projectName       = metadata.Name
			projectVersion    = metadata.Version
			useDockerCompose  bool
			dockerComposeFile string
			useEnvFile        bool
			envBaseFile       string
			selectedPackages  []string
			extraPackages     string
		)

		for _, f := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
			if _, err := os.Stat(filepath.Join(cwd, f)); err == nil {
				dockerComposeFile = f
				useDockerCompose = true
				break
			}
		}

		if matches, _ := filepath.Glob(filepath.Join(cwd, ".env*")); len(matches) > 0 {
			useEnvFile = true
			envBaseFile = filepath.Base(matches[0])
		} else {
			envBaseFile = ".env"
		}

		ui.Section("Configuration Wizard")

		group1 := huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Value(&projectName),
			huh.NewInput().
				Title("Project Version").
				Value(&projectVersion),
		)

		suggestions := discovery.SuggestedPackages(metadata.Language)

		var group2 *huh.Group
		if len(suggestions) > 0 {
			options := make([]huh.Option[string], len(suggestions))
			for i, pkg := range suggestions {
				options[i] = huh.NewOption(pkg, pkg).Selected(true)
			}
			group2 = huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title(fmt.Sprintf("Nix packages — suggested for %s (toggle with x)", metadata.Language)).
					Description("Pre-selected based on your detected language. Deselect anything you don't need.").
					Options(options...).
					Value(&selectedPackages),
			)
		} else {
			group2 = huh.NewGroup(
				huh.NewInput().
					Title("Nix packages").
					Description("Comma-separated list of packages to install (e.g. nodejs, go, python3)").
					Placeholder("nodejs, go, python3").
					Value(&extraPackages),
			)
		}

		group3 := huh.NewGroup(
			huh.NewInput().
				Title("Additional Nix packages").
				Description("Comma-separated extras to add on top — leave blank to skip").
				Placeholder("e.g. git, curl, jq").
				Value(&extraPackages),
		)

		group4 := huh.NewGroup(
			huh.NewConfirm().
				Title("Do you want to use Docker Compose?").
				Value(&useDockerCompose),
			huh.NewInput().
				Title("Docker Compose File Path").
				Value(&dockerComposeFile).
				Placeholder("e.g. docker-compose.yml"),
		)

		group5 := huh.NewGroup(
			huh.NewConfirm().
				Title("Do you want Derrick to manage Environment Variables?").
				Value(&useEnvFile),
			huh.NewInput().
				Title("Base Env File").
				Value(&envBaseFile).
				Placeholder("e.g. .env"),
		)

		var form *huh.Form
		if len(suggestions) > 0 {
			form = huh.NewForm(group1, group2, group3, group4, group5)
		} else {
			form = huh.NewForm(group1, group2, group4, group5)
		}

		err = form.Run()
		if err != nil {
			ui.FailFastf("Wizard aborted: %v", err)
		}

		pkgSet := make(map[string]struct{})
		for _, p := range selectedPackages {
			pkgSet[p] = struct{}{}
		}
		for _, p := range strings.Split(extraPackages, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				pkgSet[p] = struct{}{}
			}
		}

		var nixPackages []config.NixPackage
		for p := range pkgSet {
			nixPackages = append(nixPackages, config.NixPackage{Name: p})
		}

		if useDockerCompose {
			ui.Infof("Ensure Docker can run without root (rootless or correct group membership).")
			if dockerComposeFile == "" {
				dockerComposeFile = "docker-compose.yml"
			}
		} else {
			dockerComposeFile = ""
		}

		if !useEnvFile {
			envBaseFile = ""
		}

		cfg := config.ProjectConfig{
			Name:    projectName,
			Version: projectVersion,
		}

		cfg.Nix.Packages = nixPackages

		if useDockerCompose {
			cfg.Docker.Compose = dockerComposeFile
		}

		if useEnvFile {
			cfg.EnvManagement = config.EnvManagement{
				BaseFile:      envBaseFile,
				PromptMissing: true,
			}
		}

		yamlData, err := yaml.Marshal(&cfg)
		if err != nil {
			ui.FailFastf("Failed to generate YAML: %v", err)
		}

		err = os.WriteFile(targetConfig, yamlData, 0644)
		if err != nil {
			ui.FailFastf("Failed to write %s: %v", targetConfig, err)
		}

		ui.Successf("Successfully created %s!", filepath.Base(targetConfig))
		ui.Infof("You can now run 'derrick start' to boot your local environment.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
