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
