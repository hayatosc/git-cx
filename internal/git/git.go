package git

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"git-cx/internal/execx"
)

// ErrNoStagedChanges is returned when there are no staged changes.
var ErrNoStagedChanges = errors.New("no staged changes: please run 'git add' first")

// Runner executes git commands.
type Runner struct {
	runner execx.Runner
}

// NewRunner creates a Runner with the default executor.
func NewRunner() Runner {
	return Runner{runner: execx.DefaultRunner{}}
}

// NewRunnerWithExecutor creates a Runner with a custom executor.
func NewRunnerWithExecutor(r execx.Runner) Runner {
	return Runner{runner: r}
}

// StagedDiff returns the staged diff output.
func (r Runner) StagedDiff(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "git", "diff", "--cached", "--no-color")
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return "", ErrNoStagedChanges
	}
	return out, nil
}

// StagedStat returns the --stat output of the staged diff.
func (r Runner) StagedStat(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "git", "diff", "--cached", "--stat", "--no-color")
	if err != nil {
		return "", fmt.Errorf("git diff --stat: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// Commit executes `git commit -m <message>`.
func (r Runner) Commit(ctx context.Context, message string) error {
	_, err := r.run(ctx, "git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// ConfigGet reads a git config value. Returns "" if not set.
func (r Runner) ConfigGet(ctx context.Context, key string) string {
	out, err := r.run(ctx, "git", "config", "--get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ConfigGetFromFile reads a git config value from a config file.
// Returns "" and nil if the key is not set.
func (r Runner) ConfigGetFromFile(ctx context.Context, path, key string) (string, error) {
	out, err := r.run(ctx, "git", "config", "--file", path, "--get", key)
	if err != nil {
		if isGitConfigNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("git config --file %q --get %s: %w", path, key, err)
	}
	return strings.TrimSpace(out), nil
}

// ConfigGetAll reads multiple values for a key (e.g. repeated keys).
func (r Runner) ConfigGetAll(ctx context.Context, key string) []string {
	out, err := r.run(ctx, "git", "config", "--get-all", key)
	if err != nil {
		return nil
	}
	var result []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	return result
}

// ConfigGetAllFromFile reads multiple values for a key from a config file.
// Returns nil and nil if the key is not set.
func (r Runner) ConfigGetAllFromFile(ctx context.Context, path, key string) ([]string, error) {
	out, err := r.run(ctx, "git", "config", "--file", path, "--get-all", key)
	if err != nil {
		if isGitConfigNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("git config --file %s --get-all %s: %w", path, key, err)
	}
	var result []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

// ConfigSet writes a git config value globally.
func (r Runner) ConfigSet(ctx context.Context, key, value string) error {
	_, err := r.run(ctx, "git", "config", "--global", key, value)
	if err != nil {
		return fmt.Errorf("git config --global: %w", err)
	}
	return nil
}

func (r Runner) run(ctx context.Context, name string, args ...string) (string, error) {
	result, err := r.runner.Run(ctx, name, args...)
	if err != nil {
		msg := strings.TrimSpace(result.Stderr)
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return result.Stdout, nil
}

func isGitConfigNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "exit status 1")
}

// ConfigListFromFile lists entries in a gitconfig-format file.
func (r Runner) ConfigListFromFile(ctx context.Context, path string) (string, error) {
	out, err := r.run(ctx, "git", "config", "--file", path, "--list")
	if err != nil {
		return "", fmt.Errorf("git config --file %s --list: %w", path, err)
	}
	return strings.TrimSpace(out), nil
}
