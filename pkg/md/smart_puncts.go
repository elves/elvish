package md

import (
	"strings"
	"unicode"
)

// SmartPunctsCodec wraps another codec, converting certain ASCII punctuations to
// nicer Unicode counterparts:
//
//   - A straight double quote (") is converted to a left double quote (“) when
//     it follows a whitespace, or a right double quote (”) when it follows a
//     non-whitespace.
//
//   - A straight single quote (') is converted to a left single quote (‘) when
//     it follows a whitespace, or a right single quote or apostrophe (’) when
//     it follows a non-whitespace.
//
//   - A run of two dashes (--) is converted to an en-dash (–).
//
//   - A run of three dashes (---) is converted to an em-dash (—).
//
//   - A run of three dot (...) is converted to an ellipsis (…).
//
// Start of lines are considered to be whitespaces.
type SmartPunctsCodec struct{ Inner Codec }

func (c SmartPunctsCodec) Do(op Op) { c.Inner.Do(applySmartPunctsToOp(op)) }

func applySmartPunctsToOp(op Op) Op {
	for i := range op.Content {
		inlineOp := &op.Content[i]
		switch inlineOp.Type {
		case OpText, OpLinkStart, OpLinkEnd, OpImage:
			inlineOp.Text = applySmartPuncts(inlineOp.Text)
			if inlineOp.Type == OpImage {
				inlineOp.Alt = applySmartPuncts(inlineOp.Alt)
			}
		}
	}
	return op
}

var applySimpleSmartPuncts = strings.NewReplacer(
	"--", "–", "---", "—", "...", "…").Replace

func applySmartPuncts(s string) string {
	return applySimpleSmartPuncts(applySmartQuotes(s))
}

func applySmartQuotes(s string) string {
	if !strings.ContainsAny(s, `'"`) {
		return s
	}
	var sb strings.Builder
	// Start of line is considered to be whitespace
	prev := ' '
	for _, r := range s {
		if r == '"' {
			if unicode.IsSpace(prev) {
				sb.WriteRune('“')
			} else {
				sb.WriteRune('”')
			}
		} else if r == '\'' {
			if unicode.IsSpace(prev) {
				sb.WriteRune('‘')
			} else {
				sb.WriteRune('’')
			}
		} else {
			sb.WriteRune(r)
		}
		prev = r
	}
	return sb.String()
}
