# godate

A simple command to print dates with Go-style formatting

Usage:

	godate [flags] [[time [+-]duration...]...]

or:

	godate [-alias] tz [name...]

## Flags
-   -abs
    	suppress filling incomplete info from current time
-   -alias
    	when printing time zone matches, also print time zone aliases
-   -f string
    	read times from named file, one per line; - means stdin
-   -i string
    	interpret argument times as this Go-style format (or name) (default "any")
-   -itz string
    	interpret argument times in this time zone location (default local)
-   -o string
    	use Go-style time format string (or name) (default "rfc3339nano")
-   -otz string
    	print times in this time zone location (default local)
-   -u	default to UTC time zone rather than local

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

    ansic       Mon Jan _2 15:04:05 2006
    git         Mon Jan _2 15:04:05 2006 -0700
    go          2006-01-02 15:04:05.999999999 -0700 MST
    kitchen     3:04PM
    rfc1123     Mon, 02 Jan 2006 15:04:05 MST
    rfc1123z    Mon, 02 Jan 2006 15:04:05 -0700
    rfc3339     2006-01-02T15:04:05Z07:00
    rfc3339nano 2006-01-02T15:04:05.999999999Z07:00
    rfc822      02 Jan 06 15:04 MST
    rfc822z     02 Jan 06 15:04 -0700
    rfc850      Monday, 02-Jan-06 15:04:05 MST
    rubydate    Mon Jan 02 15:04:05 -0700 2006
    stamp       Jan _2 15:04:05
    stampmicro  Jan _2 15:04:05.000000
    stampmilli  Jan _2 15:04:05.000
    stampnano   Jan _2 15:04:05.000000000
    unix        custom
    unixdate    Mon Jan _2 15:04:05 MST 2006
    unixmilli   custom
    unixnano    custom

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
