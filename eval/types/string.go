package types

import (
	"errors"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

// String is just a string.
type String string

var (
	_ Value    = String("")
	_ ListLike = String("")
)

var ErrReplacementMustBeString = errors.New("replacement must be string")

func (String) Kind() string {
	return "string"
}

func (s String) Repr(int) string {
	return parse.Quote(string(s))
}

func (s String) Equal(rhs interface{}) bool {
	return s == rhs
}

func (s String) Hash() uint32 {
	return hash.String(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Len() int {
	return len(string(s))
}

func (s String) IndexOne(idx Value) (Value, error) {
	i, j, err := s.index(idx)
	if err != nil {
		return nil, err
	}
	return s[i:j], nil
}

func (s String) Assoc(idx, v Value) Value {
	i, j, err := s.index(idx)
	maybeThrow(err)
	repl, ok := v.(String)
	if !ok {
		throw(ErrReplacementMustBeString)
	}
	return s[:i] + repl + s[j:]
}

func (s String) index(idx Value) (int, int, error) {
	slice, i, j, err := ParseAndFixListIndex(ToString(idx), len(s))
	if err != nil {
		return 0, 0, err
	}
	r, size := utf8.DecodeRuneInString(string(s[i:]))
	if r == utf8.RuneError {
		return 0, 0, ErrBadIndex
	}
	if slice {
		if r, _ := utf8.DecodeLastRuneInString(string(s[:j])); r == utf8.RuneError {
			return 0, 0, ErrBadIndex
		}
		return i, j, nil
	}
	return i, i + size, nil
}

func (s String) Iterate(f func(v Value) bool) {
	for _, r := range s {
		b := f(String(string(r)))
		if !b {
			break
		}
	}
}
