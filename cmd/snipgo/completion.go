package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate completion script",
	Long: `Generate shell completion scripts for snipgo.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, execute once:
  # Linux:
  snipgo completion zsh > "${fpath[1]}/_snipgo"
  
  # macOS:
  snipgo completion zsh > $(brew --prefix)/share/zsh/site-functions/_snipgo
`,
	Args: cobra.NoArgs,
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	Long: `Generate the autocompletion script for zsh shell.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, add to your ~/.zshrc:
  echo 'source <(snipgo completion zsh)' >> ~/.zshrc

Or install to a system-wide location:
  snipgo completion zsh > ~/.zsh/completions/_snipgo
  echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  echo 'autoload -U compinit && compinit' >> ~/.zshrc
`,
	Args: cobra.NoArgs,
	RunE: runCompletionZsh,
}

func runCompletionZsh(cmd *cobra.Command, args []string) error {
	return rootCmd.GenZshCompletion(os.Stdout)
}

