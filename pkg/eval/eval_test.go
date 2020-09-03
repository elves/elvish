package eval

import (
	"bytes"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/testutil"
)

func TestBuiltinPid(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	builtinPid := vals.ToString(builtinNs["pid"].Get())
	if builtinPid != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, builtinPid, pid)
	}
}

func TestNumBgJobs(t *testing.T) {
	Test(t,
		That("put $num-bg-jobs").Puts("0"),
		// TODO(xiaq): Test cases where $num-bg-jobs > 0. This cannot be done
		// with { put $num-bg-jobs }& because the output channel may have
		// already been closed when the closure is run.
	)
}

func TestEvalTimeDeprecate(t *testing.T) {
	restore := prog.SetShowDeprecations(true)
	defer restore()
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	TestWithSetup(t, func(ev *Evaler) {
		ev.Global.AddGoFn("", "dep", func(fm *Frame) {
			fm.Deprecate("deprecated", nil)
		})
	},
		That("dep").PrintsStderrWith("deprecated"),
		// Deprecation message is only shown once.
		That("dep 2> tmp.txt; dep").DoesNothing(),
	)
}

func TestCompileTimeDeprecation(t *testing.T) {
	restore := prog.SetShowDeprecations(true)
	defer restore()

	ev := NewEvaler()
	r, w := mustPipe()
	_, err := ev.ParseAndCompile(parse.Source{Code: "ord a"}, w)
	if err != nil {
		t.Errorf("got err %v, want nil", err)
	}
	w.Close()
	warnings := mustReadAllAndClose(r)
	wantWarning := []byte(`the "ord" command is deprecated`)
	if !bytes.Contains(warnings, wantWarning) {
		t.Errorf("got warnings %q, want warnings to contain %q", warnings, wantWarning)
	}
}

func TestMiscEval(t *testing.T) {
	Test(t,
		// Pseudo-namespace E:
		That("E:FOO=lorem; put $E:FOO").Puts("lorem"),
		That("del E:FOO; put $E:FOO").Puts(""),
	)
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	r := evalAndCollect(t, NewEvaler(), texts)
	wantOuts := []interface{}{"hello"}
	if r.exception != nil {
		t.Errorf("eval %s => %v, want nil", texts, r.exception)
	}
	if !reflect.DeepEqual(r.valueOut, wantOuts) {
		t.Errorf("eval %s outputs %v, want %v", texts, r.valueOut, wantOuts)
	}
}

func TestConcurrentEval(t *testing.T) {
	// Run this test with "go test -race".
	ev := NewEvaler()
	src := parse.Source{Name: "[test]"}
	op, err := ev.ParseAndCompile(src, nil)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ev.Eval(op, EvalCfg{})
		wg.Done()
	}()
	go func() {
		ev.Eval(op, EvalCfg{})
		wg.Done()
	}()
	wg.Wait()
}
