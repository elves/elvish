package main

import (
	"go/scanner"
	"go/token"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/ui"
)

// Augments elvdoc.HighlightCodeBlock with additional syntax highlighting that
// we don't want to be part of the Elvish binary. This currently just includes
// Go; sh/bash is another candidate.
func highlightCodeBlock(info, code string) ui.Text {
	if language, _, _ := strings.Cut(info, " "); language == "go" {
		return highlightGo(code)
	}
	return elvdoc.HighlightCodeBlock(info, code)
}

func highlightGo(code string) ui.Text {
	lexer, posBase := lexGo(code)
	var regions []ui.StylingRegion
	for {
		pos, tok, lit := lexer.Scan()
		if tok == token.EOF {
			break
		}
		if styling := styleGoToken(tok); styling != nil {
			from := int(pos) - posBase
			// Note that lit is "" for all operator tokens like "{" and "+". We
			// don't currently highlight them, but if we do we should use
			// tok.String() instead of lit for them.
			to := from + len(lit)
			region := ui.StylingRegion{
				Ranging: diag.Ranging{From: from, To: to}, Styling: styling,
			}
			regions = append(regions, region)
		}
	}
	return ui.StyleRegions(code, regions)
}

// We don't use the full parser here, both because the scanner is sufficient for
// highlighting, and we often highlight snippets of Go that are not necessarily
// suitable for either go/parser.ParseFile or go/parser.ParseExpr.
func lexGo(code string) (lexer scanner.Scanner, posBase int) {
	fset := token.NewFileSet()
	file := fset.AddFile("main.go", -1, len(code))
	lexer.Init(file, []byte(code), nil, scanner.ScanComments)
	return lexer, file.Base()
}

func styleGoToken(tok token.Token) ui.Styling {
	switch tok {
	case token.ILLEGAL:
		return ui.Stylings(ui.FgBrightWhite, ui.BgRed)
	case token.COMMENT:
		return ui.FgCyan
	case token.CHAR, token.STRING:
		return ui.FgYellow
	default:
		if tok.IsKeyword() {
			return ui.FgBlue
		}
		return nil
	}
}
