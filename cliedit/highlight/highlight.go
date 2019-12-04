// Package highlight provides an Elvish syntax highlighter.
package highlight

import (
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Dep keeps dependencies for highlighting code.
type Dep struct {
	Check      func(n *parse.Chunk) error
	HasCommand func(name string) bool
}

// Information collected about a command region, used for asynchronou
// highlighting.
type cmdRegion struct {
	seg int
	cmd string
}

// Highlights a piece of Elvish code.
func highlight(code string, dep Dep, lateCb func(styled.Text)) (styled.Text, []error) {
	var errors []error

	n, errParse := parse.AsChunk("[interactive]", code)
	if errParse != nil {
		for _, err := range errParse.(*parse.MultiError).Entries {
			if err.Context.Begin != len(code) {
				errors = append(errors, err)
			}
		}
	}

	if dep.Check != nil {
		err := dep.Check(n)
		if err != nil && err.(diag.Ranger).Range().From != len(code) {
			errors = append(errors, err)
			// TODO: Highlight the region with errors.
		}
	}

	var text styled.Text
	regions := getRegions(n)
	lastEnd := 0
	var cmdRegions []cmdRegion

	for _, r := range regions {
		if r.begin > lastEnd {
			// Add inter-region text.
			text = append(text, styled.PlainSegment(code[lastEnd:r.begin]))
		}

		regionCode := code[r.begin:r.end]
		transformer := ""
		if r.typ == commandRegion {
			if dep.HasCommand != nil {
				// Do not highlight now, but collect the index of the region and the
				// segment.
				cmdRegions = append(cmdRegions, cmdRegion{len(text), regionCode})
			} else {
				// Treat all commands as good commands.
				transformer = transformerForGoodCommand
			}
		} else {
			transformer = transformerFor[r.typ]
		}
		seg := styled.PlainSegment(regionCode)
		if transformer != "" {
			styled.FindTransformer(transformer)(seg)
		}

		text = append(text, seg)
		lastEnd = r.end
	}
	if len(code) > lastEnd {
		// Add text after the last region as unstyled.
		text = append(text, styled.PlainSegment(code[lastEnd:]))
	}

	// Style command regions asynchronously, and call lateCb with the results.
	if dep.HasCommand != nil && len(cmdRegions) > 0 {
		go func() {
			newText := text.Clone()
			for _, cmdRegion := range cmdRegions {
				transformer := ""
				if dep.HasCommand(cmdRegion.cmd) {
					transformer = transformerForGoodCommand
				} else {
					transformer = transformerForBadCommand
				}
				styled.FindTransformer(transformer)(newText[cmdRegion.seg])
			}
			lateCb(newText)
		}()
	}

	return text, errors
}
