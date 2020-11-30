package schedule

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const exclude = "exclude"

const (
	indexVersion  = 1
	indexDateFrom = 2
	indexDateTo   = 3
	indexFrom     = 4
	indexTo       = 5
	indexWeekdays = 6
	indexExclude  = 7
)

type DayOfTheWeek string

func (d DayOfTheWeek) String() string {
	return string(d)
}

func (d DayOfTheWeek) Weekday() (time.Weekday, bool) {
	switch d {
	case Sunday:
		return time.Sunday, true
	case Monday:
		return time.Monday, true
	case Tuesday:
		return time.Tuesday, true
	case Wednesday:
		return time.Wednesday, true
	case Thursday:
		return time.Thursday, true
	case Friday:
		return time.Friday, true
	case Saturday:
		return time.Saturday, true
	default:
		return 0, false
	}
}

func getDayOfTheWeek(item time.Weekday) (DayOfTheWeek, bool) {
	switch item {
	case time.Sunday:
		return Sunday, true
	case time.Monday:
		return Monday, true
	case time.Tuesday:
		return Tuesday, true
	case time.Wednesday:
		return Wednesday, true
	case time.Thursday:
		return Thursday, true
	case time.Friday:
		return Friday, true
	case time.Saturday:
		return Saturday, true
	default:
		return "", false
	}
}

func getDayOfTheWeekBytes(b []byte) (DayOfTheWeek, bool) {
	switch {
	case bytes.Equal(bSunday, b):
		return Sunday, true
	case bytes.Equal(bMonday, b):
		return Monday, true
	case bytes.Equal(bTuesday, b):
		return Tuesday, true
	case bytes.Equal(bWednesday, b):
		return Wednesday, true
	case bytes.Equal(bThursday, b):
		return Thursday, true
	case bytes.Equal(bFriday, b):
		return Friday, true
	case bytes.Equal(bSaturday, b):
		return Saturday, true
	default:
		return "", false
	}
}

const (
	Sunday    DayOfTheWeek = "Su"
	Monday    DayOfTheWeek = "Mo"
	Tuesday   DayOfTheWeek = "Tu"
	Wednesday DayOfTheWeek = "We"
	Thursday  DayOfTheWeek = "Th"
	Friday    DayOfTheWeek = "Fr"
	Saturday  DayOfTheWeek = "Sa"
)

var (
	bSunday    = []byte(Sunday)
	bMonday    = []byte(Monday)
	bTuesday   = []byte(Tuesday)
	bWednesday = []byte(Wednesday)
	bThursday  = []byte(Thursday)
	bFriday    = []byte(Friday)
	bSaturday  = []byte(Saturday)
)

// SuMoTuWeThFrSa
// version:date-from:date-to:weekdays:from-time:to-time:exclude|include
type Schedule []byte

func New(from, to Time, weekdays ...time.Weekday) Schedule {
	return DateRange("", "", from, to, weekdays...)
}

func (s Schedule) index(n int) (int, int, bool) {
	if n < 1 {
		return 0, 0, false
	}

	offset, previous := -1, 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			previous = offset + 1
			offset = i

			n--
			if n == 0 {
				return previous, offset, previous != offset
			}
		}
	}

	if n != 1 {
		return 0, 0, false
	}

	previous = offset + 1
	offset = len(s)
	return previous, offset, previous != offset
}

func (s Schedule) validate() error {
	if _, _, ok := s.index(indexVersion); !ok {
		return fmt.Errorf("invalid Schedule: missing version")
	}
	if _, _, ok := s.index(indexTo); !ok {
		return fmt.Errorf("invalid Schedule: missing to")
	}
	return nil
}

// MarshalDynamoDBAttributeValue marshals Schedule for dynamodb
func (s Schedule) MarshalDynamoDBAttributeValue(item *dynamodb.AttributeValue) error {
	*item = dynamodb.AttributeValue{
		S: aws.String(string(s)),
	}
	return nil
}

// UnmarshalDynamoDBAttributeValue unmarshals Schedule for dynamodb
func (s *Schedule) UnmarshalDynamoDBAttributeValue(item *dynamodb.AttributeValue) error {
	if item == nil || item.S == nil {
		return fmt.Errorf("dynamodb.AttributeValue not a Schedule:  missing S key")
	}

	v := Schedule(*item.S)
	if err := v.validate(); err != nil {
		return err
	}

	*s = v
	return nil
}

