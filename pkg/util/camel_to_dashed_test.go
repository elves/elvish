package util

import "testing"

var tests = []struct {
	camel string
	want  string
}{
	{"CamelCase", "camel-case"},
	{"camelCase", "-camel-case"},
	{"123", "123"},
	{"你好", "你好"},
}

func TestCamelToDashed(t *testing.T) {
	for _, test := range tests {
		camel, want := test.camel, test.want
		dashed := CamelToDashed(camel)
		if dashed != want {
			t.Errorf("CamelToDashed(%q) => %q, want %q", camel, dashed, want)
		}
	}
}
