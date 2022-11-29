package shell

import (
	"testing"

	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestScript(t *testing.T) {
	setupCleanHomePaths(t)
	testutil.InTempDir(t)
	must.WriteFile("hello.elv", "echo hello")
	must.WriteFile("invalid-utf8.elv", "\xff")

	Test(t, &Program{},
		ThatElvish("hello.elv").WritesStdout("hello\n"),
		ThatElvish("-c", "echo hello").WritesStdout("hello\n"),

		ThatElvish("invalid-utf8.elv").
			ExitsWith(2).
			WritesStderrContaining("cannot read script"),
		ThatElvish("non-existent.elv").
			ExitsWith(2).
			WritesStderrContaining("cannot read script"),

		// parse error
		ThatElvish("-c", "echo [").
			ExitsWith(2).
			WritesStderrContaining("Parse error"),
		// parse error with -compileonly
		ThatElvish("-compileonly", "-c", "echo [").
			ExitsWith(2).
			WritesStderrContaining("Parse error"),
		// parse error with -compileonly -json
		ThatElvish("-compileonly", "-json", "-c", "echo [").
			ExitsWith(2).
			WritesStdout(`[{"fileName":"code from -c","start":6,"end":6,"message":"should be ']'"}]`+"\n"),
		// multiple parse errors with -compileonly -json
		ThatElvish("-compileonly", "-json", "-c", "echo [{").
			ExitsWith(2).
			WritesStdout(`[{"fileName":"code from -c","start":7,"end":7,"message":"should be ',' or '}'"},{"fileName":"code from -c","start":7,"end":7,"message":"should be ']'"}]`+"\n"),

		// compilation error
		ThatElvish("-c", "echo $a").
			ExitsWith(2).
			WritesStderrContaining("Compilation error"),
		// compilation error with -compileonly
		ThatElvish("-compileonly", "-c", "echo $a").
			ExitsWith(2).
			WritesStderrContaining("Compilation error"),
		// compilation error with -compileonly -json
		ThatElvish("-compileonly", "-json", "-c", "echo $a").
			ExitsWith(2).
			WritesStdout(`[{"fileName":"code from -c","start":5,"end":7,"message":"variable $a not found"}]`+"\n"),
		// parse error and compilation error with -compileonly
		ThatElvish("-compileonly", "-json", "-c", "echo [$a").
			ExitsWith(2).
			WritesStdout(`[{"fileName":"code from -c","start":8,"end":8,"message":"should be ']'"},{"fileName":"code from -c","start":6,"end":8,"message":"variable $a not found"}]`+"\n"),

		// exception
		ThatElvish("-c", "fail failure").
			ExitsWith(2).
			WritesStdout("").
			WritesStderrContaining("fail failure"),
		// exception with -compileonly
		ThatElvish("-compileonly", "-c", "fail failure").
			ExitsWith(0),
	)
}
