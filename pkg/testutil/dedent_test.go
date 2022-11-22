package testutil

import "testing"

var dedentTests = []struct {
	name string
	in   string
	out  string
}{
	{
		name: "no leading newline, no trailing newline",
		in:   " \n  foo\n bar",
		out:  "\n foo\nbar",
	},
	{
		name: "leading newline, no trailing newline",
		in: `
			a
			 b
			c`,
		out: "a\n b\nc",
	},
	{
		name: "leading newline and trailing newline",
		in: `
			a
			 b
			c
			`,
		out: "a\n b\nc\n",
	},
}

func TestDedent(t *testing.T) {
	for _, tc := range dedentTests {
		got := Dedent(tc.in)
		if got != tc.out {
			t.Errorf("Dedent(%q) -> %q, want %q", tc.in, got, tc.out)
		}
	}
}

func TestDedentPanicsOnBadInput(t *testing.T) {
	x := Recover(func() {
		Dedent(`
			a
		b`)
	})
	if x == nil {
		t.Errorf("Dedent did not panic")
	}
}
