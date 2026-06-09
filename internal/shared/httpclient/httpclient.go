// Package httpclient is a small resilient client for service-to-service calls:
// shared timeout, bounded retries on transient errors, and {success,data} unwrap.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	http    *http.Client
	retries int
}

func New() *Client {
	return &Client{
		http:    &http.Client{Timeout: 10 * time.Second},
		retries: 2,
	}
}

// GetRaw GETs addr+path with the given Authorization header, retries transient
// failures (network errors / 5xx), and returns the unwrapped `data` field.
func (cl *Client) GetRaw(ctx context.Context, addr, path, authToken string) (json.RawMessage, error) {
	url := "http://" + addr + path
	var lastErr error
	for attempt := 0; attempt <= cl.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		if authToken != "" {
			req.Header.Set("Authorization", authToken)
		}

		resp, err := cl.http.Do(req)
		if err != nil {
			lastErr = err
			continue // retry transient network errors
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("%s returned status %d", path, resp.StatusCode)
			continue // retry server errors
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("%s returned status %d", path, resp.StatusCode) // don't retry client errors
		}

		var env struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &env); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", path, err)
		}
		if !env.Success {
			return nil, fmt.Errorf("%s returned success=false", path)
		}
		return env.Data, nil
	}
	return nil, lastErr
}

// PostJSON POSTs body (JSON-encoded) to addr+path with the given extra headers,
// retries transient failures (network errors / 5xx), unwraps the `data` field,
// and decodes it into out (out may be nil). Use for service-to-service writes.
func (cl *Client) PostJSON(ctx context.Context, addr, path string, headers map[string]string, body, out any) error {
	url := "http://" + addr + path
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}
	var lastErr error
	for attempt := 0; attempt <= cl.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := cl.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("%s returned status %d", path, resp.StatusCode)
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("%s returned status %d", path, resp.StatusCode)
		}

		if out == nil {
			return nil
		}
		var env struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &env); err != nil {
			return fmt.Errorf("unmarshal %s: %w", path, err)
		}
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("decode %s data: %w", path, err)
		}
		return nil
	}
	return lastErr
}

// GetJSON is GetRaw plus decoding the unwrapped data into out.
func (cl *Client) GetJSON(ctx context.Context, addr, path, authToken string, out any) error {
	data, err := cl.GetRaw(ctx, addr, path, authToken)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s data: %w", path, err)
	}
	return nil
}
