package ui

import (
	"strconv"
	"strings"
)

type sgrTokenizer struct {
	text string

	styling Styling
	content string
}

const sgrPrefix = "\033["

func (st *sgrTokenizer) Next() bool {
	for strings.HasPrefix(st.text, sgrPrefix) {
		trimmed := strings.TrimPrefix(st.text, sgrPrefix)
		// Find the terminator of this sequence.
		termIndex := strings.IndexFunc(trimmed, func(r rune) bool {
			return r != ';' && (r < '0' || r > '9')
		})
		if termIndex == -1 {
			// The string ends with an unterminated escape sequence; ignore
			// it.
			st.text = ""
			return false
		}
		term := trimmed[termIndex]
		sgr := trimmed[:termIndex]
		st.text = trimmed[termIndex+1:]
		if term == 'm' {
			st.styling = StylingFromSGR(sgr)
			st.content = ""
			return true
		}
		// If the terminator is not 'm'; we have seen a non-SGR escape sequence;
		// ignore it and continue.
	}
	if st.text == "" {
		return false
	}
	// Parse a content segment until the next SGR prefix.
	content := ""
	nextSGR := strings.Index(st.text, sgrPrefix)
	if nextSGR == -1 {
		content = st.text
	} else {
		content = st.text[:nextSGR]
	}
	st.text = st.text[len(content):]
	st.styling = nil
	st.content = content
	return true
}

func (st *sgrTokenizer) Token() (Styling, string) {
	return st.styling, st.content
}

// ParseSGREscapedText parses SGR-escaped text into a Text. It also removes
// non-SGR CSI sequences sequences in the text.
func ParseSGREscapedText(s string) Text {
	var text Text
	var style Style

	tokenizer := sgrTokenizer{text: s}
	for tokenizer.Next() {
		styling, content := tokenizer.Token()
		if styling != nil {
			styling.transform(&style)
		}
		if content != "" {
			text = append(text, &Segment{style, content})
		}
	}
	return text
}

var sgrStyling = map[int]Styling{
	0: Reset,
	1: Bold,
	2: Dim,
	4: Underlined,
	5: Blink,
	7: Inverse,
}

// StyleFromSGR builds a Style from an SGR sequence.
func StyleFromSGR(s string) Style {
	var ret Style
	StylingFromSGR(s).transform(&ret)
	return ret
}

// StylingFromSGR builds a Style from an SGR sequence.
func StylingFromSGR(s string) Styling {
	styling := jointStyling{}
	codes := getSGRCodes(s)
	if len(codes) == 0 {
		return Reset
	}
	for len(codes) > 0 {
		code := codes[0]
		consume := 1
		var moreStyling Styling

		switch {
		case sgrStyling[code] != nil:
			moreStyling = sgrStyling[code]
		case 30 <= code && code <= 37:
			moreStyling = Fg(ansiColor(code - 30))
		case 40 <= code && code <= 47:
			moreStyling = Bg(ansiColor(code - 40))
		case 90 <= code && code <= 97:
			moreStyling = Fg(ansiBrightColor(code - 90))
		case 100 <= code && code <= 107:
			moreStyling = Bg(ansiBrightColor(code - 100))
		case code == 38 && len(codes) >= 3 && codes[1] == 5:
			moreStyling = Fg(xterm256Color(codes[2]))
			consume = 3
		case code == 48 && len(codes) >= 3 && codes[1] == 5:
			moreStyling = Bg(xterm256Color(codes[2]))
			consume = 3
		case code == 38 && len(codes) >= 5 && codes[1] == 2:
			moreStyling = Fg(trueColor{
				uint8(codes[2]), uint8(codes[3]), uint8(codes[4])})
			consume = 5
		case code == 48 && len(codes) >= 5 && codes[1] == 2:
			moreStyling = Bg(trueColor{
				uint8(codes[2]), uint8(codes[3]), uint8(codes[4])})
			consume = 5
		case code == 39:
			moreStyling = FgDefault
		case code == 49:
			moreStyling = BgDefault
		default:
			// Do nothing; skip this code
		}
		codes = codes[consume:]
		if moreStyling != nil {
			styling = append(styling, moreStyling)
		}
	}
	return styling
}

func getSGRCodes(s string) []int {
	var codes []int
	for _, part := range strings.Split(s, ";") {
		if part == "" {
			codes = append(codes, 0)
		} else {
			code, err := strconv.Atoi(part)
			if err == nil {
				codes = append(codes, code)
			}
		}
	}
	return codes
}
