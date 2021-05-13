package schedule

import (
	"testing"
	"time"
)

func TestAddDate(t *testing.T) {
	var (
		now       = time.Now().In(time.UTC)
		today     = now.Format("2006-01-02")
		tomorrow  = now.AddDate(0, 0, 1).Format("2006-01-02")
		yesterday = now.AddDate(0, 0, -1).Format("2006-01-02")
	)
	testCases := map[string]struct {
		Time      time.Time
		Days      int
		Schedules []Schedule
		Want      string
	}{
		"nop": {
			Time: now,
			Want: now.Format("2006-01-02"),
		},
		"tomorrow": {
			Time: now,
			Days: 1,
			Want: now.AddDate(0, 0, 1).Format("2006-01-02"),
		},
		"yesterday": {
			Time: now,
			Days: -1,
			Want: now.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"holiday - yesterday": {
			Time:      now,
			Days:      -1,
			Schedules: Schedules{ExcludeDateRange(yesterday, yesterday)},
			Want:      now.AddDate(0, 0, -2).Format("2006-01-02"),
		},
		"holiday - yesterday and today": {
			Time:      now,
			Days:      -1,
			Schedules: Schedules{ExcludeDateRange(yesterday, today)},
			Want:      now.AddDate(0, 0, -2).Format("2006-01-02"),
		},
		"holiday - yesterday and day before": {
			Time:      now.AddDate(0, 0, 1),
			Days:      -1,
			Schedules: Schedules{ExcludeDateRange(yesterday, today)},
			Want:      now.AddDate(0, 0, -2).Format("2006-01-02"),
		},
		"holiday - tomorrow": {
			Time:      now,
			Days:      1,
			Schedules: Schedules{ExcludeDateRange(tomorrow, tomorrow)},
			Want:      now.AddDate(0, 0, 2).Format("2006-01-02"),
		},
		"holiday - today and tomorrow": {
			Time:      now,
			Days:      1,
			Schedules: Schedules{ExcludeDateRange(today, tomorrow)},
			Want:      now.AddDate(0, 0, 2).Format("2006-01-02"),
		},
		"holiday - tomorrow and day after": {
			Time:      now.AddDate(0, 0, -1),
			Days:      1,
			Schedules: Schedules{ExcludeDateRange(today, tomorrow)},
			Want:      now.AddDate(0, 0, 2).Format("2006-01-02"),
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			date := AddDate(tc.Time, tc.Days, tc.Schedules...)
			if got, want := date.Format("2006-01-02"), tc.Want; got != want {
				t.Fatalf("got %v; want %v", got, want)
			}
		})
	}
}
