package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hayatosc/git-cx/internal/git"
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
	if providers := getAllConfigValues(entries, "cx.providers"); len(providers) > 0 {
		cfg.Providers = providers
	}
	if v := getFirstConfigValue(entries, "cx.model"); v != "" {
		cfg.Model = v
	}
	if v := getFirstConfigValue(entries, "cx.candidates"); v != "" {
		n, err := parseIntConfig("cx.candidates", v)
		if err != nil {
			return err
		}
		cfg.Candidates = n
	}
	if v := getFirstConfigValue(entries, "cx.timeout"); v != "" {
		n, err := parseIntConfig("cx.timeout", v)
		if err != nil {
			return err
		}
		cfg.Timeout = n
	}
	if v := getFirstConfigValue(entries, "cx.command"); v != "" {
		cfg.Command = v
	}
	if v := getFirstConfigValue(entries, "cx.apiBaseUrl"); v != "" {
		cfg.API.BaseURL = v
	}

	if err := applyProviderOverridesFromEntries(cfg, entries); err != nil {
		return err
	}

	if v := getFirstConfigValue(entries, "cx.commit.useEmoji"); v != "" {
		b, err := parseBoolConfig("cx.commit.useEmoji", v)
		if err != nil {
			return err
		}
		cfg.Commit.UseEmoji = b
	}
	if v := getFirstConfigValue(entries, "cx.commit.maxSubjectLength"); v != "" {
		n, err := parseIntConfig("cx.commit.maxSubjectLength", v)
		if err != nil {
			return err
		}
		cfg.Commit.MaxSubjectLength = n
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

func parseIntConfig(key, value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	n, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, trimmed, err)
	}
	return n, nil
}

func parseBoolConfig(key, value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	b, ok := parseGitBool(trimmed)
	if !ok {
		return false, fmt.Errorf("invalid %s %q: expected true/false/on/off/yes/no", key, trimmed)
	}
	return b, nil
}

func applyProviderOverridesFromEntries(cfg *Config, entries map[string][]string) error {
	if cfg.ProviderOptions == nil {
		cfg.ProviderOptions = map[string]ProviderConfig{}
	}
	for key, values := range entries {
		const prefix = "cx.providers."
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		rest := strings.TrimPrefix(key, prefix)
		parts := strings.SplitN(rest, ".", 2)
		if len(parts) != 2 {
			continue
		}
		name, field := parts[0], parts[1]
		if name == "" || len(values) == 0 {
			continue
		}
		value := values[0]
		pc := cfg.ProviderOptions[name]
		switch field {
		case "model":
			pc.Model = value
		case "candidates":
			n, err := parseIntConfig(key, value)
			if err != nil {
				return err
			}
			pc.Candidates = n
		case "timeout":
			n, err := parseIntConfig(key, value)
			if err != nil {
				return err
			}
			pc.Timeout = n
		case "command":
			pc.Command = value
		case "apiBaseUrl":
			pc.APIBaseURL = value
		default:
		}
		if hasProviderConfig(pc) {
			cfg.ProviderOptions[name] = pc
		}
	}
	return nil
}
