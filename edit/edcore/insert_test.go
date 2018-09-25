package edcore

import (
	"testing"

	"github.com/elves/elvish/tt"
)

// The string below is carefully chosen to test all word, small-word,
// and alphanumeric move/kill functions, because it contains features
// to set the different movement behaviors apart.  For convenience,
// the string is annotated with carets (^) to indicate the first
// indices of runs of word/small-word/alphanumeric characters (and
// hence return values of corresponding move functions), and dots (.)
// to indicate trailing characters in a run of non-skip characters of
// the same category.  Indices are also annotated for convenience.
//
//   cd ~/downloads; rm -rf 2018aug07-pics/*;
//   ^. ^........... ^. ^.. ^................  (word)
//   ^. ^.^........^ ^. ^^. ^........^^...^..  (small-word)
//   ^.   ^........  ^.  ^. ^........ ^...     (alphanumeric)
//   01234567890123456789012345678901234567890
//   0         1         2         3         4
//
//   word boundaries:         0 3      16 19    23
//   small-word boundaries:   0 3 5 14 16 19 20 23 32 33 37
//   alphanumeric boundaries: 0   5    16    20 23    33
//
var moveTestBuffer = "cd ~/downloads; rm -rf 2018aug07-pics/*;"

var (
	moveDotLeftTests = tt.Table{
		tt.Args("foo", 0).Rets(0),
		tt.Args("bar", 3).Rets(2),
		tt.Args("精灵", 0).Rets(0),
		tt.Args("精灵", 3).Rets(0),
		tt.Args("精灵", 6).Rets(3),
	}
	moveDotRightTests = tt.Table{
		tt.Args("foo", 0).Rets(1),
		tt.Args("bar", 3).Rets(3),
		tt.Args("精灵", 0).Rets(3),
		tt.Args("精灵", 3).Rets(6),
		tt.Args("精灵", 6).Rets(6),
	}
	// word boundaries: 0 3 16 19 23
	moveDotLeftWordTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(0),
		tt.Args(moveTestBuffer, 1).Rets(0),
		tt.Args(moveTestBuffer, 2).Rets(0),
		tt.Args(moveTestBuffer, 3).Rets(0),
		tt.Args(moveTestBuffer, 4).Rets(3),
		tt.Args(moveTestBuffer, 16).Rets(3),
		tt.Args(moveTestBuffer, 19).Rets(16),
		tt.Args(moveTestBuffer, 23).Rets(19),
		tt.Args(moveTestBuffer, 40).Rets(23),
	}
	moveDotRightWordTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(3),
		tt.Args(moveTestBuffer, 1).Rets(3),
		tt.Args(moveTestBuffer, 2).Rets(3),
		tt.Args(moveTestBuffer, 3).Rets(16),
		tt.Args(moveTestBuffer, 16).Rets(19),
		tt.Args(moveTestBuffer, 19).Rets(23),
		tt.Args(moveTestBuffer, 23).Rets(40),
	}

	// small-word boundaries: 0 3 5 14 16 19 20 23 32 33 37
	moveDotLeftSmallWordTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(0),
		tt.Args(moveTestBuffer, 1).Rets(0),
		tt.Args(moveTestBuffer, 2).Rets(0),
		tt.Args(moveTestBuffer, 3).Rets(0),
		tt.Args(moveTestBuffer, 4).Rets(3),
		tt.Args(moveTestBuffer, 5).Rets(3),
		tt.Args(moveTestBuffer, 14).Rets(5),
		tt.Args(moveTestBuffer, 16).Rets(14),
		tt.Args(moveTestBuffer, 19).Rets(16),
		tt.Args(moveTestBuffer, 20).Rets(19),
		tt.Args(moveTestBuffer, 23).Rets(20),
		tt.Args(moveTestBuffer, 32).Rets(23),
		tt.Args(moveTestBuffer, 33).Rets(32),
		tt.Args(moveTestBuffer, 37).Rets(33),
		tt.Args(moveTestBuffer, 40).Rets(37),
	}
	moveDotRightSmallWordTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(3),
		tt.Args(moveTestBuffer, 1).Rets(3),
		tt.Args(moveTestBuffer, 2).Rets(3),
		tt.Args(moveTestBuffer, 3).Rets(5),
		tt.Args(moveTestBuffer, 5).Rets(14),
		tt.Args(moveTestBuffer, 14).Rets(16),
		tt.Args(moveTestBuffer, 16).Rets(19),
		tt.Args(moveTestBuffer, 19).Rets(20),
		tt.Args(moveTestBuffer, 20).Rets(23),
		tt.Args(moveTestBuffer, 23).Rets(32),
		tt.Args(moveTestBuffer, 32).Rets(33),
		tt.Args(moveTestBuffer, 33).Rets(37),
		tt.Args(moveTestBuffer, 37).Rets(40),
	}

	// alphanumeric boundaries: 0 5 16 20 23 33
	moveDotLeftAlphanumericTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(0),
		tt.Args(moveTestBuffer, 1).Rets(0),
		tt.Args(moveTestBuffer, 2).Rets(0),
		tt.Args(moveTestBuffer, 3).Rets(0),
		tt.Args(moveTestBuffer, 4).Rets(0),
		tt.Args(moveTestBuffer, 5).Rets(0),
		tt.Args(moveTestBuffer, 6).Rets(5),
		tt.Args(moveTestBuffer, 16).Rets(5),
		tt.Args(moveTestBuffer, 20).Rets(16),
		tt.Args(moveTestBuffer, 23).Rets(20),
		tt.Args(moveTestBuffer, 33).Rets(23),
		tt.Args(moveTestBuffer, 40).Rets(33),
	}
	moveDotRightAlphanumericTests = tt.Table{
		tt.Args(moveTestBuffer, 0).Rets(5),
		tt.Args(moveTestBuffer, 1).Rets(5),
		tt.Args(moveTestBuffer, 2).Rets(5),
		tt.Args(moveTestBuffer, 3).Rets(5),
		tt.Args(moveTestBuffer, 4).Rets(5),
		tt.Args(moveTestBuffer, 5).Rets(16),
		tt.Args(moveTestBuffer, 16).Rets(20),
		tt.Args(moveTestBuffer, 20).Rets(23),
		tt.Args(moveTestBuffer, 23).Rets(33),
		tt.Args(moveTestBuffer, 33).Rets(40),
	}
)

func TestMoveDotRune(t *testing.T) {
	tt.Test(t, tt.Fn("moveDotLeft", moveDotLeft), moveDotLeftTests)
	tt.Test(t, tt.Fn("moveDotRight", moveDotRight), moveDotRightTests)
}

func TestMoveDotWord(t *testing.T) {
	tt.Test(t,
		tt.Fn("moveDotLeftWord", moveDotLeftWord),
		moveDotLeftWordTests,
	)
	tt.Test(t,
		tt.Fn("moveDotRightWord", moveDotRightWord),
		moveDotRightWordTests,
	)
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

func TestMoveDotAlphanumeric(t *testing.T) {
	tt.Test(t,
		tt.Fn("moveDotLeftAlphanumeric", moveDotLeftAlphanumeric),
		moveDotLeftAlphanumericTests,
	)
	tt.Test(t,
		tt.Fn("moveDotRightAlphanumeric", moveDotRightAlphanumeric),
		moveDotRightAlphanumericTests,
	)
}
