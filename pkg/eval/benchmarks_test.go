package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/parse"
)

func BenchmarkEval_Empty(b *testing.B) {
	benchmarkEval(b.N, "")
}

func BenchmarkEval_NopCommand(b *testing.B) {
	benchmarkEval(b.N, "nop")
}

func BenchmarkEval_PutCommand(b *testing.B) {
	benchmarkEval(b.N, "put x")
}

func BenchmarkEval_ForLoop100WithEmptyBody(b *testing.B) {
	benchmarkEval(b.N, "for x [(range 100)] { }")
}

func BenchmarkEval_EachLoop100WithEmptyBody(b *testing.B) {
	benchmarkEval(b.N, "range 100 | each [x]{ }")
}

func benchmarkEval(n int, code string) {
	ev := NewEvaler()
	src := parse.Source{Name: "[benchmark]", Code: code}
	op, err := ev.ParseAndCompile(src, nil)
	if err != nil {
		panic(err)
	}
	for i := 0; i < n; i++ {
		ev.Eval(op, EvalCfg{})
	}
}

func BenchmarkOutputCapture_Overhead(b *testing.B) {
	benchmarkOutputCapture(b.N, func(fm *Frame) {})
}

func BenchmarkOutputCapture_Values(b *testing.B) {
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.OutputChan() <- "test"
	})
}

func BenchmarkOutputCapture_Bytes(b *testing.B) {
	bytesToWrite := []byte("test")
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.OutputFile().Write(bytesToWrite)
	})
}

func BenchmarkOutputCapture_Mixed(b *testing.B) {
	bytesToWrite := []byte("test")
	benchmarkOutputCapture(b.N, func(fm *Frame) {
		fm.OutputChan() <- false
		fm.OutputFile().Write(bytesToWrite)
	})
}

func benchmarkOutputCapture(n int, f func(*Frame)) {
	ev := NewEvaler()
	defer ev.Close()
	fm := NewTopFrame(ev, parse.Source{Name: "[benchmark]"}, []*Port{{}, {}, {}})
	for i := 0; i < n; i++ {
		captureOutput(fm, func(fm *Frame) error {
			f(fm)
			return nil
		})
	}
}
