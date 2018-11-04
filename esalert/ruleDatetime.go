package main

import (
	"github.com/tehmoon/errors"
	"time"
)

type RuleDatetime struct {
	// Defaults to 1 minute
	Minus string `json:"minus"`
	Plus string `json:"plus"`

	// Accepts: second, minute, hour, day, week
	Round string `json:"round"`
	Date string `json:"date"`

	// Defaults to time.RFC3339
	Layout string `json:"layout"`
}

func (rdt RuleDatetime) Time() (*time.Time, error) {
	var (
		err error
		t time.Time
		round, layout string = rdt.Round, rdt.Layout
		minus, plus time.Duration = 0, 0
	)

	if layout == "" {
		layout = time.RFC3339
	}

	switch rdt.Date {
		case "":
			t = time.Now()
		case "now":
			t = time.Now()
		default:
			t, err = time.Parse(layout, rdt.Date)
			if err != nil {
				return nil, errors.Wrap(err, "Error parsing \"date\" field")
			}
	}

	if rdt.Round == "" {
		round = "minute"
	}

	t = roundDownTime(t, round)

	if rdt.Minus != "" {
		minus, err = time.ParseDuration(rdt.Minus)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing \"minus\" field")
		}
	}

	if rdt.Plus != "" {
		plus, err = time.ParseDuration(rdt.Plus)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing \"plus\" field")
		}
	}

	t = t.Add(-1 * minus)
	t = t.Add(plus)

	return &t, nil
}

func roundDownTime(t time.Time, round string) (time.Time) {
	var truncate time.Duration

	switch round {
		case "nanosecond":
			truncate = time.Nanosecond
		case "second":
			truncate = time.Second
		case "minute":
			truncate = time.Minute
		case "hour":
			truncate = time.Hour
		case "day":
			truncate = 24 * time.Hour
		case "week":
			truncate = 7 * 24 * time.Hour
	}

	return t.Truncate(truncate)
}
