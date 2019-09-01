package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/tt"
)

var bufferBuiltinsTests = []struct {
	name      string
	bufBefore codearea.CodeBuffer
	bufAfter  codearea.CodeBuffer
}{
	{
		"move-left",
		codearea.CodeBuffer{Content: "ab", Dot: 1},
		codearea.CodeBuffer{Content: "ab", Dot: 0},
	},
	{
		"move-right",
		codearea.CodeBuffer{Content: "ab", Dot: 1},
		codearea.CodeBuffer{Content: "ab", Dot: 2},
	},
	{
		"kill-left",
		codearea.CodeBuffer{Content: "ab", Dot: 1},
		codearea.CodeBuffer{Content: "b", Dot: 0},
	},
	{
		"kill-right",
		codearea.CodeBuffer{Content: "ab", Dot: 1},
		codearea.CodeBuffer{Content: "a", Dot: 1},
	},
}

func TestBufferBuiltins(t *testing.T) {
	app := &cli.App{}
	builtins := bufferBuiltins(app)
	buf := &app.CodeArea.State.CodeBuffer

	for _, test := range bufferBuiltinsTests {
		t.Run(test.name, func(t *testing.T) {
			*buf = test.bufBefore
			fn := builtins[test.name].(func())
			fn()
			if *buf != test.bufAfter {
				t.Errorf("got buf %v, want %v", *buf, test.bufAfter)
			}
		})
	}
}

// Tests for pure movers.

func TestMoveDotLeftRight(t *testing.T) {
	tt.Test(t, tt.Fn("moveDotLeft", moveDotLeft), tt.Table{
		tt.Args("foo", 0).Rets(0),
		tt.Args("bar", 3).Rets(2),
		tt.Args("精灵", 0).Rets(0),
		tt.Args("精灵", 3).Rets(0),
		tt.Args("精灵", 6).Rets(3),
	})
	tt.Test(t, tt.Fn("moveDotRight", moveDotRight), tt.Table{
		tt.Args("foo", 0).Rets(1),
		tt.Args("bar", 3).Rets(3),
		tt.Args("精灵", 0).Rets(3),
		tt.Args("精灵", 3).Rets(6),
		tt.Args("精灵", 6).Rets(6),
	})
}

func TestMoveDotSOLEOL(t *testing.T) {
	buffer := "abc\ndef"
	// Index:
	//         012 34567
	tt.Test(t, tt.Fn("moveDotSOL", moveDotSOL), tt.Table{
		tt.Args(buffer, 0).Rets(0),
		tt.Args(buffer, 1).Rets(0),
		tt.Args(buffer, 2).Rets(0),
		tt.Args(buffer, 3).Rets(0),
		tt.Args(buffer, 4).Rets(4),
		tt.Args(buffer, 5).Rets(4),
		tt.Args(buffer, 6).Rets(4),
		tt.Args(buffer, 7).Rets(4),
	})
	tt.Test(t, tt.Fn("moveDotEOL", moveDotEOL), tt.Table{
		tt.Args(buffer, 0).Rets(3),
		tt.Args(buffer, 1).Rets(3),
		tt.Args(buffer, 2).Rets(3),
		tt.Args(buffer, 3).Rets(3),
		tt.Args(buffer, 4).Rets(7),
		tt.Args(buffer, 5).Rets(7),
		tt.Args(buffer, 6).Rets(7),
		tt.Args(buffer, 7).Rets(7),
	})
}

