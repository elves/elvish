package eval

import (
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
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
	src := NewInternalElvishSource(true, "[test]", "")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ev.EvalSourceInTTY(src)
		wg.Done()
	}()
	go func() {
		ev.EvalSourceInTTY(src)
		wg.Done()
	}()
	wg.Wait()
}

func BenchmarkOutputCaptureOverhead(b *testing.B) {
	op := effectOp{funcOp(func(*Frame) error { return nil }), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureValues(b *testing.B) {
	op := effectOp{funcOp(func(fm *Frame) error {
		fm.ports[1].Chan <- "test"
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureBytes(b *testing.B) {
	bytesToWrite := []byte("test")
	op := effectOp{funcOp(func(fm *Frame) error {
		fm.ports[1].File.Write(bytesToWrite)
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureMixed(b *testing.B) {
	bytesToWrite := []byte("test")
	op := effectOp{funcOp(func(fm *Frame) error {
		fm.ports[1].Chan <- false
		fm.ports[1].File.Write(bytesToWrite)
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func benchmarkOutputCapture(op effectOp, n int) {
	ev := NewEvaler()
	defer ev.Close()
	ec := NewTopFrame(ev, NewInternalGoSource("[benchmark]"), []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		pcaptureOutput(ec, op)
	}
}
