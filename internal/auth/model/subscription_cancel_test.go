package model

import (
	"testing"
	"time"
)

func TestCanCancelSubscription(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		sub  *Subscription
		ok   bool
	}{
		{"nil sub", nil, false},
		{"active future", &Subscription{Status: "active", ExpiresAt: &future}, true},
		{"active no expiry", &Subscription{Status: "active", ExpiresAt: nil}, false},
		{"active expired", &Subscription{Status: "active", ExpiresAt: &past}, false},
		{"trial not cancelable", &Subscription{Status: "trial", ExpiresAt: &future}, false},
		{"expired status", &Subscription{Status: "expired", ExpiresAt: &future}, false},
	}
	for _, c := range cases {
		err := CanCancelSubscription(c.sub, now)
		if c.ok && err != nil {
			t.Fatalf("%s: want ok, got %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%s: want error, got nil", c.name)
		}
	}
}

func TestCanResumeSubscription(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		sub  *Subscription
		ok   bool
	}{
		{"nil sub", nil, false},
		{"flagged future", &Subscription{ExpiresAt: &future, CancelAtPeriodEnd: true}, true},
		{"not flagged", &Subscription{ExpiresAt: &future, CancelAtPeriodEnd: false}, false},
		{"flagged but expired", &Subscription{ExpiresAt: &past, CancelAtPeriodEnd: true}, false},
		{"flagged no expiry", &Subscription{ExpiresAt: nil, CancelAtPeriodEnd: true}, false},
	}
	for _, c := range cases {
		err := CanResumeSubscription(c.sub, now)
		if c.ok && err != nil {
			t.Fatalf("%s: want ok, got %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%s: want error, got nil", c.name)
		}
	}
}
