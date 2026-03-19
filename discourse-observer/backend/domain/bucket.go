// Spec: specs/api/api-contract.md (AC-17, AC-18)
// Tests: backend/domain/trend_test.go
package domain

import (
	"fmt"
	"time"
)

// Granularity returns "daily" for short periods, "weekly" for long ones.
// Periods under 90 days use daily; 90 days and above use weekly.
func Granularity(period string, from, to *time.Time) string {
	if from != nil && to != nil {
		days := to.Sub(*from).Hours() / 24
		if days >= 90 {
			return "weekly"
		}
		return "daily"
	}
	switch period {
	case "7d", "30d":
		return "daily"
	default:
		return "weekly"
	}
}

// BucketKey returns the bucket key for a timestamp at the given granularity.
func BucketKey(t time.Time, granularity string) string {
	if granularity == "weekly" {
		return DayString(MondayOf(t))
	}
	return DayString(t)
}

// MondayOf returns the Monday of the ISO week containing t (UTC).
func MondayOf(t time.Time) time.Time {
	t = t.UTC()
	offset := (int(t.Weekday()) + 6) % 7
	return time.Date(t.Year(), t.Month(), t.Day()-offset, 0, 0, 0, 0, time.UTC)
}

// DayString formats a time as "2006-01-02".
func DayString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// FormatBucketLabel returns a human-readable label for a bucket key.
func FormatBucketLabel(bucketKey, granularity string) string {
	t, err := time.Parse("2006-01-02", bucketKey)
	if err != nil {
		return bucketKey
	}
	label := t.Format("Jan 2")
	if granularity == "weekly" {
		return fmt.Sprintf("Week of %s", label)
	}
	return label
}

// GenerateBucketKeys returns all bucket keys from start to end inclusive.
func GenerateBucketKeys(start, end, granularity string) []string {
	s, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil
	}
	e, err := time.Parse("2006-01-02", end)
	if err != nil {
		return nil
	}

	var keys []string
	for cur := s; !cur.After(e); cur = advanceBucket(cur, granularity) {
		keys = append(keys, DayString(cur))
	}
	return keys
}

func advanceBucket(t time.Time, granularity string) time.Time {
	if granularity == "weekly" {
		return t.AddDate(0, 0, 7)
	}
	return t.AddDate(0, 0, 1)
}
