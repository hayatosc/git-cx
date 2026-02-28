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

	"github.com/hayatosc/git-cx/internal/config"
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
func NewAPIProvider(cfg *config.Config) *APIProvider {
	return &APIProvider{
		baseURL:    cfg.API.BaseURL,
		apiKey:     cfg.API.Key,
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
	}
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
		return nil, fmt.Errorf("api base URL is not set (cx.apiBaseUrl) for api provider")
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

	decoded, err := p.request(ctx, requestBody)
	if err != nil {
		return nil, err
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

func (p *APIProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	if strings.TrimSpace(p.baseURL) == "" {
		return "", "", fmt.Errorf("api base URL is not set (cx.apiBaseUrl) for api provider")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	prompt := buildDetailPrompt(req)
	requestBody := apiRequest{
		Model: p.model,
		Messages: []apiMessage{
			{Role: "user", Content: prompt},
		},
	}

	decoded, err := p.request(ctx, requestBody)
	if err != nil {
		return "", "", err
	}
	if len(decoded.Choices) == 0 {
		return "", "", fmt.Errorf("api response missing choices")
	}
	body, footer := parseDetailOutput(decoded.Choices[0].Message.Content)
	return body, footer, nil
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

func (p *APIProvider) request(ctx context.Context, requestBody apiRequest) (apiResponse, error) {
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return apiResponse{}, fmt.Errorf("failed to encode request: %w", err)
	}

	endpoint, err := joinURL(p.baseURL, "/chat/completions")
	if err != nil {
		return apiResponse{}, fmt.Errorf("invalid api base URL: %w", err)
	}
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return apiResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(p.apiKey) != "" {
		reqHTTP.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return apiResponse{}, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(resp.Status)
		var decodedErr apiResponse
		if err := json.Unmarshal(data, &decodedErr); err == nil {
			if decodedErr.Error != nil && strings.TrimSpace(decodedErr.Error.Message) != "" {
				msg = decodedErr.Error.Message
			}
			return apiResponse{}, fmt.Errorf("api request failed: %s", msg)
		}
		raw := strings.TrimSpace(string(data))
		const maxErrorBodyLen = 512
		if len(raw) > maxErrorBodyLen {
			raw = raw[:maxErrorBodyLen] + "..."
		}
		return apiResponse{}, fmt.Errorf("api request failed: status %s, could not parse error body as JSON: %s", resp.Status, raw)
	}
	var decoded apiResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return apiResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}
	return decoded, nil
}
