package core

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// LogicalDate is the study day a moment belongs to: the day rolls over at
// rolloverHour, so 01:30 still counts for the day before.
func LogicalDate(now time.Time, rolloverHour int) string {
	return now.Add(-time.Duration(rolloverHour) * time.Hour).Format(dateLayout)
}

// ParseDate reads a "2006-01-02" date as local midnight.
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, time.Local)
}

func mustDate(s string) time.Time {
	t, err := ParseDate(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// AddDays shifts a date by n calendar days.
func AddDays(date string, n int) string {
	t := mustDate(date)
	return time.Date(t.Year(), t.Month(), t.Day()+n, 0, 0, 0, 0, time.Local).Format(dateLayout)
}

// DaysBetween is the number of calendar days from a to b (b - a).
func DaysBetween(a, b string) int {
	ta, tb := mustDate(a), mustDate(b)
	ua := time.Date(ta.Year(), ta.Month(), ta.Day(), 0, 0, 0, 0, time.UTC)
	ub := time.Date(tb.Year(), tb.Month(), tb.Day(), 0, 0, 0, 0, time.UTC)
	return int(ub.Sub(ua).Hours() / 24)
}

// MinuteOfDay is the number of minutes after local midnight.
func MinuteOfDay(t time.Time) int { return t.Hour()*60 + t.Minute() }

// AtMinute is the local time m minutes after midnight on date.
func AtMinute(date string, m int) time.Time {
	t := mustDate(date)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, m, 0, 0, time.Local)
}

// Clock formats minutes after midnight as "20:20". Values outside a day wrap.
func Clock(m int) string {
	m = ((m % 1440) + 1440) % 1440
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

// ClockOf formats a time as "20:20".
func ClockOf(t time.Time) string { return Clock(MinuteOfDay(t)) }

// Duration formats minutes as "40 min", "1 h" or "1 h 41 min".
func Duration(min int) string {
	if min < 0 {
		min = 0
	}
	h, m := min/60, min%60
	switch {
	case h == 0:
		return fmt.Sprintf("%d min", m)
	case m == 0:
		return fmt.Sprintf("%d h", h)
	default:
		return fmt.Sprintf("%d h %d min", h, m)
	}
}

// DateLabel formats a date as "Tue 24 Nov".
func DateLabel(date string) string {
	t := mustDate(date)
	if t.IsZero() {
		return date
	}
	return t.Format("Mon 2 Jan")
}

// Weekday is the short weekday name, "Tue".
func Weekday(date string) string {
	t := mustDate(date)
	if t.IsZero() {
		return ""
	}
	return t.Format("Mon")
}

func ceilDiv(a, b int) int {
	if a <= 0 {
		return 0
	}
	return (a + b - 1) / b
}
