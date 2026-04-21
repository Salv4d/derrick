package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Salv4d/derrick/internal/dashboard"
	"github.com/Salv4d/derrick/internal/ui"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve [dir...]",
	Short: "Start the web dashboard",
	Long: `Starts a local web dashboard that shows project status and lets you
start and stop projects from the browser.

Pass one or more project directories as arguments. If no directories are given
the current working directory is used. Each directory must contain a derrick.yaml.

Examples:
  derrick serve                        # dashboard for the current project
  derrick serve ../api ../frontend     # dashboard for multiple projects
  derrick serve --port 8080 .`,
	Run: func(cmd *cobra.Command, args []string) {
		srv, err := dashboard.New(args)
		if err != nil {
			ui.FailFast(err)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err := srv.Serve(ctx, servePort); err != nil {
			ui.FailFast(err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 7700, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
