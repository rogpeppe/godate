package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	local  = flag.Bool("l", false, "print local time (default is UTC)")
	format = flag.String("f", "rfc3339nano", "use Go-style time format string (or name)")
	in     = flag.String("in", "unix", "interpret argument time as this Go-style format (or name)")
)

var knownFormats = map[string]string{
	"ansic":       time.ANSIC,
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
	"unixnano":    "custom",
}

func main() {
	flag.Usage = usage
	flag.Parse()
	var times []time.Time
	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			t, err := parseTime(*in, arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
				os.Exit(2)
			}
			times = append(times, t)
		}
	} else {
		times = []time.Time{
			time.Now(),
		}
	}
	for _, t := range times {
		if !*local {
			t = t.UTC()
		}
		fmt.Printf("%s\n", formatTime(t, *format))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: godate [flags] [time...]
Flags:
`[1:])
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, `

By default godate prints the current time in RFC3339 format in
the UTC time zone. The -f flag can be used to change the format
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

The unix and unixnano formats are special cases that print the number of seconds
or nanoseconds since the Unix epoch (Jan 1st 1970). The "go" format is the
format used by the time package to print times by default.

When one or more arguments are provided, they will be used as the time
to print instead of the current time. The -in flag can be used to specify
what format to interpret these arguments in. Again, unix and unixnano
can be used to specify input in seconds or nanoseconds since the Unix epoch.
`[1:])
	os.Exit(2)
}

func parseTime(format string, s string) (time.Time, error) {
	if format1, ok := knownFormats[strings.ToLower(format)]; ok {
		if format1 == "custom" {
			return parseUnix(format, s)
		}
		format = format1
	}
	return time.Parse(format, s)
}

func parseUnix(format, s string) (time.Time, error) {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	switch format {
	case "unix":
		return time.Unix(ts, 0), nil
	case "unixnano":
		return time.Unix(0, ts), nil
	default:
		panic("unknown unix time format")
	}
}

func formatTime(t time.Time, format string) string {
	if format1, ok := knownFormats[strings.ToLower(format)]; ok {
		if format1 == "custom" {
			return formatUnix(t, format)
		}
		format = format1
	}
	return t.Format(format)
}

func formatUnix(t time.Time, format string) string {
	switch format {
	case "unix":
		return fmt.Sprint(t.Unix())
	case "unixnano":
		return fmt.Sprint(t.UnixNano())
	default:
		panic("unknown unix time format")
	}
}
