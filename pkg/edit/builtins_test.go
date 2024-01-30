package edit

import (
	"io"
	"strings"
	"testing"

	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

func TestBindingTable(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `var called = $false`)
	evals(f.Evaler, `var m = (edit:binding-table [&a={ set called = $true }])`)
	_, ok := getGlobal(f.Evaler, "m").(bindingsMap)
	if !ok {
		t.Errorf("edit:binding-table did not create BindingMap variable")
	}
}

func TestCloseMode(t *testing.T) {
	f := setup(t)

	f.Editor.app.PushAddon(tk.Empty{})
	evals(f.Evaler, `edit:close-mode`)

	if addons := f.Editor.app.CopyState().Addons; len(addons) > 0 {
		t.Errorf("got addons %v, want nil or empty slice", addons)
	}
}

func TestInsertRaw(t *testing.T) {
	f := setup(t)

	f.TTYCtrl.Inject(term.K('V', ui.Ctrl))
	wantBuf := f.MakeBuffer(
		"~> ", term.DotHere, "\n",
		" RAW ", Styles,
		"*****",
	)
	f.TTYCtrl.TestBuffer(t, wantBuf)
	// Since we do not use real terminals in the test, we cannot have a
	// realistic test case against actual raw inputs. However, we can still
	// check that the builtin command does call the SetRawInput method with 1.
	if raw := f.TTYCtrl.RawInput(); raw != 1 {
		t.Errorf("RawInput() -> %d, want 1", raw)
	}

	// Raw mode does not respond to non-key events.
	f.TTYCtrl.Inject(term.MouseEvent{})
	f.TTYCtrl.TestBuffer(t, wantBuf)

	// Raw mode is dismissed after a single key event.
	f.TTYCtrl.Inject(term.K('+'))
	f.TestTTY(t,
		"~> +", Styles,
		"   v", term.DotHere,
	)
}

func TestEndOfHistory(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `edit:end-of-history`)
	f.TestTTYNotes(t, "End of history")
}

func TestKey(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `var k = (edit:key a)`)
	wantK := ui.K('a')
	if k, _ := f.Evaler.Global().Index("k"); k != wantK {
		t.Errorf("$k is %v, want %v", k, wantK)
	}
}

func TestRedraw(t *testing.T) {
	f := setup(t)

	evals(f.Evaler,
		`set edit:current-command = echo`,
		`edit:redraw`)
	f.TestTTY(t,
		"~> echo", Styles,
		"   vvvv", term.DotHere)
	evals(f.Evaler, `edit:redraw &full=$true`)
	// TODO(xiaq): Test that this is actually a full redraw.
	f.TestTTY(t,
		"~> echo", Styles,
		"   vvvv", term.DotHere)
}

func TestClear(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `set edit:current-command = echo`, `edit:clear`)
	f.TestTTY(t,
		"~> echo", Styles,
		"   vvvv", term.DotHere)
	if cleared := f.TTYCtrl.ScreenCleared(); cleared != 1 {
		t.Errorf("screen cleared %v times, want 1", cleared)
	}
}

func TestNotify(t *testing.T) {
	f := setup(t)
	evals(f.Evaler, "edit:notify string")
	f.TestTTYNotes(t, "string")

	evals(f.Evaler, "edit:notify (styled styled red)")
	f.TestTTYNotes(t,
		"styled", Styles,
		"!!!!!!")

	evals(f.Evaler, "var err = ?(edit:notify [])")
	if _, hasErr := getGlobal(f.Evaler, "err").(error); !hasErr {
		t.Errorf("calling edit:notify with [] did not result in error")
		// TODO: Test the exact error
	}
}

func TestReturnCode(t *testing.T) {
	f := setup(t)

	codeArea(f.Editor.app).MutateState(func(s *tk.CodeAreaState) {
		s.Buffer.Content = "test code"
	})
	evals(f.Evaler, `edit:return-line`)
	code, err := f.Wait()
	if code != "test code" {
		t.Errorf("got code %q, want %q", code, "test code")
	}
	if err != nil {
		t.Errorf("got err %v, want nil", err)
	}
}

