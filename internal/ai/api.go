package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git-cx/internal/config"
)

// APIProvider calls an OpenAI-compatible API endpoint.
type APIProvider struct {
	baseURL    string
	apiKey     string
	model      string
	candidates int
	timeout    int
}

// NewAPIProvider creates an APIProvider from config.
func NewAPIProvider(cfg *config.Config) (*APIProvider, error) {
	return &APIProvider{
		baseURL:    cfg.API.BaseURL,
		apiKey:     cfg.API.Key,
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
	}, nil
}

func (p *APIProvider) Name() string { return "api" }

type apiRequest struct {
	Model    string       `json:"model"`
	Messages []apiMessage `json:"messages"`
	N        int          `json:"n,omitempty"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *APIProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	if strings.TrimSpace(p.baseURL) == "" {
		return nil, fmt.Errorf("api base URL is not set (cx.api.baseUrl) for api provider")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	prompt := buildPrompt(req)
	requestBody := apiRequest{
		Model: p.model,
		Messages: []apiMessage{
			{Role: "user", Content: prompt},
		},
	}
	if p.candidates > 1 {
		requestBody.N = p.candidates
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	endpoint, err := joinURL(p.baseURL, "/chat/completions")
	if err != nil {
		return nil, fmt.Errorf("invalid api base URL: %w", err)
	}
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(p.apiKey) != "" {
		reqHTTP.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var decoded apiResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(resp.Status)
		if decoded.Error != nil && strings.TrimSpace(decoded.Error.Message) != "" {
			msg = decoded.Error.Message
		}
		return nil, fmt.Errorf("api request failed: %s", msg)
	}

	var candidates []string
	for _, choice := range decoded.Choices {
		candidates = append(candidates, parseOutput(choice.Message.Content, p.candidates)...)
		if p.candidates > 0 && len(candidates) >= p.candidates {
			break
		}
	}
	if p.candidates > 0 && len(candidates) > p.candidates {
		candidates = candidates[:p.candidates]
	}
	return candidates, nil
}

func joinURL(baseURL, path string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("base URL must include scheme and host")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed.String(), nil
}
