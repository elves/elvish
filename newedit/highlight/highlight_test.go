package highlight

import (
	"testing"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

type matchErrorCount struct {
	count int
}

func (m matchErrorCount) Match(v tt.RetValue) bool {
	return len(v.([]error)) == m.count
}

func TestHighlight(t *testing.T) {
	var noErrors []error
	tt.Test(t, tt.Fn("Highlight", Highlight), tt.Table{
		Args("ls").Rets(styled.Text{
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
		}, noErrors),
		Args(" ls\n").Rets(styled.Text{
			styled.UnstyledSegment(" "),
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
			styled.UnstyledSegment("\n"),
		}, noErrors),
		Args("ls $").Rets(styled.Text{
			&styled.Segment{styled.Style{Foreground: "green"}, "ls"},
			styled.UnstyledSegment(" "),
			&styled.Segment{styled.Style{Foreground: "magenta"}, "$"},
		}, matchErrorCount{1}),
	})
}
