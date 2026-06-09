// Package email is a tiny Resend (https://resend.com) transactional-email client.
// It is intentionally best-effort: when no API key is configured Send is a no-op
// so callers (e.g. payment activation) are never blocked by email delivery.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

type Client struct {
	apiKey string
	from   string
	http   *http.Client
}

// New builds a client. from is the RFC-5322 sender, e.g. `Suluk <billing@suluk.id>`.
func New(apiKey, from string) *Client {
	if from == "" {
		from = "Suluk <onboarding@resend.dev>"
	}
	return &Client{
		apiKey: apiKey,
		from:   from,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

// Enabled reports whether an API key is configured.
func (c *Client) Enabled() bool { return c != nil && c.apiKey != "" }

// Send delivers an HTML email. It is a no-op (returns nil) when not Enabled, so
// it never blocks a critical path; callers should still log a returned error.
func (c *Client) Send(ctx context.Context, to, subject, html string) error {
	if !c.Enabled() {
		return nil
	}
	payload, err := json.Marshal(map[string]any{
		"from":    c.from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
