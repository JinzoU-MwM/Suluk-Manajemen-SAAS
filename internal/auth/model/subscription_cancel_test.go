package model

import (
	"errors"
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
		{"active future already cancelled", &Subscription{Status: "active", ExpiresAt: &future, CancelAtPeriodEnd: true}, false},
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

// An already-cancelled (still active, unexpired) sub must be rejected with a
// distinct error, not the generic "nothing to cancel" -- so the API can tell the
// user it is already set to cancel rather than implying no paid sub exists.
func TestCanCancelSubscriptionAlreadyCanceledIsDistinct(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	sub := &Subscription{Status: "active", ExpiresAt: &future, CancelAtPeriodEnd: true}

	err := CanCancelSubscription(sub, now)
	if err == nil {
		t.Fatal("want error for already-cancelled sub, got nil")
	}
	if errors.Is(err, ErrNothingToCancel) {
		t.Fatal("want a distinct already-cancelled error, got ErrNothingToCancel")
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
