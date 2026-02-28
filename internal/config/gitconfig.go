package config

import (
	"context"
	"strconv"
	"strings"

	"git-cx/internal/git"
)

// ApplyGitConfigFile merges values from a gitconfig-format file into cfg.
func ApplyGitConfigFile(ctx context.Context, runner git.Runner, cfg *Config, path string) {
	if v := runner.ConfigGetFromFile(ctx, path, "cx.provider"); v != "" {
		cfg.Provider = v
	}
	if v := runner.ConfigGetFromFile(ctx, path, "cx.model"); v != "" {
		cfg.Model = v
	}
	if v := runner.ConfigGetFromFile(ctx, path, "cx.candidates"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Candidates = n
		}
	}
	if v := runner.ConfigGetFromFile(ctx, path, "cx.timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = n
		}
	}
	if v := runner.ConfigGetFromFile(ctx, path, "cx.command"); v != "" {
		cfg.Command = v
	}

	if v := runner.ConfigGetFromFile(ctx, path, "cx.commit.useEmoji"); v != "" {
		cfg.Commit.UseEmoji = strings.ToLower(v) == "true"
	}
	if v := runner.ConfigGetFromFile(ctx, path, "cx.commit.maxSubjectLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Commit.MaxSubjectLength = n
		}
	}
	if scopes := runner.ConfigGetAllFromFile(ctx, path, "cx.commit.scopes"); len(scopes) > 0 {
		cfg.Commit.Scopes = scopes
	}
}
