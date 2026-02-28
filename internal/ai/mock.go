package ai

import "context"

// MockProvider is a test double for AI providers.
type MockProvider struct {
	NameValue  string
	Candidates []string
	Err        error
	LastReq    *GenerateRequest
}

func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	_ = ctx
	m.LastReq = &req
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Candidates, nil
}

func (m *MockProvider) Name() string {
	if m.NameValue != "" {
		return m.NameValue
	}
	return "mock"
}
