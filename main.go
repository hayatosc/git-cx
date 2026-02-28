package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"git-cx/internal/ai"
	"git-cx/internal/config"
	"git-cx/internal/git"
	"git-cx/internal/tui"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "git cx",
		Short: "AI-powered git commit with Conventional Commits",
		Long:  "git-cx generates Conventional Commits messages using an AI CLI provider and presents them in an interactive TUI.",
		RunE:  runCommit,
	}

	root.AddCommand(newConfigCmd())
	root.AddCommand(newVersionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCommit(_ *cobra.Command, _ []string) error {
	cfg := config.Load()

	diff, err := git.StagedDiff()
	if err != nil {
		if errors.Is(err, git.ErrNoStagedChanges) {
			fmt.Fprintln(os.Stderr, "Error: no staged changes. Run 'git add' first.")
			os.Exit(1)
		}
		return fmt.Errorf("failed to get staged diff: %w", err)
	}

	stat, _ := git.StagedStat()

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize AI provider: %w", err)
	}

	m := tui.New(cfg, provider, diff, stat)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("git-cx %s\n", version)
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or set git-cx configuration",
		Long: `Show current git-cx configuration read from git config.

Use 'git config --global cx.<key> <value>' to set values.

Example:
  git config --global cx.provider gemini
  git config --global cx.model gemini-3.0-flash
  git config --global cx.candidates 3
  git config --global cx.timeout 30
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.Load()
			fmt.Printf("provider:                  %s\n", cfg.Provider)
			fmt.Printf("model:                     %s\n", cfg.Model)
			fmt.Printf("candidates:                %d\n", cfg.Candidates)
			fmt.Printf("timeout:                   %d\n", cfg.Timeout)
			if cfg.Command != "" {
				fmt.Printf("command:                   %s\n", cfg.Command)
			}
			fmt.Printf("commit.useEmoji:           %v\n", cfg.Commit.UseEmoji)
			fmt.Printf("commit.maxSubjectLength:   %d\n", cfg.Commit.MaxSubjectLength)
			if len(cfg.Commit.Scopes) > 0 {
				fmt.Printf("commit.scopes:             %v\n", cfg.Commit.Scopes)
			}
			return nil
		},
	}
	return cmd
}
