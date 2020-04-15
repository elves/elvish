package shell

import (
	"testing"
)

func TestScript_ScriptFile(t *testing.T) {
	f := setup()
	defer f.cleanup()

	writeFile("a.elv", "echo hello")

	Script(f.fds(), []string{"a.elv"}, &ScriptConfig{})

	if out := f.getOut(); out != "hello\n" {
		t.Errorf("got out %q", out)
	}
}
