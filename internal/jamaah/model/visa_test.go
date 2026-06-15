package model

import "testing"

func TestCanTransitionVisa(t *testing.T) {
	cases := []struct {
		from, to VisaStatus
		want     bool
	}{
		{VisaDraft, VisaSubmitted, true},
		{VisaSubmitted, VisaApproved, true},
		{VisaSubmitted, VisaRejected, true},
		{VisaRejected, VisaSubmitted, true},  // resubmit
		{VisaApproved, VisaExpired, true},
		{VisaExpired, VisaSubmitted, true},    // renew
		{VisaDraft, VisaApproved, false},      // can't skip submit
		{VisaApproved, VisaRejected, false},   // decided is final (except expiry)
		{VisaDraft, VisaDraft, false},         // no self-transition
		{VisaSubmitted, VisaDraft, false},     // no going back to draft
	}
	for _, c := range cases {
		if got := CanTransitionVisa(c.from, c.to); got != c.want {
			t.Errorf("CanTransitionVisa(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
