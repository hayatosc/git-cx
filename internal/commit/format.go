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
