package eval

import "strings"

// ParseVariableRef parses a variable reference.
func ParseVariableRef(text string) (explode bool, ns string, name string) {
	return parseVariableRef(text, true)
}

// ParseIncompleteVariableRef parses an incomplete variable reference.
func ParseIncompleteVariableRef(text string) (explode bool, ns string, name string) {
	return parseVariableRef(text, false)
}

func parseVariableRef(text string, complete bool) (explode bool, ns string, name string) {
	explodePart, nsPart, name := splitVariableRef(text, complete)
	ns = nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}
	return explodePart != "", ns, name
}

// SplitVariableRef splits a variable reference into three parts: an optional
// explode operator (either "" or "@"), a namespace part, and a name part.
func SplitVariableRef(text string) (explodePart, nsPart, name string) {
	return splitVariableRef(text, true)
}

// SplitIncompleteVariableRef splits an incomplete variable reference into three
// parts: an optional explode operator (either "" or "@"), a namespace part, and
// a name part.
func SplitIncompleteVariableRef(text string) (explodePart, nsPart, name string) {
	return splitVariableRef(text, false)
}

func splitVariableRef(text string, complete bool) (explodePart, nsPart, name string) {
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
	if complete && i == len(qname)-1 {
		i = strings.LastIndexByte(qname[:len(qname)-1], ':')
	}
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
