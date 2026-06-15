package model

import "testing"

func TestCanTransitionDeparture(t *testing.T) {
	cases := []struct {
		from, to DepartureStatus
		want     bool
	}{
		{DepartureDraft, DepartureSiap, true},
		{DepartureDraft, DepartureBatal, true},
		{DepartureSiap, DepartureBerangkat, true},
		{DepartureSiap, DepartureDraft, true}, // reopen
		{DepartureBerangkat, DepartureSelesai, true},
		{DepartureBatal, DepartureDraft, true},     // reactivate
		{DepartureDraft, DepartureBerangkat, false}, // can't skip siap
		{DepartureBerangkat, DepartureBatal, false}, // can't cancel a departed kloter
		{DepartureSelesai, DepartureDraft, false},   // terminal
		{DepartureSiap, DepartureSiap, false},       // no self-transition
	}
	for _, c := range cases {
		if got := CanTransitionDeparture(c.from, c.to); got != c.want {
			t.Errorf("CanTransitionDeparture(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
