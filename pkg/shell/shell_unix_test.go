//go:build unix

package shell

import (
	"fmt"
	"strings"
	"syscall"
	"testing"
	"time"

	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestSignal_USR1(t *testing.T) {
	Test(t, &Program{},
		ThatElvish("-c", killCmd("USR1")).WritesStderrContaining("src.elv.sh/pkg/shell"))
}

func TestSignal_Ignored(t *testing.T) {
	testutil.InTempDir(t)

	Test(t, &Program{},
		ThatElvish("-log", "logCHLD", "-c", killCmd("CHLD")).DoesNothing())

	wantLogCHLD := "signal " + syscall.SIGCHLD.String()
	if logCHLD := must.ReadFileString("logCHLD"); !strings.Contains(logCHLD, wantLogCHLD) {
		t.Errorf("want log when getting SIGCHLD to contain %q; got:\n%s", wantLogCHLD, logCHLD)
	}
}

func killCmd(name string) string {
	// Add a delay after kill to ensure that the signal is handled.
	return fmt.Sprintf("kill -%v $pid; sleep %v", name, testutil.Scaled(10*time.Millisecond))
}
