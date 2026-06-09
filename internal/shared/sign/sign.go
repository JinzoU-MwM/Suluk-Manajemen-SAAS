// Package sign provides small HMAC-SHA256 tokens for tamper-proof links shared
// between services (e.g. the subscription-invoice PDF download link embedded in
// an email). Both the producing and verifying service use the same secret
// (INTERNAL_API_KEY), so a token minted by one verifies in the other.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Token returns hex(HMAC-SHA256(key, msg)).
func Token(msg, key string) string {
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}

// Valid reports whether sig is the correct token for msg under key, using a
// constant-time comparison. Returns false when key or sig is empty.
func Valid(msg, key, sig string) bool {
	if key == "" || sig == "" {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(Token(msg, key)))
}
