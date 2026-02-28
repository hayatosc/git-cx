package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNoStagedChanges is returned when there are no staged changes.
var ErrNoStagedChanges = errors.New("no staged changes: please run 'git add' first")

// StagedDiff returns the staged diff output.
func StagedDiff() (string, error) {
	out, err := run("git", "diff", "--cached", "--no-color")
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return "", ErrNoStagedChanges
	}
	return out, nil
}

// StagedStat returns the --stat output of the staged diff.
func StagedStat() (string, error) {
	out, err := run("git", "diff", "--cached", "--stat", "--no-color")
	if err != nil {
		return "", fmt.Errorf("git diff --stat: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// Commit executes `git commit -m <message>`.
func Commit(message string) error {
	_, err := run("git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// ConfigGet reads a git config value. Returns "" if not set.
func ConfigGet(key string) string {
	out, err := run("git", "config", "--get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ConfigGetAll reads multiple values for a key (e.g. repeated keys).
func ConfigGetAll(key string) []string {
	out, err := run("git", "config", "--get-all", key)
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

// ConfigSet writes a git config value globally.
func ConfigSet(key, value string) error {
	_, err := run("git", "config", "--global", key, value)
	return err
}

// run executes a command and returns combined stdout/stderr.
func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return stdout.String(), nil
}
