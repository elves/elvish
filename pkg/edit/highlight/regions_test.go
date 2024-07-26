package highlight

import (
	"testing"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestGetRegions(t *testing.T) {
	lsCommand := region{0, 2, semanticRegion, commandRegion}

	tt.Test(t, getRegionsFromString,
		Args("").Rets([]region(nil)),
		Args("ls").Rets([]region{
			lsCommand,
		}),

		// Lexical regions.

		Args("ls a").Rets([]region{
			lsCommand,
			{3, 4, lexicalRegion, barewordRegion}, // a
		}),
		Args("ls 'a'").Rets([]region{
			lsCommand,
			{3, 6, lexicalRegion, singleQuotedRegion}, // 'a'
		}),
		Args(`ls "a"`).Rets([]region{
			lsCommand,
			{3, 6, lexicalRegion, doubleQuotedRegion}, // 'a'
		}),
		Args("ls $x").Rets([]region{
			lsCommand,
			{3, 5, lexicalRegion, variableRegion}, // $x
		}),
		Args("ls x*y").Rets([]region{
			lsCommand,
			{3, 4, lexicalRegion, barewordRegion}, // x
			{4, 5, lexicalRegion, wildcardRegion}, // *
			{5, 6, lexicalRegion, barewordRegion}, // y
		}),
		Args("ls ~user/x").Rets([]region{
			lsCommand,
			{3, 4, lexicalRegion, tildeRegion},     // ~
			{4, 10, lexicalRegion, barewordRegion}, // user/x
		}),
		Args("ls # comment").Rets([]region{
			lsCommand,
			{2, 12, lexicalRegion, commentRegion}, // # comment
		}),

		// The "var" special command
		Args("var x = foo").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // var
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, semanticRegion, keywordRegion},  // =
			{8, 11, lexicalRegion, barewordRegion}, // foo
		}),

		// The "set" special command
		Args("set x = foo").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // var
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, semanticRegion, keywordRegion},  // =
			{8, 11, lexicalRegion, barewordRegion}, // foo
		}),

		// The "tmp" special command
		Args("tmp x = foo").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // tmp
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, semanticRegion, keywordRegion},  // =
			{8, 11, lexicalRegion, barewordRegion}, // foo
		}),

		// The "del" special command
		Args("del x y").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // tmp
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, semanticRegion, variableRegion}, // y
		}),

		// The "if" special command.

		Args("if x { }").Rets([]region{
			{0, 2, semanticRegion, commandRegion}, // if
			{3, 4, lexicalRegion, barewordRegion}, // x
			{5, 6, lexicalRegion, "{"},
			{7, 8, lexicalRegion, "}"},
		}),
		Args("if x { } else { }").Rets([]region{
			{0, 2, semanticRegion, commandRegion}, // if
			{3, 4, lexicalRegion, barewordRegion}, // x
			{5, 6, lexicalRegion, "{"},
			{7, 8, lexicalRegion, "}"},
			{9, 13, semanticRegion, keywordRegion}, // else
			{14, 15, lexicalRegion, "{"},
			{16, 17, lexicalRegion, "}"},
		}),
		Args("if x { } elif y { }").Rets([]region{
			{0, 2, semanticRegion, commandRegion}, // if
			{3, 4, lexicalRegion, barewordRegion}, // x
			{5, 6, lexicalRegion, "{"},
			{7, 8, lexicalRegion, "}"},
			{9, 13, semanticRegion, keywordRegion},  // elif
			{14, 15, lexicalRegion, barewordRegion}, // y
			{16, 17, lexicalRegion, "{"},
			{18, 19, lexicalRegion, "}"},
		}),

		// The "for" special command.

		Args("for x [] { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // for
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, lexicalRegion, "["},
			{7, 8, lexicalRegion, "]"},
			{9, 10, lexicalRegion, "{"},
			{11, 12, lexicalRegion, "}"},
		}),
		Args("for x [] { } else { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion},  // for
			{4, 5, semanticRegion, variableRegion}, // x
			{6, 7, lexicalRegion, "["},
			{7, 8, lexicalRegion, "]"},
			{9, 10, lexicalRegion, "{"},
			{11, 12, lexicalRegion, "}"},
			{13, 17, semanticRegion, keywordRegion}, // else
			{18, 19, lexicalRegion, "{"},
			{20, 21, lexicalRegion, "}"},
		}),

		// The "try" special command.

		Args("try { } except e { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 14, semanticRegion, keywordRegion},   // except
			{15, 16, semanticRegion, variableRegion}, // e
			{17, 18, lexicalRegion, "{"},
			{19, 20, lexicalRegion, "}"},
		}),

		Args("try { } except e { } else { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 14, semanticRegion, keywordRegion},   // except
			{15, 16, semanticRegion, variableRegion}, // e
			{17, 18, lexicalRegion, "{"},
			{19, 20, lexicalRegion, "}"},
			{21, 25, semanticRegion, keywordRegion}, // else
			{26, 27, lexicalRegion, "{"},
			{28, 29, lexicalRegion, "}"},
		}),

		// Regression test for b.elv.sh/1358.
		Args("try { } except { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 14, semanticRegion, keywordRegion}, // except
			{15, 16, lexicalRegion, "{"},
			{17, 18, lexicalRegion, "}"},
		}),

		Args("try { } catch e { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 13, semanticRegion, keywordRegion},   // catch
			{14, 15, semanticRegion, variableRegion}, // e
			{16, 17, lexicalRegion, "{"},
			{18, 19, lexicalRegion, "}"},
		}),

		Args("try { } catch e { } else { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 13, semanticRegion, keywordRegion},   // catch
			{14, 15, semanticRegion, variableRegion}, // e
			{16, 17, lexicalRegion, "{"},
			{18, 19, lexicalRegion, "}"},
			{20, 24, semanticRegion, keywordRegion}, // else
			{25, 26, lexicalRegion, "{"},
			{27, 28, lexicalRegion, "}"},
		}),

		// Regression test for b.elv.sh/1358.
		Args("try { } catch { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 13, semanticRegion, keywordRegion}, // catch
			{14, 15, lexicalRegion, "{"},
			{16, 17, lexicalRegion, "}"},
		}),

		Args("try { } finally { }").Rets([]region{
			{0, 3, semanticRegion, commandRegion}, // try
			{4, 5, lexicalRegion, "{"},
			{6, 7, lexicalRegion, "}"},
			{8, 15, semanticRegion, keywordRegion}, // finally
			{16, 17, lexicalRegion, "{"},
			{18, 19, lexicalRegion, "}"},
		}),
	)
}

func getRegionsFromString(code string) []region {
	// Ignore error.
	tree, _ := parse.Parse(parse.SourceForTest(code), parse.Config{})
	return getRegions(tree.Root)
}
