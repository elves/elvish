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
		"create-windows-special-files-or-skip", createWindowsSpecialFileOrSkip,
		"create-regular-and-symlink-or-skip", func(t *testing.T) {
			testutil.ApplyDir(testutil.Dir{"regular": ""})
			err := os.Symlink("regular", "symlink")
			if err != nil {
				// On Windows we may or may not be able to create a symlink.
				t.Skipf("symlink: %v", err)
			}
		},
		"apply-test-dir-with-symlinks-or-skip", func(t *testing.T) {
			testutil.ApplyDir(testutil.Dir{
				"d": testutil.Dir{
					"f": "",
				},
			})
			var symlinks = []struct {
				path   string
				target string
			}{
				{"d/s-f", "f"},
				{"s-d", "d"},
				{"s-d-f", "d/f"},
				{"s-bad", "bad"},
			}

			for _, link := range symlinks {
				err := os.Symlink(link.target, link.path)
				if err != nil {
					// Creating symlinks requires a special permission on Windows. If
					// the user doesn't have that permission, just skip the whole test.
					t.Skip(err)
				}
			}
		},
	)
}
