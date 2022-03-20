package vals

import (
	"errors"
	"unicode/utf8"
)

var errIndexNotAtRuneBoundary = errors.New("index not at rune boundary")

func indexString(s string, index any) (string, error) {
	i, j, err := convertStringIndex(index, s)
	if err != nil {
		return "", err
	}
	return s[i:j], nil
}

func convertStringIndex(rawIndex any, s string) (int, int, error) {
	index, err := ConvertListIndex(rawIndex, len(s))
	if err != nil {
		return 0, 0, err
	}
	if index.Slice {
		lower, upper := index.Lower, index.Upper
		if startsWithRuneBoundary(s[lower:]) && endsWithRuneBoundary(s[:upper]) {
			return lower, upper, nil
		}
		return 0, 0, errIndexNotAtRuneBoundary
	}
	// Not slice
	r, size := utf8.DecodeRuneInString(s[index.Lower:])
	if r == utf8.RuneError {
		return 0, 0, errIndexNotAtRuneBoundary
	}
	return index.Lower, index.Lower + size, nil
}

func startsWithRuneBoundary(s string) bool {
	if s == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r != utf8.RuneError
}

func endsWithRuneBoundary(s string) bool {
	if s == "" {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return r != utf8.RuneError
}
