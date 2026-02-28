package ai

import "fmt"

const maxDiffLen = 4000

// buildPrompt constructs the prompt string sent to the AI provider.
func buildPrompt(req GenerateRequest) string {
	base := fmt.Sprintf(`You are a commit message generator. Based on the following git diff, generate %d commit message suggestions in Conventional Commits format.

Rules:
- Format: <type>(<scope>): <subject>
- type must be one of: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
- scope is optional
- subject must be lowercase, imperative mood, no period at end
- subject must be concise (under 72 characters)
- Output ONLY the commit messages, one per line, no numbering, no explanation

`, req.Candidates)

	if req.CommitType != "" {
		base += fmt.Sprintf("Commit type is already selected: %s\n", req.CommitType)
	}
	if req.Scope != "" {
		base += fmt.Sprintf("Scope is already selected: %s\n", req.Scope)
	}

	if req.Stat != "" {
		base += fmt.Sprintf("\nChanged files:\n%s\n", req.Stat)
	}

	diff := req.Diff
	truncated := false
	if len(diff) > maxDiffLen {
		diff = diff[:maxDiffLen]
		truncated = true
	}
	base += fmt.Sprintf("\nGit diff:\n```\n%s\n```", diff)
	if truncated {
		base += "\n(diff truncated)"
	}
	return base
}
