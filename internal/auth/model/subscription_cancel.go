package model

import (
	"errors"
	"time"
)

// ErrNothingToCancel is returned when there is no paid, active, unexpired
// subscription to flag cancel-at-period-end. Handlers map it to 400.
var ErrNothingToCancel = errors.New("no active subscription to cancel")

// ErrNothingToResume is returned when there is no pending cancel to undo.
var ErrNothingToResume = errors.New("no cancellable subscription to resume")

// CanCancelSubscription reports whether the org's subscription may be set to
// cancel-at-period-end: it must be a paid, active sub that has not expired.
// Trials lapse on their own and are intentionally not cancelable.
func CanCancelSubscription(sub *Subscription, now time.Time) error {
	if sub == nil || sub.Status != "active" || sub.ExpiresAt == nil || !sub.ExpiresAt.After(now) {
		return ErrNothingToCancel
	}
	return nil
}

// CanResumeSubscription reports whether a pending cancel can be undone: the sub
// must exist, still be unexpired, and currently be flagged cancel-at-period-end.
func CanResumeSubscription(sub *Subscription, now time.Time) error {
	if sub == nil || sub.ExpiresAt == nil || !sub.ExpiresAt.After(now) || !sub.CancelAtPeriodEnd {
		return ErrNothingToResume
	}
	return nil
}
