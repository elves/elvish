package eval

import (
	"src.elv.sh/pkg/parse"
)

func onePrimary(cn *parse.Compound) *parse.Primary {
	if len(cn.Indexings) == 1 && len(cn.Indexings[0].Indicies) == 0 {
		return cn.Indexings[0].Head
	}
	return nil
}

func oneString(cn *parse.Compound) (string, bool) {
	pn := onePrimary(cn)
	if pn != nil {
		switch pn.Type {
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			return pn.Value, true
		}
	}
	return "", false
}

// mustString musts that a Compound contains exactly one Primary of type
// Variable.
func mustString(cp *compiler, cn *parse.Compound, msg string) string {
	s, ok := oneString(cn)
	if !ok {
		cp.errorpf(cn, msg)
	}
	return s
}
