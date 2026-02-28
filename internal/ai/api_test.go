package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"git-cx/internal/config"
)

func TestAPIProviderGenerate_SendsRequestAndParsesResponse(t *testing.T) {
	var captured apiRequest
	var serverErrors []string
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var localErrors []string
		var localCaptured apiRequest
		if r.Method != http.MethodPost {
			localErrors = append(localErrors, "method")
		}
		if r.URL.Path != "/v1/chat/completions" {
			localErrors = append(localErrors, "path")
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			localErrors = append(localErrors, "auth")
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			localErrors = append(localErrors, "content-type")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			localErrors = append(localErrors, "read-body")
		}
		if err := json.Unmarshal(body, &localCaptured); err != nil {
			localErrors = append(localErrors, "decode-body")
		}
		mu.Lock()
		if len(localErrors) > 0 {
			serverErrors = append(serverErrors, localErrors...)
		} else {
			captured = localCaptured
		}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"feat: one\nfix: two"}}]}`)
	}))
	defer server.Close()

	cfg := &config.Config{
		Model:      "gpt-5",
		Candidates: 2,
		Timeout:    2,
		API: config.APIConfig{
			BaseURL: server.URL + "/v1/",
		},
	}
	provider := NewAPIProvider(cfg)
	provider.apiKey = "test-key"

	got, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 2})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	mu.Lock()
	errorsCopy := append([]string{}, serverErrors...)
	capturedCopy := captured
	mu.Unlock()
	if len(errorsCopy) > 0 {
		t.Fatalf("server received invalid request: %s", strings.Join(errorsCopy, ", "))
	}
	if capturedCopy.Model != "gpt-5" {
		t.Fatalf("unexpected model: %s", capturedCopy.Model)
	}
	if capturedCopy.N != 2 {
		t.Fatalf("unexpected n: %d", capturedCopy.N)
	}
	if len(capturedCopy.Messages) != 1 || capturedCopy.Messages[0].Role != "user" {
		t.Fatalf("unexpected messages: %#v", capturedCopy.Messages)
	}
	if len(got) != 2 || got[0] != "feat: one" || got[1] != "fix: two" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}

func TestAPIProviderGenerate_ReturnsErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"invalid key"}}`)
	}))
	defer server.Close()

	cfg := &config.Config{
		Model:      "gpt-5",
		Candidates: 1,
		Timeout:    2,
		API: config.APIConfig{
			BaseURL: server.URL,
		},
	}
	provider := NewAPIProvider(cfg)
	provider.apiKey = "test-key"

	_, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 1})
	if err == nil || !strings.Contains(err.Error(), "invalid key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
