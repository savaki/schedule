package schedule

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/tj/assert"
)

func TestSchedule_index(t *testing.T) {
	t.Run("date aware", func(t *testing.T) {
		const (
			dateFrom = "2020-01-01"
			dateTo   = "2020-01-15"
		)

		schedule := ExcludeDateRange(dateFrom, dateTo, time.Sunday, time.Monday)
		testCases := map[string]struct {
			n    int
			want string
			ok   bool
		}{
			"version": {
				n:    indexVersion,
				want: "1",
				ok:   true,
			},
			"date from": {
				n:    indexDateFrom,
				want: dateFrom,
				ok:   true,
			},
			"date to": {
				n:    indexDateTo,
				want: dateTo,
				ok:   true,
			},
			"weekdays": {
				n:    indexWeekdays,
				want: "SuMo",
				ok:   true,
			},
			"exclude": {
				n:    indexExclude,
				want: exclude,
				ok:   true,
			},
			"invalid": {
				n:  8,
				ok: false,
			},
			"too low": {
				n:  -1,
				ok: false,
			},
		}

		for label, tc := range testCases {
			t.Run(label, func(t *testing.T) {
				from, to, ok := schedule.index(tc.n)
				assert.Equal(t, tc.ok, ok)
				assert.Equal(t, tc.want, string(schedule[from:to]))
			})
		}
	})

	t.Run("generic", func(t *testing.T) {
		schedule := New(800, 1100, time.Sunday, time.Monday)
		testCases := map[string]struct {
			n    int
			want string
			ok   bool
		}{
			"version": {
				n:    1,
				want: "1",
				ok:   true,
			},
			"date from": {
				n:  2,
				ok: false,
			},
			"date to": {
				n:  3,
				ok: false,
			},
			"from": {
				n:    4,
				want: "0800",
				ok:   true,
			},
			"to": {
				n:    5,
				want: "1100",
				ok:   true,
			},
			"weekdays": {
				n:    6,
				want: "SuMo",
				ok:   true,
			},
			"exclude": {
				n:  7,
				ok: false,
			},
			"invalid": {
				n:  8,
				ok: false,
			},
			"too low": {
				n:  -1,
				ok: false,
			},
		}

		for label, tc := range testCases {
			t.Run(label, func(t *testing.T) {
				from, to, ok := schedule.index(tc.n)
				assert.Equal(t, tc.ok, ok)
				assert.Equal(t, tc.want, string(schedule[from:to]))
			})
		}
	})
}

func TestSchedule_DynamoDB(t *testing.T) {
	want := ExcludeDateRange("2006-01-02", "2006-01-03", 800, 1100, time.Sunday, time.Monday)
	item, err := dynamodbattribute.Marshal(want)
	assert.Nil(t, err)

	var got Schedule
	err = dynamodbattribute.Unmarshal(item, &got)
	assert.Nil(t, err)

	assert.Equal(t, want, got)
}

func TestOrder(t *testing.T) {
	v := []string{
		"1:::",
		"1:2006",
	}
	sort.Strings(v)
}

func TestSchedule_String(t *testing.T) {
	assert.Equal(t, "1:::0800:0100:Mo:", New(800, 100, time.Monday).String())
}

func TestDayOfTheWeek(t *testing.T) {
	testCases := map[string]struct {
		DayOfTheWeek DayOfTheWeek
		Weekday      time.Weekday
	}{
		"Sunday": {
			DayOfTheWeek: Sunday,
			Weekday:      time.Sunday,
		},
		"Monday": {
			DayOfTheWeek: Monday,
			Weekday:      time.Monday,
		},
		"Tuesday": {
			DayOfTheWeek: Tuesday,
			Weekday:      time.Tuesday,
		},
		"Wednesday": {
			DayOfTheWeek: Wednesday,
			Weekday:      time.Wednesday,
		},
		"Thursday": {
			DayOfTheWeek: Thursday,
			Weekday:      time.Thursday,
		},
		"Friday": {
			DayOfTheWeek: Friday,
			Weekday:      time.Friday,
		},
		"Saturday": {
			DayOfTheWeek: Saturday,
			Weekday:      time.Saturday,
		},
	}
	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			w, ok := tc.DayOfTheWeek.Weekday()
			assert.True(t, ok)
			assert.EqualValues(t, tc.Weekday, w)

			d, ok := getDayOfTheWeek(tc.Weekday)
			assert.True(t, ok)
			assert.Equal(t, tc.DayOfTheWeek, d)
		})
	}
}

