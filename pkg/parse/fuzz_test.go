package parse

import (
	"testing"
	"unicode/utf8"
)

func FuzzParse_CrashOrVerySlow(f *testing.F) {
	f.Add("echo")
	f.Add("put $x")
	f.Add("put foo bar | each {|x| echo $x }")
	f.Fuzz(func(t *testing.T, code string) {
		Parse(Source{Name: "fuzz", Code: code}, Config{})
	})
}

func FuzzError_Partial(f *testing.F) {
	for _, test := range testCases {
		f.Add(test.code)
	}
	f.Fuzz(func(t *testing.T, code string) {
		if !utf8.ValidString(code) {
			t.SkipNow()
		}
		_, err := Parse(Source{Name: "fuzz", Code: code}, Config{})
		if err != nil {
			t.SkipNow()
		}
		// If code has no parse error, then every prefix of it (as long as it's
		// valid UTF-8) should have either no parse errors or only partial parse
		// errors.
		for i := range code {
			if i == 0 {
				continue
			}
			prefix := code[:i]
			_, err := Parse(Source{Name: "fuzz", Code: prefix}, Config{})
			if err == nil {
				continue
			}
			for _, err := range UnpackErrors(err) {
				if !err.Partial {
					t.Errorf("prefix %q of valid %q has non-partial error: %v", prefix, code, err)
				}
			}
		}
	})
}
