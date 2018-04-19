# godate

A simple command to print dates with Go-style formatting

Usage: godate [flags] [time...]

## Flags

- -f *string*
    	use Go-style time format string (or name) (default "rfc3339nano")
- -in *string*
    	interpret argument time as this Go-style format (or name) (default "unix")
- -l
	print local time (default is UTC)

By default godate prints the current time in RFC3339 format in
the UTC time zone. The -f flag can be used to change the format
that is printed (see https://golang.org/pkg/time/#Time.Format
for details). The reference date is:

	Mon Jan 2 15:04:05 -0700 MST 2006

The format may be the name of one of the predefined format
constants in the time package (case-insensitive), in which case that format will be used.
The supported formats are these:

    ansic       Mon Jan _2 15:04:05 2006
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
    unixnano    custom

The unix and unixnano formats are special cases that print the number of seconds
or nanoseconds since the Unix epoch (Jan 1st 1970). The "go" format is the
format used by the time package to print times by default.

When one or more arguments are provided, they will be used as the time
to print instead of the current time. The -in flag can be used to specify
what format to interpret these arguments in. Again, unix and unixnano
can be used to specify input in seconds or nanoseconds since the Unix epoch.
