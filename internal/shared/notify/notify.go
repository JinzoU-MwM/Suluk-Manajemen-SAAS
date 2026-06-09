// Package notify is a tiny best-effort client for pushing in-app notifications
// to auth-service's internal endpoint. All sends are fire-and-forget: a failure
// (or missing config) never affects the caller's primary operation.
package notify

import (
	"context"

	"github.com/jamaah-in/v2/internal/shared/httpclient"
)

type Client struct {
	httpc    *httpclient.Client
	authAddr string
	key      string
}

// New builds a notify client. authAddr is the auth-service address (host:port)
// and key is the shared INTERNAL_API_KEY.
func New(authAddr, key string) *Client {
	return &Client{httpc: httpclient.New(), authAddr: authAddr, key: key}
}

// Send pushes a notification. orgID is required; userID may be "" to broadcast
// to the whole org. Errors are intentionally swallowed (best-effort).
func (c *Client) Send(ctx context.Context, orgID, userID, severity, title, message string) {
	if c == nil || c.authAddr == "" || c.key == "" || orgID == "" {
		return
	}
	body := map[string]any{
		"org_id":   orgID,
		"user_id":  userID,
		"severity": severity,
		"title":    title,
		"message":  message,
	}
	_ = c.httpc.PostJSON(ctx, c.authAddr, "/api/v1/internal/notifications",
		map[string]string{"X-Internal-Key": c.key}, body, nil)
}
