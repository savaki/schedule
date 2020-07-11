package schedule

import (
	"fmt"
	"strconv"
	"time"
)

const Midnight Time = 0

type Time int32

func NewTime(hour, minute int) Time {
	if hour < 0 || hour > 23 {
		panic(fmt.Errorf("invalid hour, %v", hour))
	}
	if minute < 0 || minute > 59 {
		panic(fmt.Errorf("invalid minute, %v", minute))
	}

	return Time(hour*100 + minute)
}

func NewTimeFromDate(date time.Time) Time {
	return NewTime(date.Hour(), date.Minute())
}

func (t Time) Append(buffer []byte) []byte {
	h, m := t.Hour(), t.Minute()
	if h < 10 {
		buffer = append(buffer, '0')
	}
	buffer = strconv.AppendInt(buffer, int64(h), 10)
	if m < 10 {
		buffer = append(buffer, '0')
	}
	buffer = strconv.AppendInt(buffer, int64(m), 10)
	return buffer
}

func (t Time) Hour() int {
	return int(t / 100)
}

func (t Time) Int32() int32 {
	return int32(t)
}

func (t Time) Int64() int64 {
	return int64(t)
}

func (t Time) Minute() int {
	return int(t % 100)
}

func (t Time) String() string {
	buffer := make([]byte, 0, 5)
	h, m := t.Hour(), t.Minute()
	if h < 10 {
		buffer = append(buffer, '0')
	}
	buffer = strconv.AppendInt(buffer, int64(h), 10)
	buffer = append(buffer, ':')
	if m < 10 {
		buffer = append(buffer, '0')
	}
	buffer = strconv.AppendInt(buffer, int64(m), 10)
	return string(buffer)
}

func (t Time) Align(v time.Time) time.Time {
	return time.Date(v.Year(), v.Month(), v.Day(), t.Hour(), t.Minute(), 0, 0, v.Location())
}
