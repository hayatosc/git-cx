package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/hayatosc/git-cx/internal/git"
)

// Config holds all git-cx configuration.
type Config struct {
	Provider   string
	Model      string
	Candidates int
	Timeout    int
	Command    string // for custom provider: supports {prompt} placeholder
	API        APIConfig
	Commit     CommitConfig
}

// APIConfig holds API provider settings.
type APIConfig struct {
	BaseURL string
	Key     string
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
// merges the gitconfig-format file at that path on top.
func LoadWithFile(ctx context.Context, runner git.Runner, path string) (*Config, error) {
	cfg := loadBase(ctx, runner)
	if path != "" {
		if err := ApplyGitConfigFile(ctx, runner, cfg, path); err != nil {
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
	if v := runner.ConfigGet(ctx, "cx.apiBaseUrl"); v != "" {
		cfg.API.BaseURL = v
	}
	if cfg.API.BaseURL == "" {
		if v := runner.ConfigGet(ctx, "cx.api.baseUrl"); v != "" {
			cfg.API.BaseURL = v
		}
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		cfg.API.Key = v
	}

	// Commit formatting
	if v := runner.ConfigGet(ctx, "cx.commit.useEmoji"); v != "" {
		if b, ok := parseGitBool(v); ok {
			cfg.Commit.UseEmoji = b
		}
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
	case "gemini", "copilot", "claude", "codex", "api", "custom":
	default:
		return fmt.Errorf("unknown provider: %q (valid providers: gemini, copilot, claude, codex, api, custom; set via 'git config cx.provider PROVIDER')", c.Provider)
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
	if c.Provider == "api" {
		if strings.TrimSpace(c.API.BaseURL) == "" {
			return fmt.Errorf("cx.apiBaseUrl is not set (required for api provider)")
		}
		if err := validateBaseURL(c.API.BaseURL); err != nil {
			return fmt.Errorf("cx.apiBaseUrl is invalid: %w", err)
		}
		if strings.TrimSpace(c.Model) == "" {
			return fmt.Errorf("cx.model is not set (required for api provider)")
		}
	}
	if c.Commit.MaxSubjectLength < 0 {
		return fmt.Errorf("commit.maxSubjectLength must be >= 0")
	}
	return nil
}

func validateBaseURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("base URL must include scheme and host")
	}
	return nil
}

func parseGitBool(value string) (bool, bool) {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "yes", "on":
		return true, true
	case "no", "off":
		return false, true
	}
	b, err := strconv.ParseBool(trimmed)
	if err != nil {
		return false, false
	}
	return b, true
}
