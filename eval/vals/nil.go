package vals

import "strings"

// JsonNil Represents the nil value, without actually being nil which
// otherwise causes problems doing equality testing.  It is present
// for JSON interop.  Previously, a null in JSON would be turned into
// an empty string, but that does not work well when dealing with code
// that expects an empty string to be distinct from null.
type JsonNil struct{}

func (j JsonNil) Repr(i int) string {
	res := "$-json-nil"
	if i > 0 {
		res = strings.Repeat(" ", i) + res
	}
	return res
}

func (j JsonNil) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (j JsonNil) UnmarshalJSON([]byte) error {
	return nil
}
