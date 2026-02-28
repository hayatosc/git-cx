package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/app"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/git"
	"github.com/hayatosc/git-cx/internal/tui"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "git cx",
		Short: "AI-powered git commit with Conventional Commits",
		Long:  "git-cx generates Conventional Commits messages using an AI provider and presents them in an interactive TUI.",
		RunE:  runCommit,
	}

	root.PersistentFlags().String("config", "", "path to gitconfig-format config file")
	root.PersistentFlags().String("provider", "", "AI provider (gemini, copilot, claude, codex, api, custom)")
	root.PersistentFlags().String("model", "", "model name passed to the provider")
	root.PersistentFlags().Int("candidates", 0, "number of commit message candidates")
	root.PersistentFlags().Int("timeout", 0, "request timeout in seconds")
	root.PersistentFlags().String("command", "", "command template for custom provider")
	root.PersistentFlags().String("api-base-url", "", "base URL for api provider")
	root.PersistentFlags().String("api-key", "", "(deprecated) ignored; use OPENAI_API_KEY env instead")
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
	flags := cmd.Flags()
	path, err := flags.GetString("config")
	if err != nil {
		return nil, fmt.Errorf("failed to read config flag: %w", err)
	}
	cfg, err := config.LoadWithFile(ctx, runner, path)
	if err != nil {
		return nil, err
	}
	if flags.Changed("provider") {
		v, err := flags.GetString("provider")
		if err != nil {
			return nil, fmt.Errorf("failed to read provider flag: %w", err)
		}
		cfg.Provider = v
	}
	if flags.Changed("model") {
		v, err := flags.GetString("model")
		if err != nil {
			return nil, fmt.Errorf("failed to read model flag: %w", err)
		}
		cfg.Model = v
	}
	if flags.Changed("candidates") {
		v, err := flags.GetInt("candidates")
		if err != nil {
			return nil, fmt.Errorf("failed to read candidates flag: %w", err)
		}
		cfg.Candidates = v
	}
	if flags.Changed("timeout") {
		v, err := flags.GetInt("timeout")
		if err != nil {
			return nil, fmt.Errorf("failed to read timeout flag: %w", err)
		}
		cfg.Timeout = v
	}
	if flags.Changed("command") {
		v, err := flags.GetString("command")
		if err != nil {
			return nil, fmt.Errorf("failed to read command flag: %w", err)
		}
		cfg.Command = v
	}
	if flags.Changed("api-base-url") {
		v, err := flags.GetString("api-base-url")
		if err != nil {
			return nil, fmt.Errorf("failed to read api-base-url flag: %w", err)
		}
		cfg.API.BaseURL = v
	}
	if flags.Changed("api-key") {
		// deprecated: ignore for now
	}
	if flags.Changed("use-emoji") {
		v, err := flags.GetBool("use-emoji")
		if err != nil {
			return nil, fmt.Errorf("failed to read use-emoji flag: %w", err)
		}
		cfg.Commit.UseEmoji = v
	}
	if flags.Changed("max-subject-length") {
		v, err := flags.GetInt("max-subject-length")
		if err != nil {
			return nil, fmt.Errorf("failed to read max-subject-length flag: %w", err)
		}
		cfg.Commit.MaxSubjectLength = v
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
	  git config --global cx.apiBaseUrl https://api.openai.com/v1
	  # OPENAI_API_KEY=... git cx
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
			if strings.TrimSpace(cfg.API.BaseURL) != "" {
				fmt.Printf("apiBaseUrl:                %s\n", cfg.API.BaseURL)
			}
			keyStatus := "<not set>"
			if strings.TrimSpace(cfg.API.Key) != "" {
				keyStatus = "<set>"
			}
			fmt.Printf("apiKey (OPENAI_API_KEY):   %s\n", keyStatus)
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
