package main

import "testing"

func TestInGitHook(t *testing.T) {
	t.Setenv("GIT_DIR", "/tmp/repo/.git")
	t.Setenv("GIT_INDEX_FILE", "/tmp/repo/.git/index")

	if !inGitHook() {
		t.Fatalf("expected hook environment to be detected")
	}
}

func TestInGitHookMissingEnv(t *testing.T) {
	t.Setenv("GIT_DIR", "")
	t.Setenv("GIT_INDEX_FILE", "")

	if inGitHook() {
		t.Fatalf("expected hook detection to be false without git env vars")
	}
}

func TestInGitHookPartialEnv(t *testing.T) {
	t.Setenv("GIT_DIR", "/tmp/repo/.git")
	t.Setenv("GIT_INDEX_FILE", "")

	if inGitHook() {
		t.Fatalf("expected hook detection to require both env vars")
	}
}
