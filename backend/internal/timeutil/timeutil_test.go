package timeutil

import (
	"testing"
	"time"
)

func TestBookingTimeRules(t *testing.T) {
	aligned := time.Date(2026, time.August, 29, 8, 30, 0, 0, time.Local)
	if !IsHalfHour(aligned) {
		t.Fatal("expected half-hour aligned time")
	}
	if !IsValidDuration(90) || IsValidDuration(45) || IsValidDuration(210) {
		t.Fatal("duration validation did not enforce 30-minute increments and bounds")
	}
	end := aligned.Add(90 * time.Minute)
	if got := OccupiedEnd(end); !got.Equal(end.Add(30 * time.Minute)) {
		t.Fatal("cleaning buffer was not applied")
	}
}