// DateFrom extracts the from date from the schedule
func (s Schedule) DateFrom() (string, bool) {
	i, j, ok := s.index(indexDateFrom)
	if !ok {
		return "", false
	}

	return string(s[i:j]), true
}

// DateTo extracts the from date from the schedule
func (s Schedule) DateTo() (string, bool) {
	i, j, ok := s.index(indexDateTo)
	if !ok {
		return "", false
	}

	return string(s[i:j]), true
}

// ContainsDate matches the provided date (but not time)
func (s Schedule) Contains(date time.Time) bool {
	if !s.ContainsWeekday(date.Weekday()) {
		return false
	}

	var (
		fi, fj, fok = s.index(indexDateFrom)
		ti, tj, tok = s.index(indexDateTo)
		buf         [len(DateLayout)]byte
		str         = date.AppendFormat(buf[:0], DateLayout)
	)

	switch {
	case fok && tok:
		from := bytes.Compare(s[fi:fj], str)
		to := bytes.Compare(s[ti:tj], str)
		return from <= 0 && to >= 0

	default:
		return true
	}
}

// ContainsWeekday matches the weekday only
func (s Schedule) ContainsWeekday(weekday time.Weekday) bool {
	if i, j, ok := s.index(indexWeekdays); ok {
		d, _ := getDayOfTheWeek(weekday)
		return strings.Contains(string(s[i:j]), d.String())
	}
	return true
}

