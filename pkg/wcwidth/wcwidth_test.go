package wcwidth

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestOf(t *testing.T) {
	tt.Test(t, Of,
		Args("\u0301").Rets(0), // Combining acute accent
		Args("a").Rets(1),
		Args("Ω").Rets(1),
		Args("好").Rets(2),
		Args("か").Rets(2),

		Args("abc").Rets(3),
		Args("你好").Rets(4),
	)
}

func TestOverride(t *testing.T) {
	r := '❱'
	oldw := OfRune(r)
	w := oldw + 1

	Override(r, w)
	if OfRune(r) != w {
		t.Errorf("Wcwidth(%q) != %d after OverrideWcwidth", r, w)
	}
	Unoverride(r)
	if OfRune(r) != oldw {
		t.Errorf("Wcwidth(%q) != %d after UnoverrideWcwidth", r, oldw)
	}
}

func TestOverride_NegativeWidthRemovesOverride(t *testing.T) {
	Override('x', 2)
	Override('x', -1)
	if OfRune('x') != 1 {
		t.Errorf("Override with negative width did not remove override")
	}
}

func TestConcurrentOverride(t *testing.T) {
	go Override('x', 2)
	_ = OfRune('x')
}

func TestTrim(t *testing.T) {
	tt.Test(t, Trim,
		Args("abc", 1).Rets("a"),
		Args("abc", 2).Rets("ab"),
		Args("abc", 3).Rets("abc"),
		Args("abc", 4).Rets("abc"),

		Args("你好", 1).Rets(""),
		Args("你好", 2).Rets("你"),
		Args("你好", 3).Rets("你"),
		Args("你好", 4).Rets("你好"),
		Args("你好", 5).Rets("你好"),
	)
}

func TestForce(t *testing.T) {
	tt.Test(t, Force,
		// Trimming
		Args("abc", 2).Rets("ab"),
		Args("你好", 2).Rets("你"),
		// Padding
		Args("abc", 4).Rets("abc "),
		Args("你好", 5).Rets("你好 "),
		// Trimming and Padding
		Args("你好", 3).Rets("你 "),
	)
}

func TestTrimEachLine(t *testing.T) {
	tt.Test(t, TrimEachLine,
		Args("abcdefg\n你好", 3).Rets("abc\n你"),
	)
}
