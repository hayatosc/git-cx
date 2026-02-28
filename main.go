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

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	path, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("failed to read config flag: %w", err)
	}
	cfg, err := config.LoadWithFile(path)
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("provider") {
		v, _ := cmd.Flags().GetString("provider")
		cfg.Provider = v
	}
	if cmd.Flags().Changed("model") {
		v, _ := cmd.Flags().GetString("model")
		cfg.Model = v
	}
	if cmd.Flags().Changed("candidates") {
		v, _ := cmd.Flags().GetInt("candidates")
		cfg.Candidates = v
	}
	if cmd.Flags().Changed("timeout") {
		v, _ := cmd.Flags().GetInt("timeout")
		cfg.Timeout = v
	}
	if cmd.Flags().Changed("command") {
		v, _ := cmd.Flags().GetString("command")
		cfg.Command = v
	}
	if cmd.Flags().Changed("use-emoji") {
		v, _ := cmd.Flags().GetBool("use-emoji")
		cfg.Commit.UseEmoji = v
	}
	if cmd.Flags().Changed("max-subject-length") {
		v, _ := cmd.Flags().GetInt("max-subject-length")
		cfg.Commit.MaxSubjectLength = v
	}
	return cfg, nil
}

func runCommit(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(cmd)
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
