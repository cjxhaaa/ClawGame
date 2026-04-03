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
		{name: "weekday rating season", now: time.Date(2026, time.April, 1, 8, 59, 0, 0, loc), want: "rating_open"},
		{name: "saturday knockout pending", now: time.Date(2026, time.April, 4, 8, 59, 0, 0, loc), want: "knockout_pending"},
		{name: "saturday bracket in progress", now: time.Date(2026, time.April, 4, 9, 7, 0, 0, loc), want: "knockout_in_progress"},
		{name: "saturday results live after finals", now: time.Date(2026, time.April, 4, 9, 36, 0, 0, loc), want: "knockout_results_live"},
		{name: "sunday rest day", now: time.Date(2026, time.April, 5, 10, 0, 0, 0, loc), want: "rest_day"},
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
