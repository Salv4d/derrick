package main

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd emits shell completion scripts. It delegates to cobra's
// built-in generators so new subcommands are picked up automatically.
var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Generate shell completion script",
	Long: `Generate shell completion for Derrick.

Load completions in the current session:

  Bash:       source <(derrick completion bash)
  Zsh:        source <(derrick completion zsh)
  Fish:       derrick completion fish | source
  PowerShell: derrick completion powershell | Out-String | Invoke-Expression

Persist completions across sessions:

  Bash (Linux):   derrick completion bash > /etc/bash_completion.d/derrick
  Bash (macOS):   derrick completion bash > $(brew --prefix)/etc/bash_completion.d/derrick
  Zsh:            derrick completion zsh  > "${fpath[1]}/_derrick"
  Fish:           derrick completion fish > ~/.config/fish/completions/derrick.fish`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
