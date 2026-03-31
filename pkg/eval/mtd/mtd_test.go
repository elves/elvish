package mtd_test

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/mtd"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/tt"
)

var Next = mtd.New[func(*eval.Frame, any, int) (int, error)]("Next")

type errorlessImpl struct{}

func (errorlessImpl) Next(i int) int { return i + 1 }

type errorfulImpl struct{}

var errSample = errors.New("sample error")

func (errorfulImpl) Next(i int) (int, error) { return i * 2, errSample }

var Args = tt.Args

func TestCall(t *testing.T) {
	ev := eval.NewEvaler()
	fm := ev.CallFrame("test")
	tt.Test(t, Next.Call,
		// Native Go method without final error return
		Args(fm, errorlessImpl{}, 100).Rets(101, nil),
		// Native Go method, with final error return
		Args(fm, errorfulImpl{}, 100).Rets(200, errSample),
		// Elvish map
		Args(fm,
			vals.MakeMap("next",
				eval.NewGoFn("<next>", func(i int) int { return i + 10 })),
			100).
			Rets(110, nil),
		// TODO: More test cases
	)
}
