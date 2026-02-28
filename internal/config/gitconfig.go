package config

import (
	"context"
	"fmt"
	"strconv"

	"git-cx/internal/git"
)

// ApplyGitConfigFile merges values from a gitconfig-format file into cfg.
func ApplyGitConfigFile(ctx context.Context, runner git.Runner, cfg *Config, path string) error {
	if err := assertReadableGitConfigFile(ctx, runner, path); err != nil {
		return err
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.provider"); err != nil {
		return err
	} else if v != "" {
		cfg.Provider = v
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.model"); err != nil {
		return err
	} else if v != "" {
		cfg.Model = v
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.candidates"); err != nil {
		return err
	} else if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Candidates = n
		}
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.timeout"); err != nil {
		return err
	} else if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = n
		}
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.command"); err != nil {
		return err
	} else if v != "" {
		cfg.Command = v
	}

	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.commit.useEmoji"); err != nil {
		return err
	} else if v != "" {
		if b, ok := parseGitBool(v); ok {
			cfg.Commit.UseEmoji = b
		}
	}
	if v, err := runner.ConfigGetFromFile(ctx, path, "cx.commit.maxSubjectLength"); err != nil {
		return err
	} else if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Commit.MaxSubjectLength = n
		}
	}
	if scopes, err := runner.ConfigGetAllFromFile(ctx, path, "cx.commit.scopes"); err != nil {
		return err
	} else if len(scopes) > 0 {
		cfg.Commit.Scopes = scopes
	}
	return nil
}

func assertReadableGitConfigFile(ctx context.Context, runner git.Runner, path string) error {
	_, err := runner.ConfigListFromFile(ctx, path)
	if err != nil {
		return fmt.Errorf("read git config file %q: %w", path, err)
	}
	return nil
}