func TestReturnEOF(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, `edit:return-eof`)
	if _, err := f.Wait(); err != io.EOF {
		t.Errorf("got err %v, want %v", err, io.EOF)
	}
}

func TestSmartEnter_InsertsNewlineWhenIncomplete(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "put [", Dot: 5})
	evals(f.Evaler, `edit:smart-enter`)
	wantBuf := tk.CodeBuffer{Content: "put [\n", Dot: 6}
	if buf := codeArea(f.Editor.app).CopyState().Buffer; buf != wantBuf {
		t.Errorf("got code buffer %v, want %v", buf, wantBuf)
	}
}

func TestSmartEnter_AcceptsCodeWhenWholeBufferIsComplete(t *testing.T) {
	f := setup(t)

	f.SetCodeBuffer(tk.CodeBuffer{Content: "put []", Dot: 5})
	evals(f.Evaler, `edit:smart-enter`)
	wantCode := "put []"
	if code, _ := f.Wait(); code != wantCode {
		t.Errorf("got return code %q, want %q", code, wantCode)
	}
}

// TODO: Test that smart-enter applies autofix.

var bufferBuiltinsTests = []struct {
	name      string
	bufBefore tk.CodeBuffer
	bufAfter  tk.CodeBuffer
}{
	{
		"move-dot-left",
		tk.CodeBuffer{Content: "ab", Dot: 1},
		tk.CodeBuffer{Content: "ab", Dot: 0},
	},
	{
		"move-dot-right",
		tk.CodeBuffer{Content: "ab", Dot: 1},
		tk.CodeBuffer{Content: "ab", Dot: 2},
	},
	{
		"kill-rune-left",
		tk.CodeBuffer{Content: "ab", Dot: 1},
		tk.CodeBuffer{Content: "b", Dot: 0},
	},
	{
		"kill-rune-right",
		tk.CodeBuffer{Content: "ab", Dot: 1},
		tk.CodeBuffer{Content: "a", Dot: 1},
	},
	{
		"transpose-rune with empty buffer",
		tk.CodeBuffer{Content: "", Dot: 0},
		tk.CodeBuffer{Content: "", Dot: 0},
	},
	{
		"transpose-rune with dot at beginning",
		tk.CodeBuffer{Content: "abc", Dot: 0},
		tk.CodeBuffer{Content: "bac", Dot: 2},
	},
	{
		"transpose-rune with dot in middle",
		tk.CodeBuffer{Content: "abc", Dot: 1},
		tk.CodeBuffer{Content: "bac", Dot: 2},
	},
	{
		"transpose-rune with dot at end",
		tk.CodeBuffer{Content: "abc", Dot: 3},
		tk.CodeBuffer{Content: "acb", Dot: 3},
	},
	{
		"transpose-rune with one character and dot at end",
		tk.CodeBuffer{Content: "a", Dot: 1},
		tk.CodeBuffer{Content: "a", Dot: 1},
	},
	{
		"transpose-rune with one character and dot at beginning",
		tk.CodeBuffer{Content: "a", Dot: 0},
		tk.CodeBuffer{Content: "a", Dot: 0},
	},
	{
		"transpose-word with dot at beginning",
		tk.CodeBuffer{Content: "ab  bc cd", Dot: 0},
		tk.CodeBuffer{Content: "bc  ab cd", Dot: 6},
	},
	{
		"transpose-word with dot in between words",
		tk.CodeBuffer{Content: "ab  bc cd", Dot: 6},
		tk.CodeBuffer{Content: "ab  cd bc", Dot: 9},
	},
	{
		"transpose-word with dot at end",
		tk.CodeBuffer{Content: "ab  bc cd", Dot: 9},
		tk.CodeBuffer{Content: "ab  cd bc", Dot: 9},
	},
	{
		"transpose-word with dot in the middle of a word",
		tk.CodeBuffer{Content: "ab  bc cd", Dot: 5},
		tk.CodeBuffer{Content: "bc  ab cd", Dot: 6},
	},
	{
		"transpose-word with one word",
		tk.CodeBuffer{Content: " ab  ", Dot: 4},
		tk.CodeBuffer{Content: " ab  ", Dot: 4},
	},
	{
		"transpose-word with no words",
		tk.CodeBuffer{Content: " \t\n  ", Dot: 4},
		tk.CodeBuffer{Content: " \t\n  ", Dot: 4},
	},
	{
		"transpose-word with complex input",
		tk.CodeBuffer{Content: "cd ~/downloads;", Dot: 4},
		tk.CodeBuffer{Content: "~/downloads; cd", Dot: 15},
	},
	{
		"transpose-small-word",
		tk.CodeBuffer{Content: "cd ~/downloads;", Dot: 4},
		tk.CodeBuffer{Content: "~/ cddownloads;", Dot: 5},
	},
	{
		"transpose-alnum-word",
		tk.CodeBuffer{Content: "cd ~/downloads;", Dot: 4},
		tk.CodeBuffer{Content: "downloads ~/cd;", Dot: 14},
	},
}