func TestMoveDotUpDown(t *testing.T) {
	buffer := "abc\n精灵语\ndef"
	// Index:
	//         012 34 7 0  34567
	// + 10 *  0        1

	tt.Test(t, tt.Fn("moveDotUp", moveDotUp), tt.Table{
		tt.Args(buffer, 0).Rets(0),  // a -> a
		tt.Args(buffer, 1).Rets(1),  // b -> b
		tt.Args(buffer, 2).Rets(2),  // c -> c
		tt.Args(buffer, 3).Rets(3),  // EOL1 -> EOL1
		tt.Args(buffer, 4).Rets(0),  // 精 -> a
		tt.Args(buffer, 7).Rets(2),  // 灵 -> c
		tt.Args(buffer, 10).Rets(3), // 语 -> EOL1
		tt.Args(buffer, 13).Rets(3), // EOL2 -> EOL1
		tt.Args(buffer, 14).Rets(4), // d -> 精
		tt.Args(buffer, 15).Rets(4), // e -> 精 (jump left half width)
		tt.Args(buffer, 16).Rets(7), // f -> 灵
		tt.Args(buffer, 17).Rets(7), // EOL3 -> 灵 (jump left half width)
	})

	tt.Test(t, tt.Fn("moveDotDown", moveDotDown), tt.Table{
		tt.Args(buffer, 0).Rets(4),   // a -> 精
		tt.Args(buffer, 1).Rets(4),   // b -> 精 (jump left half width)
		tt.Args(buffer, 2).Rets(7),   // c -> 灵
		tt.Args(buffer, 3).Rets(7),   // EOL1 -> 灵 (jump left half width)
		tt.Args(buffer, 4).Rets(14),  // 精 -> d
		tt.Args(buffer, 7).Rets(16),  // 灵 -> f
		tt.Args(buffer, 10).Rets(17), // 语 -> EOL3
		tt.Args(buffer, 13).Rets(17), // EOL2 -> EOL3
		tt.Args(buffer, 14).Rets(14), // d -> d
		tt.Args(buffer, 15).Rets(15), // e -> e
		tt.Args(buffer, 16).Rets(16), // f -> f
		tt.Args(buffer, 17).Rets(17), // EOL3 -> EOL3
	})
}

// Word movement tests.

// The string below is carefully chosen to test all word, small-word, and
// alnum-word move/kill functions, because it contains features to set the
// different movement behaviors apart.
//
// The string is annotated with carets (^) to indicate the beginning of words,
// and periods (.) to indicate trailing runes of words. Indicies are also
// annotated.
//
//   cd ~/downloads; rm -rf 2018aug07-pics/*;
//   ^. ^........... ^. ^.. ^................  (word)
//   ^. ^.^........^ ^. ^^. ^........^^...^..  (small-word)
//   ^.   ^........  ^.  ^. ^........ ^...     (alnum-word)
//   01234567890123456789012345678901234567890
//   0         1         2         3         4
//
//   word boundaries:         0 3      16 19    23
//   small-word boundaries:   0 3 5 14 16 19 20 23 32 33 37
//   alnum-word boundaries:   0   5    16    20 23    33
//
var wordMoveTestBuffer = "cd ~/downloads; rm -rf 2018aug07-pics/*;"

