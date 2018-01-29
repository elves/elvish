package eval

import "strings"

// ParseVariableRef parses a variable reference.
func ParseVariableRef(text string) (explode bool, ns string, name string) {
	explodePart, nsPart, name := SplitVariableRef(text)
	ns = nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}
	return explodePart != "", ns, name
}

// SplitVariableRef splits a variable reference into three parts: an optional
// explode operator (either "" or "@"), a namespace part, and a name part.
func SplitVariableRef(text string) (explodePart, nsPart, name string) {
	if text == "" {
		return "", "", ""
	}
	e, qname := "", text
	if text[0] == '@' {
		e = "@"
		qname = text[1:]
	}
	if qname == "" {
		return e, "", ""
	}
	i := strings.LastIndexByte(qname, ':')
	return e, qname[:i+1], qname[i+1:]
}

// MakeVariableRef builds a variable reference.
func MakeVariableRef(explode bool, ns string, name string) string {
	prefix := ""
	if explode {
		prefix = "@"
	}
	if ns != "" {
		prefix += ns + ":"
	}
	return prefix + name
}
