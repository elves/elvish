package eval

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

var benchmarks = []struct {
	name string
	code string
}{
	{"empty", ""},
	{"nop", "nop"},
	{"nop-nop", "nop | nop"},
	{"put-x", "put x"},
	{"for-100", "for x [(range 100)] { }"},
	{"range-100", "range 100 | each {|_| }"},
	{"read-local", "var x = val; nop $x"},
	{"read-upval", "var x = val; { nop $x }"},
}

func BenchmarkEval(b *testing.B) {
	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			ev := NewEvaler()
			src := parse.Source{Name: "[benchmark]", Code: bench.code}

			tree, err := parse.Parse(src, parse.Config{})
			if err != nil {
				panic(err)
			}
			op, _, err := compile(ev.builtin.static(), ev.global.static(), nil, tree, nil)
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
		})
	}
}
