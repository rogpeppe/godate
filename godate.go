package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rogpeppe/godate/timeformat"
)

//go:generate bash getzones.bash

// Possible TODOs:
// - support for rounding and truncation (how would that work syntactically?).
//
// 	godate now +5m round:1h  trunc:1day
//
// 	rounding for date durations might be hard.

var (
	outFormat = flag.String("o", "rfc3339nano", "use Go-style time format string (or name)")
	inFormat  = flag.String("i", "any", "interpret argument times as this Go-style format (or name)")
	file      = flag.String("f", "", "read times from named file, one per line; - means stdin")
	tzIn      = flag.String("itz", "", "interpret argument times in this time zone location (default local)")
	tzOut     = flag.String("otz", "", "print times in this time zone location (default local)")
	alias     = flag.Bool("alias", false, "when printing time zone matches, also print time zone aliases")
	utc       = flag.Bool("u", false, "default to UTC time zone rather than local")
	abs       = flag.Bool("abs", false, "suppress filling incomplete info from current time")
)

var knownFormats = map[string]string{
	"ansic":       time.ANSIC,
	"git":         "Mon Jan _2 15:04:05 2006 -0700",
	"unixdate":    time.UnixDate,
	"rubydate":    time.RubyDate,
	"rfc822":      time.RFC822,
	"rfc822z":     time.RFC822Z,
	"rfc850":      time.RFC850,
	"rfc1123":     time.RFC1123,
	"rfc1123z":    time.RFC1123Z,
	"rfc3339":     time.RFC3339,
	"rfc3339nano": time.RFC3339Nano,
	"kitchen":     time.Kitchen,
	"stamp":       time.Stamp,
	"stampmilli":  time.StampMilli,
	"stampmicro":  time.StampMicro,
	"stampnano":   time.StampNano,
	"go":          "2006-01-02 15:04:05.999999999 -0700 MST",
	"unix":        "custom",
	"unixmilli":   "custom",
	"unixmicro":   "custom",
	"unixnano":    "custom",
	"any":         "custom",
}

func main() {
	flag.Usage = usage
	flag.Parse()
	formatTime, err := formatter()
	if err != nil {
		fatalf("%v", err)
	}
	parseTime, err := timeParser()
	if err != nil {
		fatalf("%v", err)
	}
	if *file != "" {
		if flag.NArg() > 0 {
			fatalf("cannot provide arguments with -file flag")
		}
		f := os.Stdin
		if *file != "-" {
			var err error
			f, err = os.Open(*file)
			if err != nil {
				fatalf("%v, err")
			}
		}
		for scanner := bufio.NewScanner(f); scanner.Scan(); {
			t, err := parseTime(scanner.Text())
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse error on %q: %v\n", scanner.Text(), err)
				continue
			}
			fmt.Printf("%s\n", formatTime(t))
		}
		return
	}
	var times []time.Time
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"now"}
	}
	if args[0] == "tz" {
		printZones(args[1:])
		return
	}
	i := 0
	for i < len(args) {
		arg := args[i]
		t, err := parseTime(arg)
		if err != nil {
			fatalf("parse error on %q: %v", arg, err)
		}
		i++
		for i < len(args) {
			arg := args[i]
			if arg != "" && (arg[0] == '-' || arg[0] == '+') {
				d, err := parseDelta(arg)
				if err != nil {
					fatalf("parse error on duration %q: %v", arg, err)
				}
				t = d.add(t)
				i++
			} else {
				break
			}
		}
		times = append(times, t)
	}
	for _, t := range times {
		fmt.Printf("%s\n", formatTime(t))
	}
}

type delta struct {
	year, month, day int
	duration         time.Duration
}

