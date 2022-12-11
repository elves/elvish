package filter

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
	"src.elv.sh/pkg/ui"
)

func Highlight(q string) (ui.Text, []ui.Text) {
	n, _ := parseFilter(q)
	w := walker{}
	w.walk(n)
	text := ui.StyleRegions(q, w.regions)
	// TODO: Add errors.
	return text, nil
}

type walker struct {
	regions []ui.StylingRegion
}

func (w *walker) emit(r diag.Ranger, s ui.Styling) {
	region := ui.StylingRegion{Ranging: r.Range(), Styling: s}
	w.regions = append(w.regions, region)
}

func (w *walker) walk(n parse.Node) {
	switch n := n.(type) {
	case *parse.Sep:
		w.walkSep(n)
	case *parse.Primary:
		w.walkPrimary(n)
	}
	for _, ch := range parse.Children(n) {
		w.walk(ch)
	}
}

func (w *walker) walkSep(n *parse.Sep) {
	text := parse.SourceText(n)
	trimmed := strings.TrimLeftFunc(text, parse.IsWhitespace)
	if trimmed == "" {
		// Whitespace; nothing to do.
		return
	}
	// Metacharacter; style it bold.
	w.emit(n, ui.Bold)
}

func (w *walker) walkPrimary(n *parse.Primary) {
	switch n.Type {
	case parse.Bareword:
		// Barewords are unstyled.
	case parse.SingleQuoted, parse.DoubleQuoted:
		w.emit(n, ui.FgYellow)
	case parse.List:
		if len(n.Elements) == 0 {
			w.emit(n, ui.FgRed)
			return
		}
		headNode := n.Elements[0]
		head, ok := cmpd.StringLiteral(headNode)
		if !ok {
			w.emit(headNode, ui.FgRed)
		}
		switch head {
		case "re", "and", "or":
			w.emit(headNode, ui.FgGreen)
		default:
			w.emit(headNode, ui.FgRed)
		}
	default:
		// Unsupported primary type.
		w.emit(n, ui.FgRed)
	}
}
