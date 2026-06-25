package service

import (
	"reflect"
	"testing"
)

func TestDueReminder(t *testing.T) {
	cases := []struct {
		name      string
		daysLeft  float64
		sent      []int
		wantEmail int
		wantMark  []int
		wantOk    bool
	}{
		{"H-7 first", 7, nil, 7, []int{7}, true},
		{"H-3 after 7 sent", 3, []int{7}, 3, []int{3, 7}, true},
		{"H-1 after 7,3 sent", 1, []int{7, 3}, 1, []int{1, 3, 7}, true},
		{"between 7 and 3, already sent 7", 5, []int{7}, 0, nil, false},
		{"downtime catch-up at 2 days, none sent", 2, nil, 3, []int{3, 7}, true},
		{"all sent", 0.5, []int{1, 3, 7}, 0, nil, false},
		{"far out, nothing applicable", 10, nil, 0, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			email, mark, ok := dueReminder(c.daysLeft, c.sent)
			if email != c.wantEmail || ok != c.wantOk || !reflect.DeepEqual(mark, c.wantMark) {
				t.Errorf("dueReminder(%v,%v) = (%d,%v,%v), want (%d,%v,%v)",
					c.daysLeft, c.sent, email, mark, ok, c.wantEmail, c.wantMark, c.wantOk)
			}
		})
	}
}
