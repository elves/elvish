package util

import (
	"bytes"
	"unicode"
)

// CamelToDashed converts a CamelCaseIdentifier to a dash-separated-identifier,
// or a camelCaseIdentifier to a -dash-separated-identifier.
func CamelToDashed(camel string) string {
	var buf bytes.Buffer
	for i, r := range camel {
		if (i == 0 && unicode.IsLower(r)) || (i > 0 && unicode.IsUpper(r)) {
			buf.WriteRune('-')
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return buf.String()
}
