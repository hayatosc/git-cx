package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"git-cx/internal/git"
)

// Config holds all git-cx configuration.
type Config struct {
	Provider   string
	Model      string
	Candidates int
	Timeout    int
	Command    string // for custom provider: supports {prompt} placeholder
	Commit     CommitConfig
}

// CommitConfig holds commit message formatting settings.
type CommitConfig struct {
	UseEmoji         bool
	MaxSubjectLength int
	Scopes           []string
}

// Load reads config from git config, falling back to defaults.
func Load(ctx context.Context, runner git.Runner) (*Config, error) {
	cfg := loadBase(ctx, runner)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadWithFile reads config from git config and, if path is non-empty,
// merges the TOML file at that path on top.
func LoadWithFile(ctx context.Context, runner git.Runner, path string) (*Config, error) {
	cfg := loadBase(ctx, runner)
	if path != "" {
		if err := ApplyTOML(cfg, path); err != nil {
			return nil, fmt.Errorf("failed to load config file %q: %w", path, err)
		}
	}
	return cfg, cfg.Validate()
}

func loadBase(ctx context.Context, runner git.Runner) *Config {
	cfg := DefaultConfig()

	if v := runner.ConfigGet(ctx, "cx.provider"); v != "" {
		cfg.Provider = v
	}
	if v := runner.ConfigGet(ctx, "cx.model"); v != "" {
		cfg.Model = v
	}
	if v := runner.ConfigGet(ctx, "cx.candidates"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Candidates = n
		}
	}
	if v := runner.ConfigGet(ctx, "cx.timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = n
		}
	}
	if v := runner.ConfigGet(ctx, "cx.command"); v != "" {
		cfg.Command = v
	}

	// Commit formatting
	if v := runner.ConfigGet(ctx, "cx.commit.useEmoji"); v != "" {
		cfg.Commit.UseEmoji = strings.ToLower(v) == "true"
	}
	if v := runner.ConfigGet(ctx, "cx.commit.maxSubjectLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Commit.MaxSubjectLength = n
		}
	}
	if scopes := runner.ConfigGetAll(ctx, "cx.commit.scopes"); len(scopes) > 0 {
		cfg.Commit.Scopes = scopes
	}

	return cfg
}

// Validate checks config values for consistency.
func (c *Config) Validate() error {
	switch c.Provider {
	case "gemini", "copilot", "custom":
	default:
		return fmt.Errorf("unknown provider: %q", c.Provider)
	}
	if c.Candidates <= 0 {
		return fmt.Errorf("candidates must be greater than 0")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	if c.Provider == "custom" && strings.TrimSpace(c.Command) == "" {
		return fmt.Errorf("cx.command is not set (required for custom provider)")
	}
	if c.Commit.MaxSubjectLength < 0 {
		return fmt.Errorf("commit.maxSubjectLength must be >= 0")
	}
	return nil
}
