package app

import (
	"context"
	"errors"
	"strings"

	"git-cx/internal/ai"
	"git-cx/internal/commit"
	"git-cx/internal/config"
	"git-cx/internal/git"
)

// CommitService coordinates commit flow.
type CommitService struct {
	cfg      *config.Config
	provider ai.Provider
	git      git.Runner
}

// NewCommitService builds a service with dependencies.
func NewCommitService(cfg *config.Config, provider ai.Provider, gitRunner git.Runner) *CommitService {
	return &CommitService{cfg: cfg, provider: provider, git: gitRunner}
}

// StagedChanges returns staged diff and stat.
func (s *CommitService) StagedChanges(ctx context.Context) (string, string, error) {
	diff, err := s.git.StagedDiff(ctx)
	if err != nil {
		return "", "", err
	}
	stat, err := s.git.StagedStat(ctx)
	if err != nil {
		return "", "", err
	}
	return diff, stat, nil
}

// GenerateCandidates generates commit message candidates.
func (s *CommitService) GenerateCandidates(ctx context.Context, diff, stat, commitType, scope string) ([]string, error) {
	req := ai.GenerateRequest{
		Diff:       diff,
		Stat:       stat,
		CommitType: commitType,
		Scope:      scope,
		Candidates: s.cfg.Candidates,
	}
	return s.provider.Generate(ctx, req)
}

// BuildMessage formats commit message.
func (s *CommitService) BuildMessage(c *commit.ConventionalCommit) string {
	return commit.BuildMessage(c, s.cfg.Commit.UseEmoji, s.cfg.Commit.MaxSubjectLength)
}

// Commit executes git commit.
func (s *CommitService) Commit(ctx context.Context, message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message is empty")
	}
	if err := s.git.Commit(ctx, message); err != nil {
		return err
	}
	return nil
}
