package world

import (
	"testing"
	"time"
)

func TestCurrentArenaStatusDailySchedule(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)

	testCases := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "signup open before nine", now: time.Date(2026, time.April, 1, 8, 59, 0, 0, loc), want: "signup_open"},
		{name: "seeding after nine", now: time.Date(2026, time.April, 1, 9, 2, 0, 0, loc), want: "signup_locked"},
		{name: "main bracket in progress", now: time.Date(2026, time.April, 1, 9, 7, 0, 0, loc), want: "in_progress"},
		{name: "results live after finals", now: time.Date(2026, time.April, 1, 9, 36, 0, 0, loc), want: "results_live"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := currentArenaStatus(tc.now)
			if got.Code != tc.want {
				t.Fatalf("expected status %q, got %q", tc.want, got.Code)
			}
		})
	}
}
