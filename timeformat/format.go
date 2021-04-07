package timeformat

import (
	"fmt"
	"strings"
	"time"
)

// Compononents holds a bitmask specifying a set of time components,
// as held in a time format string.
type Components int

const (
	Year Components = 1 << iota
	Month
	Day
	Hour
	Minute
	Second
	TZOffset
	TZName
)

func (c Components) String() string {
	var buf strings.Builder
	for i := Components(1); i <= TZName && c != 0; i <<= 1 {
		if (c & i) != 0 {
			if buf.Len() > 0 {
				buf.WriteByte('|')
			}
			var s string
			switch i {
			case Year:
				s = "year"
			case Month:
				s = "month"
			case Day:
				s = "day"
			case Hour:
				s = "hour"
			case Minute:
				s = "minute"
			case Second:
				s = "second"
			case TZOffset:
				s = "tzoffset"
			case TZName:
				s = "tzname"
			}
			if s != "" {
				buf.WriteString(s)
				c &^= i
			}
		}
	}
	if c != 0 {
		if buf.Len() > 0 {
			buf.WriteByte('|')
		}
		fmt.Fprintf(&buf, "%#x", uint64(c))
	}
	return buf.String()
}

// Date specifies a date. Each field corresponds to one argument
// to the time.Date function.
type Date struct {
	Year       int
	Month      time.Month
	Day        int
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
	Location   *time.Location
}

// SetComponents sets all the components of d from d1
// that are specified in the which bitmask.
// The location is set if either TZOffset or TZName are present.
func (d *Date) SetComponents(d1 *Date, which Components) {
	if which&Year != 0 {
		d.Year = d1.Year
	}
	if which&Month != 0 {
		d.Month = d1.Month
	}
	if which&Day != 0 {
		d.Day = d1.Day
	}
	if which&Hour != 0 {
		d.Hour = d1.Hour
	}
	if which&Minute != 0 {
		d.Minute = d1.Minute
	}
	if which&Second != 0 {
		d.Second = d1.Second
		d.Nanosecond = d1.Nanosecond
	}
	if which&(TZName|TZOffset) != 0 {
		d.Location = d1.Location
	}
}

// Time returns the time corresponding to the given date.
func (d *Date) Time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, d.Nanosecond, d.Location)
}

// TimeDate returns the date information for the given time.
func TimeDate(t time.Time) *Date {
	return &Date{
		Year:       t.Year(),
		Month:      t.Month(),
		Day:        t.Day(),
		Hour:       t.Hour(),
		Minute:     t.Minute(),
		Second:     t.Second(),
		Nanosecond: t.Nanosecond(),
		Location:   t.Location(),
	}
}

var stdComponents = []Components{
	stdLongMonth: Month,
	stdMonth:     Month,
	stdNumMonth:  Month,
	stdZeroMonth: Month,
	// Note: it's not possible to find the specified weekday
	// from time.Parse. To do better, we'd need to reimplement
	// time.Parse and I'm not quite ready to do that yet.
	stdLongWeekDay:  0,
	stdWeekDay:      0,
	stdDay:          Day,
	stdUnderDay:     Day,
	stdZeroDay:      Day,
	stdUnderYearDay: Day,
	stdZeroYearDay:  Day,
	stdHour:         Hour,
	stdHour12:       Hour,
	stdZeroHour12:   Hour,
	stdMinute:       Minute,
	stdZeroMinute:   Minute,
	stdSecond:       Second,
	stdZeroSecond:   Second,
	stdLongYear:     Year,
	stdYear:         Year,
	// It really doesn't make sense to have PM/AM as its own
	// component.
	stdPM:                    0,
	stdpm:                    0,
	stdTZ:                    TZName,
	stdISO8601TZ:             TZOffset,
	stdISO8601SecondsTZ:      TZOffset,
	stdISO8601ShortTZ:        TZOffset,
	stdISO8601ColonTZ:        TZOffset,
	stdISO8601ColonSecondsTZ: TZOffset,
	stdNumTZ:                 TZOffset,
	stdNumSecondsTz:          TZOffset,
	stdNumShortTZ:            TZOffset,
	stdNumColonTZ:            TZOffset,
	stdNumColonSecondsTZ:     TZOffset,
	stdFracSecond0:           0,
	stdFracSecond9:           0,
}

// LayoutComponents returns a bitmask of all the format
// components held in the given time layout string.
func LayoutComponents(layout string) Components {
	var c Components
	for layout != "" {
		_, std, suffix := nextStdChunk(layout)
		if std == 0 {
			break
		}
		layout = suffix
		c |= stdComponents[std]
	}
	return c
}
