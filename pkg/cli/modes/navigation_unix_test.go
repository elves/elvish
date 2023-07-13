//go:build unix

package modes

import (
	"os"
	"syscall"
	"testing"

	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestNavigation_NoPreviewForNamedPipes(t *testing.T) {
	// A previous implementation tried to call os.Open on named pipes, which can
	// block indefinitely. Ensure that this no longer happens.
	testutil.InTempDir(t)
	must.OK(os.Mkdir("d", 0777))
	must.OK(syscall.Mkfifo("d/pipe", 0666))
	must.OK(os.Chdir("d"))
	f := setupNav(t)
	defer f.Stop()

	// Use the default real FS cursor.
	startNavigation(f.App, NavigationSpec{})
	f.TestTTY(t,
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d    pipe          no preview for named\n", Styles,
		"#### ++++++++++++++ !!!!!!!!!!!!!!!!!!!!",
		"                     pipe", Styles,
		"                    !!!!!",
	)
}
