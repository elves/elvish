package parse

import (
	"testing"
)

func FuzzParse(f *testing.F) {
	f.Add("echo")
	f.Add("put $x")
	f.Add("put foo bar | each {|x| echo $x }")
	f.Fuzz(func(t *testing.T, code string) {
		Parse(Source{Name: "fuzz", Code: code}, Config{})
	})
}
