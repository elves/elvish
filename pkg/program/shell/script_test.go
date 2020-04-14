package shell

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/util"
)

func TestScript(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()
	err := ioutil.WriteFile("a.elv", []byte("echo hello > out"), 0600)
	if err != nil {
		panic(err)
	}

	Script(
		[3]*os.File{eval.DevNull, eval.DevNull, eval.DevNull},
		[]string{"a.elv"}, &ScriptConfig{})

	out, err := ioutil.ReadFile("out")
	if err != nil {
		panic(err)
	}
	if string(out) != "hello\n" {
		t.Errorf("got out %q", out)
	}
}
