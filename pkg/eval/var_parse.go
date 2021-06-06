package eval

import "strings"

// SplitSigil splits any leading sigil from a qualified variable name.
func SplitSigil(ref string) (sigil string, qname string) {
	if ref == "" {
		return "", ""
	}
	// TODO: Support % (and other sigils?) if https://b.elv.sh/584 is implemented for map explosion.
	switch ref[0] {
	case '@':
		return ref[:1], ref[1:]
	default:
		return "", ref
	}
}

// SplitQName splits a qualified name into the first namespace segment and the
// rest.
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

// SplitIncompleteQNameNs splits an incomplete qualified variable name into the
// namespace part and the name part.
func SplitIncompleteQNameNs(qname string) (ns, name string) {
	colon := strings.LastIndexByte(qname, ':')
	// If colon is -1, colon+1 will be 0, rendering an empty ns.
	return qname[:colon+1], qname[colon+1:]
}