func TestBufferBuiltins(t *testing.T) {
	f := setup(t)
	app := f.Editor.app

	for _, test := range bufferBuiltinsTests {
		t.Run(test.name, func(t *testing.T) {
			codeArea(app).MutateState(func(s *tk.CodeAreaState) {
				s.Buffer = test.bufBefore
			})
			cmd := strings.Split(test.name, " ")[0]
			evals(f.Evaler, "edit:"+cmd)
			if buf := codeArea(app).CopyState().Buffer; buf != test.bufAfter {
				t.Errorf("got buf %v, want %v", buf, test.bufAfter)
			}
		})
	}
}

// Builtins that expect the focused widget to be code areas. This
// includes some builtins defined in files other than builtins.go.
var focusedWidgetNotCodeAreaTests = []string{
	"edit:insert-raw",
	"edit:smart-enter",
	"edit:move-dot-right", // other buffer builtins not tested
	"edit:completion:start",
	"edit:history:start",
}

func TestBuiltins_FocusedWidgetNotCodeArea(t *testing.T) {
	for _, code := range focusedWidgetNotCodeAreaTests {
		t.Run(code, func(t *testing.T) {
			f := setup(t)
			f.Editor.app.PushAddon(tk.Label{})

			evals(f.Evaler, code)
			f.TestTTYNotes(t,
				"error: "+modes.ErrFocusedWidgetNotCodeArea.Error(), Styles,
				"!!!!!!")
		})
	}
}

// Tests for pure movers.

func TestMoveDotLeftRight(t *testing.T) {
	tt.Test(t, moveDotLeft,
		Args("foo", 0).Rets(0),
		Args("bar", 3).Rets(2),
		Args("精灵", 0).Rets(0),
		Args("精灵", 3).Rets(0),
		Args("精灵", 6).Rets(3),
	)

	tt.Test(t, moveDotRight,
		Args("foo", 0).Rets(1),
		Args("bar", 3).Rets(3),
		Args("精灵", 0).Rets(3),
		Args("精灵", 3).Rets(6),
		Args("精灵", 6).Rets(6),
	)
}

func TestMoveDotSOLEOL(t *testing.T) {
	buffer := "abc\ndef"
	// Index:
	//         012 34567
	tt.Test(t, moveDotSOL,
		Args(buffer, 0).Rets(0),
		Args(buffer, 1).Rets(0),
		Args(buffer, 2).Rets(0),
		Args(buffer, 3).Rets(0),
		Args(buffer, 4).Rets(4),
		Args(buffer, 5).Rets(4),
		Args(buffer, 6).Rets(4),
		Args(buffer, 7).Rets(4),
	)
	tt.Test(t, moveDotEOL,
		Args(buffer, 0).Rets(3),
		Args(buffer, 1).Rets(3),
		Args(buffer, 2).Rets(3),
		Args(buffer, 3).Rets(3),
		Args(buffer, 4).Rets(7),
		Args(buffer, 5).Rets(7),
		Args(buffer, 6).Rets(7),
		Args(buffer, 7).Rets(7),
	)
}

