package highlight

import (
	"sort"
	"strings"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

var sourceText = parse.SourceText

// Represents a region to be highlighted.
type region struct {
	Begin int
	End   int
	// Regions can be lexical or semantic. Lexical regions always correspond to
	// a leaf node in the parse tree, either a parse.Primary node or a parse.Sep
	// node. Semantic regions may span several leaves and override all lexical
	// regions in it.
	Kind regionKind
	// In lexical regions for Primary nodes, this field corresponds to the Type
	// field of the node (e.g. "bareword", "single-quoted"). In lexical regions
	// for Sep nodes, this field is simply the source text itself (e.g. "(",
	// "|"), except for comments, which have Type == "comment".
	//
	// In semantic regions, this field takes a value from a fixed list (see
	// below).
	Type string
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
	// A region of parse or compilation error.
	errorRegion = "error"
)

func getRegions(n parse.Node) []region {
	regions := getRegionsInner(n)
	regions = fixRegions(regions)
	return regions
}

func getRegionsInner(n parse.Node) []region {
	var regions []region
	emitRegions(n, func(n parse.Node, kind regionKind, typ string) {
		regions = append(regions, region{n.Range().From, n.Range().To, kind, typ})
	})
	return regions
}

func fixRegions(regions []region) []region {
	// Sort regions by the begin position, putting semantic regions before
	// lexical regions.
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].Begin < regions[j].Begin {
			return true
		}
		if regions[i].Begin == regions[j].Begin {
			return regions[i].Kind == semanticRegion && regions[j].Kind == lexicalRegion
		}
		return false
	})
	// Remove overlapping regions, preferring the ones that appear earlier.
	var newRegions []region
	lastEnd := 0
	for _, r := range regions {
		if r.Begin < lastEnd {
			continue
		}
		newRegions = append(newRegions, r)
		lastEnd = r.End
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
	for _, child := range parse.Children(n) {
		emitRegions(child, f)
	}
}

func emitRegionsInForm(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Special forms.
	// TODO: This only highlights bareword special commands, however currently
	// quoted special commands are also possible (e.g `"if" $true { }` is
	// accepted).
	head := sourceText(n.Head)
	switch head {
	case "var", "set", "tmp":
		emitRegionsInAssign(n, f)
	case "del":
		emitRegionsInDel(n, f)
	case "if":
		emitRegionsInIf(n, f)
	case "for":
		emitRegionsInFor(n, f)
	case "try":
		emitRegionsInTry(n, f)
	}
	if isBarewordCompound(n.Head) {
		f(n.Head, semanticRegion, commandRegion)
	}
}

func emitRegionsInAssign(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight all LHS, and = as a keyword.
	for _, arg := range n.Args {
		if parse.SourceText(arg) == "=" {
			f(arg, semanticRegion, keywordRegion)
			break
		}
		emitVariableRegion(arg, f)
	}
}

func emitRegionsInDel(n *parse.Form, f func(parse.Node, regionKind, string)) {
	for _, arg := range n.Args {
		emitVariableRegion(arg, f)
	}
}

func emitVariableRegion(n *parse.Compound, f func(parse.Node, regionKind, string)) {
	// Only handle valid LHS here. Invalid LHS will result in a compile error
	// and highlighted as an error accordingly.
	if n != nil && len(n.Indexings) == 1 && n.Indexings[0].Head != nil {
		f(n.Indexings[0].Head, semanticRegion, variableRegion)
	}
}

func isBarewordCompound(n *parse.Compound) bool {
	return len(n.Indexings) == 1 && len(n.Indexings[0].Indices) == 0 && n.Indexings[0].Head.Type == parse.Bareword
}

func emitRegionsInIf(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight all "elif" and "else".
	for i := 2; i < len(n.Args); i += 2 {
		arg := n.Args[i]
		if s := sourceText(arg); s == "elif" || s == "else" {
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
	if 3 < len(n.Args) && sourceText(n.Args[3]) == "else" {
		f(n.Args[3], semanticRegion, keywordRegion)
	}
}

func emitRegionsInTry(n *parse.Form, f func(parse.Node, regionKind, string)) {
	// Highlight "except", the exception variable after it, "else" and
	// "finally".
	i := 1
	matchKW := func(text string) bool {
		if i < len(n.Args) && sourceText(n.Args[i]) == text {
			f(n.Args[i], semanticRegion, keywordRegion)
			return true
		}
		return false
	}
	if matchKW("except") || matchKW("catch") {
		if i+1 < len(n.Args) && isStringLiteral(n.Args[i+1]) {
			f(n.Args[i+1], semanticRegion, variableRegion)
			i += 3
		} else {
			i += 2
		}
	}
	if matchKW("else") {
		i += 2
	}
	matchKW("finally")
}

func isStringLiteral(n *parse.Compound) bool {
	_, ok := cmpd.StringLiteral(n)
	return ok
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
	text := sourceText(n)
	trimmed := strings.TrimLeftFunc(text, parse.IsWhitespace)
	switch {
	case trimmed == "":
		// Don't do anything; whitespaces do not get highlighted.
	case strings.HasPrefix(trimmed, "#"):
		f(n, lexicalRegion, commentRegion)
	default:
		f(n, lexicalRegion, text)
	}
}