func TestSchedule_DateFrom(t *testing.T) {
	dateFrom := "2006-01-01"
	dateTo := "2006-02-01"
	s := DateRange(dateFrom, dateTo, 800, 1200, time.Sunday)

	got, ok := s.DateFrom()
	assert.True(t, ok)
	assert.Equal(t, dateFrom, got)

	got, ok = s.DateTo()
	assert.True(t, ok)
	assert.Equal(t, dateTo, got)
}

func TestSchedule_Weekdays(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		s := New(800, 1200, time.Tuesday, time.Thursday)

		got := s.Weekdays()
		assert.Equal(t, []time.Weekday{time.Tuesday, time.Thursday}, got)
	})

	t.Run("none", func(t *testing.T) {
		s := New(800, 1200)

		got := s.Weekdays()
		assert.Len(t, got, 0)
	})
}

func TestSimpleSchedule(t *testing.T) {
	s := New(800, 1200, time.Monday)
	fmt.Println(s)
}

func TestSchedule_Contains(t *testing.T) {
	schedule := DateRange("2020-06-01", "2020-06-30", 900, 1700,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
	)

	testCases := map[string]struct {
		Date string
		Want bool
	}{
		"included weekday - wed": {
			Date: "2020-06-03",
			Want: true,
		},
		"excluded weekday - sun": {
			Date: "2020-06-07",
			Want: false,
		},
		"excluded date": {
			Date: "2020-05-31",
			Want: false,
		},
		"included date": {
			Date: "2020-06-03",
			Want: true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			date, err := time.Parse(DateLayout, tc.Date)
			assert.Nil(t, err)

			got := schedule.Contains(date)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestSchedule_ContainsWeekday(t *testing.T) {
	testCases := map[string]struct {
		Schedule string
		Weekday  time.Weekday
		Want     bool
	}{
		"Wed": {
			Schedule: "1:::0900:1700:MoTuWeThFr:",
			Weekday:  time.Wednesday,
			Want:     true,
		},
		"Sun": {
			Schedule: "1:::0900:1700:MoTuWeThFr:",
			Weekday:  time.Sunday,
			Want:     false,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			s := Schedule(tc.Schedule)
			got := s.ContainsWeekday(tc.Weekday)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestContainsDate(t *testing.T) {
	schedule := DateRange("2020-06-01", "2020-06-30", 900, 1700,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
	)

	testCases := map[string]struct {
		Schedule Schedule
		Date     string
		Want     bool
	}{
		"included weekday - wed": {
			Schedule: schedule,
			Date:     "2020-06-03",
			Want:     true,
		},
		"excluded weekday - sun": {
			Schedule: schedule,
			Date:     "2020-06-07",
			Want:     false,
		},
		"excluded date": {
			Schedule: schedule,
			Date:     "2020-05-31",
			Want:     false,
		},
		"included date": {
			Schedule: schedule,
			Date:     "2020-06-03",
			Want:     true,
		},
		"any date": {
			Schedule: New(900, 1700), // 9-5 everyday
			Date:     "2020-06-03",
			Want:     true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			date, err := time.Parse(DateLayout, tc.Date)
			assert.Nil(t, err)

			got := ContainsDate(date, tc.Schedule)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestContainsWeekday(t *testing.T) {
	testCases := map[string]struct {
		Schedule string
		Weekday  time.Weekday
		Want     bool
	}{
		"Wed": {
			Schedule: "1:::::MoTuWeThFr:",
			Weekday:  time.Wednesday,
			Want:     true,
		},
		"Sun": {
			Schedule: "1:::::MoTuWeThFr:",
			Weekday:  time.Sunday,
			Want:     false,
		},
		"Any": {
			Schedule: "1::::::",
			Weekday:  time.Sunday,
			Want:     true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			s := Schedule(tc.Schedule)
			got := ContainsWeekday(tc.Weekday, s)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func BenchmarkContainsDate(b *testing.B) {
	var (
		date     = time.Date(2020, time.June, 3, 0, 0, 0, 0, time.Local)
		got      int64
		schedule = DateRange("2020-06-01", "2020-06-30", 900, 1700,
			time.Monday,
			time.Tuesday,
			time.Wednesday,
			time.Thursday,
			time.Friday,
		)
	)

	for i := 0; i < b.N; i++ {
		if ContainsDate(date, schedule) {
			got++
		}
	}
	if got == 0 {
		b.Fatalf("got %v; want > 0", got)
	}
}

func Test_getDayOfTheWeekBytes(t *testing.T) {
	daysOfTheWeek := []DayOfTheWeek{
		Sunday,
		Monday,
		Tuesday,
		Wednesday,
		Thursday,
		Friday,
		Saturday,
	}

	for _, want := range daysOfTheWeek {
		got, ok := getDayOfTheWeekBytes([]byte(want))
		assert.True(t, ok)
		assert.Equal(t, want, got)
	}

	_, ok := getDayOfTheWeekBytes([]byte("nope"))
	assert.False(t, ok)
}

func TestSchedules_Serialize(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		want := Schedules{
			New(800, 1200),
			New(1300, 1700),
		}

		item, err := dynamodbattribute.Marshal(want)
		assert.Nil(t, err)

		var got Schedules
		err = dynamodbattribute.Unmarshal(item, &got)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("nil", func(t *testing.T) {
		var want Schedules
		item, err := dynamodbattribute.Marshal(want)
		assert.Nil(t, err)

		var got Schedules
		err = dynamodbattribute.Unmarshal(item, &got)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
}

func TestSchedules_After(t *testing.T) {
	ss := Schedules{
		New(730, 830),
		New(1300, 1700),
		New(900, 1200),
	}

	date := time.Date(2020, time.July, 20, 11, 0, 0, 0, time.Local)
	got, err := ss.After(date)
	assert.Nil(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, NewTimeSlot(1100, 1200), got[0])
	assert.Equal(t, NewTimeSlot(1300, 1700), got[1])
}

func TestSchedules_TimeSlots(t *testing.T) {
	date := time.Date(2020, time.July, 20, 11, 0, 0, 0, time.Local)
	dateStr := date.Format("2006-01-02")

	testCases := map[string]struct {
		Input Schedules
		Want  []TimeSlot
	}{
		"all": {
			Input: Schedules{
				New(1300, 1700),
				New(900, 1200),
			},
			Want: []TimeSlot{
				NewTimeSlot(900, 1200),
				NewTimeSlot(1300, 1700),
			},
		},
		"union": {
			Input: Schedules{
				New(1200, 1700),
				New(900, 1200),
			},
			Want: []TimeSlot{
				NewTimeSlot(900, 1700),
			},
		},
		"exclude date": {
			Input: Schedules{
				New(1200, 1700),
				New(900, 1200),
				ExcludeDateRange(dateStr, dateStr),
			},
			Want: nil,
		},
		"holiday schedule takes precedence": {
			Input: Schedules{
				New(1200, 1700),
				New(900, 1200),
				DateRange(dateStr, dateStr, 1400, 1600),
			},
			Want: []TimeSlot{
				NewTimeSlot(1400, 1600),
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got, err := tc.Input.TimeSlots(date)
			assert.Nil(t, err)
			assert.Equal(t, tc.Want, got)
		})
	}
}

func TestSchedules_Next(t *testing.T) {
	date := time.Date(2020, time.July, 20, 11, 0, 0, 0, time.Local)

	testCases := map[string]struct {
		Input  Schedules
		Buffer time.Duration
		Want   time.Time
	}{
		"same day": {
			Input:  Schedules{New(900, 1700)},
			Buffer: 30 * time.Minute,
			Want:   date,
		},
		"later that day": {
			Input:  Schedules{New(1400, 1700)},
			Buffer: 30 * time.Minute,
			Want:   NewTime(14, 0).Align(date),
		},
		"next day": {
			Input:  Schedules{New(900, 1000)},
			Buffer: 30 * time.Minute,
			Want:   NewTime(9, 0).Align(date).AddDate(0, 0, 1),
		},
		"require at least buffer remaining": {
			Input:  Schedules{New(1130, 1145), New(1200, 1230)},
			Buffer: 30 * time.Minute,
			Want:   NewTime(12, 0).Align(date),
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got, err := tc.Input.Next(date, tc.Buffer)
			assert.Nil(t, err)
			assert.Equal(t, tc.Want.Format(time.RFC822), got.Format(time.RFC822))
		})
	}
}

func TestExclude(t *testing.T) {
	s := ExcludeDateRange("2020-07-01", "2020-07-15")
	assert.True(t, s.IsExclude())
}
