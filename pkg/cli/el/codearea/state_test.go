package codearea

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestApplyPending(t *testing.T) {
	applyPending := func(s State) State {
		s.ApplyPending()
		return s
	}
	tt.Test(t, tt.Fn("applyPending", applyPending), tt.Table{
		tt.Args(State{Buffer: Buffer{}, Pending: Pending{0, 0, "ls"}}).
			Rets(State{Buffer: Buffer{Content: "ls", Dot: 2}, Pending: Pending{}}),
		tt.Args(State{Buffer: Buffer{"x", 1}, Pending: Pending{0, 0, "ls"}}).
			Rets(State{Buffer: Buffer{Content: "lsx", Dot: 3}, Pending: Pending{}}),
		// No-op when Pending is empty.
		tt.Args(State{Buffer: Buffer{"x", 1}}).
			Rets(State{Buffer: Buffer{Content: "x", Dot: 1}}),
		// HideRPrompt is kept intact.
		tt.Args(State{Buffer: Buffer{"x", 1}, HideRPrompt: true}).
			Rets(State{Buffer: Buffer{Content: "x", Dot: 1}, HideRPrompt: true}),
	})
}
