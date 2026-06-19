// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for hd CLI.

To load completions:

Bash:
  $ source <(hd completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ hd completion bash > /etc/bash_completion.d/hd
  # macOS:
  $ hd completion bash > $(brew --prefix)/etc/bash_completion.d/hd

Zsh:
  # If shell completion is already enabled in your environment,
  # you can load completions for each session by executing once:
  $ hd completion zsh > "${fpath[1]}/_hd"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ hd completion fish | source
  # To load completions for each session, execute once:
  $ hd completion fish > ~/.config/fish/completions/hd.fish
`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "bash":
			if err := rootCmd.GenBashCompletion(os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating bash completion: %v\n", err)
				os.Exit(1)
			}
		case "zsh":
			if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating zsh completion: %v\n", err)
				os.Exit(1)
			}
		case "fish":
			if err := rootCmd.GenFishCompletion(os.Stdout, true); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating fish completion: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
