package program

import (
	"bytes"
	"fmt"
)

func quoteJSON(s string) string {
	var b bytes.Buffer
	b.WriteRune('"')
	for _, r := range s {
		if r == '\\' {
			b.WriteString(`\\`)
		} else if r == '"' {
			b.WriteString(`\"`)
		} else if r < 0x20 {
			fmt.Fprintf(&b, `\u%04x`, r)
		} else {
			b.WriteRune(r)
		}
	}
	b.WriteRune('"')
	return b.String()
}
