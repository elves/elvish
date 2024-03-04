package md_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"src.elv.sh/pkg/must"
)

type testCase struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
	Section  string `json:"section"`
	Name     string // supplemental cases only
}

func (tc *testCase) testName() string {
	if tc.Name != "" {
		return fmt.Sprintf("%s/%s", tc.Section, tc.Name)
	}
	return fmt.Sprintf("%s/Example %d", tc.Section, tc.Example)
}

func (tc *testCase) skipReason() string {
	switch tc.Section {
	case "Tabs",
		"Setext headings",
		"Link reference definitions":
		return "section not supported"
	}
	switch tc.Example {
	case 59, 115, 141, 300:
		return "setext heading not supported"
	case 23, 33, 317,
		// Link
		527, 528, 529, 530, 531, 532, 533, 534, 535, 536, 537, 538, 539, 540, 541, 542, 543, 544, 545, 549, 550, 553, 554, 555, 556, 557, 558, 559, 560, 561, 562, 563, 564, 565, 566, 567, 568, 569, 570, 571, 573, 576, 577,
		// Image
		582, 583, 584, 585, 586, 587, 588, 589, 591, 592, 593:
		return "link reference definitions not supported"
	case 294, 296, 307, 318, 319, 320, 321, 323:
		return "tight list not supported"
	}
	return ""
}

func (tc *testCase) skipIfNotSupported(t *testing.T) {
	if reason := tc.skipReason(); reason != "" {
		t.Skip(reason)
	}
}

//go:embed spec/spec.json
var specJSON []byte

var specTestCases = readSpecTestCases(specJSON)

var htmlTestCases = concat(specTestCases, supplementalHTMLTestCases)

