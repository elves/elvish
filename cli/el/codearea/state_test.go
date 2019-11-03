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
		tt.Args(State{Buffer{}, Pending{0, 0, "ls"}}).
			Rets(State{Buffer{Content: "ls", Dot: 2}, Pending{}}),
		tt.Args(State{Buffer{"x", 1}, Pending{0, 0, "ls"}}).
			Rets(State{Buffer{Content: "lsx", Dot: 3}, Pending{}}),
	})
}
