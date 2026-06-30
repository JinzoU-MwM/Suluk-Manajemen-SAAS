package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GoogleClaims are the id_token claims we consume from "Sign in with Google".
// sub/iss/aud/exp come from the embedded RegisteredClaims.
type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	jwt.RegisteredClaims
}

// Google publishes its id_token signing keys (JWKS) here. Keys rotate; we cache
// them briefly and refetch on a cache-miss (rotated kid).
const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

// Google sets `iss` to one of these two forms.
var validGoogleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

type googleKeyCache struct {
	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

var googleKeys = &googleKeyCache{}

func (kc *googleKeyCache) get(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if kc.keys == nil || time.Now().After(kc.expires) {
		if err := kc.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	if key, ok := kc.keys[kid]; ok {
		return key, nil
	}
	// kid not found: keys may have just rotated — refresh once and retry.
	if err := kc.refreshLocked(ctx); err != nil {
		return nil, err
	}
	key, ok := kc.keys[kid]
	if !ok {
		return nil, fmt.Errorf("google jwks: no key for kid %q", kid)
	}
	return key, nil
}

func (kc *googleKeyCache) refreshLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleCertsURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch google jwks: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch google jwks: status %d", resp.StatusCode)
	}

	var doc struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
			Kty string `json:"kty"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("decode google jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 | int(b)
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	if len(keys) == 0 {
		return fmt.Errorf("google jwks: no usable RSA keys")
	}

	kc.keys = keys
	kc.expires = time.Now().Add(time.Hour)
	return nil
}

// VerifyGoogleIDToken verifies a Google "Sign in with Google" id_token against
// Google's published keys and returns its claims. It enforces the RS256
// signature, the audience (must equal clientID), the issuer, and expiry. The
// caller is responsible for the email_verified check and account decisions.
func VerifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*GoogleClaims, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("google login is not configured")
	}

	claims := &GoogleClaims{}
	keyFunc := func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("google id_token: missing kid")
		}
		return googleKeys.get(ctx, kid)
	}

	_, err := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(clientID),
		jwt.WithExpirationRequired(),
	).ParseWithClaims(idToken, claims, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("verify google id_token: %w", err)
	}

	if !validGoogleIssuers[strings.TrimSpace(claims.Issuer)] {
		return nil, fmt.Errorf("google id_token: unexpected issuer %q", claims.Issuer)
	}
	if claims.Email == "" || claims.Subject == "" {
		return nil, fmt.Errorf("google id_token: missing email or subject")
	}
	return claims, nil
}
