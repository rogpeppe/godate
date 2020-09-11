package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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
	inFormat  = flag.String("i", "unix", "interpret argument times as this Go-style format (or name)")
	file      = flag.String("f", "", "read times from named file, one per line; - means stdin")
	tzIn      = flag.String("itz", "", "interpret argument times in this time zone location (default local)")
	tzOut     = flag.String("otz", "", "print times in this time zone location (default local)")
	alias     = flag.Bool("alias", false, "when printing time zone matches, also print time zone aliases")
	utc       = flag.Bool("u", false, "default to UTC time zone rather than local")
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
	"unixnano":    "custom",
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

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage:
	godate [flags] [[time [+-]duration...]...]
or:
	godate tz [name...]
Flags:
`[1:])
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, `

This command parses and prints times in arbitrary formats and time zones.
Each argument is a time followed by an arbitrary number of offset
arguments adjusting the time. Godate reads all the times
according to the format specified by the -i flag, adjusts them by
the offsets, and prints them in the format specified by the -o flag.
The special time "now" is recognized as the current time.

As a special case, if the first argument is "tz", then godate prints all
the available time zones (note: this uses an internal list and may not
exactly match the system-provided time zones). If any arguments are
provided after "tz", only time zones matching those arguments (see below
for timezone matching behavior) are printed.

The format for a duration is either as accepted by Go's ParseDuration
function (see https://golang.org/pkg/time/#Time.ParseDuration for details)
or a similar format that specifies years (year, y), months (month, mo),
weeks (week, w) or days (day, d). For example, this would print
the local time 1 month and 3 days hence and 20 minutes before the
current time:

	godate now +1month3days -20m

Note that year, month, and week durations cannot be mixed with
other duration kinds in the same argument.

By default godate prints the current time in RFC3339 format in
the local time zone. The -o flag can be used to change the format
that is printed (see https://golang.org/pkg/time/#Time.Format
for details). The reference date is:

	Mon Jan 2 15:04:05 -0700 MST 2006

The format may be the name of one of the predefined format
constants in the time package (case-insensitive), in which case that format will be used.
The supported formats are these:

`[1:])
	type format struct {
		name   string
		format string
	}
	var formats []format
	for name, f := range knownFormats {
		formats = append(formats, format{name, f})
	}
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].name < formats[j].name
	})
	w := tabwriter.NewWriter(os.Stderr, 4, 4, 1, ' ', 0)
	for _, f := range formats {
		fmt.Fprintf(w, "\t%s\t%s\n", f.name, f.format)
	}
	w.Flush()

	fmt.Fprintf(os.Stderr, `

The unix, unixmilla and unixnano formats are special cases that print the number of seconds,
milliseconds or nanoseconds since the Unix epoch (Jan 1st 1970). The "go" format is the
format used by the time package to print times by default.

When one or more arguments are provided, they will be used as the time
to print instead of the current time. The -in flag can be used to specify
what format to interpret these arguments in. Again, unix and unixnano
can be used to specify input in seconds or nanoseconds since the Unix epoch.

Time zones can be specified with the -itz and -otz flags. As a convenience,
if the specified zone does not exactly match one of the known zones,
a case-insensitive match is tried, and then a substring match.
If the result is unambiguous, the matching time zone is used
(for example "-otz london" can be used to select the "Europe/London"
time zone).
`[1:])
	os.Exit(2)
}

func timeParser() (func(s string) (time.Time, error), error) {
	tz, err := loadLocation(*tzIn)
	if err != nil {
		return nil, err
	}
	if tz == nil {
		tz = time.Local
	}
	format := *inFormat
	var parser func(s string) (time.Time, error)
	if format1, ok := knownFormats[strings.ToLower(format)]; ok {
		if format1 == "custom" {
			parser = func(s string) (time.Time, error) {
				return parseCustom(format, s, tz)
			}
		} else {
			format = format1
		}
	}
	if parser == nil {
		parser = func(s string) (time.Time, error) {
			return time.ParseInLocation(format, s, tz)
		}
	}
	now := time.Now().In(tz)
	return func(s string) (time.Time, error) {
		if s == "now" {
			return now, nil
		}
		return parser(s)
	}, nil
}

func parseCustom(format, s string, tz *time.Location) (time.Time, error) {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unix?: %v", err)
	}
	switch format {
	case "unix":
		return time.Unix(ts, 0).In(tz), nil
	case "unixnano":
		return time.Unix(0, ts).In(tz), nil
	case "unixmilli":
		return time.Unix(ts/1000, (ts%1000)*1e6).In(tz), nil
	default:
		panic("unknown unix time format")
	}
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
	case "unixnano":
		return fmt.Sprint(t.UnixNano())
	default:
		panic("unknown unix time format")
	}
}

func printZones(args []string) {
	if len(args) == 0 {
		args = []string{""}
	}
	var tzs []string
	zones := make(map[string]bool)
	for _, arg := range args {
		for _, tz := range zoneMatch(arg) {
			zones[tz] = true
		}
	}
	if len(zones) == 0 {
		fatalf("no matching time zones found")
	}
	tzs = make([]string, 0, len(zones))
	for zone := range zones {
		tzs = append(tzs, zone)
	}
	sort.Strings(tzs)
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
	for _, tz := range tzs {
		linked := zoneNames[tz]
		if !*alias || linked == "" {
			fmt.Fprintf(w, "%s\n", tz)
		} else {
			fmt.Fprintf(w, "%s\t%s\n", tz, linked)
		}
	}
	w.Flush()
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
