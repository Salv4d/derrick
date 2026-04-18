package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// runCmd creates an ephemeral environment loaded with specific Nix packages.
var runCmd = &cobra.Command{
	Use:   "run [packages...] [-- command args...]",
	Short: "Run a command in an ephemeral Nix environment with ad-hoc packages",
	Long: `derrick run creates a temporary Nix environment with the specified packages
and either executes a command or opens an interactive shell.

Key differences from 'derrick shell':
- 'run' uses packages specified on the command line (no derrick.yaml needed)
- 'run' is ephemeral by default (use --save to persist)
- 'shell' uses packages from your project's derrick.yaml

Examples:
  derrick run nodejs -- npm test
  derrick run python3 -- python -c "print('hello')"
  derrick run --save go git nodejs  # creates a saved environment directory`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var packages []string
		var execArgs []string
		separatorFound := false
		for _, arg := range args {
			if arg == "--" {
				separatorFound = true
				continue
			}
			if separatorFound {
				execArgs = append(execArgs, arg)
			} else {
				packages = append(packages, arg)
			}
		}

		if len(packages) == 0 {
			ui.FailFast(fmt.Errorf("at least one package is required"))
		}

		if saveEnv && rmFiles {
			ui.FailFast(fmt.Errorf("flags --save and --rm-files cannot be used together"))
		}

		cwd, err := os.Getwd()
		if err != nil {
			ui.FailFast(err)
		}

		var workDir string
		var isolateDir bool
		var flakeOutDir string

		if rmFiles {
			workDir, err = os.MkdirTemp("", "derrick-run-*")
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to create temp directory for sandbox: %v", err))
			}
			flakeOutDir = filepath.Join(workDir, ".derrick")
			isolateDir = true
			ui.Infof("Running in ephemeral isolated directory (will be deleted on exit): %s", workDir)
		} else if saveEnv {
			envName := fmt.Sprintf("derrick-env-%d", time.Now().Unix())
			workDir = filepath.Join(cwd, envName)
			err = os.MkdirAll(workDir, 0755)
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to create save directory: %v", err))
			}
			flakeOutDir = filepath.Join(workDir, ".derrick")
			ui.Infof("Saving environment to isolated directory: %s", workDir)
		} else {
			workDir = cwd
			flakeOutDir = filepath.Join(workDir, fmt.Sprintf(".derrick-tmp-%d", time.Now().UnixNano()))
			ui.Infof("Running in local directory, will discard Nix configuration on exit.")
		}

		if rmFiles {
			defer func() {
				ui.Infof("Cleaning up ephemeral files: %s", workDir)
				os.RemoveAll(workDir)
			}()
		} else {
			defer os.RemoveAll(flakeOutDir)
		}

		if workDir != cwd {
			if err := os.Chdir(workDir); err != nil {
				ui.FailFast(fmt.Errorf("failed to change directory to %s: %v", workDir, err))
			}
			defer os.Chdir(cwd)
		}

		ui.Taskf("Resolving packages: %v", packages)

		var nixPkgs []config.NixPackage
		for _, arg := range packages {
			nixPkgs = append(nixPkgs, config.NixPackage{Name: arg})
		}

		if err := engine.BootEnvironment("", nixPkgs, config.DefaultNixRegistry, flakeOutDir); err != nil {
			ui.FailFast(fmt.Errorf("failed to initialize run environment: %v", err))
		}

		if saveEnv {
			yamlPath := filepath.Join(workDir, "derrick.yaml")
			var cfg config.ProjectConfig
			cfg.Name = filepath.Base(workDir)
			cfg.Version = "0.1.0"
			cfg.Nix.Packages = nixPkgs
			cfg.Nix.Registry = config.DefaultNixRegistry

			data, err := yaml.Marshal(&cfg)
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to generate derrick.yaml: %v", err))
			}
			if err := os.WriteFile(yamlPath, data, 0644); err != nil {
				ui.FailFast(fmt.Errorf("failed to write derrick.yaml: %v", err))
			}
			ui.Successf("Successfully wrote derrick.yaml to %s", yamlPath)
		}

		ui.Successf("Environment loaded. Entering %s...", func() string {
			if len(execArgs) > 0 {
				return fmt.Sprintf("command [%s]", strings.Join(execArgs, " "))
			}
			return "shell"
		}())
		if isolateDir {
			ui.Warning("NOTE: Any files created here will be permanently deleted on exit (--rm-files is active).")
		} else if !saveEnv {
			ui.Warning("NOTE: This is an ephemeral environment. Nix packages will be discarded on exit.")
			ui.Info("Use --save to persist the environment in a directory.")
		}

		eng := engine.NewShellEngine()
		if len(execArgs) > 0 {
			if err := eng.EnterSandbox(flakeOutDir, execArgs); err != nil {
				ui.FailFast(fmt.Errorf("command execution failed: %v", err))
			}
		} else {
			ui.Info("For a persistent project environment, use 'derrick shell' instead.")
			ui.Success("Sandbox ready. Use Ctrl+D to exit.")
			if err := eng.EnterSandbox(flakeOutDir, nil); err != nil {
				ui.Warningf("Sandbox session ended: %v", err)
			} else {
				ui.Success("Sandbox session closed.")
			}
		}
	},
}

func init() {
	runCmd.Flags().BoolVar(&saveEnv, "save", false, "Creates a persisted derrick directory + derrick.yaml for these packages")
	runCmd.Flags().BoolVar(&rmFiles, "rm-files", false, "Executes the sandbox in an isolated OS temp directory, wiping any created files on exit")
	rootCmd.AddCommand(runCmd)
}
