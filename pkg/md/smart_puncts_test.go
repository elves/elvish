package md_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
)

var smartPunctsTestCases = []testCase{
	{
		Name:     "Simple smart punctuations",
		Markdown: `a -- b --- c...`,
		HTML: dedent(`
			<p>a – b –- c…</p>
			`),
	},
	{
		Name:     "Smart quotes",
		Markdown: `It's "foo" and 'bar'.`,
		HTML: dedent(`
			<p>It’s “foo” and ‘bar’.</p>
			`),
	},
	{
		Name: "Link and image title",
		Markdown: dedent(`
			[link text](a.html "--")
			![img alt](a.png "--")
			`),
		HTML: dedent(`
			<p><a href="a.html" title="–">link text</a>
			<img src="a.png" alt="img alt" title="–" /></p>
			`),
	},
	{
		Name:     "Link alt",
		Markdown: `![img -- alt](a.png)`,
		HTML: dedent(`
			<p><img src="a.png" alt="img – alt" /></p>
			`),
	},
	{
		Name:     "Code span is unchanged",
		Markdown: "`a -- b`",
		HTML: dedent(`
			<p><code>a -- b</code></p>
			`),
	},
	{
		Name: "Non-inline content is unchanged",
		Markdown: dedent(`
			~~~
			a -- b
			~~~
			`),
		HTML: dedent(`
			<pre><code>a -- b
			</code></pre>
			`),
	},
}

func TestSmartPuncts(t *testing.T) {
	for _, tc := range smartPunctsTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var htmlCodec HTMLCodec
			Render(tc.Markdown, SmartPunctsCodec{&htmlCodec})
			got := htmlCodec.String()
			if diff := cmp.Diff(tc.HTML, got); diff != "" {
				t.Errorf("input:\n%s\ndiff (-want +got):\n%s",
					hr+"\n"+tc.Markdown+hr, diff)
			}
		})
	}
}
