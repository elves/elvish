package strutil

import (
	"strings"
	"unicode"
)

// CamelToDashed converts a CamelCaseIdentifier to a dash-separated-identifier,
// or a camelCaseIdentifier to a -dash-separated-identifier. All-cap words
// are converted to lower case; HTTP becomes http and HTTPRequest becomes
// http-request.
func CamelToDashed(camel string) string {
	var sb strings.Builder
	runes := []rune(camel)
	for i, r := range runes {
		if (i == 0 && unicode.IsLower(r)) ||
			(0 < i && i < len(runes)-1 &&
				unicode.IsUpper(r) && unicode.IsLower(runes[i+1])) {
			sb.WriteRune('-')
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return sb.String()
}
