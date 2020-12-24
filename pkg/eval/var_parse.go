package eval

import "strings"

// SplitVariableRef splits a variable reference into the sigil and the
// (qualified) name.
func SplitVariableRef(ref string) (sigil string, qname string) {
	if ref == "" {
		return "", ""
	}
	switch ref[0] {
	case '@':
		// TODO(xiaq): Support % later.
		return ref[:1], ref[1:]
	default:
		return "", ref
	}
}

func SplitQName(qname string) (first, rest string) {
	colon := strings.IndexByte(qname, ':')
	if colon == -1 {
		return qname, ""
	}
	return qname[:colon+1], qname[colon+1:]
}

// SplitQNameSegs splits a qualified name into namespace segments.
func SplitQNameSegs(qname string) []string {
	segs := strings.SplitAfter(qname, ":")
	if len(segs) > 0 && segs[len(segs)-1] == "" {
		segs = segs[:len(segs)-1]
	}
	return segs
}

// SplitQNameNs splits a qualified variable name into the namespace part and the
// name part.
func SplitQNameNs(qname string) (ns, name string) {
	if qname == "" {
		return "", ""
	}
	colon := strings.LastIndexByte(qname[:len(qname)-1], ':')
	// If colon is -1, colon+1 will be 0, rendering an empty ns.
	return qname[:colon+1], qname[colon+1:]
}

// SplitQNameNsIncomplete splits an incomplete qualified variable name into
// the namespace part and the name part.
func SplitQNameNsIncomplete(qname string) (ns, name string) {
	colon := strings.LastIndexByte(qname, ':')
	// If colon is -1, colon+1 will be 0, rendering an empty ns.
	return qname[:colon+1], qname[colon+1:]
}

// SplitQNameNsFirst splits a qualified variable name into the first part and
// the rest.
func SplitQNameNsFirst(qname string) (ns, rest string) {
	colon := strings.IndexByte(qname, ':')
	if colon == len(qname)-1 {
		// Unqualified variable ending with colon ($name:).
		return "", qname
	}
	// If colon is -1, colon+1 will be 0, rendering an empty ns.
	return qname[:colon+1], qname[colon+1:]
}

// SplitIncompleteQNameFirstNs splits an incomplete qualified variable name
// into the first part and the rest.
func SplitIncompleteQNameFirstNs(qname string) (ns, rest string) {
	colon := strings.IndexByte(qname, ':')
	// If colon is -1, colon+1 will be 0, rendering an empty ns.
	return qname[:colon+1], qname[colon+1:]
}

// SplitQNameNsSegs splits a qualified name into namespace segments.
func SplitQNameNsSegs(qname string) []string {
	segs := strings.SplitAfter(qname, ":")
	if len(segs) > 0 && segs[len(segs)-1] == "" {
		segs = segs[:len(segs)-1]
	}
	return segs
}
