package model

import (
	"errors"
	"time"
)

// ErrNothingToCancel is returned when there is no paid, active, unexpired
// subscription to flag cancel-at-period-end. Handlers map it to 400.
var ErrNothingToCancel = errors.New("no active subscription to cancel")

// ErrAlreadyCanceled is returned when the subscription is already flagged
// cancel-at-period-end, so there is nothing more to cancel. Handlers map it to 400.
var ErrAlreadyCanceled = errors.New("subscription is already set to cancel")

// ErrNothingToResume is returned when there is no pending cancel to undo.
var ErrNothingToResume = errors.New("no cancellable subscription to resume")

// CanCancelSubscription reports whether the org's subscription may be set to
// cancel-at-period-end: it must be a paid, active sub that has not expired and is
// not already flagged. Trials lapse on their own and are intentionally not cancelable.
func CanCancelSubscription(sub *Subscription, now time.Time) error {
	if sub == nil || sub.Status != "active" || sub.ExpiresAt == nil || !sub.ExpiresAt.After(now) {
		return ErrNothingToCancel
	}
	if sub.CancelAtPeriodEnd {
		return ErrAlreadyCanceled
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
