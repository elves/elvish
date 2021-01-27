package complete

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

var parent = parse.Parent

var completers = []completer{
	completeCommand,
	completeIndex,
	completeRedir,
	completeVariable,
	completeArg,
}

type completer func(parse.Node, Config) (*context, []RawItem, error)

type context struct {
	name     string
	seed     string
	quote    parse.PrimaryType
	interval diag.Ranging
}

func completeArg(n parse.Node, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	if sep, ok := n.(*parse.Sep); ok {
		if form, ok := parent(sep).(*parse.Form); ok && form.Head != nil {
			// Case 1: starting a new argument.
			ctx := &context{"argument", "", parse.Bareword, range0(n.Range().To)}
			args := purelyEvalForm(form, "", n.Range().To, ev)
			items, err := cfg.ArgGenerator(args)
			return ctx, items, err
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := parent(compound).(*parse.Form); ok {
				if form.Head != nil && form.Head != compound {
					// Case 2: in an incomplete argument.
					ctx := &context{"argument", seed, primary.Type, compound.Range()}
					args := purelyEvalForm(form, seed, compound.Range().From, ev)
					items, err := cfg.ArgGenerator(args)
					return ctx, items, err
				}
			}
		}
	}
	return nil, nil, errNoCompletion
}

func completeCommand(n parse.Node, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	generateForEmpty := func(pos int) (*context, []RawItem, error) {
		ctx := &context{"command", "", parse.Bareword, range0(pos)}
		items, err := generateCommands("", ev)
		return ctx, items, err
	}

	if is(n, aChunk) {
		// Case 1: The leaf is a Chunk. That means that the chunk is empty
		// (nothing entered at all) and it is a correct place for completing a
		// command.
		return generateForEmpty(n.Range().To)
	}
	if is(n, aSep) {
		parent := parent(n)
		switch {
		case is(parent, aChunk), is(parent, aPipeline):
			// Case 2: Just after a newline, semicolon, or a pipe.
			return generateForEmpty(n.Range().To)
		case is(parent, aPrimary):
			ptype := parent.(*parse.Primary).Type
			if ptype == parse.OutputCapture || ptype == parse.ExceptionCapture {
				// Case 3: At the beginning of output or exception capture.
				return generateForEmpty(n.Range().To)
			}
		}
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := parent(compound).(*parse.Form); ok {
				if form.Head == compound {
					// Case 4: At an already started command.
					ctx := &context{
						"command", seed, primary.Type, compound.Range()}
					items, err := generateCommands(seed, ev)
					return ctx, items, err
				}
			}
		}
	}
	return nil, nil, errNoCompletion
}

// NOTE: This now only supports a single level of indexing; for instance,
// $a[<Tab> is supported, but $a[x][<Tab> is not.
func completeIndex(n parse.Node, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	generateForEmpty := func(v interface{}, pos int) (*context, []RawItem, error) {
		ctx := &context{"index", "", parse.Bareword, range0(pos)}
		return ctx, generateIndices(v), nil
	}

	if is(n, aSep) {
		if is(parent(n), aIndexing) {
			// We are just after an opening bracket.
			indexing := parent(n).(*parse.Indexing)
			if len(indexing.Indicies) == 1 {
				if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
					return generateForEmpty(indexee, n.Range().To)
				}
			}
		}
		if is(parent(n), aArray) {
			array := parent(n)
			if is(parent(array), aIndexing) {
				// We are after an existing index and spaces.
				indexing := parent(array).(*parse.Indexing)
				if len(indexing.Indicies) == 1 {
					if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
						return generateForEmpty(indexee, n.Range().To)
					}
				}
			}
		}
	}

	if is(n, aPrimary) {
		primary := n.(*parse.Primary)
		compound, seed := primaryInSimpleCompound(primary, ev)
		if compound != nil {
			if is(parent(compound), aArray) {
				array := parent(compound)
				if is(parent(array), aIndexing) {
					// We are just after an incomplete index.
					indexing := parent(array).(*parse.Indexing)
					if len(indexing.Indicies) == 1 {
						if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
							ctx := &context{
								"index", seed, primary.Type, compound.Range()}
							return ctx, generateIndices(indexee), nil
						}
					}
				}
			}
		}
	}
	return nil, nil, errNoCompletion
}

func completeRedir(n parse.Node, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	if is(n, aSep) {
		if is(parent(n), aRedir) {
			// Empty redirection target.
			ctx := &context{"redir", "", parse.Bareword, range0(n.Range().To)}
			items, err := generateFileNames("", false)
			return ctx, items, err
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if is(parent(compound), &parse.Redir{}) {
				// Non-empty redirection target.
				ctx := &context{
					"redir", seed, primary.Type, compound.Range()}
				items, err := generateFileNames(seed, false)
				return ctx, items, err
			}
		}
	}
	return nil, nil, errNoCompletion
}

func completeVariable(n parse.Node, cfg Config) (*context, []RawItem, error) {
	ev := cfg.PureEvaler
	primary, ok := n.(*parse.Primary)
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

func range0(pos int) diag.Ranging {
	return diag.Ranging{From: pos, To: pos}
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}
