package md_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"src.elv.sh/pkg/md"
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
		return tc.Name
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
		526, 527, 528, 529, 530, 531, 532, 533, 534, 535, 536, 537, 538, 539, 540, 541, 542, 543, 544, 548, 549, 552, 553, 554, 555, 556, 557, 558, 559, 560, 561, 562, 563, 564, 565, 566, 567, 568, 569, 570, 572, 575, 576,
		// Image
		581, 582, 583, 584, 585, 586, 587, 588, 590, 591, 592:
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

var testCases []testCase

// When adding supplemental test cases, check a reference implementation to
// determine the expected output. https://spec.commonmark.org/dingus (which uses
// https://github.com/commonmark/commonmark.js) is the most convenient.
var supplementalCases = []testCase{
	{
		Name:     "Fenced code blocks supplemental/Empty line in list item",
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
		Name: "HTML blocks supplemental/Closed by lack of blockquote marker",
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
		Name: "HTML blocks supplemental/Closed by insufficient list item indentation",
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
		Name: "Blockquotes supplemental/Increasing level",
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
		Name: "Blockquotes supplemental/Reducing level",
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
		Name: "List items supplemental/Two leading empty lines with spaces",
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
		Name:     "Links supplemental/Backslash and entity in destination",
		Markdown: `[a](\&gt;)`,
		HTML:     `<p><a href="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Name:     "Links supplemental/Backslash and entity in title",
		Markdown: `[a](b (\&gt;))`,
		HTML:     `<p><a href="b" title="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Name:     "Links supplemental/Unmatched ( in destination, with title",
		Markdown: `[a](http://( "b")`,
		HTML:     "<p>[a](http://( &quot;b&quot;)</p>\n",
	},
	{
		Name:     "Links supplemental/Unescaped ( in title started with (",
		Markdown: `[a](b (()))`,
		HTML:     "<p>[a](b (()))</p>\n",
	},
	{
		Name:     "Links supplemental/Literal & in destination",
		Markdown: `[a](http://b?c&d)`,
		HTML:     `<p><a href="http://b?c&amp;d">a</a></p>` + "\n",
	},
	{
		Name: "Image supplemental/Omit hard line break tag in alt",
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
		Name:     "Image supplemental/Keep raw HTML in alt",
		Markdown: "![a <a></a>](b.png)",
		HTML:     `<p><img src="b.png" alt="a &lt;a&gt;&lt;/a&gt;" /></p>` + "\n",
	},
	// CommonMark.js has a bug and will not generate the expected output:
	// https://github.com/commonmark/commonmark.js/issues/263
	{
		Name:     "Autolinks supplemental/Entity",
		Markdown: `<http://&gt;>`,
		HTML:     `<p><a href="http://%3E">http://&gt;</a></p>` + "\n",
	},
	{
		Name:     "Raw HTML supplemental/unclosed <",
		Markdown: `a<`,
		HTML:     "<p>a&lt;</p>\n",
	},
	{
		Name:     "Raw HTML supplemental/unclosed <!--",
		Markdown: `a<!--`,
		HTML:     "<p>a&lt;!--</p>\n",
	},
	{
		Name:     "Soft line breaks supplemental/trailing spaces in last line",
		Markdown: "a  \n",
		HTML:     "<p>a</p>\n",
	},
}

func init() {
	must.OK(json.Unmarshal(specJSON, &testCases))
	testCases = append(testCases, supplementalCases...)
}

var (
	hr = strings.Repeat("═", 40)
)

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

type codecStringer interface {
	md.Codec
	String() string
}

func render(markdown string, codec codecStringer) string {
	md.Render(markdown, codec)
	return codec.String()
}
