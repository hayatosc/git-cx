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
		Subject:    "add feature",
		Candidates: 2,
	}
	got := buildPrompt(req)
	if !containsAll(got, []string{
		"generate 2 commit message suggestions",
		"Commit type is already selected: feat",
		"Scope is already selected: core",
		"Subject is already selected: add feature",
		"Changed files:\n",
		"a | 1 +",
		"Git diff:",
	}) {
		t.Fatalf("prompt missing expected content:\n%s", got)
	}
}

func TestBuildDetailPrompt_IncludesSubject(t *testing.T) {
	req := GenerateRequest{
		Diff:       "diff --git a/a b/a",
		Stat:       "a | 1 +",
		CommitType: "feat",
		Scope:      "core",
		Subject:    "add feature",
		Candidates: 1,
	}
	got := buildDetailPrompt(req)
	if !containsAll(got, []string{
		"generate a commit body and footer",
		"Commit type is already selected: feat",
		"Scope is already selected: core",
		"Subject is already selected: add feature",
		"Changed files:\n",
		"Git diff:",
	}) {
		t.Fatalf("detail prompt missing expected content:\n%s", got)
	}
}

func TestParseDetailOutput(t *testing.T) {
	input := "Body:\nline1\nline2\nFooter:\nRefs: #1\nReviewed-by: bot"
	body, footer := parseDetailOutput(input)
	if body != "line1\nline2" {
		t.Fatalf("unexpected body: %q", body)
	}
	if footer != "Refs: #1\nReviewed-by: bot" {
		t.Fatalf("unexpected footer: %q", footer)
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