func (s Schedule) From() (Time, error) {
	i, j, ok := s.index(indexFrom)
	if !ok {
		return 0, fmt.Errorf("invalid from time, %s", s)
	}

	v, err := strconv.ParseInt(string(s[i:j]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid from time, %s: %w", s, err)
	}

	return Time(v), nil
}

// HasDateRange indicates Schedule is scoped to a date range
func (s Schedule) HasDateRange() bool {
	_, _, ok := s.index(indexDateFrom)
	return ok
}

func (s Schedule) IsExclude() bool {
	_, _, ok := s.index(indexExclude)
	return ok
}

func (s Schedule) String() string {
	return string(s)
}

func (s Schedule) TimeSlot() (TimeSlot, error) {
	from, err := s.From()
	if err != nil {
		return TimeSlot{}, err
	}
	to, err := s.To()
	if err != nil {
		return TimeSlot{}, err
	}

	return TimeSlot{
		From: from,
		To:   to,
	}, nil
}

func (s Schedule) To() (Time, error) {
	i, j, ok := s.index(indexTo)
	if !ok {
		return 0, fmt.Errorf("invalid to time, %s", s)
	}

	v, err := strconv.ParseInt(string(s[i:j]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid to time, %s: %w", s, err)
	}

	return Time(v), nil
}

func (s Schedule) Weekdays() []time.Weekday {
	i, j, ok := s.index(indexWeekdays)
	if !ok {
		return nil
	}

	days := s[i:j]

	var ww []time.Weekday
	for k := 0; k+1 < len(days); k += 2 {
		d, ok := getDayOfTheWeekBytes(days[k : k+2])
		if !ok {
			continue
		}
		if w, ok := d.Weekday(); ok {
			ww = append(ww, w)
		}
	}

	return ww
}

type Schedules []Schedule

func (s Schedules) After(date time.Time) ([]TimeSlot, error) {
	return After(date, s...)
}

// Contains returns true if Schedules contains the requested Schedule
func (s Schedules) Contains(want Schedule) bool {
	for _, item := range s {
		if reflect.DeepEqual(item, want) {
			return true
		}
	}
	return false
}

func (s Schedules) MarshalDynamoDBAttributeValue(item *dynamodb.AttributeValue) error {
	var ss []string
	for _, v := range s {
		ss = append(ss, v.String())
	}

	item.SS = aws.StringSlice(ss)

	return nil
}

func (s Schedules) Next(date time.Time, sans ...TimeSlot) (time.Time, error) {
	return Next(date, s, sans...)
}

// TimeSlots returns the set of time slots for the date requested
func (s Schedules) TimeSlots(date time.Time) ([]TimeSlot, error) {
	return TimeSlots(date, s...)
}

func (s *Schedules) UnmarshalDynamoDBAttributeValue(item *dynamodb.AttributeValue) error {
	var v Schedules
	for _, s := range aws.StringValueSlice(item.SS) {
		v = append(v, Schedule(s))
	}

	*s = v

	return nil
}

func buildSchedule(dateFrom string, dateTo string, from Time, to Time, weekdays []time.Weekday) []byte {
	buffer := make([]byte, 0, 64)
	buffer = append(buffer, '1') // version
	buffer = append(buffer, ':')
	buffer = append(buffer, dateFrom...)
	buffer = append(buffer, ':')
	buffer = append(buffer, dateTo...)
	buffer = append(buffer, ':')
	buffer = from.Append(buffer)
	buffer = append(buffer, ':')
	buffer = to.Append(buffer)
	buffer = append(buffer, ':')
	for _, w := range weekdays {
		if d, ok := getDayOfTheWeek(w); ok {
			buffer = append(buffer, d...)
		}
	}
	buffer = append(buffer, ':')
	return buffer
}

func alignMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// After returns the set of TimeSlots that occur on the date provided AND
// after the time provided.  Results will be unioned and sorted
func After(date time.Time, ss ...Schedule) ([]TimeSlot, error) {
	timeSlots, err := TimeSlots(date, ss...)
	if err != nil {
		return nil, err
	}

	var (
		want    = NewTimeFromDate(date)
		matches []TimeSlot
	)

	for _, timeSlot := range timeSlots {
		switch {
		case timeSlot.To <= want:
			continue
		case timeSlot.From <= want:
			matches = append(matches, NewTimeSlot(want, timeSlot.To))
		default:
			matches = append(matches, timeSlot)
		}
	}

	return matches, nil
}

// ContainsDate returns true if the specified date and time is contained in any
// of the schedules provided
func ContainsDate(date time.Time, ss ...Schedule) bool {
	for _, s := range ss {
		if s.Contains(date) {
			return true
		}
	}
	return false
}

// ContainsWeekday returns true if the weekday is contained within the provided schedule
func ContainsWeekday(weekday time.Weekday, ss ...Schedule) bool {
	for _, s := range ss {
		if s.ContainsWeekday(weekday) {
			return true
		}
	}
	return false
}

// DateRange returns a new Schedule from a date range.  Useful for special holiday
// hours.
func DateRange(dateFrom, dateTo string, from, to Time, weekdays ...time.Weekday) Schedule {
	buffer := buildSchedule(dateFrom, dateTo, from, to, weekdays)
	return Schedule(buffer)
}

// ExcludeDateRange defines an excluded Schedule for a date range.  Useful for
// holiday closed hours
func ExcludeDateRange(dateFrom, dateTo string, weekdays ...time.Weekday) Schedule {
	buffer := buildSchedule(dateFrom, dateTo, 0, 0, weekdays)
	buffer = append(buffer, exclude...)
	return Schedule(buffer)
}

// Next returns the next time available from the Schedule provided with at least
// buffer duration remaining in the schedule
func Next(date time.Time, ss Schedules, sans ...TimeSlot) (time.Time, error) {
	const daysOut = 7
	for i := 0; i < daysOut; i++ {
		d := date.AddDate(0, 0, i)
		timeSlots, err := After(d, ss...)
		if err != nil {
			return time.Time{}, err
		}

		if len(sans) > 0 {
			timeSlots = SubAll(timeSlots, sans)
		}

		for _, timeSlot := range timeSlots {
			return timeSlot.From.Align(d), nil
		}

		if i == 0 {
			date = alignMidnight(date)
		}
	}

	return time.Time{}, fmt.Errorf("no time matches in next %v days", daysOut)
}

// TimeSlots returns the timeslots for the date
func TimeSlots(date time.Time, ss ...Schedule) ([]TimeSlot, error) {
	var regular, holiday []TimeSlot
	for _, s := range ss {
		if !s.Contains(date) {
			continue
		}
		if s.IsExclude() {
			return nil, nil // excluded date
		}

		timeSlot, err := s.TimeSlot()
		if err != nil {
			return nil, err
		}

		if s.HasDateRange() {
			holiday = append(holiday, timeSlot)
			continue
		}

		regular = append(regular, timeSlot)
	}

	slots := holiday
	if len(slots) == 0 {
		slots = regular
	}

	slots = Union(slots...)

	sort.Slice(slots, func(i, j int) bool {
		return slots[i].From < slots[j].From
	})

	return slots, nil
}
