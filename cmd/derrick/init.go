package main

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/discovery"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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

		targetConfig := filepath.Join(cwd, "derrick.yaml")
		if _, err := os.Stat(targetConfig); err == nil {
			var overwrite bool
			err := huh.NewConfirm().
				Title("A derrick.yaml file already exists. Overwrite?").
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

		var (
			projectName       = metadata.Name
			projectVersion    = metadata.Version
			useDockerCompose  bool
			dockerComposeFile string
			useEnvFile        bool
			envBaseFile       string
		)

		// Check for existing docker-compose
		possibleDockerFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
		for _, f := range possibleDockerFiles {
			if _, err := os.Stat(filepath.Join(cwd, f)); err == nil {
				dockerComposeFile = f
				useDockerCompose = true
				break
			}
		}

		// Check for existing .env
		if matches, _ := filepath.Glob(filepath.Join(cwd, ".env*")); len(matches) > 0 {
			useEnvFile = true
			envBaseFile = filepath.Base(matches[0])
		} else {
			envBaseFile = ".env"
		}

		ui.Section("Configuration Wizard")

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project Name").
					Value(&projectName),
				huh.NewInput().
					Title("Project Version").
					Value(&projectVersion),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to use Docker Compose?").
					Value(&useDockerCompose),
				huh.NewInput().
					Title("Docker Compose File Path").
					Value(&dockerComposeFile).
					Placeholder("e.g. docker-compose.yml"),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want Derrick to manage Environment Variables?").
					Value(&useEnvFile),
				huh.NewInput().
					Title("Base Env File").
					Value(&envBaseFile).
					Placeholder("e.g. .env"),
			),
		)

		err = form.Run()
		if err != nil {
			ui.FailFastf("Wizard aborted: %v", err)
		}

		if useDockerCompose {
			ui.Infof("🐳 Notice: Ensure Docker is installed and can be run without root (e.g. rootless or correct group permissions) to avoid friction.")
			if dockerComposeFile == "" {
				dockerComposeFile = "docker-compose.yml"
			}
		} else {
			dockerComposeFile = "" // ensure empty if user selected NO
		}

		if !useEnvFile {
			envBaseFile = ""
		}

		cfg := config.ProjectConfig{
			Name:    projectName,
			Version: projectVersion,
		}

		if useDockerCompose {
			cfg.Dependencies.DockerCompose = dockerComposeFile
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