func TestMoveDotUpDown(t *testing.T) {
	buffer := "abc\n精灵语\ndef"
	// Index:
	//         012 34 7 0  34567
	// + 10 *  0        1

	tt.Test(t, moveDotUp,
		Args(buffer, 0).Rets(0),  // a -> a
		Args(buffer, 1).Rets(1),  // b -> b
		Args(buffer, 2).Rets(2),  // c -> c
		Args(buffer, 3).Rets(3),  // EOL1 -> EOL1
		Args(buffer, 4).Rets(0),  // 精 -> a
		Args(buffer, 7).Rets(2),  // 灵 -> c
		Args(buffer, 10).Rets(3), // 语 -> EOL1
		Args(buffer, 13).Rets(3), // EOL2 -> EOL1
		Args(buffer, 14).Rets(4), // d -> 精
		Args(buffer, 15).Rets(4), // e -> 精 (jump left half width)
		Args(buffer, 16).Rets(7), // f -> 灵
		Args(buffer, 17).Rets(7), // EOL3 -> 灵 (jump left half width)
	)

	tt.Test(t, moveDotDown,
		Args(buffer, 0).Rets(4),   // a -> 精
		Args(buffer, 1).Rets(4),   // b -> 精 (jump left half width)
		Args(buffer, 2).Rets(7),   // c -> 灵
		Args(buffer, 3).Rets(7),   // EOL1 -> 灵 (jump left half width)
		Args(buffer, 4).Rets(14),  // 精 -> d
		Args(buffer, 7).Rets(16),  // 灵 -> f
		Args(buffer, 10).Rets(17), // 语 -> EOL3
		Args(buffer, 13).Rets(17), // EOL2 -> EOL3
		Args(buffer, 14).Rets(14), // d -> d
		Args(buffer, 15).Rets(15), // e -> e
		Args(buffer, 16).Rets(16), // f -> f
		Args(buffer, 17).Rets(17), // EOL3 -> EOL3
	)
}

// Word movement tests.

// The string below is carefully chosen to test all word, small-word, and
// alnum-word move/kill functions, because it contains features to set the
// different movement behaviors apart.
//
// The string is annotated with carets (^) to indicate the beginning of words,
// and periods (.) to indicate trailing runes of words. Indices are also
// annotated.
//
//	cd ~/downloads; rm -rf 2018aug07-pics/*;
//	^. ^........... ^. ^.. ^................  (word)
//	^. ^.^........^ ^. ^^. ^........^^...^..  (small-word)
//	^.   ^........  ^.  ^. ^........ ^...     (alnum-word)
//	01234567890123456789012345678901234567890
//	0         1         2         3         4
//
//	word boundaries:         0 3      16 19    23
//	small-word boundaries:   0 3 5 14 16 19 20 23 32 33 37
//	alnum-word boundaries:   0   5    16    20 23    33
var wordMoveTestBuffer = "cd ~/downloads; rm -rf 2018aug07-pics/*;"

