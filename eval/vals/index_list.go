package vals

import (
	"errors"
	"strconv"
	"strings"

	"github.com/xiaq/persistent/vector"
)

var (
	errIndexMustBeString = errors.New("index must be string")
	errIndexMustBeNumber = errors.New("index or slice component must be number")
	errIndexOutOfRange   = errors.New("index out of range")
)

type listIndexable interface {
	Lener
	Index(int) (interface{}, bool)
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
	// Bounds are already checked.
	value, _ := l.Index(index.Lower)
	return value, nil
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
		return 0, errIndexMustBeNumber
	}
	return i, nil
}
