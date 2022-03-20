package tt

import (
	"math/big"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"src.elv.sh/pkg/persistent/hashmap"
	"src.elv.sh/pkg/persistent/vector"
)

// CommonCmpOpt is cmp.Option shared between tt and evaltest.
var CommonCmpOpt = cmp.Options([]cmp.Option{
	cmp.Transformer("transformList", transformList),
	cmp.Transformer("transformMap", transformMap),
	cmp.Comparer(func(x, y *big.Int) bool { return x.Cmp(y) == 0 }),
	cmp.Comparer(func(x, y *big.Rat) bool { return x.Cmp(y) == 0 }),
})

var cmpopt = cmp.Options([]cmp.Option{
	cmpopts.EquateErrors(),
	CommonCmpOpt,
})

func transformList(l vector.Vector) []any {
	res := make([]any, 0, l.Len())
	for it := l.Iterator(); it.HasElem(); it.Next() {
		res = append(res, it.Elem())
	}
	return res
}

func transformMap(m hashmap.Map) map[any]any {
	res := make(map[any]any, m.Len())
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		res[k] = v
	}
	return res
}
