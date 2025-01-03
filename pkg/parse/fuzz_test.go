package parse

import (
	"testing"
	"unicode/utf8"

	"src.elv.sh/pkg/diag"
)

func FuzzParse_CrashOrVerySlow(f *testing.F) {
	f.Add("echo")
	f.Add("put $x")
	f.Add("put foo bar | each {|x| echo $x }")
	f.Fuzz(func(t *testing.T, code string) {
		Parse(Source{Name: "fuzz", Code: code}, Config{})
	})
}

func FuzzPartialError(f *testing.F) {
	for _, test := range testCases {
		f.Add(test.code)
	}
	fuzzPartialError(f, func(src Source) []*Error {
		_, err := Parse(src, Config{})
		return UnpackErrors(err)
	})
}

func fuzzPartialError[T diag.ErrorTag](f *testing.F, fn func(src Source) []*diag.Error[T]) {
	f.Fuzz(func(t *testing.T, code string) {
		if !utf8.ValidString(code) {
			t.SkipNow()
		}
		errs := fn(Source{Name: "fuzz.elv", Code: code})
		if len(errs) > 0 {
			t.SkipNow()
		}
		// If code has no error, then every prefix of it (as long as it's valid
		// UTF-8) should have either no errors or only partial errors.
		for i := range code {
			if i == 0 {
				continue
			}
			prefix := code[:i]
			errs := fn(Source{Name: "fuzz.elv", Code: prefix})
			for _, err := range errs {
				if !err.Partial {
					t.Errorf("prefix %q of valid %q has non-partial error: %v", prefix, code, err)
				}
			}
		}
	})
}
