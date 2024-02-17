package md_test

import (
	"fmt"
	"strings"
	"testing"

	"src.elv.sh/pkg/diff"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

// Most of the parser behavior is tested indirectly via the HTML output. This
// file only covers behavior not observable from the HTML output.

type lineNoCodec struct{ strings.Builder }

func (c *lineNoCodec) Do(op md.Op) {
	fmt.Fprintln(&c.Builder, op.Type, op.LineNo)
}

var lineNoTestInput = testutil.Dedent(`
	---

	# line 3

	~~~line 5
	foo
	~~~

	    line 9
	    foo

	<pre>line 12
	</pre>

	<hr>line 15

	line 17
	more lines

	> line 20
	> more lines
	>
	> line 23

	- line 25

	  line 27

	- line 29

	1. line 31

	2. line 33
	`)

var lineNoTestOutput = testutil.Dedent(`
	OpThematicBreak 1
	OpHeading 3
	OpCodeBlock 5
	OpCodeBlock 9
	OpHTMLBlock 12
	OpHTMLBlock 15
	OpParagraph 17
	OpBlockquoteStart 20
	OpParagraph 20
	OpParagraph 23
	OpBlockquoteEnd 24
	OpBulletListStart 25
	OpListItemStart 25
	OpParagraph 25
	OpParagraph 27
	OpListItemEnd 29
	OpListItemStart 29
	OpParagraph 29
	OpListItemEnd 31
	OpBulletListEnd 31
	OpOrderedListStart 31
	OpListItemStart 31
	OpParagraph 31
	OpListItemEnd 33
	OpListItemStart 33
	OpParagraph 33
	OpListItemEnd 34
	OpOrderedListEnd 34
	`)

func TestLineNo(t *testing.T) {
	got := md.RenderString(lineNoTestInput, &lineNoCodec{})
	if want := lineNoTestOutput; want != got {
		t.Errorf("%s", diff.Diff("want", want, "got", got))
	}
}
