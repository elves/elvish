// Package highlight provides an Elvish syntax highlighter.
package highlight

import (
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Highlight highlights a piece of Elvish code.
func Highlight(code string) (styled.Text, []error) {
	var errors []error

	n, errParse := parse.AsChunk("[interactive]", code)
	if errParse != nil {
		errors = append(errors, errParse)
	}
	// TODO: Add compilation errors.
	var text styled.Text
	regions := getRegions(n)
	lastEnd := 0
	for _, r := range regions {
		if r.begin > lastEnd {
			text = append(text, styled.UnstyledSegment(code[lastEnd:r.begin]))
		}
		// TODO: Style commandRegion.
		seg := styled.UnstyledSegment(code[r.begin:r.end])
		transformer := transformerFor[r.typ]
		if transformer != "" {
			styled.FindTransformer(transformer)(seg)
		}
		text = append(text, seg)
		lastEnd = r.end
	}
	if len(code) > lastEnd {
		text = append(text, styled.UnstyledSegment(code[lastEnd:]))
	}
	return text, errors
}
