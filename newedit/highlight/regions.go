package highlight

import (
	"sort"
	"strings"

	"github.com/elves/elvish/parse"
)

// Represents a region to be highlighted.
type region struct {
	begin int
	end   int
	// Regions can be lexical or semantic. Lexical regions always correspond to
	// a leaf node in the parse tree, either a parse.Primary node or a parse.Sep
	// node. Semantic regions may span several leaves and override all lexical
	// regions in it.
	kind regionKind
	// In lexical regions for Primary nodes, this field corresponds to the Type
	// field of the node (e.g. "bareword", "single-quoted"). In lexical regions
	// for Sep nodes, this field is simply the source text itself (e.g. "(",
	// "|"), except for comments, which have typ == "comment".
	//
	// In semantic regions, this field takes a value from a fixed list (see
	// below).
	typ string
}

type regionKind int

// Region kinds.
const (
	lexicalRegion regionKind = iota
	semanticRegion
)

// Lexical region types.
const (
	barewordRegion     = "bareword"
	singleQuotedRegion = "single-quoted"
	doubleQuotedRegion = "double-quoted"
	variableRegion     = "variable" // Could also be semantic.
	wildcardRegion     = "wildcard"
	tildeRegion        = "tilde"
	// A comment region. Note that this is the only type of Sep leaf node that
	// is not identified by its text.
	commentRegion = "comment"
)

// Semantic region types.
const (
	// A region when a string literal (bareword, single-quoted or double-quoted)
	// appears as a command.
	commandRegion = "command"
	// A region for keywords in special forms, like "else" in an "if" form.
	keywordRegion = "keyword"
)

func getRegions(n parse.Node) []region {
	var regions []region
	emitRegions(n, func(n parse.Node, kind regionKind, typ string) {
		regions = append(regions, region{n.Begin(), n.End(), kind, typ})
	})
	// Sort the regions by the begin position, putting semantic regions before
	// lexical regions.
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].begin < regions[j].begin {
			return true
		}
		if regions[i].begin == regions[j].begin {
			return regions[i].kind == semanticRegion && regions[j].kind == lexicalRegion
		}
		return false
	})
	// Remove overlapping regions, preferring the ones that appear earlier.
	var newRegions []region
	lastEnd := 0
	for _, r := range regions {
		if r.begin < lastEnd {
			continue
		}
		newRegions = append(newRegions, r)
		lastEnd = r.end
	}
	return newRegions
}

func emitRegions(n parse.Node, f func(parse.Node, regionKind, string)) {
	switch n := n.(type) {
	case *parse.Form:
		emitRegionsInForm(n, f)
	case *parse.Primary:
		emitRegionsInPrimary(n, f)
	case *parse.Sep:
		emitRegionsInSep(n, f)
	}
	for _, child := range n.Children() {
		emitRegions(child, f)
	}
}

func emitRegionsInForm(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Left hands of temporary assignments.
	for _, an := range n.Assignments {
		if an.Left != nil && an.Left.Head != nil {
			f(an.Left.Head, semanticRegion, variableRegion)
		}
	}
	// Left hands of ordinary assignments.
	for _, cn := range n.Vars {
		// NOTE: In a well-formed assignment, cn.Indexings will have at most 1
		// element. For instance, "x[0]y[1] = value" is not well-formed and will
		// result in a parse error.
		if len(cn.Indexings) > 0 && cn.Indexings[0].Head != nil {
			f(cn.Indexings[0].Head, semanticRegion, variableRegion)
		}
	}
	if n.Head == nil {
		return
	}
	f(n.Head, semanticRegion, commandRegion)
	// Special forms.
	// TODO: This only highlights bareword special commands, however currently
	// quoted special commands are also possible (e.g `"if" $true { }` is
	// accepted).
	switch n.Head.SourceText() {
	case "if":
		emitRegionsInIf(n, f)
	case "for":
		emitRegionsInFor(n, f)
	case "try":
		emitRegionsInTry(n, f)
	}
}

func emitRegionsInIf(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight all "elif" and "else".
	for i := 2; i < len(n.Args); i += 2 {
		arg := n.Args[i]
		if arg.SourceText() == "elif" || arg.SourceText() == "else" {
			f(arg, semanticRegion, keywordRegion)
		}
	}
}

func emitRegionsInFor(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight the iterating variable.
	if 0 < len(n.Args) && len(n.Args[0].Indexings) > 0 {
		f(n.Args[0].Indexings[0].Head, semanticRegion, variableRegion)
	}
	// Highlight "else".
	if 3 < len(n.Args) && n.Args[3].SourceText() == "else" {
		f(n.Args[3], semanticRegion, keywordRegion)
	}
}

func emitRegionsInTry(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight "except", the exception variable after it, "else" and
	// "finally".
	i := 1
	matchKW := func(text string) bool {
		if i < len(n.Args) && n.Args[i].SourceText() == text {
			f(n.Args[i], semanticRegion, keywordRegion)
			return true
		}
		return false
	}
	if matchKW("except") {
		if i+1 < len(n.Args) && len(n.Args[i+1].Indexings) > 0 {
			f(n.Args[i+1], semanticRegion, variableRegion)
		}
		i += 3
	}
	if matchKW("else") {
		i += 2
	}
	matchKW("finally")
}

func emitRegionsInPrimary(n *parse.Primary, f func(parse.Node, regionKind, string)) {
	switch n.Type {
	case parse.Bareword:
		f(n, lexicalRegion, barewordRegion)
	case parse.SingleQuoted:
		f(n, lexicalRegion, singleQuotedRegion)
	case parse.DoubleQuoted:
		f(n, lexicalRegion, doubleQuotedRegion)
	case parse.Variable:
		f(n, lexicalRegion, variableRegion)
	case parse.Wildcard:
		f(n, lexicalRegion, wildcardRegion)
	case parse.Tilde:
		f(n, lexicalRegion, tildeRegion)
	}
}

func emitRegionsInSep(n *parse.Sep, f func(parse.Node, regionKind, string)) {
	text := n.SourceText()
	switch {
	case strings.TrimSpace(text) == "":
		// Don't do anything; whitespaces do not get highlighted.
	case strings.HasPrefix(text, "#"):
		f(n, lexicalRegion, commentRegion)
	default:
		f(n, lexicalRegion, text)
	}
}
