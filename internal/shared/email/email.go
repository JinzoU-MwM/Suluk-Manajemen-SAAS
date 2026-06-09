// Package email sends transactional email via SMTP (preferred) or Resend.
// It is best-effort: when nothing is configured Send is a no-op so callers
// (e.g. payment activation) are never blocked by email delivery.
package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

// Config selects the transport. If SMTPHost is set, SMTP is used; otherwise if
// ResendAPIKey is set, Resend is used.
type Config struct {
	From string // RFC-5322 sender, e.g. `Suluk <admin@suluk.site>`

	// SMTP
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string

	// Resend (fallback)
	ResendAPIKey string
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) *Client {
	if cfg.From == "" {
		cfg.From = "Suluk <onboarding@resend.dev>"
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: 15 * time.Second}}
}

// Enabled reports whether any transport is configured.
func (c *Client) Enabled() bool {
	return c != nil && (c.cfg.SMTPHost != "" || c.cfg.ResendAPIKey != "")
}

// Send delivers an HTML email. No-op (nil) when not Enabled.
func (c *Client) Send(ctx context.Context, to, subject, html string) error {
	if !c.Enabled() {
		return nil
	}
	if c.cfg.SMTPHost != "" {
		return c.sendSMTP(to, subject, html)
	}
	return c.sendResend(ctx, to, subject, html)
}

// fromAddress extracts the bare email from the configured From header.
func (c *Client) fromAddress() string {
	if a, err := mail.ParseAddress(c.cfg.From); err == nil {
		return a.Address
	}
	return c.cfg.From
}

func (c *Client) sendSMTP(to, subject, html string) error {
	host := c.cfg.SMTPHost
	port := c.cfg.SMTPPort
	if port == 0 {
		port = 465
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	from := c.fromAddress()

	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", c.cfg.From)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(html)
	msg := b.Bytes()

	auth := smtp.PlainAuth("", c.cfg.SMTPUser, c.cfg.SMTPPass, host)

	// Port 465 = implicit TLS (SMTPS): dial TLS first, then speak SMTP.
	if port == 465 {
		dialer := &net.Dialer{Timeout: 15 * time.Second}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("smtp mail: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := w.Write(msg); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return client.Quit()
	}

	// Port 587/25 = STARTTLS handled by net/smtp.SendMail.
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func (c *Client) sendResend(ctx context.Context, to, subject, html string) error {
	payload, err := json.Marshal(map[string]any{
		"from": c.cfg.From, "to": []string{to}, "subject": subject, "html": html,
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
	req.Header.Set("Authorization", "Bearer "+c.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