var (
	// word boundaries: 0 3 16 19 23
	moveDotLeftWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(0),
		tt.Args(wordMoveTestBuffer, 1).Rets(0),
		tt.Args(wordMoveTestBuffer, 2).Rets(0),
		tt.Args(wordMoveTestBuffer, 3).Rets(0),
		tt.Args(wordMoveTestBuffer, 4).Rets(3),
		tt.Args(wordMoveTestBuffer, 16).Rets(3),
		tt.Args(wordMoveTestBuffer, 19).Rets(16),
		tt.Args(wordMoveTestBuffer, 23).Rets(19),
		tt.Args(wordMoveTestBuffer, 40).Rets(23),
	}
	moveDotRightWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(3),
		tt.Args(wordMoveTestBuffer, 1).Rets(3),
		tt.Args(wordMoveTestBuffer, 2).Rets(3),
		tt.Args(wordMoveTestBuffer, 3).Rets(16),
		tt.Args(wordMoveTestBuffer, 16).Rets(19),
		tt.Args(wordMoveTestBuffer, 19).Rets(23),
		tt.Args(wordMoveTestBuffer, 23).Rets(40),
	}

	// small-word boundaries: 0 3 5 14 16 19 20 23 32 33 37
	moveDotLeftSmallWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(0),
		tt.Args(wordMoveTestBuffer, 1).Rets(0),
		tt.Args(wordMoveTestBuffer, 2).Rets(0),
		tt.Args(wordMoveTestBuffer, 3).Rets(0),
		tt.Args(wordMoveTestBuffer, 4).Rets(3),
		tt.Args(wordMoveTestBuffer, 5).Rets(3),
		tt.Args(wordMoveTestBuffer, 14).Rets(5),
		tt.Args(wordMoveTestBuffer, 16).Rets(14),
		tt.Args(wordMoveTestBuffer, 19).Rets(16),
		tt.Args(wordMoveTestBuffer, 20).Rets(19),
		tt.Args(wordMoveTestBuffer, 23).Rets(20),
		tt.Args(wordMoveTestBuffer, 32).Rets(23),
		tt.Args(wordMoveTestBuffer, 33).Rets(32),
		tt.Args(wordMoveTestBuffer, 37).Rets(33),
		tt.Args(wordMoveTestBuffer, 40).Rets(37),
	}
	moveDotRightSmallWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(3),
		tt.Args(wordMoveTestBuffer, 1).Rets(3),
		tt.Args(wordMoveTestBuffer, 2).Rets(3),
		tt.Args(wordMoveTestBuffer, 3).Rets(5),
		tt.Args(wordMoveTestBuffer, 5).Rets(14),
		tt.Args(wordMoveTestBuffer, 14).Rets(16),
		tt.Args(wordMoveTestBuffer, 16).Rets(19),
		tt.Args(wordMoveTestBuffer, 19).Rets(20),
		tt.Args(wordMoveTestBuffer, 20).Rets(23),
		tt.Args(wordMoveTestBuffer, 23).Rets(32),
		tt.Args(wordMoveTestBuffer, 32).Rets(33),
		tt.Args(wordMoveTestBuffer, 33).Rets(37),
		tt.Args(wordMoveTestBuffer, 37).Rets(40),
	}

	// alnum-word boundaries: 0 5 16 20 23 33
	moveDotLeftAlnumWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(0),
		tt.Args(wordMoveTestBuffer, 1).Rets(0),
		tt.Args(wordMoveTestBuffer, 2).Rets(0),
		tt.Args(wordMoveTestBuffer, 3).Rets(0),
		tt.Args(wordMoveTestBuffer, 4).Rets(0),
		tt.Args(wordMoveTestBuffer, 5).Rets(0),
		tt.Args(wordMoveTestBuffer, 6).Rets(5),
		tt.Args(wordMoveTestBuffer, 16).Rets(5),
		tt.Args(wordMoveTestBuffer, 20).Rets(16),
		tt.Args(wordMoveTestBuffer, 23).Rets(20),
		tt.Args(wordMoveTestBuffer, 33).Rets(23),
		tt.Args(wordMoveTestBuffer, 40).Rets(33),
	}
	moveDotRightAlnumWordTests = tt.Table{
		tt.Args(wordMoveTestBuffer, 0).Rets(5),
		tt.Args(wordMoveTestBuffer, 1).Rets(5),
		tt.Args(wordMoveTestBuffer, 2).Rets(5),
		tt.Args(wordMoveTestBuffer, 3).Rets(5),
		tt.Args(wordMoveTestBuffer, 4).Rets(5),
		tt.Args(wordMoveTestBuffer, 5).Rets(16),
		tt.Args(wordMoveTestBuffer, 16).Rets(20),
		tt.Args(wordMoveTestBuffer, 20).Rets(23),
		tt.Args(wordMoveTestBuffer, 23).Rets(33),
		tt.Args(wordMoveTestBuffer, 33).Rets(40),
	}
)

func TestMoveDotWord(t *testing.T) {
	tt.Test(t, tt.Fn("moveDotLeftWord", moveDotLeftWord), moveDotLeftWordTests)
	tt.Test(t, tt.Fn("moveDotRightWord", moveDotRightWord), moveDotRightWordTests)
}

func TestMoveDotSmallWord(t *testing.T) {
	tt.Test(t,
		tt.Fn("moveDotLeftSmallWord", moveDotLeftSmallWord),
		moveDotLeftSmallWordTests,
	)
	tt.Test(t,
		tt.Fn("moveDotRightSmallWord", moveDotRightSmallWord),
		moveDotRightSmallWordTests,
	)
}

func TestMoveDotAlnumWord(t *testing.T) {
	tt.Test(t,
		tt.Fn("moveDotLeftAlnumWord", moveDotLeftAlnumWord),
		moveDotLeftAlnumWordTests,
	)
	tt.Test(t,
		tt.Fn("moveDotRightAlnumWord", moveDotRightAlnumWord),
		moveDotRightAlnumWordTests,
	)
}
