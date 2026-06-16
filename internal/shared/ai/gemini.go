// Package ai is a thin, text-only Gemini client shared by services that want a
// natural-language generation step (e.g. the accounting copilot). It mirrors the
// HTTP shape used by ai-ocr-service but without the vision/OCR specifics.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// defaultModel is the Gemini model used when GEMINI_MODEL is unset. We default to
// 2.5-flash because gemini-2.0-flash has a free-tier quota of 0 on newer keys.
const defaultModel = "gemini-2.5-flash"

type Client struct {
	apiKey string
	model  string
	httpc  *http.Client
}

// New returns a client, or nil when no key is configured. A nil client is safe:
// callers should treat (nil → AI unavailable) as a graceful, non-fatal state.
func New(apiKey string) *Client {
	if apiKey == "" {
		return nil
	}
	model := defaultModel
	if m := os.Getenv("GEMINI_MODEL"); m != "" {
		model = m
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpc:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Available reports whether a usable client is configured.
func (c *Client) Available() bool { return c != nil && c.apiKey != "" }

type genReq struct {
	Contents []content `json:"contents"`
	Config   *genCfg   `json:"generationConfig,omitempty"`
}
type content struct {
	Parts []part `json:"parts"`
}
type part struct {
	Text string `json:"text"`
}
type genCfg struct {
	Temperature     float64   `json:"temperature"`
	MaxOutputTokens int       `json:"maxOutputTokens"`
	Thinking        *thinkCfg `json:"thinkingConfig,omitempty"`
}

// thinkCfg disables the 2.5-series "thinking" pass. Without this, thinking tokens
// consume the maxOutputTokens budget and the visible answer is truncated mid-text.
type thinkCfg struct {
	ThinkingBudget int `json:"thinkingBudget"`
}
type genResp struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GenerateText sends a single prompt and returns the model's text reply.
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if !c.Available() {
		return "", fmt.Errorf("ai client not configured")
	}
	body, _ := json.Marshal(genReq{
		Contents: []content{{Parts: []part{{Text: prompt}}}},
		Config:   &genCfg{Temperature: 0.3, MaxOutputTokens: 1024, Thinking: &thinkCfg{ThinkingBudget: 0}},
	})
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.model)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini api call: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var gr genResp
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("parse gemini response: %w", err)
	}
	if gr.Error != nil {
		return "", fmt.Errorf("gemini api error (%s): %s", gr.Error.Status, gr.Error.Message)
	}
	if resp.StatusCode != 200 || len(gr.Candidates) == 0 {
		return "", fmt.Errorf("gemini api returned status %d", resp.StatusCode)
	}
	text := ""
	for _, p := range gr.Candidates[0].Content.Parts {
		text += p.Text
	}
	return text, nil
}
