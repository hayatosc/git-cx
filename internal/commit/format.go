package commit

import "strings"

// Format returns the full commit message string from a ConventionalCommit.
func Format(c *ConventionalCommit, useEmoji bool, maxSubjectLen int) string {
	var sb strings.Builder

	// Build header: type(scope)!: subject
	sb.WriteString(c.Type)
	if c.Scope != "" {
		sb.WriteString("(")
		sb.WriteString(c.Scope)
		sb.WriteString(")")
	}
	if c.Breaking {
		sb.WriteString("!")
	}
	sb.WriteString(": ")

	subject := c.Subject
	if maxSubjectLen > 0 && len(subject) > maxSubjectLen {
		subject = subject[:maxSubjectLen]
	}
	if useEmoji {
		if emoji, ok := typeEmojis[c.Type]; ok {
			sb.WriteString(emoji)
			sb.WriteString(" ")
		}
	}
	sb.WriteString(subject)

	if c.Body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(c.Body)
	}

	if c.Footer != "" {
		sb.WriteString("\n\n")
		sb.WriteString(c.Footer)
	}

	return sb.String()
}

// BuildMessage decides whether to format or use raw subject.
func BuildMessage(c *ConventionalCommit, useEmoji bool, maxSubjectLen int) string {
	if isConventionalHeader(c.Subject) {
		result := c.Subject
		if c.Body != "" {
			result += "\n\n" + c.Body
		}
		if c.Footer != "" {
			result += "\n\n" + c.Footer
		}
		return result
	}
	if c.Type == "" {
		result := c.Subject
		if c.Body != "" {
			result += "\n\n" + c.Body
		}
		if c.Footer != "" {
			result += "\n\n" + c.Footer
		}
		return result
	}
	return Format(c, useEmoji, maxSubjectLen)
}

func isConventionalHeader(s string) bool {
	for _, t := range CommitTypes {
		if t == "auto" {
			continue
		}
		if len(s) > len(t) && s[:len(t)] == t {
			rest := s[len(t):]
			if len(rest) > 0 && (rest[0] == '(' || rest[0] == ':' || rest[0] == '!') {
				return true
			}
		}
	}
	return false
}

// typeEmojis maps commit types to gitmoji-style emojis.
var typeEmojis = map[string]string{
	"feat":     "✨",
	"fix":      "🐛",
	"docs":     "📝",
	"style":    "💄",
	"refactor": "♻️",
	"perf":     "⚡️",
	"test":     "✅",
	"build":    "🔧",
	"ci":       "👷",
	"chore":    "🔨",
	"revert":   "⏪",
}
