package parse_test

import (
	"testing"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

func TestSourceAsMap(t *testing.T) {
	vals.TestValue(t, parse.Source{Name: "[tty]", Code: "echo"}).
		Kind("map").
		Repr("[&code=echo &is-file=$false &name='[tty]']").
		AllKeys("name", "code", "is-file")

	vals.TestValue(t, parse.Source{Name: "/etc/rc.elv", Code: "echo", IsFile: true}).
		Index("is-file", true)
}
