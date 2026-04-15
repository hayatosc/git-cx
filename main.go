package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/app"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/git"
	"github.com/hayatosc/git-cx/internal/tui"
)

var version = "dev"

// inGitHook reports whether git-cx is being invoked from a Git hook by
// checking Git-provided environment variables.
func inGitHook() bool {
	return os.Getenv("GIT_DIR") != "" && os.Getenv("GIT_INDEX_FILE") != ""
}

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
	root.PersistentFlags().Bool("use-emoji", false, "prefix commit type with emoji")
	root.PersistentFlags().Int("max-subject-length", 0, "max length of commit subject line")
	root.PersistentFlags().Bool("dry-run", false, "preview commit message without actually committing")

	root.AddCommand(newConfigCmd())
	root.AddCommand(newVersionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func applyStringFlag(flags *pflag.FlagSet, name string, dest *string) error {
	if !flags.Changed(name) {
		return nil
	}
	v, err := flags.GetString(name)
	if err != nil {
		return fmt.Errorf("failed to read %s flag: %w", name, err)
	}
	*dest = v
	return nil
}

func applyIntFlag(flags *pflag.FlagSet, name string, dest *int) error {
	if !flags.Changed(name) {
		return nil
	}
	v, err := flags.GetInt(name)
	if err != nil {
		return fmt.Errorf("failed to read %s flag: %w", name, err)
	}
	*dest = v
	return nil
}

func applyBoolFlag(flags *pflag.FlagSet, name string, dest *bool) error {
	if !flags.Changed(name) {
		return nil
	}
	v, err := flags.GetBool(name)
	if err != nil {
		return fmt.Errorf("failed to read %s flag: %w", name, err)
	}
	*dest = v
	return nil
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
	for _, fn := range []func() error{
		func() error { return applyStringFlag(flags, "provider", &cfg.Provider) },
		func() error { return applyStringFlag(flags, "model", &cfg.Model) },
		func() error { return applyIntFlag(flags, "candidates", &cfg.Candidates) },
		func() error { return applyIntFlag(flags, "timeout", &cfg.Timeout) },
		func() error { return applyStringFlag(flags, "command", &cfg.Command) },
		func() error { return applyStringFlag(flags, "api-base-url", &cfg.API.BaseURL) },
		func() error { return applyBoolFlag(flags, "use-emoji", &cfg.Commit.UseEmoji) },
		func() error { return applyIntFlag(flags, "max-subject-length", &cfg.Commit.MaxSubjectLength) },
	} {
		if err := fn(); err != nil {
			return nil, err
		}
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

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	commitService := app.NewCommitService(cfg, provider, gitRunner)
	diff, stat, err := commitService.StagedChanges(ctx)
	if err != nil {
		if errors.Is(err, git.ErrNoStagedChanges) {
			if !dryRun {
				fmt.Fprintln(os.Stderr, "Error: no staged changes. Run 'git add' first.")
				os.Exit(1)
			}
			diff, _ = gitRunner.UnstagedDiff(ctx)
			stat, _ = gitRunner.UnstagedStat(ctx)
			if strings.TrimSpace(diff) == "" {
				diff, _ = gitRunner.LastCommitDiff(ctx)
				stat, _ = gitRunner.LastCommitStat(ctx)
			}
		} else {
			return fmt.Errorf("failed to get staged diff: %w", err)
		}
	}

	m := tui.New(commitService, diff, stat, dryRun)

	hookMode := inGitHook()
	opts := []tea.ProgramOption{}
	if !hookMode {
		opts = append(opts, tea.WithAltScreen())
	} else {
		fmt.Fprintln(os.Stderr, "git-cx: detected git hook environment, keeping output on the main screen.")
	}

	p := tea.NewProgram(m, opts...)
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	if final, ok := result.(tui.Model); ok {
		if out := strings.TrimSpace(final.LogOutput()); out != "" {
			fmt.Fprintln(os.Stderr, out)
		}
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
