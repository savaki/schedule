package schedule

import (
	"bytes"
	"sort"
	"time"
)

const DateLayout = "2006-01-02"

// TimeSlot provides a range from a to b
type TimeSlot struct {
	From Time // From time
	To   Time // To time
}

// NewTimeSlot returns a new TimeSlot
func NewTimeSlot(from, to Time) TimeSlot {
	return TimeSlot{
		From: from,
		To:   to,
	}
}

// Contains indicates the TimeSlot completely contains the provided TimeSlot
func (t TimeSlot) Contains(v TimeSlot) bool {
	return t.From <= v.From && t.To >= v.To
}

// Duration of this TimeSlot
func (t TimeSlot) Duration() time.Duration {
	h := t.To.Hour() - t.From.Hour()
	m := t.To.Minute() - t.From.Minute()

	if m < 0 {
		m += 60
		h--
	}

	return time.Duration(h*60+m) * time.Minute
}

// Sub subtracts the TimeSlot provided the current TimeSlot
// and returns the remaining TimeSlots
func (t TimeSlot) Sub(v TimeSlot) []TimeSlot {
	switch {
	case !t.Contains(v):
		return []TimeSlot{t}

	case t.From == v.From && t.To == v.To: // exact
		return nil

	case t.From == v.From: // sub head
		return []TimeSlot{
			{
				From: v.To,
				To:   t.To,
			},
		}

	case t.To == v.To: // sub tail
		return []TimeSlot{
			{
				From: t.From,
				To:   v.From,
			},
		}

	default:
		return []TimeSlot{
			{
				From: t.From,
				To:   v.From,
			},
			{
				From: v.To,
				To:   t.To,
			},
		}
	}
}

// Overlaps returns true if the any portion of the provide TimeSlot
// overlaps the current time Slot
func (t TimeSlot) Overlaps(v TimeSlot) bool {
	return t.From <= v.From && t.To >= v.From
}

// Union merges two time slots. Any time between the two TimeSlots
// will be included in the union
func (t TimeSlot) Union(v TimeSlot) TimeSlot {
	return TimeSlot{
		From: min(t.From, v.From),
		To:   max(t.To, v.To),
	}
}

func Hours(date time.Time, schedules ...Schedule) ([]TimeSlot, bool) {
	sort.Slice(schedules, func(i, j int) bool {
		return bytes.Compare(schedules[i], schedules[j]) < 0
	})

	// exclude takes precedence
	for _, s := range schedules {
		if s.IsExclude() && s.Contains(date) {
			return nil, false // exclude trumps all other schedules
		}
	}

	// check other elements
	for _, s := range schedules {
		if s.IsExclude() {
			continue
		}
		if match := s.Contains(date); !match {
			continue
		}

		block, err := s.TimeSlot()
		if err != nil {
			return nil, false
		}

		return []TimeSlot{block}, true
	}
	return nil, false
}

func Availability(date time.Time, schedules []Schedule, reserved []TimeSlot) []TimeSlot {
	blocks, ok := Hours(date, schedules...)
	if !ok {
		return nil
	}

	for _, r := range reserved {
		blocks = Sub(blocks, r)
	}

	return blocks
}

// Sub returns the block of time slots sans the provided time slot
func Sub(blocks []TimeSlot, v TimeSlot) []TimeSlot {
	var results []TimeSlot
	for index, block := range blocks {
		if block.Contains(v) {
			results = append(results, block.Sub(v)...)
			results = append(results, blocks[index+1:]...)
			break
		}
		results = append(results, block)
	}
	return results
}

// SubAll removes all sans time slots from the block provided
func SubAll(blocks, sans []TimeSlot) []TimeSlot {
	for _, s := range sans {
		blocks = Sub(blocks, s)
	}
	return blocks
}

func Union(blocks ...TimeSlot) []TimeSlot {
	sort.Slice(blocks, func(i, j int) bool {
		ii, jj := blocks[i], blocks[j]
		if ii.From == jj.From {
			return ii.To > jj.To
		}
		return ii.From < jj.From
	})

	var results []TimeSlot
	for i := 0; i < len(blocks); i++ {
		v := blocks[i]
		for ; i+1 < len(blocks) && v.Overlaps(blocks[i+1]); i++ {
			v = v.Union(blocks[i+1])
		}
		results = append(results, v)
	}
	return results
}

func max(a, b Time) Time {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b Time) Time {
	if a < b {
		return a
	} else {
		return b
	}
}
