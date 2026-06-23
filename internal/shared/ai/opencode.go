package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type openCodeGenerator struct {
	apiKey  string
	model   string
	baseURL string
	httpc   *http.Client
}

func newOpenCode(apiKey, model, baseURL string) *openCodeGenerator {
	if model == "" {
		model = "gpt-5-nano"
	}
	if baseURL == "" {
		baseURL = "https://opencode.ai/zen/v1"
	}
	return &openCodeGenerator{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *openCodeGenerator) Available() bool { return g != nil && g.apiKey != "" }

type ocReq struct {
	Model       string      `json:"model"`
	Messages    []ocMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

type ocMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string (text) or []any (vision parts)
}

type ocResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *openCodeGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	if !g.Available() {
		return "", fmt.Errorf("opencode generator not configured")
	}
	body, _ := json.Marshal(ocReq{
		Model:       g.model,
		Messages:    []ocMessage{{Role: "user", Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   1024,
	})
	raw, err := ocPost(ctx, g.httpc, g.baseURL+"/chat/completions", g.apiKey, body)
	if err != nil {
		return "", err
	}
	var r ocResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", fmt.Errorf("parse opencode response: %w", err)
	}
	if r.Error != nil {
		return "", fmt.Errorf("opencode api error: %s", r.Error.Message)
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("opencode returned no choices")
	}
	return r.Choices[0].Message.Content, nil
}

// ocPost POSTs JSON with Bearer auth and retries on 429 / >=500 / network error
// (max 3 attempts, capped exponential backoff). Returns the final response body.
func ocPost(ctx context.Context, httpc *http.Client, url, apiKey string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := httpc.Do(req)
		if err != nil {
			lastErr = err
			ocBackoff(ctx, attempt)
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("opencode api status %d", resp.StatusCode)
			ocBackoff(ctx, attempt)
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("opencode api status %d: %s", resp.StatusCode, string(raw))
		}
		return raw, nil
	}
	return nil, fmt.Errorf("opencode api failed after retries: %w", lastErr)
}

func ocBackoff(ctx context.Context, attempt int) {
	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > 10*time.Second {
		d = 10 * time.Second
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
