package shell

import (
	"strings"
	"syscall"
	"testing"

	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestSignal_USR1(t *testing.T) {
	Test(t, &Program{},
		ThatElvish("-c", "kill -USR1 $pid").WritesStderrContaining("src.elv.sh/pkg/shell"))
}

func TestSignal_Ignored(t *testing.T) {
	testutil.InTempDir(t)

	Test(t, &Program{},
		ThatElvish("-log", "logCHLD", "-c", "kill -CHLD $pid").DoesNothing())

	wantLogCHLD := "signal " + syscall.SIGCHLD.String()
	if logCHLD := must.ReadFileString("logCHLD"); !strings.Contains(logCHLD, wantLogCHLD) {
		t.Errorf("want log when getting SIGCHLD to contain %q; got:\n%s", wantLogCHLD, logCHLD)
	}
}
