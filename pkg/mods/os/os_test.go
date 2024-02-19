package os_test

import (
	"embed"
	"net"
	"os"
	"strconv"
	"testing"

	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"umask", func(t *testing.T, arg string) {
			testutil.Umask(t, must.OK1(strconv.Atoi(arg)))
		},
		"mkfifo-or-skip", mkFifoOrSkip,
		"mksock-or-skip", func(t *testing.T, s string) {
			listener, err := net.Listen("unix", "./sock")
			if err != nil {
				t.Skipf("can't listen to UNIX socket: %v", err)
			}
			t.Cleanup(func() { listener.Close() })
		},
		"only-if-can-create-symlink", func(t *testing.T) {
			testutil.ApplyDir(testutil.Dir{"test-file": ""})
			err := os.Symlink("test-file", "test-symlink")
			if err != nil {
				// On Windows we may or may not be able to create a symlink.
				t.Skipf("symlink: %v", err)
			}
			must.OK(os.Remove("test-file"))
			must.OK(os.Remove("test-symlink"))
		},
		"create-windows-special-files-or-skip", createWindowsSpecialFileOrSkip,
	)
}
