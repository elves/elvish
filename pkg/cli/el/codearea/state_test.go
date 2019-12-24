package codearea

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestApplyPending(t *testing.T) {
	applyPending := func(s CodeAreaState) CodeAreaState {
		s.ApplyPending()
		return s
	}
	tt.Test(t, tt.Fn("applyPending", applyPending), tt.Table{
		tt.Args(CodeAreaState{Buffer: CodeBuffer{}, Pending: PendingCode{0, 0, "ls"}}).
			Rets(CodeAreaState{Buffer: CodeBuffer{Content: "ls", Dot: 2}, Pending: PendingCode{}}),
		tt.Args(CodeAreaState{Buffer: CodeBuffer{"x", 1}, Pending: PendingCode{0, 0, "ls"}}).
			Rets(CodeAreaState{Buffer: CodeBuffer{Content: "lsx", Dot: 3}, Pending: PendingCode{}}),
		// No-op when Pending is empty.
		tt.Args(CodeAreaState{Buffer: CodeBuffer{"x", 1}}).
			Rets(CodeAreaState{Buffer: CodeBuffer{Content: "x", Dot: 1}}),
		// HideRPrompt is kept intact.
		tt.Args(CodeAreaState{Buffer: CodeBuffer{"x", 1}, HideRPrompt: true}).
			Rets(CodeAreaState{Buffer: CodeBuffer{Content: "x", Dot: 1}, HideRPrompt: true}),
	})
}
