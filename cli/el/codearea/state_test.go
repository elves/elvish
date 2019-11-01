package codearea

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestApplyPending(t *testing.T) {
	applyPending := func(s State) State {
		s.ApplyPending()
		return s
	}
	tt.Test(t, tt.Fn("applyPending", applyPending), tt.Table{
		tt.Args(State{CodeBuffer{}, PendingCode{0, 0, "ls"}}).
			Rets(State{CodeBuffer{Content: "ls", Dot: 2}, PendingCode{}}),
		tt.Args(State{CodeBuffer{"x", 1}, PendingCode{0, 0, "ls"}}).
			Rets(State{CodeBuffer{Content: "lsx", Dot: 3}, PendingCode{}}),
	})
}
