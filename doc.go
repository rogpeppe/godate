package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

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

When the input time is missing some parts, any more significant parts
will be filled in using the current time. So, for example,
"godate -i 15:04 17:01" will print a time with the current date
but the time 15:04. Less significant parts will be left zero,
so "godate -i 2006 1973" will print "1973-01-01T00:00:00Z".
Using the -abs flag suppresses this behavior.

The default input time format is the special format "any" which
interprets the time according to the first format that parses OK from
the following list or, if the time consists only of digits, the
first of unixnano, unixmilli or unix that parses as a time outside
January 1970:

	2006
	2006-01-02
	2006-01-02T15:04:05Z
	2006-01-02 15:04:05Z
	2006-01-02T15:04:05
	2006-01-02 15:04:05
	01-02 15:04
	Jan 1
	Jan 1 15:04
	Jan 1 15:04:05
	1 Jan
	1 Jan 15:04
	1 Jan 15:04:05
	15:04
	15:04:05

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

The format may also be the name of one of the predefined format
constants in the time package (case-insensitive), in which case that format will be used.
The supported predefined names are:

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

The unix, unixmilli, unixmicro and unixnano formats are special cases that print the number of seconds,
milliseconds, microseconds or nanoseconds since the Unix epoch (Jan 1st 1970). The "go" format is the
format used by the time package to print times by default.

When one or more arguments are provided, they will be used as the time
to print instead of the current time. The -in flag can be used to specify
what format to interpret these arguments in. Again, unix, unixmilli, unixmicro and unixnano
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
