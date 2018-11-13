package rules

import (
	"github.com/tehmoon/errors"
	"time"
	"../util"
)

type DateTime struct {
	plus time.Duration
	minus time.Duration
	layout string
	round string
	date string
}

func NewDateTime(config *DateTimeConfig) (dt *DateTime, err error) {
	dt = &DateTime{
		plus: time.Duration(0),
		minus: time.Duration(0),
		round: "minute",
		layout: time.RFC3339,
		date: "now",
	}

	if config.Round != "" {
		dt.round = config.Round
	}

	if config.Plus != "" {
		dt.plus, err = time.ParseDuration(config.Plus)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %q field", "plus")
		}
	}

	if config.Minus != "" {
		dt.minus, err = time.ParseDuration(config.Minus)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %q field", "minus")
		}
	}

	if config.Layout != "" {
		dt.layout = config.Layout
	}

	if config.Date != "" {
		dt.date = config.Date
	}

	if dt.date != "now" {
		_, err = time.Parse(dt.layout, dt.date)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %q field", "date")
		}
	}

	_, err = roundDownTime(time.Now(), dt.round)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing %q field", "round")
	}

	return dt, nil
}

func (dt DateTime) String(now *time.Time) (str string) {
	return util.FormatTime(dt.Time(now))
}

// Will return a time.Time representation of the DateTime object
// Pass a pointer to a time.Time if you want to set the DateTime.Date
// when it is set to now. The pointer is only to check against nil.
func (dt DateTime) Time(now *time.Time) (t time.Time) {
	switch dt.date {
		case "now":
			if now != nil {
				t = *now
			} else {
				t = time.Now()
			}

		default:
			// Already checked in NewDateTime()
			t, _ = time.Parse(dt.layout, dt.date)
	}

	// Already checked in NewDateTime()
	t, _ = roundDownTime(t, dt.round)
	t = t.Add(-1 * dt.minus)
	t = t.Add(dt.plus)

	return t
}

func roundDownTime(t time.Time, round string) (time.Time, error) {
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
		default:
			return time.Time{}, errors.Errorf("Round %q is not supported", round)
	}

	return t.Truncate(truncate), nil
}
