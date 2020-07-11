package schedule

import (
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestClock_String(t *testing.T) {
	assert.Equal(t, "00:00", Midnight.String())
}

func TestNewTime(t *testing.T) {
	tm := NewTime(8, 30)
	got := tm.Align(time.Now())
	assert.Equal(t, tm.Hour(), got.Hour())
	assert.Equal(t, tm.Minute(), got.Minute())
	assert.Equal(t, 0, got.Second())
}

func TestTime_Int64(t *testing.T) {
	assert.EqualValues(t, 830, NewTime(8, 30).Int64())
}

func TestTime_Int32(t *testing.T) {
	assert.EqualValues(t, 830, NewTime(8, 30).Int32())
}

func TestTime_Add(t *testing.T) {
	tm := Time(900)
	testCases := map[string]struct {
		Input time.Duration
		Want  Time
	}{
		"0": {
			Input: 0,
			Want:  tm,
		},
		"< 1hr": {
			Input: 30 * time.Minute,
			Want:  930,
		},
		"= 1hr": {
			Input: 60 * time.Minute,
			Want:  1000,
		},
		"> 1hr": {
			Input: 150 * time.Minute,
			Want:  1130,
		},
		"day wrap": {
			Input: 1410 * time.Minute,
			Want:  830,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := tm.Add(tc.Input)
			assert.Equal(t, got, tc.Want)
		})
	}
}
