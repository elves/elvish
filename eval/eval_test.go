package eval

import (
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/eval/vals"
)

func TestBuiltinPid(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	builtinPid := vals.ToString(builtinNs["pid"].Get())
	if builtinPid != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, builtinPid, pid)
	}
}

var miscEvalTests = []Test{
	// Pseudo-namespaces local: and up:
	{"x=lorem; []{local:x=ipsum; put $up:x $local:x}",
		want{out: strs("lorem", "ipsum")}},
	{"x=lorem; []{up:x=ipsum; put $x}; put $x",
		want{out: strs("ipsum", "ipsum")}},
	// Pseudo-namespace E:
	{"E:FOO=lorem; put $E:FOO", want{out: strs("lorem")}},
	{"del E:FOO; put $E:FOO", want{out: strs("")}},
}

func TestMiscEval(t *testing.T) {
	runTests(t, miscEvalTests)
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	outs, _, err := evalAndCollect(t, NewEvaler(), texts, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %s => %v, want nil", texts, err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %s outputs %v, want %v", texts, outs, wanted)
	}
}

func BenchmarkOutputCaptureOverhead(b *testing.B) {
	op := Op{funcOp(func(*Frame) error { return nil }), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureValues(b *testing.B) {
	op := Op{funcOp(func(fm *Frame) error {
		fm.ports[1].Chan <- "test"
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureBytes(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{funcOp(func(fm *Frame) error {
		fm.ports[1].File.Write(bytesToWrite)
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureMixed(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{funcOp(func(fm *Frame) error {
		fm.ports[1].Chan <- false
		fm.ports[1].File.Write(bytesToWrite)
		return nil
	}), 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func benchmarkOutputCapture(op Op, n int) {
	ev := NewEvaler()
	defer ev.Close()
	ec := NewTopFrame(ev, NewInternalSource("[benchmark]"), []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		pcaptureOutput(ec, op)
	}
}
