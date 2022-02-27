package complete

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

var completers = []func(nodePath, Config) (*context, []RawItem, error){
	completeCommand,
	completeIndex,
	completeRedir,
	completeVariable,
	completeArg,
}

type context struct {
	name     string
	seed     string
	quote    parse.PrimaryType
	interval diag.Ranging
}

func completeArg(np nodePath, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler

	var form *parse.Form
	if np.match(aSep, store(&form)) && form.Head != nil {
		// Case 1: starting a new argument.
		ctx := &context{"argument", "", parse.Bareword, range0(np[0].Range().To)}
		args := purelyEvalForm(form, "", np[0].Range().To, ev)
		items, err := generateArgs(args, cfg)
		return ctx, items, err
	}

	expr := simpleExpr(ev)
	if np.match(expr, store(&form)) && form.Head != nil && form.Head != expr.compound {
		// Case 2: in an incomplete argument.
		ctx := &context{"argument", expr.s, expr.quote, expr.compound.Range()}
		args := purelyEvalForm(form, expr.s, expr.compound.Range().From, ev)
		items, err := generateArgs(args, cfg)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

func completeCommand(np nodePath, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	generateForEmpty := func(pos int) (*context, []RawItem, error) {
		ctx := &context{"command", "", parse.Bareword, range0(pos)}
		items, err := generateCommands("", ev)
		return ctx, items, err
	}

	if np.match(aChunk) {
		// Case 1: The leaf is a Chunk. That means that the chunk is empty
		// (nothing entered at all) and it is a correct place for completing a
		// command.
		return generateForEmpty(np[0].Range().To)
	}
	if np.match(aSep, aChunk) || np.match(aSep, aPipeline) {
		// Case 2: Just after a newline, semicolon, or a pipe.
		return generateForEmpty(np[0].Range().To)
	}

	var primary *parse.Primary
	if np.match(aSep, store(&primary)) {
		t := primary.Type
		if t == parse.OutputCapture || t == parse.ExceptionCapture || t == parse.Lambda {
			// Case 3: At the beginning of output, exception capture or lambda.
			//
			// TODO: Don't trigger after "{|".
			return generateForEmpty(np[0].Range().To)
		}
	}

	expr := simpleExpr(ev)
	var form *parse.Form
	if np.match(expr, store(&form)) && form.Head == expr.compound {
		// Case 4: At an already started command.
		ctx := &context{"command", expr.s, expr.quote, expr.compound.Range()}
		items, err := generateCommands(expr.s, ev)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

// NOTE: This now only supports a single level of indexing; for instance,
// $a[<Tab> is supported, but $a[x][<Tab> is not.
func completeIndex(np nodePath, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	generateForEmpty := func(v interface{}, pos int) (*context, []RawItem, error) {
		ctx := &context{"index", "", parse.Bareword, range0(pos)}
		return ctx, generateIndices(v), nil
	}

	var indexing *parse.Indexing
	if np.match(aSep, store(&indexing)) || np.match(aSep, aArray, store(&indexing)) {
		// We are at a new index, either directly after the opening bracket, or
		// after an existing index and some spaces.
		if len(indexing.Indices) == 1 {
			if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
				return generateForEmpty(indexee, np[0].Range().To)
			}
		}
	}

	expr := simpleExpr(ev)
	if np.match(expr, aArray, store(&indexing)) {
		// We are just after an incomplete index.
		if len(indexing.Indices) == 1 {
			if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
				ctx := &context{
					"index", expr.s, expr.quote, expr.compound.Range()}
				return ctx, generateIndices(indexee), nil
			}
		}
	}

	return nil, nil, errNoCompletion
}

func completeRedir(np nodePath, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	if np.match(aSep, aRedir) {
		// Empty redirection target.
		ctx := &context{"redir", "", parse.Bareword, range0(np[0].Range().To)}
		items, err := generateFileNames("", false)
		return ctx, items, err
	}

	expr := simpleExpr(ev)
	if np.match(expr, aRedir) {
		// Non-empty redirection target.
		ctx := &context{"redir", expr.s, expr.quote, expr.compound.Range()}
		items, err := generateFileNames(expr.s, false)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

func completeVariable(np nodePath, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	primary, ok := np[0].(*parse.Primary)
	if !ok || primary.Type != parse.Variable {
		return nil, nil, errNoCompletion
	}
	sigil, qname := eval.SplitSigil(primary.Value)
	ns, nameSeed := eval.SplitIncompleteQNameNs(qname)
	// Move past "$", "@" and "<ns>:".
	begin := primary.Range().From + 1 + len(sigil) + len(ns)

	ctx := &context{
		"variable", nameSeed, parse.Bareword,
		diag.Ranging{From: begin, To: primary.Range().To}}

	var items []RawItem
	ev.EachVariableInNs(ns, func(varname string) {
		items = append(items, noQuoteItem(parse.QuoteVariableName(varname)))
	})

	ev.EachNs(func(thisNs string) {
		// This is to match namespaces that are "nested" under the current
		// namespace.
		if hasProperPrefix(thisNs, ns) {
			items = append(items, noQuoteItem(parse.QuoteVariableName(thisNs[len(ns):])))
		}
	})

	return ctx, items, nil
}

func purelyEvalForm(form *parse.Form, seed string, upto int, ev PureEvaler) []string {
	// Find out head of the form and preceding arguments.
	// If form.Head is not a simple compound, head will be "", just what we want.
	head, _ := ev.PurelyEvalPartialCompound(form.Head, -1)
	words := []string{head}
	for _, compound := range form.Args {
		if compound.Range().From >= upto {
			break
		}
		if arg, ok := ev.PurelyEvalCompound(compound); ok {
			// TODO(xiaq): Arguments that are not simple compounds are simply ignored.
			words = append(words, arg)
		}
	}

	words = append(words, seed)
	return words
}

func range0(pos int) diag.Ranging {
	return diag.Ranging{From: pos, To: pos}
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}
