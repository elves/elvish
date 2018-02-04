package types

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/xiaq/persistent/vector"
)

// Getter wraps the Get method.
type Getter interface {
	// Get retrieves the value corresponding to the specified key in the
	// container. It returns the value (if any), and whether it actually
	// exists.
	Get(k interface{}) (v interface{}, ok bool)
}

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(idx interface{}) (interface{}, error)
}

var (
	errIndexMustBeString = errors.New("index must be string")
	errNotIndexable      = errors.New("not indexable")
	errBadIndex          = errors.New("bad index")
	errIndexOutOfRange   = errors.New("index out of range")
)

type noSuchKeyError struct {
	key interface{}
}

// NoSuchKey returns an error indicating that a key is not found in a map-like
// value.
func NoSuchKey(k interface{}) error {
	return noSuchKeyError{k}
}

func (err noSuchKeyError) Error() string {
	return "no such key: " + Repr(err.key, NoPretty)
}

// Index indexes a value with the given key. It is implemented for the builtin
// type string, and types satisfying the listIndexable, Getter or Indexer
// interface. For other types, it returns a nil value and a non-nil error.
func Index(a, k interface{}) (interface{}, error) {
	switch a := a.(type) {
	case Indexer:
		return a.Index(k)
	case string:
		return indexString(a, k)
	case listIndexable:
		return indexList(a, k)
	case Getter:
		v, ok := a.Get(k)
		if !ok {
			return nil, NoSuchKey(k)
		}
		return v, nil
	default:
		return nil, errNotIndexable
	}
}

// MustIndex indexes i with k and returns the value. If the operation
// resulted in an error, it panics. It is useful when the caller knows that the
// key must be present.
func MustIndex(i Indexer, k interface{}) interface{} {
	v, err := i.Index(k)
	if err != nil {
		panic(err)
	}
	return v
}

func indexString(s string, index interface{}) (string, error) {
	i, j, err := convertStringIndex(index, s)
	if err != nil {
		return "", err
	}
	return s[i:j], nil
}

func convertStringIndex(rawIndex interface{}, s string) (int, int, error) {
	index, err := ConvertListIndex(rawIndex, len(s))
	if err != nil {
		return 0, 0, err
	}
	r, size := utf8.DecodeRuneInString(s[index.Lower:])
	if r == utf8.RuneError {
		return 0, 0, errBadIndex
	}
	if index.Slice {
		if r, _ := utf8.DecodeLastRuneInString(s[:index.Upper]); r == utf8.RuneError {
			return 0, 0, errBadIndex
		}
		return index.Lower, index.Upper, nil
	}
	return index.Lower, index.Lower + size, nil
}

type listIndexable interface {
	Lener
	Nth(int) interface{}
	SubVector(int, int) vector.Vector
}

var _ listIndexable = vector.Vector(nil)

func indexList(l listIndexable, rawIndex interface{}) (interface{}, error) {
	index, err := ConvertListIndex(rawIndex, l.Len())
	if err != nil {
		return nil, err
	}
	if index.Slice {
		return l.SubVector(index.Lower, index.Upper), nil
	}
	return l.Nth(index.Lower), nil
}

// ListIndex represents a (converted) list index.
type ListIndex struct {
	Slice bool
	Lower int
	Upper int
}

// ConvertListIndex parses a list index, check whether it is valid, and returns
// the converted structure.
func ConvertListIndex(rawIndex interface{}, n int) (*ListIndex, error) {
	s, ok := rawIndex.(string)
	if !ok {
		return nil, errIndexMustBeString
	}
	slice, i, j, err := parseListIndex(s, n)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i += n
	}
	if j < 0 {
		j += n
	}
	if slice {
		if !(0 <= i && i <= j && j <= n) {
			return nil, errIndexOutOfRange
		}
	} else {
		if !(0 <= i && i < n) {
			return nil, errIndexOutOfRange
		}
	}
	return &ListIndex{slice, i, j}, nil
}

// ListIndex = Number |
//             Number ':' Number
func parseListIndex(s string, n int) (slice bool, i int, j int, err error) {
	colon := strings.IndexRune(s, ':')
	if colon == -1 {
		// A single number
		i, err := atoi(s)
		if err != nil {
			return false, 0, 0, err
		}
		return false, i, 0, nil
	}
	if s[:colon] == "" {
		i = 0
	} else {
		i, err = atoi(s[:colon])
		if err != nil {
			return false, 0, 0, err
		}
	}
	if s[colon+1:] == "" {
		j = n
	} else {
		j, err = atoi(s[colon+1:])
		if err != nil {
			return false, 0, 0, err
		}
	}
	// Two numbers
	return true, i, j, nil
}

// atoi is a wrapper around strconv.Atoi, converting strconv.ErrRange to
// errIndexOutOfRange.
func atoi(a string) (int, error) {
	i, err := strconv.Atoi(a)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			return 0, errIndexOutOfRange
		}
		return 0, errBadIndex
	}
	return i, nil
}
