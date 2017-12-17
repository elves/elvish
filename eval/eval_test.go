package eval

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func TestBuiltinPid(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	builtinPid := ToString(makeBuiltinNamespace(nil)["pid"].Get())
	if builtinPid != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, builtinPid, pid)
	}
}

type evalTest struct {
	text string
	want
}

type want struct {
	out      []Value
	bytesOut []byte
	err      error
}

var (
	wantNothing = want{}
	// Special value for want.err to indicate that any error, as long as not
	// nil, is OK
	errAny = errors.New("")
)

// Shorthands for values in want.out

func strs(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = String(s)
	}
	return vs
}

func bools(bs ...bool) []Value {
	vs := make([]Value, len(bs))
	for i, b := range bs {
		vs[i] = Bool(b)
	}
	return vs
}

var filesToCreate = sorted(
	"a1", "a2", "a3", "a10", "b1", "b2", "b3",
	"foo", "bar", "lorem", "ipsum",
)

func sorted(a ...string) []string {
	sort.Strings(a)
	return a
}

// To be called from init in separate test files.
func addToEvalTests(tests []evalTest) {
	evalTests = append(evalTests, tests...)
}

var evalTests = []evalTest{

	// Pseudo-namespaces local: and up:
	{"x=lorem; []{local:x=ipsum; put $up:x $local:x}",
		want{out: strs("lorem", "ipsum")}},
	{"x=lorem; []{up:x=ipsum; put $x}; put $x",
		want{out: strs("ipsum", "ipsum")}},
	// Pseudo-namespace E:
	{"E:FOO=lorem; put $E:FOO", want{out: strs("lorem")}},
	{"del E:FOO; put $E:FOO", want{out: strs("")}},
}

func mustParseAndCompile(t *testing.T, ev *Evaler, name, text string) Op {
	n, err := parse.Parse(name, text)
	if err != nil {
		t.Fatalf("Parse(%q) error: %s", text, err)
	}
	op, err := ev.Compile(n, name, text)
	if err != nil {
		t.Fatalf("Compile(Parse(%q)) error: %s", text, err)
	}
	return op
}

func TestEval(t *testing.T) {
	util.InTempDir(func(tempDir string) {
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", oldHome)
		for _, filename := range filesToCreate {
			file, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			file.Close()
		}

		for _, tt := range evalTests {
			// fmt.Printf("eval %q\n", tt.text)

			out, bytesOut, err := evalAndCollect(t, []string{tt.text}, len(tt.want.out))

			first := true
			errorf := func(format string, args ...interface{}) {
				if first {
					first = false
					t.Errorf("eval(%q) fails:", tt.text)
				}
				t.Errorf("  "+format, args...)
			}

			if !matchOut(tt.want.out, out) {
				errorf("got out=%v, want %v", out, tt.want.out)
			}
			if string(tt.want.bytesOut) != string(bytesOut) {
				errorf("got bytesOut=%q, want %q", bytesOut, tt.want.bytesOut)
			}
			if !matchErr(tt.want.err, err) {
				errorf("got err=%v, want %v", err, tt.want.err)
			}
		}
	})
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	outs, _, err := evalAndCollect(t, texts, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %s => %v, want nil", texts, err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %s outputs %v, want %v", texts, outs, wanted)
	}
}

func evalAndCollect(t *testing.T, texts []string, chsize int) ([]Value, []byte, error) {
	name := "<eval test>"
	ev := NewEvaler(api.NewClient("/invalid"), nil, dataDir, nil)

	// Collect byte output
	bytesOut := []byte{}
	pr, pw, _ := os.Pipe()
	bytesDone := make(chan struct{})
	go func() {
		for {
			var buf [64]byte
			nr, err := pr.Read(buf[:])
			bytesOut = append(bytesOut, buf[:nr]...)
			if err != nil {
				break
			}
		}
		close(bytesDone)
	}()

	// Channel output
	outs := []Value{}

	// Eval error. Only that of the last text is saved.
	var ex error

	for _, text := range texts {
		op := mustParseAndCompile(t, ev, name, text)

		outCh := make(chan Value, chsize)
		outDone := make(chan struct{})
		go func() {
			for v := range outCh {
				outs = append(outs, v)
			}
			close(outDone)
		}()

		ports := []*Port{
			{File: os.Stdin, Chan: ClosedChan},
			{File: pw, Chan: outCh},
			{File: os.Stderr, Chan: BlackholeChan},
		}

		ex = ev.eval(op, ports, name, text)
		close(outCh)
		<-outDone
	}

	pw.Close()
	<-bytesDone
	pr.Close()

	return outs, bytesOut, ex
}

func matchOut(want, got []Value) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}

func matchErr(want, got error) bool {
	if got == nil {
		return want == nil
	}
	return want == errAny || reflect.DeepEqual(got.(*Exception).Cause, want)
}

func BenchmarkOutputCaptureOverhead(b *testing.B) {
	op := Op{func(*EvalCtx) {}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureValues(b *testing.B) {
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].Chan <- String("test")
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureBytes(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].File.Write(bytesToWrite)
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func BenchmarkOutputCaptureMixed(b *testing.B) {
	bytesToWrite := []byte("test")
	op := Op{func(ec *EvalCtx) {
		ec.ports[1].Chan <- Bool(false)
		ec.ports[1].File.Write(bytesToWrite)
	}, 0, 0}
	benchmarkOutputCapture(op, b.N)
}

func benchmarkOutputCapture(op Op, n int) {
	ev := NewEvaler(api.NewClient("/invalid"), nil, "", nil)
	ec := NewTopEvalCtx(ev, "[benchmark]", "", []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		pcaptureOutput(ec, op)
	}
}
