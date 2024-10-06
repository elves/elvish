package main

import (
	"fmt"
	"regexp"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/ui"
)

func Preso(b etk.Context) (etk.View, etk.React) {
	currentVar := etk.State(b, "current", 0)
	current := currentVar.Get()
	pagesVar := etk.State(b, "pages", []ui.Text{ui.T("no content")})
	pages := pagesVar.Get()

	return etk.Box(`
			[indicator=]
			1
			page=`,
			etk.Text(ui.T(fmt.Sprintf("%d / %d", current+1, len(pages))),
				etk.DotHere),
			etk.Text(pages[current]),
		), func(e term.Event) etk.Reaction {
			switch e {
			case term.K(ui.Left):
				if current > 0 {
					currentVar.Set(current - 1)
					return etk.Consumed
				}
			case term.K(ui.Right):
				if current < len(pages)-1 {
					currentVar.Set(current + 1)
					return etk.Consumed
				}
			}
			return etk.Unused
		}
}

var thematicBreakRegexp = regexp.MustCompile(
	`(?m)^ {0,3}((?:-[ \t]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})$`)

func parsePreso(src string) []ui.Text {
	pageSrcs := thematicBreakRegexp.Split(src, -1)
	pages := make([]ui.Text, len(pageSrcs))
	for i, pageSrc := range pageSrcs {
		codec := md.TTYCodec{
			HighlightCodeBlock: elvdoc.HighlightCodeBlock,
		}
		md.Render(pageSrc, &codec)
		pages[i] = codec.Text()
	}
	return pages
}
