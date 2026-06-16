package service

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

// narrativeCache memoizes AI narratives so we don't call Gemini on every request.
// The cache key is a fingerprint of (org, prompt); because the prompt is fully
// derived from the period dates + financial metrics + anomalies, an unchanged GL
// reuses the narrative and any change in the numbers (or the day rolling over)
// naturally produces a new key and regenerates. The TTL is a safety bound so
// entries don't live indefinitely; max caps memory.
type narrativeCache struct {
	mu  sync.Mutex
	m   map[string]narrativeEntry
	ttl time.Duration
	max int
}

type narrativeEntry struct {
	text string
	at   time.Time
}

func newNarrativeCache(ttl time.Duration, max int) *narrativeCache {
	return &narrativeCache{m: make(map[string]narrativeEntry), ttl: ttl, max: max}
}

// narrativeKey fingerprints (org, prompt) into a stable cache key.
func narrativeKey(orgID uuid.UUID, prompt string) string {
	h := sha256.New()
	h.Write(orgID[:])
	h.Write([]byte{0})
	h.Write([]byte(prompt))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *narrativeCache) get(key string, now time.Time) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok {
		return "", false
	}
	if now.Sub(e.at) > c.ttl {
		delete(c.m, key)
		return "", false
	}
	return e.text, true
}

func (c *narrativeCache) put(key, text string, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.m) >= c.max {
		// Evict expired entries first; if still at capacity, reset wholesale.
		// Both paths are rare (max is generous) so a simple strategy is fine.
		for k, e := range c.m {
			if now.Sub(e.at) > c.ttl {
				delete(c.m, k)
			}
		}
		if len(c.m) >= c.max {
			c.m = make(map[string]narrativeEntry, c.max)
		}
	}
	c.m[key] = narrativeEntry{text: text, at: now}
}
