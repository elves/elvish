package eval

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

func FuzzCheck(f *testing.F) {
	f.Add("echo")
	f.Add("put $x")
	f.Add("put foo bar | each {|x| echo $x }")
	f.Fuzz(func(t *testing.T, code string) {
		NewEvaler().Check(parse.Source{Name: "[fuzz]", Code: code}, nil)
	})
}
