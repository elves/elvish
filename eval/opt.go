package eval

import (
	"errors"

	"github.com/elves/elvish/parse"
)

type OptSpec struct {
	Name    string
	Default Value
}

type OptSet struct {
	optSpecs []OptSpec
	indices  map[string]int
}

func NewOptSet(optSpecs ...OptSpec) *OptSet {
	indices := make(map[string]int)
	for i, optSpec := range optSpecs {
		indices[optSpec.Name] = i
	}
	return &OptSet{optSpecs, indices}
}

func (os *OptSet) Pick(opts map[string]Value) ([]Value, error) {
	vs := make([]Value, len(os.optSpecs))
	for k, v := range opts {
		if i, ok := os.indices[k]; ok {
			vs[i] = v
		} else {
			return nil, errors.New("unknown option " + parse.Quote(k))
		}
	}
	for i, optSpec := range os.optSpecs {
		if vs[i] == nil {
			vs[i] = optSpec.Default
		}
	}
	return vs, nil
}

func (os *OptSet) MustPick(opts map[string]Value) []Value {
	vs, err := os.Pick(opts)
	maybeThrow(err)
	return vs
}
