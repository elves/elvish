// Package complete implements the code completion algorithm for Elvish.
package complete

import (
	"errors"
	"sort"

	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/parse/parseutil"
)

// An error returned by Complete as well as the completers if there is no
// applicable completion.
var errNoCompletion = errors.New("no completion")

// Config stores the configuration required for code completion.
type Config struct {
	// A function for filtering raw candidates. If nil, no filtering is done.
	Filter func(ctxName, seed string, rawItems []RawItem) []RawItem
	// Used to generate candidates for a command argument. Defaults to
	// Filenames.
	ArgGenerator ArgGenerator
	// An interface to access the runtime. Must not be nil.
	PureEvaler PureEvaler
}

// ArgGenerator is the type of function that generates raw candidates for a
// command argument. It takes all the existing arguments, the last being the
// argument to complete, and returns raw candidates or an error.
type ArgGenerator func(args []string) ([]RawItem, error)

// Result keeps the result of the completion algorithm.
type Result struct {
	Name    string
	Replace diag.Ranging
	Items   []completion.Item
}

// RawItem represents completion items before the quoting pass.
type RawItem interface {
	String() string
	Cook(parse.PrimaryType) completion.Item
}

// PureEvaler encapsulates the functionality the completion algorithm needs from
// the language runtime.
type PureEvaler interface {
	EachExternal(func(cmd string))
	EachSpecial(func(special string))
	EachNs(func(string))
	EachVariableInNs(string, func(string))
	PurelyEvalPrimary(pn *parse.Primary) interface{}
	PurelyEvalCompound(*parse.Compound) (string, error)
	PurelyEvalPartialCompound(*parse.Compound, *parse.Indexing) (string, error)
}

// CodeBuffer is the same the type in github.com/elves/elvish/el/codearea,
// replicated here to avoid an unnecessary dependency.
type CodeBuffer struct {
	Content string
	Dot     int
}

// Complete runs the code completion algorithm in the given context, and returns
// the completion type, items and any error encountered.
func Complete(code CodeBuffer, cfg Config) (*Result, error) {
	if cfg.ArgGenerator == nil {
		cfg.ArgGenerator = GenerateFileNames
	}

	// Ignore the error; the function always returns a valid *ChunkNode.
	chunk, _ := parse.AsChunk("[interactive]", code.Content)
	leaf := parseutil.FindLeafNode(chunk, code.Dot)
	for _, completer := range completers {
		ctx, rawItems, err := completer(leaf, cfg)
		if err == errNoCompletion {
			continue
		}
		if cfg.Filter != nil {
			rawItems = cfg.Filter(ctx.name, ctx.seed, rawItems)
		}
		items := make([]completion.Item, len(rawItems))
		for i, rawCand := range rawItems {
			items[i] = rawCand.Cook(ctx.quote)
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].ToShow < items[j].ToShow
		})
		return &Result{Name: ctx.name, Items: items, Replace: ctx.interval}, nil
	}
	return nil, errNoCompletion
}
