package config

import (
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
func Load() *Config {
	cfg := DefaultConfig()

	if v := git.ConfigGet("cx.provider"); v != "" {
		cfg.Provider = v
	}
	if v := git.ConfigGet("cx.model"); v != "" {
		cfg.Model = v
	}
	if v := git.ConfigGet("cx.candidates"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Candidates = n
		}
	}
	if v := git.ConfigGet("cx.timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = n
		}
	}
	if v := git.ConfigGet("cx.command"); v != "" {
		cfg.Command = v
	}

	// Commit formatting
	if v := git.ConfigGet("cx.commit.useEmoji"); v != "" {
		cfg.Commit.UseEmoji = strings.ToLower(v) == "true"
	}
	if v := git.ConfigGet("cx.commit.maxSubjectLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Commit.MaxSubjectLength = n
		}
	}
	if scopes := git.ConfigGetAll("cx.commit.scopes"); len(scopes) > 0 {
		cfg.Commit.Scopes = scopes
	}

	return cfg
}

// LoadWithFile reads config from git config and, if path is non-empty,
// merges the TOML file at that path on top.
func LoadWithFile(path string) (*Config, error) {
	cfg := Load()
	if path != "" {
		if err := ApplyTOML(cfg, path); err != nil {
			return nil, fmt.Errorf("failed to load config file %q: %w", path, err)
		}
	}
	return cfg, nil
}
