package vals

import (
	"errors"
	"unicode/utf8"
)

var errIndexNotAtRuneBoundary = errors.New("index not at rune boundary")

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
		return 0, 0, errIndexNotAtRuneBoundary
	}
	if index.Slice {
		if r, _ := utf8.DecodeLastRuneInString(s[:index.Upper]); r == utf8.RuneError {
			return 0, 0, errIndexNotAtRuneBoundary
		}
		return index.Lower, index.Upper, nil
	}
	return index.Lower, index.Lower + size, nil
}
