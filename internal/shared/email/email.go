// Package email sends transactional email via SMTP (preferred) or Resend.
// It is best-effort: when nothing is configured Send is a no-op so callers
// (e.g. payment activation) are never blocked by email delivery.
package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"io"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"regexp"
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
	// Validate the recipient and strip any CRLF from header inputs to prevent
	// SMTP header injection (a "victim@x.com\r\nBcc: ..." address could otherwise
	// inject arbitrary headers/body). Normalize `to` to its bare address.
	addr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	to = addr.Address
	subject = stripCRLF(subject)
	if c.cfg.SMTPHost != "" {
		return c.sendSMTP(to, subject, html)
	}
	return c.sendResend(ctx, to, subject, html)
}

// stripCRLF removes carriage returns and newlines so a value is safe to place
// in a single email header line.
func stripCRLF(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

// fromAddress extracts the bare email from the configured From header.
func (c *Client) fromAddress() string {
	if a, err := mail.ParseAddress(c.cfg.From); err == nil {
		return a.Address
	}
	return c.cfg.From
}

// buildMessage assembles an RFC-5322 multipart/alternative message (plain text +
// HTML) with Date and Message-ID headers and a MIME-encoded subject. HTML-only
// messages without these headers score as spam and are rejected by content
// filters (e.g. MailChannels "[CS] Message blocked"); a text part + real
// headers gets the same content delivered.
func (c *Client) buildMessage(to, subject, html string) ([]byte, error) {
	domain := "localhost"
	if at := strings.LastIndex(c.fromAddress(), "@"); at >= 0 {
		domain = c.fromAddress()[at+1:]
	}
	boundary, err := randHex(16)
	if err != nil {
		return nil, fmt.Errorf("generate MIME boundary: %w", err)
	}
	msgID, err := randHex(16)
	if err != nil {
		return nil, fmt.Errorf("generate message-id: %w", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", c.cfg.From)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", encodeSubject(subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", msgID, domain)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, htmlToText(html))
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, html)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return b.Bytes(), nil
}

// randHex returns n cryptographically-random bytes hex-encoded. It errors on a
// rand source failure rather than silently returning a zero (predictable)
// value, which would yield a fixed MIME boundary/Message-ID.
func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// encodeSubject MIME-encodes a subject only when it contains non-ASCII bytes.
func encodeSubject(s string) string {
	for _, r := range s {
		if r > 127 {
			return mime.QEncoding.Encode("UTF-8", s)
		}
	}
	return s
}

var (
	reStyleScript = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	reBreak       = regexp.MustCompile(`(?i)<br\s*/?>`)
	reBlockEnd    = regexp.MustCompile(`(?i)</(p|div|tr|h1|h2|h3|table|li)>`)
	reTag         = regexp.MustCompile(`(?s)<[^>]*>`)
	reSpaces      = regexp.MustCompile(`[ \t]+`)
	reBlankLines  = regexp.MustCompile(`\n{3,}`)
)

// htmlToText derives a readable plain-text alternative from an HTML body.
func htmlToText(h string) string {
	h = reStyleScript.ReplaceAllString(h, "")
	h = reBreak.ReplaceAllString(h, "\n")
	h = reBlockEnd.ReplaceAllString(h, "\n")
	h = reTag.ReplaceAllString(h, "")
	h = stdhtml.UnescapeString(h)
	h = reSpaces.ReplaceAllString(h, " ")
	lines := strings.Split(h, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	h = strings.Join(lines, "\n")
	h = reBlankLines.ReplaceAllString(h, "\n\n")
	return strings.TrimSpace(h)
}

func (c *Client) sendSMTP(to, subject, html string) error {
	host := c.cfg.SMTPHost
	port := c.cfg.SMTPPort
	if port == 0 {
		port = 465
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	from := c.fromAddress()

	msg, err := c.buildMessage(to, subject, html)
	if err != nil {
		return fmt.Errorf("build message: %w", err)
	}

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
		"from": c.cfg.From, "to": []string{to}, "subject": subject,
		"html": html, "text": htmlToText(html),
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
