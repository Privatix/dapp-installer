package util

import (
	"encoding/json"
	"testing"
)

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

func TestIsURL(t *testing.T) {
	var paths = map[string]bool{
		"http://localhost/dapp.zip":             true,
		"/home/dappcore/":                       false,
		"https://privatix.io/download/dapp.zip": true,
		"c:/program files/privatix/":            false,
	}
	for p, ok := range paths {
		if ok != IsURL(p) {
			t.Error(
				"For", p,
				"expected", ok,
				"got", !ok,
			)
		}
	}
}

func TestMatchAddr(t *testing.T) {
	testaddrs := []Addr{
		Addr{Address: "localhost:8000", Host: "localhost", Port: "8000"},
		Addr{Address: "localhost:8888", Host: "localhost", Port: "8888"},
	}
	teststr := `
	"SessionServer": {
		"Addr": "localhost:8000",
		"TLS": null
	},
	"UI": {
		"Addr": "localhost:8888",
		"AllowedOrigins": ["*"],
		"Handler": {},
		"TLS": null
	}`
	addrs := MatchAddr(teststr)
	if len(addrs) != 2 {
		t.Error("For", teststr, "expected 2 matches", "got", len(addrs))
	}
	for i, addr := range addrs {
		if addr.Address != testaddrs[i].Address {
			t.Error("Expected", testaddrs[i], "got", addr)
		}
	}
}

func TestMergeJSON(t *testing.T) {
	dst := `{"A":1,"B":{"b":2,"bb":3,"bbbb":4},"C":true}`
	src := `{"B":{"b":2,"bb":30,"bbb":4},"C":false}`
	result := `{"A":1,"B":{"b":2,"bb":30,"bbbb":4},"C":false}`

	var dstMap, srcMap map[string]interface{}
	json.Unmarshal([]byte(dst), &dstMap)
	json.Unmarshal([]byte(src), &srcMap)

	m := &merger{}
	m.merge(dstMap, srcMap)

	r, _ := json.Marshal(dstMap)

	if string(r) != result {
		t.Error("Expected", result, "got", string(r))
	}
}

func TestMergeJSONWithExceptions(t *testing.T) {
	dst := `{"A":1,"B":{"b":2,"bb":3,"bbbb":4},"C":true}`
	src := `{"B":{"b":2,"bb":30,"bbb":4},"C":false}`
	result := `{"A":1,"B":{"b":2,"bb":3,"bbbb":4},"C":true}`

	var dstMap, srcMap map[string]interface{}
	json.Unmarshal([]byte(dst), &dstMap)
	json.Unmarshal([]byte(src), &srcMap)

	m := &merger{exceptions: []string{"bb", "C", "bbbb"}}
	m.merge(dstMap, srcMap)

	r, _ := json.Marshal(dstMap)

	if string(r) != result {
		t.Error("Expected", result, "got", string(r))
	}
}
