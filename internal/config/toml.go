package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Provider   *string     `toml:"provider"`
	Model      *string     `toml:"model"`
	Candidates *int        `toml:"candidates"`
	Timeout    *int        `toml:"timeout"`
	Command    *string     `toml:"command"`
	Commit     *tomlCommit `toml:"commit"`
}

type tomlCommit struct {
	UseEmoji                *bool    `toml:"use_emoji"`
	UseEmojiCamel           *bool    `toml:"useEmoji"`
	MaxSubjectLength        *int     `toml:"max_subject_length"`
	MaxSubjectLengthCamel   *int     `toml:"maxSubjectLength"`
	Scopes                  []string `toml:"scopes"`
}

// ApplyTOML reads the TOML file at path and merges it into cfg.
// Fields present in the TOML override fields from git config.
func ApplyTOML(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	var tc tomlConfig
	meta, err := toml.Decode(string(data), &tc)
	if err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return fmt.Errorf("parse config file: unknown keys: %v", undecoded)
	}
	if tc.Provider != nil {
		cfg.Provider = *tc.Provider
	}
	if tc.Model != nil {
		cfg.Model = *tc.Model
	}
	if tc.Candidates != nil {
		cfg.Candidates = *tc.Candidates
	}
	if tc.Timeout != nil {
		cfg.Timeout = *tc.Timeout
	}
	if tc.Command != nil {
		cfg.Command = *tc.Command
	}
	if tc.Commit != nil {
		if tc.Commit.UseEmoji != nil {
			cfg.Commit.UseEmoji = *tc.Commit.UseEmoji
		} else if tc.Commit.UseEmojiCamel != nil {
			cfg.Commit.UseEmoji = *tc.Commit.UseEmojiCamel
		}
		if tc.Commit.MaxSubjectLength != nil {
			cfg.Commit.MaxSubjectLength = *tc.Commit.MaxSubjectLength
		} else if tc.Commit.MaxSubjectLengthCamel != nil {
			cfg.Commit.MaxSubjectLength = *tc.Commit.MaxSubjectLengthCamel
		}
		if tc.Commit.Scopes != nil {
			cfg.Commit.Scopes = tc.Commit.Scopes
		}
	}
	return nil
}
