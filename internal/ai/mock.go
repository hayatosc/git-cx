package ai

import "context"

// MockProvider is a test double for AI providers.
type MockProvider struct {
	NameValue  string
	Candidates []string
	Body       string
	Footer     string
	Err        error
	LastReq    *GenerateRequest
	LastDetail *GenerateRequest
}

func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	_ = ctx
	m.LastReq = &req
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Candidates, nil
}

func (m *MockProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	_ = ctx
	m.LastDetail = &req
	if m.Err != nil {
		return "", "", m.Err
	}
	return m.Body, m.Footer, nil
}

func (m *MockProvider) Name() string {
	if m.NameValue != "" {
		return m.NameValue
	}
	return "mock"
}