var (
	// word boundaries: 0 3 16 19 23
	moveDotLeftWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(0),
		Args(wordMoveTestBuffer, 1).Rets(0),
		Args(wordMoveTestBuffer, 2).Rets(0),
		Args(wordMoveTestBuffer, 3).Rets(0),
		Args(wordMoveTestBuffer, 4).Rets(3),
		Args(wordMoveTestBuffer, 16).Rets(3),
		Args(wordMoveTestBuffer, 19).Rets(16),
		Args(wordMoveTestBuffer, 23).Rets(19),
		Args(wordMoveTestBuffer, 40).Rets(23),
	}
	moveDotRightWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(3),
		Args(wordMoveTestBuffer, 1).Rets(3),
		Args(wordMoveTestBuffer, 2).Rets(3),
		Args(wordMoveTestBuffer, 3).Rets(16),
		Args(wordMoveTestBuffer, 16).Rets(19),
		Args(wordMoveTestBuffer, 19).Rets(23),
		Args(wordMoveTestBuffer, 23).Rets(40),
	}

	// small-word boundaries: 0 3 5 14 16 19 20 23 32 33 37
	moveDotLeftSmallWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(0),
		Args(wordMoveTestBuffer, 1).Rets(0),
		Args(wordMoveTestBuffer, 2).Rets(0),
		Args(wordMoveTestBuffer, 3).Rets(0),
		Args(wordMoveTestBuffer, 4).Rets(3),
		Args(wordMoveTestBuffer, 5).Rets(3),
		Args(wordMoveTestBuffer, 14).Rets(5),
		Args(wordMoveTestBuffer, 16).Rets(14),
		Args(wordMoveTestBuffer, 19).Rets(16),
		Args(wordMoveTestBuffer, 20).Rets(19),
		Args(wordMoveTestBuffer, 23).Rets(20),
		Args(wordMoveTestBuffer, 32).Rets(23),
		Args(wordMoveTestBuffer, 33).Rets(32),
		Args(wordMoveTestBuffer, 37).Rets(33),
		Args(wordMoveTestBuffer, 40).Rets(37),
	}
	moveDotRightSmallWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(3),
		Args(wordMoveTestBuffer, 1).Rets(3),
		Args(wordMoveTestBuffer, 2).Rets(3),
		Args(wordMoveTestBuffer, 3).Rets(5),
		Args(wordMoveTestBuffer, 5).Rets(14),
		Args(wordMoveTestBuffer, 14).Rets(16),
		Args(wordMoveTestBuffer, 16).Rets(19),
		Args(wordMoveTestBuffer, 19).Rets(20),
		Args(wordMoveTestBuffer, 20).Rets(23),
		Args(wordMoveTestBuffer, 23).Rets(32),
		Args(wordMoveTestBuffer, 32).Rets(33),
		Args(wordMoveTestBuffer, 33).Rets(37),
		Args(wordMoveTestBuffer, 37).Rets(40),
	}

	// alnum-word boundaries: 0 5 16 20 23 33
	moveDotLeftAlnumWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(0),
		Args(wordMoveTestBuffer, 1).Rets(0),
		Args(wordMoveTestBuffer, 2).Rets(0),
		Args(wordMoveTestBuffer, 3).Rets(0),
		Args(wordMoveTestBuffer, 4).Rets(0),
		Args(wordMoveTestBuffer, 5).Rets(0),
		Args(wordMoveTestBuffer, 6).Rets(5),
		Args(wordMoveTestBuffer, 16).Rets(5),
		Args(wordMoveTestBuffer, 20).Rets(16),
		Args(wordMoveTestBuffer, 23).Rets(20),
		Args(wordMoveTestBuffer, 33).Rets(23),
		Args(wordMoveTestBuffer, 40).Rets(33),
	}
	moveDotRightAlnumWordTests = []*tt.Case{
		Args(wordMoveTestBuffer, 0).Rets(5),
		Args(wordMoveTestBuffer, 1).Rets(5),
		Args(wordMoveTestBuffer, 2).Rets(5),
		Args(wordMoveTestBuffer, 3).Rets(5),
		Args(wordMoveTestBuffer, 4).Rets(5),
		Args(wordMoveTestBuffer, 5).Rets(16),
		Args(wordMoveTestBuffer, 16).Rets(20),
		Args(wordMoveTestBuffer, 20).Rets(23),
		Args(wordMoveTestBuffer, 23).Rets(33),
		Args(wordMoveTestBuffer, 33).Rets(40),
	}
)

func TestMoveDotWord(t *testing.T) {
	tt.Test(t, moveDotLeftWord, moveDotLeftWordTests...)
	tt.Test(t, moveDotRightWord, moveDotRightWordTests...)
}

func TestMoveDotSmallWord(t *testing.T) {
	tt.Test(t, moveDotLeftSmallWord, moveDotLeftSmallWordTests...)
	tt.Test(t, moveDotRightSmallWord, moveDotRightSmallWordTests...)
}

func TestMoveDotAlnumWord(t *testing.T) {
	tt.Test(t, moveDotLeftAlnumWord, moveDotLeftAlnumWordTests...)
	tt.Test(t, moveDotRightAlnumWord, moveDotRightAlnumWordTests...)
}
