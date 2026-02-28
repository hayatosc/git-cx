package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"git-cx/internal/ai"
	"git-cx/internal/app"
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

	root.PersistentFlags().String("config", "", "path to TOML config file")
	root.PersistentFlags().String("provider", "", "AI provider (gemini, copilot, custom)")
	root.PersistentFlags().String("model", "", "model name passed to the provider")
	root.PersistentFlags().Int("candidates", 0, "number of commit message candidates")
	root.PersistentFlags().Int("timeout", 0, "request timeout in seconds")
	root.PersistentFlags().String("command", "", "command template for custom provider")
	root.PersistentFlags().Bool("use-emoji", false, "prefix commit type with emoji")
	root.PersistentFlags().Int("max-subject-length", 0, "max length of commit subject line")

	root.AddCommand(newConfigCmd())
	root.AddCommand(newVersionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadConfig(cmd *cobra.Command, runner git.Runner) (*config.Config, error) {
	ctx := context.Background()
	path := mustGetString(cmd, "config")
	cfg, err := config.LoadWithFile(ctx, runner, path)
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("provider") {
		cfg.Provider = mustGetString(cmd, "provider")
	}
	if cmd.Flags().Changed("model") {
		cfg.Model = mustGetString(cmd, "model")
	}
	if cmd.Flags().Changed("candidates") {
		cfg.Candidates = mustGetInt(cmd, "candidates")
	}
	if cmd.Flags().Changed("timeout") {
		cfg.Timeout = mustGetInt(cmd, "timeout")
	}
	if cmd.Flags().Changed("command") {
		cfg.Command = mustGetString(cmd, "command")
	}
	if cmd.Flags().Changed("use-emoji") {
		cfg.Commit.UseEmoji = mustGetBool(cmd, "use-emoji")
	}
	if cmd.Flags().Changed("max-subject-length") {
		cfg.Commit.MaxSubjectLength = mustGetInt(cmd, "max-subject-length")
	}
	return cfg, cfg.Validate()
}

func runCommit(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	gitRunner := git.NewRunner()

	cfg, err := loadConfig(cmd, gitRunner)
	if err != nil {
		return err
	}

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize AI provider: %w", err)
	}

	commitService := app.NewCommitService(cfg, provider, gitRunner)
	diff, stat, err := commitService.StagedChanges(ctx)
	if err != nil {
		if errors.Is(err, git.ErrNoStagedChanges) {
			fmt.Fprintln(os.Stderr, "Error: no staged changes. Run 'git add' first.")
			os.Exit(1)
		}
		return fmt.Errorf("failed to get staged diff: %w", err)
	}

	m := tui.New(commitService, diff, stat)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func mustGetString(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Errorf("failed to read %s flag: %w", name, err))
	}
	return value
}

func mustGetInt(cmd *cobra.Command, name string) int {
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(fmt.Errorf("failed to read %s flag: %w", name, err))
	}
	return value
}

func mustGetBool(cmd *cobra.Command, name string) bool {
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(fmt.Errorf("failed to read %s flag: %w", name, err))
	}
	return value
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(cmd, git.NewRunner())
			if err != nil {
				return err
			}
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
