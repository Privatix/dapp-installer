package util

import "testing"

type testpair struct {
	version string
	value   int64
}

var tests = []testpair{
	{"0.0.0", 0},
	{"0.0.1", 10000},
	{"0.0.0.5", 5},
	{"0.11.0", 1100000000},
	{"1.0.0.57", 1000000000057},
	{"1.0", 1000000000000},
}

func TestParseVersion(t *testing.T) {
	for _, pair := range tests {
		v := ParseVersion(pair.version)
		if v != pair.value {
			t.Error(
				"For", pair.version,
				"expected", pair.value,
				"got", v,
			)
		}
	}
}
