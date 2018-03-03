package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval/vals"
)

func TestBuiltinFnEnv(t *testing.T) {
	oldpath := os.Getenv("PATH")
	listSep := string(os.PathListSeparator)
	runTests(t, []Test{
		{`get-env var`, want{err: ErrMissingEnvVar}},
		{`set-env var test1`, want{}},
		{`get-env var`, want{out: strs("test1")}},
		{`put $E:var`, want{out: strs("test1")}},
		{`set-env var test2`, want{}},
		{`get-env var`, want{out: strs("test2")}},
		{`put $E:var`, want{out: strs("test2")}},

		{`set-env PATH /test-path`, want{}},
		{`put $paths`, want{out: []interface{}{
			vals.MakeList(strs("/test-path")...)}}},
		{`paths = [/test-path2 $@paths]`, want{}},
		{`get-env PATH`, want{out: strs(
			"/test-path2" + listSep + "/test-path")}},
	})
	os.Setenv("PATH", oldpath)
}
