package timeutil

import "time"

const (
	SlotMinutes   = 30
	MinBookingMin = 30
	MaxBookingMin = 180
	CleaningMin   = 30
)

func IsHalfHour(t time.Time) bool {
	return t.Second() == 0 && t.Nanosecond() == 0 && t.Minute()%SlotMinutes == 0
}

func IsValidDuration(minutes int) bool {
	return minutes >= MinBookingMin && minutes <= MaxBookingMin && minutes%SlotMinutes == 0
}

func OccupiedEnd(end time.Time) time.Time {
	return end.Add(CleaningMin * time.Minute)
}
