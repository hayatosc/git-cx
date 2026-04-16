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
	Provider        string
	Model           string
	Candidates      int
	Timeout         int
	Command         string // for custom provider: supports {prompt} placeholder
	API             APIConfig
	Commit          CommitConfig
	Providers       []string
	ProviderOptions map[string]ProviderConfig
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

// ProviderConfig holds optional provider-specific overrides.
type ProviderConfig struct {
	Model      string
	Candidates int
	Timeout    int
	Command    string
	APIBaseURL string
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

	if cfg.ProviderOptions == nil {
		cfg.ProviderOptions = map[string]ProviderConfig{}
	}

	if v := runner.ConfigGet(ctx, "cx.provider"); v != "" {
		cfg.Provider = v
	}
	if providers := runner.ConfigGetAll(ctx, "cx.providers"); len(providers) > 0 {
		cfg.Providers = providers
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

	loadProviderOverrides(ctx, runner, cfg)

	return cfg
}

func loadProviderOverrides(ctx context.Context, runner git.Runner, cfg *Config) {
	names := normalizeProviders(cfg.Providers, cfg.Provider)
	for _, name := range names {
		pc := cfg.ProviderOptions[name]
		if v := runner.ConfigGet(ctx, fmt.Sprintf("cx.providers.%s.model", name)); v != "" {
			pc.Model = v
		}
		if v := runner.ConfigGet(ctx, fmt.Sprintf("cx.providers.%s.candidates", name)); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				pc.Candidates = n
			}
		}
		if v := runner.ConfigGet(ctx, fmt.Sprintf("cx.providers.%s.timeout", name)); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				pc.Timeout = n
			}
		}
		if v := runner.ConfigGet(ctx, fmt.Sprintf("cx.providers.%s.command", name)); v != "" {
			pc.Command = v
		}
		if v := runner.ConfigGet(ctx, fmt.Sprintf("cx.providers.%s.apiBaseUrl", name)); v != "" {
			pc.APIBaseURL = v
		}
		if hasProviderConfig(pc) {
			cfg.ProviderOptions[name] = pc
		}
	}
}

// Validate checks config values for consistency.
func (c *Config) Validate() error {
	validProviders := map[string]struct{}{
		"gemini":  {},
		"copilot": {},
		"claude":  {},
		"codex":   {},
		"api":     {},
		"custom":  {},
	}

	if c.Provider == "" {
		return fmt.Errorf("provider is not set (set via 'git config cx.provider PROVIDER')")
	}
	if _, ok := validProviders[c.Provider]; !ok {
		return fmt.Errorf("unknown provider: %q (valid providers: gemini, copilot, claude, codex, api, custom; set via 'git config cx.provider PROVIDER')", c.Provider)
	}

	if c.ProviderOptions == nil {
		c.ProviderOptions = map[string]ProviderConfig{}
	}

	c.Providers = normalizeProviders(c.Providers, c.Provider)
	for _, p := range c.Providers {
		if _, ok := validProviders[p]; !ok {
			return fmt.Errorf("unknown provider in cx.providers: %q (valid providers: gemini, copilot, claude, codex, api, custom)", p)
		}
	}
	if c.Candidates <= 0 {
		return fmt.Errorf("candidates must be greater than 0")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	activeProvider := c.ProviderConfig(c.Provider)
	if c.Provider == "custom" && strings.TrimSpace(activeProvider.Command) == "" {
		return fmt.Errorf("cx.providers.%s.command is not set (required for custom provider)", c.Provider)
	}
	if c.Provider == "api" {
		if strings.TrimSpace(activeProvider.APIBaseURL) == "" {
			return fmt.Errorf("cx.providers.%s.apiBaseUrl is not set (required for api provider)", c.Provider)
		}
		if err := validateBaseURL(activeProvider.APIBaseURL); err != nil {
			return fmt.Errorf("cx.providers.%s.apiBaseUrl is invalid: %w", c.Provider, err)
		}
		if strings.TrimSpace(activeProvider.Model) == "" {
			return fmt.Errorf("cx.providers.%s.model is not set (required for api provider)", c.Provider)
		}
	}
	if c.Commit.MaxSubjectLength < 0 {
		return fmt.Errorf("commit.maxSubjectLength must be >= 0")
	}

	for _, name := range c.Providers {
		pc := c.ProviderConfig(name)
		if pc.Candidates <= 0 {
			return fmt.Errorf("cx.providers.%s.candidates must be greater than 0", name)
		}
		if pc.Timeout <= 0 {
			return fmt.Errorf("cx.providers.%s.timeout must be greater than 0", name)
		}
		switch name {
		case "custom":
			if strings.TrimSpace(pc.Command) == "" {
				return fmt.Errorf("cx.providers.%s.command is not set (required for custom provider)", name)
			}
		case "api":
			if strings.TrimSpace(pc.APIBaseURL) == "" {
				return fmt.Errorf("cx.providers.%s.apiBaseUrl is not set (required for api provider)", name)
			}
			if err := validateBaseURL(pc.APIBaseURL); err != nil {
				return fmt.Errorf("cx.providers.%s.apiBaseUrl is invalid: %w", name, err)
			}
			if strings.TrimSpace(pc.Model) == "" {
				return fmt.Errorf("cx.providers.%s.model is not set (required for api provider)", name)
			}
		}
	}
	return nil
}

// ProviderConfig returns provider-specific config merged with defaults.
func (c *Config) ProviderConfig(name string) ProviderConfig {
	base := ProviderConfig{
		Model:      c.Model,
		Candidates: c.Candidates,
		Timeout:    c.Timeout,
		Command:    c.Command,
		APIBaseURL: c.API.BaseURL,
	}
	if opt, ok := c.ProviderOptions[name]; ok {
		if opt.Model != "" {
			base.Model = opt.Model
		}
		if opt.Candidates > 0 {
			base.Candidates = opt.Candidates
		}
		if opt.Timeout > 0 {
			base.Timeout = opt.Timeout
		}
		if opt.Command != "" {
			base.Command = opt.Command
		}
		if opt.APIBaseURL != "" {
			base.APIBaseURL = opt.APIBaseURL
		}
	}
	return base
}

func hasProviderConfig(pc ProviderConfig) bool {
	return pc.Model != "" ||
		pc.Candidates != 0 ||
		pc.Timeout != 0 ||
		pc.Command != "" ||
		pc.APIBaseURL != ""
}

func normalizeProviders(list []string, primary string) []string {
	seen := map[string]bool{}
	var result []string

	if trimmed := strings.TrimSpace(primary); trimmed != "" {
		result = append(result, trimmed)
		seen[trimmed] = true
	}

	for _, p := range list {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		result = append(result, trimmed)
		seen[trimmed] = true
	}
	return result
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
