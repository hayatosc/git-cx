package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyTOML_CommitSnakeCaseKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`
[commit]
use_emoji = true
max_subject_length = 88
scopes = ["core", "cli"]
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := DefaultConfig()
	if err := ApplyTOML(cfg, path); err != nil {
		t.Fatalf("ApplyTOML error: %v", err)
	}

	if !cfg.Commit.UseEmoji {
		t.Fatalf("expected UseEmoji=true")
	}
	if cfg.Commit.MaxSubjectLength != 88 {
		t.Fatalf("expected MaxSubjectLength=88, got %d", cfg.Commit.MaxSubjectLength)
	}
	if len(cfg.Commit.Scopes) != 2 || cfg.Commit.Scopes[0] != "core" || cfg.Commit.Scopes[1] != "cli" {
		t.Fatalf("unexpected scopes: %#v", cfg.Commit.Scopes)
	}
}

func TestApplyTOML_CommitGitConfigStyleKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`
[commit]
useEmoji = true
maxSubjectLength = 88
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := DefaultConfig()
	if err := ApplyTOML(cfg, path); err != nil {
		t.Fatalf("ApplyTOML error: %v", err)
	}

	if !cfg.Commit.UseEmoji {
		t.Fatalf("expected UseEmoji=true")
	}
	if cfg.Commit.MaxSubjectLength != 88 {
		t.Fatalf("expected MaxSubjectLength=88, got %d", cfg.Commit.MaxSubjectLength)
	}
}