// When adding supplemental test cases, check a reference implementation to
// determine the expected output. https://spec.commonmark.org/dingus (which uses
// https://github.com/commonmark/commonmark.js) is the most convenient.
var supplementalHTMLTestCases = []testCase{
	{
		Section:  "ATX headings",
		Name:     "Attribute extension",
		Markdown: "# title {#id}",
		HTML: dedent(`
			<h1 id="id">title</h1>
			`),
	},
	{
		Section:  "Fenced code blocks",
		Name:     "Empty line in list item",
		Markdown: "- ```\n  a\n\n  ```\n",
		HTML: dedent(`
			<ul>
			<li>
			<pre><code>a

			</code></pre>
			</li>
			</ul>
			`),
	},
	{
		Section: "HTML blocks",
		Name:    "Closed by lack of blockquote marker",
		Markdown: dedent(`
			> <pre>

			a
			`),
		HTML: dedent(`
			<blockquote>
			<pre>
			</blockquote>
			<p>a</p>
			`),
	},
	{
		Section: "HTML blocks",
		Name:    "Closed by insufficient list item indentation",
		Markdown: dedent(`
			- <pre>
			 a
			`),
		HTML: dedent(`
			<ul>
			<li>
			<pre>
			</li>
			</ul>
			<p>a</p>
			`),
	},
	{
		Section: "Blockquotes",
		Name:    "Increasing level",
		Markdown: dedent(`
			> a
			>> b
			`),
		HTML: dedent(`
			<blockquote>
			<p>a</p>
			<blockquote>
			<p>b</p>
			</blockquote>
			</blockquote>
			`),
	},
	{
		Section: "Blockquotes",
		Name:    "Reducing level",
		Markdown: dedent(`
			>> a
			>
			> b
			`),
		HTML: dedent(`
			<blockquote>
			<blockquote>
			<p>a</p>
			</blockquote>
			<p>b</p>
			</blockquote>
			`),
	},
	{
		Section: "List items",
		Name:    "Two leading empty lines with spaces",
		Markdown: dedent(`
			- 
			  
			a
			`),
		HTML: dedent(`
			<ul>
			<li></li>
			</ul>
			<p>a</p>
			`),
	},
	{
		Section: "List",
		Name:    "Two-level bullet list with no content interrupting paragraph",
		Markdown: dedent(`
			a
			- -
			`),
		HTML: dedent(`
			<p>a</p>
			<ul>
			<li>
			<ul>
			<li></li>
			</ul>
			</li>
			</ul>
			`),
	},
	{
		Section: "List",
		Name:    "Ordered list with non-1 start in bullet list interrupting paragraph",
		Markdown: dedent(`
			a
			- 2.
			`),
		HTML: dedent(`
			<p>a</p>
			<ul>
			<li>
			<ol start="2">
			<li></li>
			</ol>
			</li>
			</ul>
			`),
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Star after letter before punctuation does not start emphasis",
		Markdown: `a*$*`,
		HTML:     `<p>a*$*</p>` + "\n",
	},
	{
		Section:  "Links",
		Name:     "Backslash and entity in destination",
		Markdown: `[a](\&gt;)`,
		HTML:     `<p><a href="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Section:  "Links",
		Name:     "Backslash and entity in title",
		Markdown: `[a](b (\&gt;))`,
		HTML:     `<p><a href="b" title="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Section:  "Links",
		Name:     "Unmatched ( in destination, with title",
		Markdown: `[a](http://( "b")`,
		HTML:     "<p>[a](http://( &quot;b&quot;)</p>\n",
	},
	{
		Section:  "Links",
		Name:     "Unescaped ( in title started with (",
		Markdown: `[a](b (()))`,
		HTML:     "<p>[a](b (()))</p>\n",
	},
	{
		Section:  "Links",
		Name:     "Literal & in destination",
		Markdown: `[a](http://b?c&d)`,
		HTML:     `<p><a href="http://b?c&amp;d">a</a></p>` + "\n",
	},
	{
		Section: "Image",
		Name:    "Omit hard line break tag in alt",
		Markdown: dedent(`
			![a\
			b](c.png)
			`),
		HTML: dedent(`
			<p><img src="c.png" alt="a
			b" /></p>
			`),
	},
	// This behavior is intentionally under-specified in the spec. The reference
	// implementations puts the raw HTML in the alt attribute, so we match their
	// behavior.
	//
	// CommonMark.js is inconsistent here and does not escape the < and > in the
	// alt attribute: https://github.com/commonmark/commonmark.js/issues/264
	{
		Section:  "Image",
		Name:     "Keep raw HTML in alt",
		Markdown: "![a <a></a>](b.png)",
		HTML:     `<p><img src="b.png" alt="a &lt;a&gt;&lt;/a&gt;" /></p>` + "\n",
	},
	// CommonMark.js has a bug and will not generate the expected output:
	// https://github.com/commonmark/commonmark.js/issues/263
	{
		Section:  "Autolinks",
		Name:     "Entity",
		Markdown: `<http://&gt;>`,
		HTML:     `<p><a href="http://%3E">http://&gt;</a></p>` + "\n",
	},
	{
		Section:  "Raw HTML",
		Name:     "unclosed <",
		Markdown: `a<`,
		HTML:     "<p>a&lt;</p>\n",
	},
	{
		Section:  "Raw HTML",
		Name:     "unclosed <!--",
		Markdown: `a<!--`,
		HTML:     "<p>a&lt;!--</p>\n",
	},
	{
		Section:  "Soft line breaks",
		Name:     "trailing spaces in last line",
		Markdown: "a  \n",
		HTML:     "<p>a</p>\n",
	},
}

func readSpecTestCases(data []byte) []testCase {
	var cases []testCase
	must.OK(json.Unmarshal(data, &cases))
	return cases
}

func concat[T any](a, b []T) []T {
	c := make([]T, 0, len(a)+len(b))
	c = append(c, a...)
	c = append(c, b...)
	return c
}

var hr = strings.Repeat("═", 40)

func hrFence(s string) string { return "\n" + hr + "\n" + s + "\n" + hr }

func dedent(text string) string {
	lines := strings.Split(strings.TrimPrefix(text, "\n"), "\n")
	line0 := lines[0]
	indent := line0[:len(line0)-len(strings.TrimLeft(lines[0], " \t"))]
	for i, line := range lines {
		if !strings.HasPrefix(line, indent) && line != "" {
			panic(fmt.Sprintf("line %d is not empty but doesn't start with %q", i, indent))
		}
		lines[i] = strings.TrimPrefix(line, indent)
	}
	return strings.Join(lines, "\n")
}

var looseListItem = regexp.MustCompile(`<li>([^<]+)</li>`)

func loosifyLists(html string) string {
	return strings.ReplaceAll(
		looseListItem.ReplaceAllString(html, "<li>\n<p>$1</p>\n</li>"),
		"<li></li>", "<li>\n</li>")
}
