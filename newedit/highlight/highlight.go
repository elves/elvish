// Package highlight provides an Elvish syntax highlighter.
package highlight

import (
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Dependencies for highlighting code.
type hlDep struct {
	Check      func(n *parse.Chunk) error
	HasCommand func(name string) bool
}

// Highlights a piece of Elvish code.
func highlight(code string, hl hlDep) (styled.Text, []error) {
	var errors []error

	n, errParse := parse.AsChunk("[interactive]", code)
	if errParse != nil {
		for _, err := range errParse.(parse.MultiError).Entries {
			if err.Context.Begin != len(code) {
				errors = append(errors, err)
			}
		}
	}

	if hl.Check != nil {
		err := hl.Check(n)
		if err != nil && err.(diag.Ranger).Range().From != len(code) {
			errors = append(errors, err)
			// TODO: Highlight the region with errors.
		}
	}

	var text styled.Text
	regions := getRegions(n)
	lastEnd := 0
	for _, r := range regions {
		if r.begin > lastEnd {
			// Add inter-region text.
			text = append(text, styled.UnstyledSegment(code[lastEnd:r.begin]))
		}

		regionCode := code[r.begin:r.end]
		transformer := ""
		if hl.HasCommand != nil && r.typ == commandRegion {
			// Commands are highlighted differently depending on whether they
			// are valid.
			if hl.HasCommand(regionCode) {
				transformer = transformerForGoodCommand
			} else {
				transformer = transformerForBadCommand
			}
		} else {
			transformer = transformerFor[r.typ]
		}
		seg := styled.UnstyledSegment(regionCode)
		if transformer != "" {
			styled.FindTransformer(transformer)(seg)
		}

		text = append(text, seg)
		lastEnd = r.end
	}
	if len(code) > lastEnd {
		// Add text after the last region as unstyled.
		text = append(text, styled.UnstyledSegment(code[lastEnd:]))
	}
	return text, errors
}
