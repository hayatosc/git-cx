package ai

import (
	"fmt"
	"strings"
)

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
	if req.Subject != "" {
		base += fmt.Sprintf("Subject is already selected: %s\n", req.Subject)
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

// buildDetailPrompt constructs the prompt for body/footer generation.
func buildDetailPrompt(req GenerateRequest) string {
	base := `You are a commit message generator. Based on the following git diff, generate a commit body and footer for the subject below.

Rules:
- Use Conventional Commits style
- Body is optional; footer is optional
- Output ONLY the result in this exact format:
Body:
<body text or empty>
Footer:
<footer text or empty>

`

	if req.CommitType != "" {
		base += fmt.Sprintf("Commit type is already selected: %s\n", req.CommitType)
	}
	if req.Scope != "" {
		base += fmt.Sprintf("Scope is already selected: %s\n", req.Scope)
	}
	if req.Subject != "" {
		base += fmt.Sprintf("Subject is already selected: %s\n", req.Subject)
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

// parseDetailOutput extracts body and footer from AI output.
func parseDetailOutput(output string) (string, string) {
	const bodyLabel = "Body:"
	const footerLabel = "Footer:"

	body := ""
	footer := ""
	section := ""

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case bodyLabel:
			section = "body"
			continue
		case footerLabel:
			section = "footer"
			continue
		}

		switch section {
		case "body":
			if body != "" {
				body += "\n"
			}
			body += line
		case "footer":
			if footer != "" {
				footer += "\n"
			}
			footer += line
		}
	}

	return strings.TrimSpace(body), strings.TrimSpace(footer)
}
