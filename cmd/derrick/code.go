package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var codeRmFiles bool

type EditorProfile struct {
	Name    string
	Package string
	Binary  string
}

var editorProfiles = []EditorProfile{
	{Name: "VS Code", Package: "vscode", Binary: "code"},
	{Name: "Cursor", Package: "code-cursor", Binary: "cursor"},
	{Name: "Neovim", Package: "neovim", Binary: "nvim"},
	{Name: "Helix", Package: "helix", Binary: "hx"},
	{Name: "Vim", Package: "vim", Binary: "vim"},
	{Name: "Emacs", Package: "emacs", Binary: "emacs"},
}

var codeCmd = &cobra.Command{
	Use:   "code [editor] [path...]",
	Short: "Launch a pre-configured Editor inside an isolated Sandbox",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader()

		cwd, _ := os.Getwd()

		var selectedEditor EditorProfile
		var targetPath string

		editorArg := ""
		if len(args) > 0 {
			editorArg = args[0]
		}

		found := false
		for _, profile := range editorProfiles {
			if editorArg == profile.Binary || editorArg == profile.Package || (editorArg == "vscode" && profile.Binary == "code") {
				selectedEditor = profile
				found = true
				break
			}
		}

		if !found {
			options := make([]huh.Option[EditorProfile], len(editorProfiles))
			for i, p := range editorProfiles {
				options[i] = huh.NewOption(fmt.Sprintf("%s (%s)", p.Name, p.Binary), p)
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[EditorProfile]().
						Title("Which Editor would you like to launch?").
						Options(options...).
						Value(&selectedEditor),
				),
			)
			if err := form.Run(); err != nil {
				ui.FailFast(err)
			}

			targetPath = "."
			if len(args) > 0 {
				targetPath = args[0]
			}
		} else {
			targetPath = "."
			if len(args) > 1 {
				targetPath = args[1]
			}
		}

		ui.Taskf("Parsing baseline configuration (%s)", configFile)
		cfg, err := config.ParseConfig(configFile, profileName)
		if err != nil {
			ui.FailFast(err)
		}

		ui.Taskf("Injecting IDE package: %s", selectedEditor.Package)
		cfg.Dependencies.NixPackages = append(cfg.Dependencies.NixPackages, config.NixPackage{Name: selectedEditor.Package})

		outDir := filepath.Join(cwd, ".derrick-code")
		if codeRmFiles {
			tmpDir, err := os.MkdirTemp("", "derrick-code-*")
			if err != nil {
				ui.FailFast(fmt.Errorf("failed to create temporary code directory: %v", err))
			}
			outDir = tmpDir
			defer os.RemoveAll(tmpDir)
			ui.Infof("Using ephemeral IDE sandbox at %s", tmpDir)
		} else {
			ui.Infof("Using persistent IDE sandbox at %s (cache preserved)", ".derrick-code")
		}

		if err := engine.BootEnvironment(configFile, cfg.Dependencies.NixPackages, cfg.Dependencies.NixRegistry, outDir); err != nil {
			ui.FailFast(err)
		}

		eng := engine.NewShellEngine()
		ui.Successf("Launching \033[1;36m%s\033[0m firmly bound to your project...", selectedEditor.Name)

		execArgs := []string{selectedEditor.Binary, targetPath}

		if err := eng.EnterSandbox(outDir, execArgs); err != nil {
			ui.FailFast(err)
		}
	},
}

func init() {
	codeCmd.Flags().BoolVar(&codeRmFiles, "rm", false, "Executes the IDE sandbox in an isolated OS temp directory, wiping generation cache on exit")
	rootCmd.AddCommand(codeCmd)
}
