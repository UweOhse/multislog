package main

import "testing"

func TestMatch(t *testing.T) {
	tab := []struct {
		pat string
		str string
		res bool
	}{
		{"a", "a", true},
		{"a", "b", false},
		{"a*", "abc", true},
		{"a*c", "abc", true},
		{"*c", "abc", true},
		{"*", "a", true},
		{"*", "ab", true},
		{"**", "ab", false},
		{"**", "a*", true},
		// match handles literal string
		{"one", "one", true},
		{"one", "", false},
		{"one", "on", false},
		{"one", "onf", false},
		{"one", "one*", false},
		{"one", "onetwo", false},

		// match handles empty string
		{"", "", true},
		{"", "x", false},

		// match handles full-line wildcard
		{"*", "", true},
		{"*", "x", true},
		{"*", "*", true},
		{"*", "one", true},

		// match handles ending wildcard
		{"one*", "one", true},
		{"one*", "one*", true},
		{"one*", "onetwo", true},
		{"one*", "", false},
		{"one*", "x", false},
		{"one*", "on", false},
		{"one*", "onf", false},

		// match handles wildcard termination
		{"* one", " one", true},
		{"* one", "x one", true},
		{"* one", "* one", true},
		{"* one", "xy one", true},
		{"* one", "one", false},
		{"* one", " two", false},
		{"* one", "  one", false},
		{"* one", "xy one ", false},

		// match handles multiple wildcards
		{"* * one", "  one", true},
		{"* * one", "x  one", true},
		{"* * one", " y one", true},
		{"* * one", "x y one", true},
		{"* * one", "one", false},
		{"* * one", " one", false},
		{"* * one", "   one", false},
		{"* *  one", "   one", true},
	}
	for _, entry := range tab {
		res := match(entry.pat, entry.str)
		if res != entry.res {
			t.Errorf("match(%s,%s) returned not %v, but %v",
				entry.pat, entry.str, entry.res, res)
		}
	}
}
