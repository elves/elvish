package edit

import (
	"errors"
	"sort"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
)

var errValueShouldBeFn = errors.New("value should be function")

func getBinding(bindingVar vartypes.Variable, k ui.Key) eval.Fn {
	binding := bindingVar.Get().(BindingTable)
	switch {
	case binding.HasKey(k):
		return binding.get(k)
	case binding.HasKey(ui.Default):
		return binding.get(ui.Default)
	default:
		return nil
	}
}

// BindingTable is a special Map that converts its key to ui.Key and ensures
// that its values satisfy eval.CallableValue.
type BindingTable struct {
	types.Map
}

var (
	_ types.Value   = BindingTable{}
	_ types.MapLike = BindingTable{}
)

// Repr returns the representation of the binding table as if it were an
// ordinary map keyed by strings.
func (bt BindingTable) Repr(indent int) string {
	var builder types.MapReprBuilder
	builder.Indent = indent

	var keys ui.Keys
	bt.Map.IterateKey(func(k types.Value) bool {
		keys = append(keys, k.(ui.Key))
		return true
	})
	sort.Sort(keys)

	for _, k := range keys {
		v, err := bt.Map.Index(k)
		if err != nil {
			panic(err)
		}
		builder.WritePair(parse.Quote(k.String()), indent+2, types.Repr(v, indent+2))
	}

	return builder.String()
}

// Index converts the index to ui.Key and uses the Index of the inner Map.
func (bt BindingTable) Index(idx types.Value) (types.Value, error) {
	return bt.Map.Index(ui.ToKey(idx))
}

func (bt BindingTable) get(k ui.Key) eval.Fn {
	v, err := bt.Map.Index(k)
	if err != nil {
		panic(err)
	}
	return v.(eval.Fn)
}

// Assoc converts the index to ui.Key, ensures that the value is CallableValue,
// uses the Assoc of the inner Map and converts the result to a BindingTable.
func (bt BindingTable) Assoc(k, v types.Value) (types.Value, error) {
	key := ui.ToKey(k)
	f, ok := v.(eval.Fn)
	if !ok {
		return nil, errValueShouldBeFn
	}
	map2, err := bt.Map.Assoc(key, f)
	if err != nil {
		return nil, err
	}
	return BindingTable{map2.(types.Map)}, nil
}

func makeBindingTable(f *eval.Frame, args []types.Value, opts map[string]types.Value) {
	var raw types.Map
	eval.ScanArgs(args, &raw)
	eval.TakeNoOpt(opts)

	converted := types.EmptyMapInner
	raw.IteratePair(func(k, v types.Value) bool {
		f, ok := v.(eval.Fn)
		if !ok {
			throw(errValueShouldBeFn)
		}
		converted = converted.Assoc(ui.ToKey(k), f)
		return true
	})

	f.OutputChan() <- BindingTable{types.NewMap(converted)}
}
