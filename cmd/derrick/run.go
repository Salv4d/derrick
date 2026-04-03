package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	saveEnv bool
	rmFiles bool
)

var runCmd = &cobra.Command{
	Use:   "run [packages...]",
	Short: "Creates an ephemeral environment loaded with specific Nix packages",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFast(err)
		}

		var workDir string
		var isolateDir bool
		var flakeOutDir string

		// Logic for directory handling based on flags
		if rmFiles {
			workDir, err = os.MkdirTemp("", "derrick-run-*")
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to create temp directory for sandbox: %v", err))
			}
			flakeOutDir = filepath.Join(workDir, ".derrick")
			isolateDir = true
			ui.Infof("Running in ephemeral isolated directory (will be deleted on exit): %s", workDir)

			defer func() {
				ui.Infof("Cleaning up ephemeral files: %s", workDir)
				os.RemoveAll(workDir)
			}()

			os.Chdir(workDir)
			defer os.Chdir(cwd)

		} else if saveEnv {
			envName := fmt.Sprintf("derrick-env-%d", time.Now().Unix())
			workDir = filepath.Join(cwd, envName)
			err = os.MkdirAll(workDir, 0755)
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to create save directory: %v", err))
			}
			flakeOutDir = filepath.Join(workDir, ".derrick")
			ui.Infof("Saving environment to isolated directory: %s", workDir)
			
			// We move there to execute
			os.Chdir(workDir)
			defer os.Chdir(cwd)

		} else {
			// standard ephemeral run in same directory (no files wiped, but .derrick flake wiped)
			workDir = cwd
			flakeOutDir = filepath.Join(workDir, fmt.Sprintf(".derrick-tmp-%d", time.Now().UnixNano()))
			ui.Infof("Running in local directory, will discard Nix configuration on exit.")
			defer func() {
				os.RemoveAll(flakeOutDir)
			}()
		}

		ui.Taskf("Resolving packages: %v", args)

		var nixPkgs []config.NixPackage
		for _, arg := range args {
			nixPkgs = append(nixPkgs, config.NixPackage{Name: arg})
		}

		// Initialize Environment
		err = engine.BootEnvironment("", nixPkgs, config.DefaultNixRegistry, flakeOutDir)
		if err != nil {
			ui.FailFast(fmt.Errorf("failed to initialize run environment: %v", err))
		}

		// Handle save state yaml generation
		if saveEnv {
			yamlPath := filepath.Join(workDir, "derrick.yaml")
			var cfg config.ProjectConfig
			cfg.Name = filepath.Base(workDir)
			cfg.Version = "0.1.0"
			cfg.Dependencies.NixPackages = nixPkgs
			cfg.Dependencies.NixRegistry = config.DefaultNixRegistry

			data, _ := yaml.Marshal(&cfg)
			os.WriteFile(yamlPath, data, 0644)
			ui.Successf("Successfully wrote derrick.yaml to %s", yamlPath)
		}

		ui.Success("Environment loaded. Entering shell...")
		// Provide an indication this is temporary if we isolate
		if isolateDir {
			ui.Warning("NOTE: Any files created here will be permanently deleted on exit (--rm-files is active).")
		}

		eng := engine.NewShellEngine()
		if err := eng.EnterSandbox(flakeOutDir, nil); err != nil {
			// Can fail with exit code if they just exit naturally with error code
			ui.Warningf("Application execution ended: %v", err)
		} else {
			ui.Success("Sandbox session closed normally.")
		}
	},
}

func init() {
	runCmd.Flags().BoolVar(&saveEnv, "save", false, "Creates a persisted derrick directory + derrick.yaml for these packages")
	runCmd.Flags().BoolVar(&rmFiles, "rm-files", false, "Executes the sandbox in an isolated OS temp directory, wiping any created files on exit")
	rootCmd.AddCommand(runCmd)
}
