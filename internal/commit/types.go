package commit

// ConventionalCommit represents a Conventional Commits message.
type ConventionalCommit struct {
	Type     string
	Scope    string
	Breaking bool
	Subject  string
	Body     string
	Footer   string
}

// CommitTypes is the ordered list of supported commit types.
var CommitTypes = []string{
	"auto",
	"feat",
	"fix",
	"docs",
	"style",
	"refactor",
	"perf",
	"test",
	"build",
	"ci",
	"chore",
	"revert",
}

// CommitTypeDescriptions maps each type to a short description.
var CommitTypeDescriptions = map[string]string{
	"auto":     "AI selects type and header",
	"feat":     "A new feature",
	"fix":      "A bug fix",
	"docs":     "Documentation only changes",
	"style":    "Changes that do not affect the meaning of the code",
	"refactor": "A code change that neither fixes a bug nor adds a feature",
	"perf":     "A code change that improves performance",
	"test":     "Adding missing tests or correcting existing tests",
	"build":    "Changes that affect the build system or external dependencies",
	"ci":       "Changes to CI configuration files and scripts",
	"chore":    "Other changes that don't modify src or test files",
	"revert":   "Reverts a previous commit",
}
