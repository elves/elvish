// Package highlight provides an Elvish syntax highlighter.
package highlight

import (
	"time"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// Config keeps configuration for highlighting code.
type Config struct {
	Check      func(n parse.Tree) error
	HasCommand func(name string) bool
}

// Information collected about a command region, used for asynchronous
// highlighting.
type cmdRegion struct {
	seg int
	cmd string
}

// MaxBlockForLate specifies the maximum wait time to block for late results.
// It can be changed for test cases.
var MaxBlockForLate = 10 * time.Millisecond

// Highlights a piece of Elvish code.
func highlight(code string, cfg Config, lateCb func(ui.Text)) (ui.Text, []error) {
	var errors []error
	var errorRegions []region

	tree, errParse := parse.Parse(parse.Source{Name: "[interactive]", Code: code}, parse.Config{})
	if errParse != nil {
		for _, err := range errParse.(*parse.Error).Entries {
			if err.Context.From != len(code) {
				errors = append(errors, err)
				errorRegions = append(errorRegions,
					region{
						err.Context.From, err.Context.To,
						semanticRegion, errorRegion})
			}
		}
	}

	if cfg.Check != nil {
		err := cfg.Check(tree)
		if r, ok := err.(diag.Ranger); ok && r.Range().From != len(code) {
			errors = append(errors, err)
			errorRegions = append(errorRegions,
				region{
					r.Range().From, r.Range().To, semanticRegion, errorRegion})
		}
	}

	var text ui.Text
	regions := getRegionsInner(tree.Root)
	regions = append(regions, errorRegions...)
	regions = fixRegions(regions)
	lastEnd := 0
	var cmdRegions []cmdRegion

	for _, r := range regions {
		if r.Begin > lastEnd {
			// Add inter-region text.
			text = append(text, &ui.Segment{Text: code[lastEnd:r.Begin]})
		}

		regionCode := code[r.Begin:r.End]
		var styling ui.Styling
		if r.Type == commandRegion {
			if cfg.HasCommand != nil {
				// Do not highlight now, but collect the index of the region and the
				// segment.
				cmdRegions = append(cmdRegions, cmdRegion{len(text), regionCode})
			} else {
				// Treat all commands as good commands.
				styling = stylingForGoodCommand
			}
		} else {
			styling = stylingFor[r.Type]
		}
		seg := &ui.Segment{Text: regionCode}
		if styling != nil {
			seg = ui.StyleSegment(seg, styling)
		}

		text = append(text, seg)
		lastEnd = r.End
	}
	if len(code) > lastEnd {
		// Add text after the last region as unstyled.
		text = append(text, &ui.Segment{Text: code[lastEnd:]})
	}

	if cfg.HasCommand != nil && len(cmdRegions) > 0 {
		// Launch a goroutine to style command regions asynchronously.
		lateCh := make(chan ui.Text)
		go func() {
			newText := text.Clone()
			for _, cmdRegion := range cmdRegions {
				var styling ui.Styling
				if cfg.HasCommand(cmdRegion.cmd) {
					styling = stylingForGoodCommand
				} else {
					styling = stylingForBadCommand
				}
				seg := &newText[cmdRegion.seg]
				*seg = ui.StyleSegment(*seg, styling)
			}
			lateCh <- newText
		}()
		// Block a short while for the late text to arrive, in order to reduce
		// flickering. Otherwise, return the text already computed, and pass the
		// late result to lateCb in another goroutine.
		select {
		case late := <-lateCh:
			return late, errors
		case <-time.After(MaxBlockForLate):
			go func() {
				lateCb(<-lateCh)
			}()
			return text, errors
		}
	}
	return text, errors
}