func parseDelta(s string) (delta, error) {
	orig := s
	dur, err := time.ParseDuration(s)
	if err == nil {
		return delta{
			duration: dur,
		}, nil
	}
	neg := false
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	if s == "" {
		return delta{}, fmt.Errorf("invalid duration %q", orig)
	}
	var d delta
	for s != "" {
		var v int32
		v, s, err = leadingInt(s)
		if err != nil {
			return delta{}, fmt.Errorf("invalid duration %q", orig)
		}
		if neg {
			v = -v
		}
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return delta{}, fmt.Errorf("missing unit in duration %q", orig)
		}
		u := s[:i]
		s = s[i:]
		switch u {
		case "y", "year", "years":
			d.year += int(v)
		case "mo", "month", "months":
			d.month += int(v)
		case "d", "day", "days":
			d.day += int(v)
		case "w", "week", "weeks":
			d.day += 7 * int(v)
		default:
			return delta{}, fmt.Errorf("time unknown unit in duration %q", orig)
		}
	}
	return d, nil
}

func (d delta) add(t time.Time) time.Time {
	if d.duration != 0 {
		return t.Add(d.duration)
	}
	return t.AddDate(d.year, d.month, d.day)
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int32, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<31-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int32(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

var errLeadingInt = errors.New("bad [0-9]*") // never printed

func timeParser() (func(s string) (time.Time, error), error) {
	tz, err := loadLocation(*tzIn)
	if err != nil {
		return nil, err
	}
	if tz == nil {
		tz = time.Local
	}
	now := time.Now().In(tz)
	format := *inFormat
	var parser func(s string) (time.Time, error)
	if format1, ok := knownFormats[strings.ToLower(format)]; ok {
		if format1 == "custom" {
			parser = func(s string) (time.Time, error) {
				return parseCustom(format, s, tz, now)
			}
		} else {
			format = format1
		}
	}
	if parser == nil {
		components := timeformat.LayoutComponents(format)
		now := time.Now()
		parser = func(s string) (time.Time, error) {
			t, err := time.ParseInLocation(format, s, tz)
			if err != nil || *abs {
				return t, err
			}
			return relativeTime(t, components, now), nil
		}
	}
	return func(s string) (time.Time, error) {
		if s == "now" {
			return now, nil
		}
		return parser(s)
	}, nil
}

var componentsBySignificance = []timeformat.Components{
	timeformat.Year,
	timeformat.Month,
	timeformat.Day,
	timeformat.Hour,
	timeformat.Minute,
	timeformat.Second,
}

func relativeTime(t time.Time, components timeformat.Components, now time.Time) time.Time {
	td := timeformat.TimeDate(t)
	nowd := timeformat.TimeDate(now)
	var toSet timeformat.Components
	for _, c := range componentsBySignificance {
		if components&c != 0 {
			break
		}
		toSet |= c
	}
	td.SetComponents(nowd, toSet)
	return td.Time()
}

func parseCustom(format, s string, tz *time.Location, now time.Time) (time.Time, error) {
	if format == "any" {
		return parseAny(s, tz, now)
	}
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unix?: %v", err)
	}
	switch format {
	case "unix":
		return time.Unix(ts, 0).In(tz), nil
	case "unixnano":
		return time.Unix(0, ts).In(tz), nil
	case "unixmicro":
		return time.Unix(ts/1e6, (ts%1e6)*1e3).In(tz), nil
	case "unixmilli":
		return time.Unix(ts/1e3, (ts%1e3)*1e6).In(tz), nil
	default:
		panic("unknown unix time format")
	}
}

var anyFormats = []string{
	"2006",
	"2006-01-02",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"01-02 15:04",
	"Jan 1",
	"Jan 1 15:04",
	"Jan 1 15:04:05",
	"1 Jan",
	"1 Jan 15:04",
	"1 Jan 15:04:05",
	"15:04",
	"15:04:05",
	"3pm",
	"3PM",
	"3:04pm",
	"3:04PM",
	"3:04:05pm",
	"3:04:05PM",
}

var unixFormats = []string{"unixnano", "unixmicro", "unixmilli", "unix"}

func parseAny(s string, tz *time.Location, now time.Time) (time.Time, error) {
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		for _, format := range unixFormats {
			t, err := parseCustom(format, s, tz, now)
			if err != nil {
				continue
			}
			if format == "unix" {
				return t, nil
			}
			tooSmall := t.Year() == 1970 && t.Month() == time.January ||
				t.Year() == 1969 && t.Month() != time.December
			if !tooSmall {
				return t, nil
			}
		}
	}
	for _, format := range anyFormats {
		t, err := time.ParseInLocation(format, s, tz)
		if err != nil {
			continue
		}
		return relativeTime(t, timeformat.LayoutComponents(format), now), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as arbitrary format", s)
}

func formatter() (func(time.Time) string, error) {
	tz, err := loadLocation(*tzOut)
	if err != nil {
		return nil, err
	}
	toTZ := func(t time.Time) time.Time {
		if tz == nil {
			return t
		}
		return t.In(tz)
	}
	format := *outFormat
	if format1, ok := knownFormats[strings.ToLower(format)]; ok {
		if format1 == "custom" {
			return func(t time.Time) string {
				return formatCustom(toTZ(t), format)
			}, nil
		}
		format = format1
	}
	return func(t time.Time) string {
		return toTZ(t).Format(format)
	}, nil
}

func loadLocation(loc string) (*time.Location, error) {
	switch strings.ToLower(loc) {
	case "local":
		return time.Local, nil
	case "utc":
		return time.UTC, nil
	case "":
		if *utc {
			return time.UTC, nil
		}
		return nil, nil
	}
	tz, err := time.LoadLocation(loc)
	if err == nil {
		return tz, nil
	}
	available := zoneMatch(loc)
	if len(available) > 1 {
		// If the zones are actually all referring to the same underlying time zone, then
		// allow it (for example, "samoa" could match both "US/Samoa" and "Pacific/Samoa"
		// but they're actually both the same)
		if !allIdenticalZones(available) {
			return nil, fmt.Errorf("ambiguous time zone %q (%d matches; use 'godate tz %s' to see them)", loc, len(available), loc)
		}
	}
	if len(available) == 0 {
		return nil, err
	}
	tz, err = time.LoadLocation(available[0])
	if err != nil {
		return nil, fmt.Errorf("time zone %s not available in system time zone database: %v", available[0], err)
	}
	return tz, nil
}

func formatCustom(t time.Time, format string) string {
	switch format {
	case "unix":
		return fmt.Sprint(t.Unix())
	case "unixmilli":
		return fmt.Sprint(int64(time.Duration(t.UnixNano()) / time.Millisecond))
	case "unixmicro":
		return fmt.Sprint(int64(time.Duration(t.UnixNano()) / time.Microsecond))
	case "unixnano":
		return fmt.Sprint(t.UnixNano())
	case "any":
		// Arbitrary.
		return t.Format(time.RFC3339)
	default:
		panic("unknown unix time format")
	}
}

func zoneMatch(tz string) []string {
	if _, ok := zoneNames[tz]; ok {
		return []string{tz}
	}
	var matches []string
	for name := range zoneNames {
		if strings.EqualFold(name, tz) {
			matches = append(matches, name)
		}
	}
	if len(matches) > 0 {
		return matches
	}
	tz = strings.ToLower(tz)
	for name := range zoneNames {
		if strings.Contains(strings.ToLower(name), tz) {
			matches = append(matches, name)
		}
	}
	return matches
}

func allIdenticalZones(tzs []string) bool {
	if len(tzs) < 2 {
		return true
	}
	ctz := canonicalTimezone(tzs[0])
	for _, tz := range tzs[1:] {
		if canonicalTimezone(tz) != ctz {
			return false
		}
	}
	return true
}

func canonicalTimezone(tz string) string {
	for {
		link := zoneNames[tz]
		if link == "" {
			return tz
		}
		tz = link
	}
}

func fatalf(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(f, a...))
	os.Exit(1)
}
