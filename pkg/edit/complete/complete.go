// Package complete implements the code completion algorithm for Elvish.
package complete

import (
	"errors"
	"sort"

	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/np"
)

// An error returned by Complete as well as the completers if there is no
// applicable completion.
var errNoCompletion = errors.New("no completion")

// Config stores the configuration required for code completion.
type Config struct {
	// A function for filtering raw candidates. If nil, no filtering is done.
	Filterer Filterer
	// Used to generate candidates for a command argument. Defaults to
	// GenerateFileNames.
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
	Items   []modes.CompletionItem
}

// RawItem represents completion items before the quoting pass.
type RawItem interface {
	String() string
	Cook(parse.PrimaryType) modes.CompletionItem
}

// CodeBuffer is the same the type in src.elv.sh/pkg/el/codearea,
// replicated here to avoid an unnecessary dependency.
type CodeBuffer struct {
	Content string
	Dot     int
}

// Complete runs the code completion algorithm in the given context, and returns
// the completion type, items and any error encountered.
func Complete(code CodeBuffer, ev *eval.Evaler, cfg Config) (*Result, error) {
	if cfg.Filterer == nil {
		cfg.Filterer = FilterPrefix
	}
	if cfg.ArgGenerator == nil {
		cfg.ArgGenerator = GenerateFileNames
	}

	// Ignore the error; the function always returns a valid *ChunkNode.
	tree, _ := parse.Parse(parse.Source{Name: "[interactive]", Code: code.Content}, parse.Config{})
	path := np.FindLeft(tree.Root, code.Dot)
	if len(path) == 0 {
		// This can happen when there is a parse error.
		return nil, errNoCompletion
	}
	for _, completer := range completers {
		ctx, rawItems, err := completer(path, ev, cfg)
		if err == errNoCompletion {
			continue
		}
		rawItems = cfg.Filterer(ctx.name, ctx.seed, rawItems)
		sort.Slice(rawItems, func(i, j int) bool {
			return rawItems[i].String() < rawItems[j].String()
		})
		items := make([]modes.CompletionItem, len(rawItems))
		for i, rawCand := range rawItems {
			items[i] = rawCand.Cook(ctx.quote)
		}
		items = dedup(items)
		return &Result{Name: ctx.name, Items: items, Replace: ctx.interval}, nil
	}
	return nil, errNoCompletion
}

func dedup(items []modes.CompletionItem) []modes.CompletionItem {
	var result []modes.CompletionItem
	for i, item := range items {
		if i == 0 || item.ToInsert != items[i-1].ToInsert {
			result = append(result, item)
		}
	}
	return result
}
