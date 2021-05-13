package schedule

import (
	"time"
)

// AddDate performs date arithmetic accounting for dates excluded by schedules
func AddDate(t time.Time, days int, ss ...Schedule) time.Time {
	type DateRange struct {
		From string
		To   string
	}

	var excludes []DateRange
	for _, s := range ss {
		if !s.IsExclude() {
			continue
		}

		dateFrom, ok := s.DateFrom()
		if !ok {
			continue
		}

		dateTo, ok := s.DateTo()
		if !ok {
			continue
		}

		excludes = append(excludes, DateRange{
			From: dateFrom,
			To:   dateTo,
		})
	}

	date := t
	delta := 1
	if days < 0 {
		delta = -1
	}

loop:
	for days != 0 {
		date = date.AddDate(0, 0, delta)
		dateStr := date.Format("2006-01-02")
		for _, ex := range excludes {
			if dateStr >= ex.From && dateStr <= ex.To {
				continue loop
			}
		}

		days -= delta
	}

	return date
}
