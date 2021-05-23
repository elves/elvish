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
		That(`
			echo haha > out3
			f = (file:open out3)
			slurp < $f
			file:close $f
		`).Puts("haha\n"),

		That(`
			p = (file:pipe)
			echo haha > $p
			file:close $p[w]
			slurp < $p
			file:close $p[r]
		`).Puts("haha\n"),

		That(`
			p = (file:pipe)
			echo Zeppelin > $p
			file:close $p[w]
			echo Sabbath > $p
			slurp < $p
			file:close $p[r]
		`).Puts("Zeppelin\n"),

		That(`
			p = (file:pipe)
			echo Legolas > $p
			file:close $p[r]
			slurp < $p
		`).Throws(AnyError),
	)
}
