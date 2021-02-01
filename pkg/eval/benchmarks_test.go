package eval

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

func BenchmarkEval_Empty(b *testing.B) {
	benchmarkEval(b, "")
}

func BenchmarkEval_NopCommand(b *testing.B) {
	benchmarkEval(b, "nop")
}

func BenchmarkEval_PutCommand(b *testing.B) {
	benchmarkEval(b, "put x")
}

func BenchmarkEval_ForLoop100WithEmptyBody(b *testing.B) {
	benchmarkEval(b, "for x [(range 100)] { }")
}

func BenchmarkEval_EachLoop100WithEmptyBody(b *testing.B) {
	benchmarkEval(b, "range 100 | each [x]{ }")
}

func BenchmarkEval_LocalVariableAccess(b *testing.B) {
	benchmarkEval(b, "x = val; nop $x")
}

func BenchmarkEval_UpVariableAccess(b *testing.B) {
	benchmarkEval(b, "x = val; { nop $x }")
}

func benchmarkEval(b *testing.B, code string) {
	ev := NewEvaler()
	src := parse.Source{Name: "[benchmark]", Code: code}

	tree, err := parse.Parse(src, parse.Config{})
	if err != nil {
		panic(err)
	}
	op, err := ev.compile(tree, ev.Global(), nil)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fm, cleanup := ev.prepareFrame(src, EvalCfg{Global: ev.Global()})
		_, exec := op.prepare(fm)
		_ = exec()
		cleanup()
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
	fm := &Frame{Evaler: ev, local: ev.Global(), up: new(Ns)}
	for i := 0; i < n; i++ {
		fm.CaptureOutput(func(fm *Frame) error {
			f(fm)
			return nil
		})
	}
}
