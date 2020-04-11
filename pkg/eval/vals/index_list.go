package vals

import (
	"errors"
	"strconv"
	"strings"

	"github.com/elves/elvish/pkg/eval/errs"
)

var (
	errIndexMustBeInteger = errors.New("index must must be integer")
)

func indexList(l List, rawIndex interface{}) (interface{}, error) {
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

func adjustAndCheckIndex(i, n int, includeN bool) (int, error) {
	if i < 0 {
		if i < -n {
			return 0, negIndexOutOfRange(strconv.Itoa(i), n)
		}
		return i + n, nil
	}
	if includeN {
		if i > n {
			return 0, posIndexOutOfRange(strconv.Itoa(i), n+1)
		}
	} else {
		if i >= n {
			return 0, posIndexOutOfRange(strconv.Itoa(i), n)
		}
	}
	return i, nil
}

// ConvertListIndex parses a list index, check whether it is valid, and returns
// the converted structure.
func ConvertListIndex(rawIndex interface{}, n int) (*ListIndex, error) {
	switch rawIndex := rawIndex.(type) {
	case float64:
		index := int(rawIndex)
		if rawIndex != float64(index) {
			return nil, errIndexMustBeInteger
		}
		index, err := adjustAndCheckIndex(index, n, false)
		if err != nil {
			return nil, err
		}
		return &ListIndex{false, index, 0}, nil
	case string:
		slice, i, j, err := parseListIndex(rawIndex, n)
		if err != nil {
			return nil, err
		}
		if !slice {
			i, err = adjustAndCheckIndex(i, n, false)
			if err != nil {
				return nil, err
			}
		} else {
			i, err = adjustAndCheckIndex(i, n, true)
			if err != nil {
				return nil, err
			}
			j0 := j
			j, err = adjustAndCheckIndex(j, n, true)
			if err != nil {
				return nil, err
			}
			if j < i {
				if j0 < 0 {
					return nil, errs.OutOfRange{
						What:     "negative slice upper index here",
						ValidLow: i - n, ValidHigh: -1, Actual: strconv.Itoa(j0)}
				} else {
					return nil, errs.OutOfRange{
						What:     "slice upper index here",
						ValidLow: i, ValidHigh: n, Actual: strconv.Itoa(j0)}
				}
			}
		}
		return &ListIndex{slice, i, j}, nil
	default:
		return nil, errIndexMustBeInteger
	}
}

// ListIndex = Number |
//             Number ':' Number
func parseListIndex(s string, n int) (slice bool, i int, j int, err error) {
	colon := strings.IndexRune(s, ':')
	if colon == -1 {
		// A single number
		i, err := atoi(s, n)
		if err != nil {
			return false, 0, 0, err
		}
		return false, i, 0, nil
	}
	if s[:colon] == "" {
		i = 0
	} else {
		i, err = atoi(s[:colon], n)
		if err != nil {
			return false, 0, 0, err
		}
	}
	if s[colon+1:] == "" {
		j = n
	} else {
		j, err = atoi(s[colon+1:], n)
		if err != nil {
			return false, 0, 0, err
		}
	}
	// Two numbers
	return true, i, j, nil
}

// atoi is a wrapper around strconv.Atoi, converting strconv.ErrRange to
// errs.OutOfRange.
func atoi(a string, n int) (int, error) {
	i, err := strconv.Atoi(a)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			if i < 0 {
				return 0, negIndexOutOfRange(a, n)
			}
			return 0, posIndexOutOfRange(a, n)
		}
		return 0, errIndexMustBeInteger
	}
	return i, nil
}

func posIndexOutOfRange(index string, n int) errs.OutOfRange {
	return errs.OutOfRange{
		What:     "index here",
		ValidLow: 0, ValidHigh: n - 1, Actual: index}
}

func negIndexOutOfRange(index string, n int) errs.OutOfRange {
	return errs.OutOfRange{
		What:     "negative index here",
		ValidLow: -n, ValidHigh: -1, Actual: index}
}
