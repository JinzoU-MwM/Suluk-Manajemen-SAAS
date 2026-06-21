package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNarrativeCache(t *testing.T) {
	c := newNarrativeCache(time.Hour, 4)
	t0 := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	org := uuid.New()
	k := narrativeKey(org, "prompt-A")

	// miss before put
	if _, ok := c.get(k, t0); ok {
		t.Fatal("expected miss on empty cache")
	}

	// hit within TTL
	c.put(k, "narasi", t0)
	if v, ok := c.get(k, t0.Add(30*time.Minute)); !ok || v != "narasi" {
		t.Fatalf("expected hit, got %q ok=%v", v, ok)
	}

	// expired after TTL
	if _, ok := c.get(k, t0.Add(2*time.Hour)); ok {
		t.Fatal("expected miss after TTL")
	}
}

func TestNarrativeKeyDistinct(t *testing.T) {
	o1, o2 := uuid.New(), uuid.New()
	if narrativeKey(o1, "p") == narrativeKey(o2, "p") {
		t.Fatal("same prompt across different orgs must not collide")
	}
	if narrativeKey(o1, "p1") == narrativeKey(o1, "p2") {
		t.Fatal("different prompts must produce different keys")
	}
	k1, k2 := narrativeKey(o1, "p"), narrativeKey(o1, "p")
	if k1 != k2 {
		t.Fatal("key must be stable for identical inputs")
	}
}

func TestNarrativeCacheBoundsMemory(t *testing.T) {
	c := newNarrativeCache(time.Hour, 4)
	t0 := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	org := uuid.New()
	for i := 0; i < 20; i++ {
		c.put(narrativeKey(org, string(rune('a'+i))), "x", t0)
	}
	c.mu.Lock()
	n := len(c.m)
	c.mu.Unlock()
	if n > 4 {
		t.Fatalf("cache exceeded max: %d", n)
	}
}
