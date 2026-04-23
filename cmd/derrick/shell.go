package main

import (
	"github.com/Salv4d/derrick/internal/engine"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

// shellCmd drops into the project's managed shell, routing through the
// resolved provider so docker, nix, and hybrid projects each get the
// correct backend: the nix dev shell for nix/hybrid, a container exec
// for docker-only.
var shellCmd = &cobra.Command{
	Use:   "shell [command...]",
	Short: "Drop into the project's managed shell or execute a command inside it",
	Long: `Opens an interactive shell using the project's provider.

For nix and hybrid projects this is the Nix dev shell (language tooling on
PATH for editors and LSP clients). For docker projects this is a 'docker
compose exec' into the configured service. Any args after the command
name are executed as a one-shot command instead of opening an interactive
session.`,
	Args: cobra.ArbitraryArgs,
	Run: RunDerrick(func(ctx *DerrickContext, cmd *cobra.Command, args []string) {
		cfg := ctx.Config
		cwd := ctx.Cwd

		provider := engine.ResolveProvider(cfg)

		if len(args) > 0 {
			ui.Infof("Executing command via %s provider at %s", provider.Name(), cwd)
		} else {
			ui.Infof("Opening %s shell at %s", provider.Name(), cwd)
		}

		if err := provider.Shell(cfg, args); err != nil {
			ui.FailFast(err)
		}
	}),
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
