package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/commit"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/git"
)

// CommitService coordinates commit flow.
type CommitService struct {
	cfg           *config.Config
	provider      ai.Provider
	providerName  string
	providers     map[string]ai.Provider
	providerOrder []string
	git           git.Runner
}

// NewCommitService builds a service with dependencies.
func NewCommitService(cfg *config.Config, providers map[string]ai.Provider, currentProvider string, gitRunner git.Runner) *CommitService {
	order := make([]string, 0, len(cfg.Providers))
	for _, name := range cfg.Providers {
		if _, ok := providers[name]; ok {
			order = append(order, name)
		}
	}
	if len(order) == 0 && currentProvider != "" {
		order = append(order, currentProvider)
	}
	if currentProvider == "" && len(order) > 0 {
		currentProvider = order[0]
	}
	provider := providers[currentProvider]
	return &CommitService{
		cfg:           cfg,
		provider:      provider,
		providerName:  currentProvider,
		providers:     providers,
		providerOrder: order,
		git:           gitRunner,
	}
}

// ProviderNames returns configured provider names in preference order.
func (s *CommitService) ProviderNames() []string {
	out := make([]string, 0, len(s.providerOrder))
	out = append(out, s.providerOrder...)
	if len(out) == 0 {
		for name := range s.providers {
			out = append(out, name)
		}
	}
	return out
}

// CurrentProvider returns the active provider name.
func (s *CommitService) CurrentProvider() string {
	return s.providerName
}

// UseProvider switches the active provider.
func (s *CommitService) UseProvider(name string) error {
	p, ok := s.providers[name]
	if !ok {
		return fmt.Errorf("unknown provider: %s", name)
	}
	s.provider = p
	s.providerName = name
	return nil
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
	if s.provider == nil {
		return nil, fmt.Errorf("provider %q is not configured", s.providerName)
	}
	return s.provider.Generate(ctx, req)
}

// GenerateDetails generates commit body and footer.
func (s *CommitService) GenerateDetails(ctx context.Context, diff, stat, commitType, scope, subject string) (string, string, error) {
	req := ai.GenerateRequest{
		Diff:       diff,
		Stat:       stat,
		CommitType: commitType,
		Scope:      scope,
		Subject:    subject,
		Candidates: 1,
	}
	if s.provider == nil {
		return "", "", fmt.Errorf("provider %q is not configured", s.providerName)
	}
	return s.provider.GenerateDetail(ctx, req)
}

// BuildMessage formats commit message.
func (s *CommitService) BuildMessage(c *commit.ConventionalCommit) string {
	return commit.BuildMessage(c, s.cfg.Commit.UseEmoji, s.cfg.Commit.MaxSubjectLength)
}

// Commit executes git commit.
func (s *CommitService) Commit(ctx context.Context, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", errors.New("commit message is empty")
	}
	return s.git.Commit(ctx, message)
}
