package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"src.elv.sh/pkg/md"
)

// A wrapper of [md.HTMLCodec] implementing generic additional features.
type htmlCodec struct {
	md.HTMLCodec
	preprocessInline func([]md.InlineOp)
	// Extensions
	numberSections, toc bool
	// Components of the current section number. Populated if numberSections or
	// toc is true (used for maintaining the sections array in the latter case).
	sectionNumbers []int
	// Tree of sections to be used in the table of content. Populated if toc is
	// true. The root node is a dummy node.
	sectionRoot section
}

type section struct {
	title    string
	id       string
	children []section
}

var (
	numberSectionsRegexp = regexp.MustCompile(`\bnumber-sections\b`)
	tocRegexp            = regexp.MustCompile(`\btoc\b`)
)

func (c *htmlCodec) Do(op md.Op) {
	c.preprocessInline(op.Content)
	switch op.Type {
	case md.OpHeading:
		id := ""
		// Only support #id since that's the only thing used in Elvish's
		// Markdown right now. More can be added if needed.
		if info := op.Info; info != "" && op.Info[0] == '#' {
			id = op.Info[1:]
		} else {
			// Generate an ID using the inline text content converted to lower
			// case.
			id = strings.ToLower(plainTextOfInlineContent(op.Content))
		}
		idHTML := html.EscapeString(processHTMLID(id))

		level := op.Number
		// An empty onclick handler is needed for :hover to work on mobile:
		// https://stackoverflow.com/a/25673064/566659
		fmt.Fprintf(c, `<h%d onclick="" id="%s">`, level, idHTML)

		// Render the content separately first; this may be used in the ToC too.
		var sb strings.Builder
		md.RenderInlineContentToHTML(&sb, op.Content)
		titleHTML := sb.String()

		// Number the section.
		if c.numberSections || c.toc {
			if level < len(c.sectionNumbers) {
				// When going from a higher section level to a lower one,
				// discard higher-level numbers. Discard higher-level section
				// numbers. For example, when going from a #### to a #, only
				// keep the first section number.
				c.sectionNumbers = c.sectionNumbers[:level]
			}
			if level == len(c.sectionNumbers) {
				c.sectionNumbers[level-1]++
			} else {
				// We are going from a lower section level to a higher one (e.g.
				// # to ##), possibly with missing levels (e.g. # to ###).
				// Populate all with 1.
				for level > len(c.sectionNumbers) {
					c.sectionNumbers = append(c.sectionNumbers, 1)
				}
			}

			if c.numberSections {
				titleHTML = sectionNumberPrefix(c.sectionNumbers) + titleHTML
			}
			if c.toc {
				// The section numbers identify a path in the section tree.
				p := &c.sectionRoot
				for _, num := range c.sectionNumbers {
					idx := num - 1
					if idx == len(p.children) {
						p.children = append(p.children, section{})
					}
					p = &p.children[idx]
				}
				p.id = idHTML
				p.title = titleHTML
			}
		}

		c.WriteString(titleHTML)

		// Add self link
		fmt.Fprintf(c,
			`<a href="#%s" class="anchor icon-link" aria-hidden="true"></a>`, idHTML)

		fmt.Fprintf(c, "</h%d>\n", op.Number)
	case md.OpHTMLBlock:
		if c.Len() == 0 && strings.HasPrefix(op.Lines[0], "<!--") {
			// Look for options.
			for _, line := range op.Lines {
				if numberSectionsRegexp.MatchString(line) {
					c.numberSections = true
				}
				if tocRegexp.MatchString(line) {
					c.toc = true
				}
			}
		}
		c.HTMLCodec.Do(op)
	default:
		c.HTMLCodec.Do(op)
	}
}

func sectionNumberPrefix(nums []int) string {
	var sb strings.Builder
	for _, num := range nums {
		fmt.Fprintf(&sb, "%d.", num)
	}
	sb.WriteByte(' ')
	return sb.String()
}

func plainTextOfInlineContent(ops []md.InlineOp) string {
	var sb strings.Builder
	for _, op := range ops {
		sb.WriteString(op.String())
	}
	return sb.String()
}

var whitespaceRun = regexp.MustCompile(`\s+`)

func processHTMLID(s string) string {
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/id
	// Only whitespaces are not allowed in ID; place them with "-".
	return whitespaceRun.ReplaceAllLiteralString(s, "-")
}

const tocBefore = `
<div id="toc-wrapper">
  <div id="toc-header"><span id="toc-status"></span> Table of content</div>
  <div id="toc">
`

const tocAfter = `
  </div>
  <script>
  (function() {
    var open = true,
	    tocHeader = document.getElementById('toc-header'),
	    tocStatus = document.getElementById('toc-status'),
        tocList = document.getElementById('toc');
    tocHeader.onclick = function() {
      open = !open;
      if (open) {
		tocStatus.className = '';
        tocList.className = '';
      } else {
		tocStatus.className = 'closed';
        tocList.className = 'no-display';
      }
    };
  })();
  </script>
</div>
`

func (c *htmlCodec) String() string {
	if !c.toc {
		return c.HTMLCodec.String()
	}
	var sb strings.Builder
	sb.WriteString(tocBefore)
	sb.WriteString("<ul>\n")
	for _, section := range c.sectionRoot.children {
		writeSection(&sb, section)
	}
	sb.WriteString("</ul>\n")
	sb.WriteString(tocAfter)

	sb.WriteString(c.HTMLCodec.String())
	return sb.String()
}

func writeSection(sb *strings.Builder, s section) {
	fmt.Fprintf(sb, `<li><a href="#%s">%s</a>`, s.id, s.title)
	if len(s.children) > 0 {
		sb.WriteString("\n<ul>\n")
		for _, child := range s.children {
			writeSection(sb, child)
		}
		sb.WriteString("</ul>\n")
	}
	sb.WriteString("</li>\n")
}
