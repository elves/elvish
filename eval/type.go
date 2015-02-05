package eval

import "bytes"

// Type represents the static information of a Value.
type Type interface {
	Default() Value
	String() string
}

// anyType is a special type that may be assigned or assigned to any another
// type.
type anyType struct {
}

func (at anyType) Default() Value {
	return str("")
}

func (at anyType) String() string {
	return "any"
}

type stringType struct {
}

func (st stringType) Default() Value {
	return str("")
}

func (st stringType) String() string {
	return "string"
}

type boolType struct {
}

func (bt boolType) Default() Value {
	return boolean(true)
}

func (bt boolType) String() string {
	return "bool"
}

type exitusType struct {
}

func (et exitusType) Default() Value {
	return success
}

func (et exitusType) String() string {
	return "exitus"
}

type tableType struct {
}

func (tt tableType) Default() Value {
	return newTable()
}

func (tt tableType) String() string {
	return "table"
}

type callableType struct {
}

func (ct callableType) Default() Value {
	return newClosure([]string{}, nil, map[string]Variable{})
}

func (ct callableType) String() string {
	return "callable"
}

type ratType struct{}

func (rt ratType) Default() Value {
	return newRat()
}

func (rt ratType) String() string {
	return "rat"
}

var typenames = map[string]Type{
	"any":      anyType{},
	"string":   stringType{},
	"exitus":   exitusType{},
	"bool":     boolType{},
	"table":    tableType{},
	"callable": callableType{},
	"rat":      ratType{},
}

func isAny(t Type) bool {
	_, ok := t.(anyType)
	return ok
}

// typeStar is either a single Type or a Kleene star of a Type.
type typeStar struct {
	t    Type
	star bool
}

func (ts typeStar) String() string {
	s := ts.t.String()
	if ts.star {
		s += "*"
	}
	return s
}

// typeRun is a run of typeStar's.
type typeRun []typeStar

func (tr typeRun) String() string {
	var buf bytes.Buffer
	for i, ts := range tr {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(ts.String())
	}
	return buf.String()
}

// count returns the least number of Type's in a typeRun, and whether the
// actual number could be larger.
func (tr typeRun) count() (ntypes int, more bool) {
	for _, ts := range tr {
		if ts.star {
			more = true
		} else {
			ntypes++
		}
	}
	return
}

// mayCountTo returns whether a run of n Type's may match the typeRun.
func (tr typeRun) mayCountTo(n int) bool {
	ntypes, more := tr.count()
	if more {
		return n >= ntypes
	}
	return n == ntypes
}

// newFixedTypeRun returns a typeRun where all typeStar's are simple
// non-starred types.
func newFixedTypeRun(ts ...Type) typeRun {
	tr := make([]typeStar, len(ts))
	for i, t := range ts {
		tr[i].t = t
	}
	return tr
}

// newFixedTypeRun returns a typeRun representing n types of t, and followed by
// a star of t if v is true.
func newHomoTypeRun(t Type, n int, v bool) typeRun {
	tr := make([]typeStar, n)
	for i := 0; i < n; i++ {
		tr[i].t = t
	}
	if v {
		tr = append(tr, typeStar{t, true})
	}
	return tr
}
