package comps

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tries to expand a simple abbreviation. This function assumes the state mutex is held.
func expandSimpleAbbr(simpleAbbr func(func(a, f string)), buf TextBuffer, streak string) (TextBuffer, bool) {
	var abbr, full string
	// Find the longest matching abbreviation.
	simpleAbbr(func(a, f string) {
		if strings.HasSuffix(streak, a) && len(a) > len(abbr) {
			abbr, full = a, f
		}
	})
	if len(abbr) > 0 {
		return TextBuffer{
			Content: buf.Content[:buf.Dot-len(abbr)] + full + buf.Content[buf.Dot:],
			Dot:     buf.Dot - len(abbr) + len(full),
		}, true
	}
	return TextBuffer{}, false
}

var commandRegex = regexp.MustCompile(`(?:^|[^^]\n|\||;|{\s|\()\s*([\p{L}\p{M}\p{N}!%+,\-./:@\\_<>*]+)(\s)$`)

// Tries to expand a command abbreviation. This function assumes the state mutex
// is held.
//
// We use a regex rather than parse.Parse() because dealing with the latter
// requires a lot of code. A simple regex is far simpler and good enough for
// this use case. The regex essentially matches commands at the start of the
// line (with potential leading whitespace) and similarly after the opening
// brace of a lambda or pipeline char.
//
// This only handles bareword commands.
func expandCmdAbbr(cmdAbbr func(func(a, f string)), buf TextBuffer) (TextBuffer, bool) {
	if buf.Dot < len(buf.Content) {
		// Command abbreviations are only expanded when inserting at the end of the buffer.
		return TextBuffer{}, false
	}

	// See if there is something that looks like a bareword at the end of the buffer.
	matches := commandRegex.FindStringSubmatch(buf.Content)
	if len(matches) == 0 {
		return TextBuffer{}, false
	}

	// Find an abbreviation matching the command.
	command, whitespace := matches[1], matches[2]
	var expansion string
	cmdAbbr(func(a, e string) {
		if a == command {
			expansion = e
		}
	})
	if expansion == "" {
		return TextBuffer{}, false
	}

	// We found a matching abbreviation -- replace it with its expansion.
	newContent := buf.Content[:buf.Dot-len(command)-1] + expansion + whitespace
	return TextBuffer{
		Content: newContent,
		Dot:     len(newContent),
	}, true
}

// Try to expand a small word abbreviation. This function assumes the state mutex is held.
func expandSmallWordAbbr(smallWordAbbr func(func(a, f string)), buf TextBuffer, streak string, trigger rune, categorizer func(rune) int) (TextBuffer, bool) {
	if buf.Dot < len(buf.Content) {
		// Word abbreviations are only expanded when inserting at the end of the buffer.
		return TextBuffer{}, false
	}
	triggerLen := len(string(trigger))
	if triggerLen >= len(streak) {
		// Only the trigger has been inserted, or a simple abbreviation was just
		// expanded. In either case, there is nothing to expand.
		return TextBuffer{}, false
	}
	// The trigger is only used to determine word boundary; when considering
	// what to expand, we only consider the part that was inserted before it.
	inserts := streak[:len(streak)-triggerLen]

	var abbr, full string
	// Find the longest matching abbreviation.
	smallWordAbbr(func(a, f string) {
		if len(a) <= len(abbr) {
			// This abbreviation can't be the longest.
			return
		}
		if !strings.HasSuffix(inserts, a) {
			// This abbreviation was not inserted.
			return
		}
		// Verify the trigger rune creates a word boundary.
		r, _ := utf8.DecodeLastRuneInString(a)
		if categorizer(trigger) == categorizer(r) {
			return
		}
		// Verify the rune preceding the abbreviation, if any, creates a word
		// boundary.
		if len(buf.Content) > len(a)+triggerLen {
			r1, _ := utf8.DecodeLastRuneInString(buf.Content[:len(buf.Content)-len(a)-triggerLen])
			r2, _ := utf8.DecodeRuneInString(a)
			if categorizer(r1) == categorizer(r2) {
				return
			}
		}
		abbr, full = a, f
	})
	if len(abbr) > 0 {
		return TextBuffer{
			Content: buf.Content[:buf.Dot-len(abbr)-triggerLen] + full + string(trigger),
			Dot:     buf.Dot - len(abbr) + len(full),
		}, true
	}
	return TextBuffer{}, false
}

// isAlnum determines if the rune is an alphanumeric character.
func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// categorizeSmallWord determines if the rune is whitespace, alphanum, or
// something else.
func categorizeSmallWord(r rune) int {
	switch {
	case unicode.IsSpace(r):
		return 0
	case isAlnum(r):
		return 1
	default:
		return 2
	}
}
