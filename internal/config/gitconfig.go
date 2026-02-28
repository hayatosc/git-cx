package config

import (
	"context"
	"fmt"
	"strconv"

	"git-cx/internal/git"
)

func getFirstConfigValue(entries map[string][]string, key string) string {
	if entries == nil {
		return ""
	}
	values, ok := entries[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}

func getAllConfigValues(entries map[string][]string, key string) []string {
	if entries == nil {
		return nil
	}
	values, ok := entries[key]
	if !ok {
		return nil
	}
	return values
}

// ApplyGitConfigFile merges values from a gitconfig-format file into cfg.
func ApplyGitConfigFile(ctx context.Context, runner git.Runner, cfg *Config, path string) error {
	entries, err := runner.ConfigListFromFile(ctx, path)
	if err != nil {
		return fmt.Errorf("read git config file %q: %w", path, err)
	}

	if v := getFirstConfigValue(entries, "cx.provider"); v != "" {
		cfg.Provider = v
	}
	if v := getFirstConfigValue(entries, "cx.model"); v != "" {
		cfg.Model = v
	}
	if v := getFirstConfigValue(entries, "cx.candidates"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Candidates = n
		}
	}
	if v := getFirstConfigValue(entries, "cx.timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = n
		}
	}
	if v := getFirstConfigValue(entries, "cx.command"); v != "" {
		cfg.Command = v
	}

	if v := getFirstConfigValue(entries, "cx.commit.useEmoji"); v != "" {
		if b, ok := parseGitBool(v); ok {
			cfg.Commit.UseEmoji = b
		}
	}
	if v := getFirstConfigValue(entries, "cx.commit.maxSubjectLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Commit.MaxSubjectLength = n
		}
	}
	if scopes := getAllConfigValues(entries, "cx.commit.scopes"); len(scopes) > 0 {
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
