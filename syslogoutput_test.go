package main

import (
	"testing"
	"log"
)
func TestSyslogParseTarget(t *testing.T) {
	tests := []struct {
	in string
	proto string
	target string
	formatter string
	framer string
	}{
		{"unix",                     "",     "", "unix",    ""},
		{"unix/rfc5424",             "",     "", "rfc5424", ""},
		{"unix/rfc5424/none",        "",     "", "rfc5424", "none"},
		{"unix/rfc5424/rfc5425",     "",     "", "rfc5424", "rfc5425"},
		{"unix/rfc5424/rfc5426",     "",     "", "rfc5424", "rfc5426"},
		{"127.0.0.1:514",            "tcp",  "127.0.0.1:514", "rfc5424", ""},
		{"tcp:127.0.0.1:514",        "tcp",  "127.0.0.1:514", "rfc5424", ""},
		{"udp:127.0.0.1:514",        "udp",  "127.0.0.1:514", "rfc5424", ""},
		{"XXX:127.0.0.1:514",        "tcp",  "XXX:127.0.0.1:514", "rfc5424", ""},
		{"tcp/XXX/YYY:127.0.0.1:514","tcp",  "127.0.0.1:514", "XXX", "YYY"},
		{"tcp/XXX:127.0.0.1:514",    "tcp",  "127.0.0.1:514", "XXX", ""},
		{"udp/XXX/YYY:127.0.0.1:514","udp",  "127.0.0.1:514", "XXX", "YYY"},
		{"udp/XXX:127.0.0.1:514",    "udp",  "127.0.0.1:514", "XXX", ""},
	}
	for _, test := range tests {
		var sc scriptLine
		sc.target=test.in
		pr,ta,fo,fr := syslogParseTarget(&sc)
		if pr!=test.proto || ta != test.target || fo!=test.formatter || fr != test.framer {
			t.Errorf("targetParse: in %s, got proto=%q target=%q formatter=%q framer=%q, expected proto=%q target=%q formatter=%q framer=%q",
				test.in, pr, ta, fo, fr, test.proto, test.target, test.formatter, test.framer)
		}
	}
}
