package eval

import (
	"testing"
)

func TestRegisterVariableGetQname(t *testing.T) {
	cp := &compiler{
		builtin: staticNs{"builtin": struct{}{}},
		scopes: []staticNs{
			{"up-scope": struct{}{}},
			{"local-scope": struct{}{}},
		},
		capture: staticNs{"capture": struct{}{}},
		begin:   0,
		end:     0,
		srcMeta: nil,
	}

	for name, exp := range map[string]bool{
		"builtin":     true,
		"local-scope": true,
		"up-scope": true,
		"not-exists": false,

		"builtin:builtin": true,
		"local:local-scope": true,
		"up:up-scope": true,

		"e:foo": true,
		"E:bar": true,
		"shared:baz": true,
	} {
		if cp.registerVariableGetQname(name) != exp {
			t.Errorf("getting \"%s\" returns %v", name, !exp)
		}
	}
}
