package file

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestFile(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("file", Ns).Ns())
	}
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	TestWithSetup(t, setup,
		That(
			"echo haha > out3", "f = (file:open out3)",
			"slurp < $f", "file:close $f").Puts("haha\n"),

		That(`p = (file:pipe)`, `echo haha > $p `, `pwclose $p`,
			`slurp < $p`, `prclose $p`).Puts("haha\n"),

		That(`p = (file:pipe)`, `echo Zeppelin > $p`, `file:pwclose $p`,
			`echo Sabbath > $p`, `slurp < $p`, `file:prclose $p`).Puts("Zeppelin\n"),

		That(`p = (file:pipe)`, `echo Legolas > $p`, `file:prclose $p`,
			`slurp < $p`).Throws(AnyError),
	)

}
