package schedule

import (
	"fmt"
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestTimeSlot_Sub(t *testing.T) {
	var (
		morning   = NewTimeSlot(800, 1200)
		afternoon = NewTimeSlot(1200, 1600)
		lunch     = NewTimeSlot(1100, 1300)
		day       = NewTimeSlot(800, 1600)
	)

	testCases := map[string]struct {
		Time TimeSlot
		Sub  TimeSlot
		Want []TimeSlot
	}{
		"no-overlap": {
			Time: morning,
			Sub:  afternoon,
			Want: []TimeSlot{morning},
		},
		"equal": {
			Time: morning,
			Sub:  morning,
			Want: nil,
		},
		"head": {
			Time: day,
			Sub:  morning,
			Want: []TimeSlot{afternoon},
		},
		"tail": {
			Time: day,
			Sub:  afternoon,
			Want: []TimeSlot{morning},
		},
		"middle": {
			Time: day,
			Sub:  lunch,
			Want: []TimeSlot{
				NewTimeSlot(day.From, lunch.From),
				NewTimeSlot(lunch.To, day.To),
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := tc.Time.Sub(tc.Sub)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestSub(t *testing.T) {
	var (
		morning   = NewTimeSlot(800, 1000)
		midday    = NewTimeSlot(1000, 1200)
		lunch     = NewTimeSlot(1200, 1400)
		afternoon = NewTimeSlot(1400, 1600)
		head      = NewTimeSlot(800, 1400)
		tail      = NewTimeSlot(1400, 2000)
	)

	testCases := map[string]struct {
		Slots []TimeSlot
		Sub   TimeSlot
		Want  []TimeSlot
	}{
		"no-overlap": {
			Slots: []TimeSlot{morning},
			Sub:   afternoon,
			Want:  []TimeSlot{morning},
		},
		"equal": {
			Slots: []TimeSlot{morning},
			Sub:   morning,
			Want:  nil,
		},
		"save tail": {
			Slots: []TimeSlot{morning, midday, lunch},
			Sub:   morning,
			Want:  []TimeSlot{midday, lunch},
		},
		"save head": {
			Slots: []TimeSlot{morning, midday, lunch},
			Sub:   midday,
			Want:  []TimeSlot{morning, lunch},
		},
		"splice middle": {
			Slots: []TimeSlot{head, tail},
			Sub:   midday,
			Want:  []TimeSlot{morning, lunch, tail},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := Sub(tc.Slots, tc.Sub)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestUnion(t *testing.T) {
	var (
		morning = NewTimeSlot(800, 1000)
		midday  = NewTimeSlot(1000, 1200)
		lunch   = NewTimeSlot(1200, 1400)
		head    = NewTimeSlot(800, 1400)
	)

	testCases := map[string]struct {
		Slots []TimeSlot
		Want  []TimeSlot
	}{
		"nil": {
			Slots: nil,
			Want:  nil,
		},
		"single": {
			Slots: []TimeSlot{head},
			Want:  []TimeSlot{head},
		},
		"pair": {
			Slots: []TimeSlot{morning, midday, lunch},
			Want:  []TimeSlot{head},
		},
		"out of order": {
			Slots: []TimeSlot{morning, lunch, midday},
			Want:  []TimeSlot{head},
		},
		"unconnected": {
			Slots: []TimeSlot{morning, lunch},
			Want:  []TimeSlot{morning, lunch},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := Union(tc.Slots...)
			assert.EqualValues(t, tc.Want, got)
		})
	}
}

func TestHours(t *testing.T) {
	var (
		date         = time.Now()
		today        = date.Format("2006-01-02")
		tomorrow     = date.AddDate(0, 0, 1).Format("2006-01-02")
		standard     = New(800, 1800, date.Weekday())
		special      = DateRange(today, today, 800, 1200, date.Weekday())
		otherSpecial = DateRange(tomorrow, tomorrow, 800, 100, date.Weekday())
		exclude      = ExcludeDateRange(today, today, date.Weekday())
	)

	testCases := map[string]struct {
		Scheduled []Schedule
		Ok        bool
		Want      []TimeSlot
	}{
		"open day": {
			Scheduled: []Schedule{standard},
			Ok:        true,
			Want: []TimeSlot{
				NewTimeSlot(800, 1800),
			},
		},
		"date override": {
			Scheduled: []Schedule{standard, special},
			Ok:        true,
			Want: []TimeSlot{
				NewTimeSlot(800, 1200),
			},
		},
		"exclude takes precedence": {
			Scheduled: []Schedule{special, standard, exclude, special},
			Ok:        false,
			Want:      nil,
		},
		"other special ignored": {
			Scheduled: []Schedule{otherSpecial},
			Ok:        false,
			Want:      nil,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got, ok := Hours(date, tc.Scheduled...)
			assert.Equal(t, tc.Ok, ok)
			assert.EqualValues(t, tc.Want, got)
		})
	}
}

func TestAvailability(t *testing.T) {
	var (
		date     = time.Now()
		standard = New(800, 1800, date.Weekday())
		tomorrow = date.AddDate(0, 0, 1).Format("2006-01-02")
		special  = DateRange(tomorrow, tomorrow, 800, 1200, date.Weekday())
	)

	testCases := map[string]struct {
		Scheduled []Schedule
		Reserved  []TimeSlot
		Want      []TimeSlot
	}{
		"open day": {
			Scheduled: []Schedule{standard},
			Want: []TimeSlot{
				NewTimeSlot(800, 1800),
			},
		},
		"day off": {
			Scheduled: []Schedule{special},
			Want:      nil,
		},
		"reservations": {
			Scheduled: []Schedule{standard},
			Reserved: []TimeSlot{
				NewTimeSlot(900, 1000),
			},
			Want: []TimeSlot{
				NewTimeSlot(800, 900),
				NewTimeSlot(1000, 1800),
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := Availability(date, tc.Scheduled, tc.Reserved)
			assert.EqualValues(t, tc.Want, got)
		})
	}
}

func TestSample(t *testing.T) {
	a := New(800, 1800, time.Monday, time.Wednesday, time.Friday)
	b := New(800, 1400, time.Tuesday, time.Thursday)
	c := DateRange("2020-02-17", "2020-02-17", 800, 1200)
	//c := ExcludeDateRange("2020-02-17", "2020-02-17")

	date, _ := time.Parse("2006-01-02", "2020-02-17")
	//blocks, isAvailable := Hours(date, a, b, c)
	//fmt.Println(blocks, isAvailable)

	reservation := NewTimeSlot(1000, 1100)
	blocks := Availability(date, []Schedule{a, b, c}, []TimeSlot{reservation})
	fmt.Println(blocks)
}
