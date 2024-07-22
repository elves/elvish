package complete

import (
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/np"
)

var completers = []func(np.Path, *eval.Evaler, Config) (*context, []RawItem, error){
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

func completeArg(p np.Path, ev *eval.Evaler, cfg Config) (*context, []RawItem, error) {
	var form *parse.Form
	if p.Match(np.Sep, np.Store(&form)) && form.Head != nil {
		// Case 1: starting a new argument.
		ctx := &context{"argument", "", parse.Bareword, range0(p[0].Range().To)}
		args := purelyEvalForm(form, "", p[0].Range().To, ev)
		items, err := generateArgs(args, ev, p, cfg)
		return ctx, items, err
	}

	var expr np.SimpleExprData
	if p.Match(np.SimpleExpr(&expr, ev), np.Store(&form)) && form.Head != nil && form.Head != expr.Compound {
		// Case 2: in an incomplete argument.
		ctx := &context{"argument", expr.Value, expr.PrimarType, expr.Compound.Range()}
		args := purelyEvalForm(form, expr.Value, expr.Compound.Range().From, ev)
		items, err := generateArgs(args, ev, p, cfg)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

func completeCommand(p np.Path, ev *eval.Evaler, cfg Config) (*context, []RawItem, error) {
	generateForEmpty := func(pos int) (*context, []RawItem, error) {
		ctx := &context{"command", "", parse.Bareword, range0(pos)}
		items, err := generateCommands("", ev, p)
		return ctx, items, err
	}

	if p.Match(np.Chunk) {
		// Case 1: The leaf is a Chunk. That means that the chunk is empty
		// (nothing entered at all) and it is a correct place for completing a
		// command.
		return generateForEmpty(p[0].Range().To)
	}
	if p.Match(np.Sep, np.Chunk) || p.Match(np.Sep, np.Pipeline) {
		// Case 2: Just after a newline, semicolon, or a pipe.
		return generateForEmpty(p[0].Range().To)
	}

	var primary *parse.Primary
	if p.Match(np.Sep, np.Store(&primary)) {
		t := primary.Type
		if t == parse.OutputCapture || t == parse.ExceptionCapture || t == parse.Lambda {
			// Case 3: At the beginning of output, exception capture or lambda.
			//
			// TODO: Don't trigger after "{|".
			return generateForEmpty(p[0].Range().To)
		}
	}

	var expr np.SimpleExprData
	var form *parse.Form
	if p.Match(np.SimpleExpr(&expr, ev), np.Store(&form)) && form.Head == expr.Compound {
		// Case 4: At an already started command.
		ctx := &context{"command", expr.Value, expr.PrimarType, expr.Compound.Range()}
		items, err := generateCommands(expr.Value, ev, p)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

// NOTE: This now only supports a single level of indexing; for instance,
// $a[<Tab> is supported, but $a[x][<Tab> is not.
func completeIndex(p np.Path, ev *eval.Evaler, cfg Config) (*context, []RawItem, error) {
	generateForEmpty := func(v any, pos int) (*context, []RawItem, error) {
		ctx := &context{"index", "", parse.Bareword, range0(pos)}
		return ctx, generateIndices(v), nil
	}

	var indexing *parse.Indexing
	if p.Match(np.Sep, np.Store(&indexing)) || p.Match(np.Sep, np.Array, np.Store(&indexing)) {
		// We are at a new index, either directly after the opening bracket, or
		// after an existing index and some spaces.
		if len(indexing.Indices) == 1 {
			if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
				return generateForEmpty(indexee, p[0].Range().To)
			}
		}
	}

	var expr np.SimpleExprData
	if p.Match(np.SimpleExpr(&expr, ev), np.Array, np.Store(&indexing)) {
		// We are just after an incomplete index.
		if len(indexing.Indices) == 1 {
			if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
				ctx := &context{
					"index", expr.Value, expr.PrimarType, expr.Compound.Range()}
				return ctx, generateIndices(indexee), nil
			}
		}
	}

	return nil, nil, errNoCompletion
}

func completeRedir(p np.Path, ev *eval.Evaler, cfg Config) (*context, []RawItem, error) {
	if p.Match(np.Sep, np.Redir) {
		// Empty redirection target.
		ctx := &context{"redir", "", parse.Bareword, range0(p[0].Range().To)}
		items, err := generateFileNames("", nil)
		return ctx, items, err
	}

	var expr np.SimpleExprData
	if p.Match(np.SimpleExpr(&expr, ev), np.Redir) {
		// Non-empty redirection target.
		ctx := &context{"redir", expr.Value, expr.PrimarType, expr.Compound.Range()}
		items, err := generateFileNames(expr.Value, nil)
		return ctx, items, err
	}

	return nil, nil, errNoCompletion
}

func completeVariable(p np.Path, ev *eval.Evaler, cfg Config) (*context, []RawItem, error) {
	primary, ok := p[0].(*parse.Primary)
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
	eachVariableInNs(ev, p, ns, func(varname string) {
		items = append(items, noQuoteItem(parse.QuoteVariableName(varname)))
	})
	if ns == "" {
		items = append(items, noQuoteItem("e:"), noQuoteItem("E:"))
	}

	return ctx, items, nil
}

func purelyEvalForm(form *parse.Form, seed string, upto int, ev *eval.Evaler) []string {
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
