package eval

import "bytes"

type Type interface {
	Default() Value
	String() string
}

type AnyType struct {
}

func (at AnyType) Default() Value {
	return NewString("")
}

func (at AnyType) String() string {
	return "any"
}

type StringType struct {
}

func (st StringType) Default() Value {
	return NewString("")
}

func (st StringType) String() string {
	return "string"
}

type BoolType struct {
}

func (bt BoolType) Default() Value {
	return Bool(true)
}

func (bt BoolType) String() string {
	return "bool"
}

type ExitusType struct {
}

func (et ExitusType) Default() Value {
	return success
}

func (et ExitusType) String() string {
	return "exitus"
}

type TableType struct {
}

func (tt TableType) Default() Value {
	return NewTable()
}

func (tt TableType) String() string {
	return "table"
}

type ClosureType struct {
}

func (st ClosureType) Default() Value {
	return NewClosure([]string{}, nil, map[string]Variable{})
}

func (ct ClosureType) String() string {
	return "closure"
}

var typenames = map[string]Type{
	"any":     AnyType{},
	"string":  StringType{},
	"exitus":  ExitusType{},
	"bool":    BoolType{},
	"table":   TableType{},
	"closure": ClosureType{},
}

func isAny(t Type) bool {
	_, ok := t.(AnyType)
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
	} else {
		return n == ntypes
	}
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
