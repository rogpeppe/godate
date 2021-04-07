package timeformat

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

var layoutComponentsTests = []struct {
	layout     string
	components Components
}{{
	layout:     time.RFC3339,
	components: Year | Month | Day | Hour | Minute | Second | TZOffset,
}, {
	layout:     time.Kitchen,
	components: Hour | Minute,
}, {
	layout:     "2006-01-02",
	components: Year | Month | Day,
}, {
	layout:     "2006",
	components: Year,
}, {
	layout:     "2006 MST",
	components: Year | TZName,
}, {
	layout:     "2006 -07:00",
	components: Year | TZOffset,
}}

func TestLayoutComponents(t *testing.T) {
	c := qt.New(t)
	for _, test := range layoutComponentsTests {
		c.Run(test.layout, func(c *qt.C) {
			c.Assert(LayoutComponents(test.layout), qt.Equals, test.components)
		})
	}
}
