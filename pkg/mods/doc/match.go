package doc

import (
	"sort"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/ui"
)

func match(markdown string, qs []string) ([]matchedBlock, bool) {
	var codec md.TextCodec
	md.Render(markdown, &codec)
	bs := codec.Blocks()
	bMatches := make([][]diag.Ranging, len(bs))
	for _, q := range qs {
		qMatchesAny := false
		for i, b := range bs {
			if from := strings.Index(b.Text, q); from != -1 {
				bMatches[i] = append(bMatches[i],
					diag.Ranging{From: from, To: from + len(q)})
				qMatchesAny = true
				break
			}
		}
		if !qMatchesAny {
			return nil, false
		}
	}
	var matched []matchedBlock
	for i, b := range bs {
		if len(bMatches[i]) > 0 {
			matched = append(matched, matchedBlock{b, sortAndMergeMatches(bMatches[i])})
		}
	}
	return matched, true
}

func sortAndMergeMatches(rs []diag.Ranging) []diag.Ranging {
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].From < rs[j].From
	})
	i := 0
	for j := 1; j < len(rs); j++ {
		if rs[j].From > rs[j-1].To {
			i++
			rs[i] = rs[j]
		} else {
			rs[i].To = rs[j].To
		}
	}
	return rs[:i+1]
}

type matchedBlock struct {
	block   md.TextBlock
	matches []diag.Ranging
}

var queryStyling = ui.Stylings(ui.Bold, ui.FgRed)

func (b matchedBlock) Show() string {
	var sb strings.Builder
	// The algorithms to highlight queries in code blocks and queries in normal
	// text are quite similar, with one subtle difference: code blocks can
	// contain newlines, but we avoid writing them to keep the output tidy.
	// Newlines can arise when two adjacent lines are shown, or when a query
	// spans multiple lines. Both get replaced with " … ".
	if b.block.Code {
		lastTo := 0
		lastLineTo := 0

		for _, m := range b.matches {
			lineFrom := lastLineStart(b.block.Text, m.From)
			if lastLineTo < lineFrom {
				sb.WriteString(b.block.Text[lastTo:lastLineTo])
				if sb.Len() > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString("… ")
				sb.WriteString(b.block.Text[lineFrom:m.From])
			} else {
				sb.WriteString(b.block.Text[lastTo:m.From])
			}
			queryText := strings.ReplaceAll(b.block.Text[m.From:m.To], "\n", " … ")
			sb.WriteString(ui.T(queryText, queryStyling).String())
			lastTo = m.To
			lastLineTo = firstLineEnd(b.block.Text, m.To)
		}
		sb.WriteString(b.block.Text[lastTo:lastLineTo])
		if lastLineTo < len(b.block.Text) {
			sb.WriteString(" …")
		}
	} else {
		lastTo := 0
		lastSentenceTo := 0

		for _, m := range b.matches {
			sentenceFrom := lastSentenceStart(b.block.Text, m.From)
			if lastSentenceTo < sentenceFrom {
				sb.WriteString(b.block.Text[lastTo:lastSentenceTo])
				sb.WriteString("… ")
				sb.WriteString(b.block.Text[sentenceFrom:m.From])
			} else {
				sb.WriteString(b.block.Text[lastTo:m.From])
			}
			sb.WriteString(ui.T(b.block.Text[m.From:m.To], queryStyling).String())
			lastTo = m.To
			lastSentenceTo = firstSentenceStart(b.block.Text, m.To)
		}
		sb.WriteString(b.block.Text[lastTo:lastSentenceTo])
		if lastSentenceTo < len(b.block.Text) {
			sb.WriteString("…")
		}
	}
	return sb.String()
}

func firstSentenceStart(s string, from int) int {
	if i := strings.Index(s[from:], ". "); i != -1 {
		return from + i + 2
	}
	return len(s)
}

func lastSentenceStart(s string, upto int) int {
	if i := strings.LastIndex(s[:upto], ". "); i != -1 {
		return i + 2
	}
	return 0
}

func firstLineEnd(s string, from int) int {
	if i := strings.Index(s[from:], "\n"); i != -1 {
		return from + i
	}
	return len(s)
}

func lastLineStart(s string, upto int) int {
	if i := strings.LastIndex(s[:upto], "\n"); i != -1 {
		return i + 1
	}
	return 0
}
