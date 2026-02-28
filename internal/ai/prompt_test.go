package ai

import (
	"strings"
	"testing"
)

func TestBuildPrompt_IncludesStatAndSelections(t *testing.T) {
	req := GenerateRequest{
		Diff:       "diff --git a/a b/a",
		Stat:       "a | 1 +",
		CommitType: "feat",
		Scope:      "core",
		Candidates: 2,
	}
	got := buildPrompt(req)
	if !containsAll(got, []string{
		"generate 2 commit message suggestions",
		"Commit type is already selected: feat",
		"Scope is already selected: core",
		"Changed files:\n",
		"a | 1 +",
		"Git diff:",
	}) {
		t.Fatalf("prompt missing expected content:\n%s", got)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
