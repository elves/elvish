// Package complete implements the code completion algorithm for Elvish.
package complete

import (
	"errors"
	"sort"

	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/parseutil"
)

type item = mode.CompletionItem

// An error returned by Complete if the config has not supplied a PureEvaler.
var errNoPureEvaler = errors.New("no PureEvaler supplied")

// An error returned by Complete as well as the completers if there is no
// applicable completion.
var errNoCompletion = errors.New("no completion")

// Config stores the configuration required for code completion.
type Config struct {
	// An interface to access the runtime. Complete will return an error if this
	// is nil.
	PureEvaler PureEvaler
	// A function for filtering raw candidates. If nil, no filtering is done.
	Filterer Filterer
	// Used to generate candidates for a command argument. Defaults to
	// Filenames.
	ArgGenerator ArgGenerator
}

// Filterer is the type of functions that filter raw candidates.
type Filterer func(ctxName, seed string, rawItems []RawItem) []RawItem

// ArgGenerator is the type of functions that generate raw candidates for a
// command argument. It takes all the existing arguments, the last being the
// argument to complete, and returns raw candidates or an error.
type ArgGenerator func(args []string) ([]RawItem, error)

// Result keeps the result of the completion algorithm.
type Result struct {
	Name    string
	Replace diag.Ranging
	Items   []mode.CompletionItem
}

// RawItem represents completion items before the quoting pass.
type RawItem interface {
	String() string
	Cook(parse.PrimaryType) mode.CompletionItem
}

// PureEvaler encapsulates the functionality the completion algorithm needs from
// the language runtime.
type PureEvaler interface {
	EachExternal(func(cmd string))
	EachSpecial(func(special string))
	EachNs(func(string))
	EachVariableInNs(string, func(string))
	PurelyEvalPrimary(pn *parse.Primary) interface{}
	PurelyEvalCompound(*parse.Compound) (string, bool)
	PurelyEvalPartialCompound(*parse.Compound, int) (string, bool)
}

// CodeBuffer is the same the type in src.elv.sh/pkg/el/codearea,
// replicated here to avoid an unnecessary dependency.
type CodeBuffer struct {
	Content string
	Dot     int
}

// Complete runs the code completion algorithm in the given context, and returns
// the completion type, items and any error encountered.
func Complete(code CodeBuffer, cfg Config) (*Result, error) {
	if cfg.PureEvaler == nil {
		return nil, errNoPureEvaler
	}
	if cfg.Filterer == nil {
		cfg.Filterer = FilterPrefix
	}
	if cfg.ArgGenerator == nil {
		cfg.ArgGenerator = GenerateFileNames
	}

	// Ignore the error; the function always returns a valid *ChunkNode.
	tree, _ := parse.Parse(parse.Source{Name: "[interactive]", Code: code.Content}, parse.Config{})
	leaf := parseutil.FindLeafNode(tree.Root, code.Dot)
	for _, completer := range completers {
		ctx, rawItems, err := completer(leaf, cfg)
		if err == errNoCompletion {
			continue
		}
		rawItems = cfg.Filterer(ctx.name, ctx.seed, rawItems)
		items := make([]mode.CompletionItem, len(rawItems))
		for i, rawCand := range rawItems {
			items[i] = rawCand.Cook(ctx.quote)
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].ToShow < items[j].ToShow
		})
		items = dedup(items)
		return &Result{Name: ctx.name, Items: items, Replace: ctx.interval}, nil
	}
	return nil, errNoCompletion
}

func dedup(items []mode.CompletionItem) []mode.CompletionItem {
	var result []mode.CompletionItem
	for i, item := range items {
		if i == 0 || item.ToInsert != items[i-1].ToInsert {
			result = append(result, item)
		}
	}
	return result
}
