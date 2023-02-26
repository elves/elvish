package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
)

func TestStyledSegment(t *testing.T) {
	Test(t,
		That("print (styled (styled-segment abc &fg-color=cyan) bold)").
			Prints("\033[;1;36mabc\033[m"),
		That("print (styled (styled-segment (styled-segment abc &fg-color=magenta) &dim=$true) cyan)").
			Prints("\033[;2;36mabc\033[m"),
		That("print (styled (styled-segment abc &inverse=$true) inverse)").
			Prints("\033[;7mabc\033[m"),
		That("print (styled (styled-segment abc) toggle-inverse)").
			Prints("\033[;7mabc\033[m"),
		That("print (styled (styled-segment abc &inverse=$true) no-inverse)").
			Prints("\033[mabc"),
		That("print (styled (styled-segment abc &inverse=$true) toggle-inverse)").
			Prints("\033[mabc"),

		That("styled-segment []").Throws(ErrorWithMessage(
			"argument to styled-segment must be a string or a styled segment")),
		That("styled-segment text &foo=bar").
			Throws(ErrorWithMessage("unrecognized option 'foo'")),
	)
}

func TestStyled(t *testing.T) {
	Test(t,
		// Transform string
		That("print (styled abc bold)").Prints("\033[;1mabc\033[m"),
		That("print (styled abc red cyan)").Prints("\033[;36mabc\033[m"),
		That("print (styled abc bg-green)").Prints("\033[;42mabc\033[m"),
		That("print (styled abc no-dim)").Prints("\033[mabc"),

		// Transform already styled text
		That("print (styled (styled abc red) blue)").
			Prints("\033[;34mabc\033[m"),
		That("print (styled (styled abc italic) red)").
			Prints("\033[;3;31mabc\033[m"),
		That("print (styled (styled abc inverse) inverse)").
			Prints("\033[;7mabc\033[m"),
		That("print (styled (styled abc inverse) no-inverse)").
			Prints("\033[mabc"),
		That("print (styled (styled abc inverse) toggle-inverse)").
			Prints("\033[mabc"),
		That("print (styled (styled abc inverse) toggle-inverse toggle-inverse)").
			Prints("\033[;7mabc\033[m"),

		// Function as transformer
		That("print (styled abc {|s| put $s })").Prints("\033[mabc"),
		That("print (styled abc {|s| styled-segment $s &bold=$true &italic=$false })").Prints("\033[;1mabc\033[m"),
		That("print (styled abc italic {|s| styled-segment $s &bold=$true &italic=$false })").Prints("\033[;1mabc\033[m"),

		That("styled abc {|_| fail bad }").Throws(eval.FailError{"bad"}),
		That("styled abc {|_| put a b }").Throws(ErrorWithMessage(
			"styling function must return a single segment; got 2 values")),
		That("styled abc {|_| put [] }").Throws(ErrorWithMessage(
			"styling function must return a segment; got list")),

		// Bad usage
		That("styled abc hopefully-never-exists").Throws(ErrorWithMessage(
			"hopefully-never-exists is not a valid style transformer")),
		That("styled []").Throws(ErrorWithMessage(
			"expected string, styled segment or styled text; got list")),
		That("styled abc []").Throws(ErrorWithMessage(
			"need string or callable; got list")),
	)
}

func TestStyled_DoesNotModifyArgument(t *testing.T) {
	Test(t,
		That("var x = (styled text); var y = (styled $x red); put $x[0][fg-color]").
			Puts("default"),
		That("var x = (styled-segment text); var y = (styled $x red); put $x[fg-color]").
			Puts("default"),
	)
}

func TestStyledConcat(t *testing.T) {
	Test(t,
		// segment + string
		That("print (styled-segment abc &fg-color=red)abc").Prints("\033[;31mabc\033[mabc"),
		// segment + segment
		That("print (styled-segment abc &bg-color=red)(styled-segment abc &fg-color=red)").Prints("\033[;41mabc\033[;31mabc\033[m"),
		// segment + text
		That("print (styled-segment abc &underlined=$true)(styled abc bright-cyan)").Prints("\033[;4mabc\033[;96mabc\033[m"),
		// segment + num
		That("print (styled-segment abc &blink)(num 44/3)").Prints("\033[;5mabc\033[m44/3"),
		That("print (styled-segment abc &blink)(num 42)").Prints("\033[;5mabc\033[m42"),
		// segment + unsupported
		That("print (styled-segment abc){ }").Throws(ErrorWithMessage("cannot concatenate ui:text-segment and fn")),
		// string + segment
		That("print abc(styled-segment abc &fg-color=red)").Prints("\033[mabc\033[31mabc\033[m"),
		// num + segment
		That("print (num 99.0)(styled-segment abc &blink)").Prints("\033[m99.0\033[5mabc\033[m"),
		That("print (num 66)(styled-segment abc &blink)").Prints("\033[m66\033[5mabc\033[m"),
		That("print (num 3/2)(styled-segment abc &blink)").Prints("\033[m3/2\033[5mabc\033[m"),
		// unsupported + segment
		That("print { }(styled-segment abc)").Throws(ErrorWithMessage("cannot concatenate fn and ui:text-segment")),

		// text + string
		That("print (styled abc blink)abc").Prints("\033[;5mabc\033[mabc"),
		// text + number
		That("print (styled abc blink)(num 13)").Prints("\033[;5mabc\033[m13"),
		That("print (styled abc blink)(num 3/4)").Prints("\033[;5mabc\033[m3/4"),
		// text + segment
		That("print (styled abc inverse)(styled-segment abc &bg-color=white)").Prints("\033[;7mabc\033[;47mabc\033[m"),
		// text + text
		That("print (styled abc bold)(styled abc dim)").Prints("\033[;1mabc\033[;2mabc\033[m"),
		// text + unsupported
		That("print (styled abc){ }").Throws(ErrorWithMessage("cannot concatenate ui:text and fn")),
		// string + text
		That("print abc(styled abc blink)").Prints("\033[mabc\033[5mabc\033[m"),
		// number + text
		That("print (num 13)(styled abc blink)").Prints("\033[m13\033[5mabc\033[m"),
		That("print (num 4/3)(styled abc blink)").Prints("\033[m4/3\033[5mabc\033[m"),
		// unsupported + text
		That("print { }(styled abc)").Throws(ErrorWithMessage("cannot concatenate fn and ui:text")),
	)
}

func TestStyledIndexing(t *testing.T) {
	Test(t,
		That("put (styled-segment abc &italic=$true &fg-color=red)[bold]").Puts(false),
		That("put (styled-segment abc &italic=$true &fg-color=red)[italic]").Puts(true),
		That("put (styled-segment abc &italic=$true &fg-color=red)[fg-color]").Puts("red"),
	)

	Test(t,
		That("put (styled abc red)[0][bold]").Puts(false),
		That("put (styled abc red)[0][bg-color]").Puts("default"),
		That("var t = (styled-segment abc &underlined=$true)(styled abc bright-cyan); put $t[1][fg-color]").Puts("bright-cyan"),
		That("var t = (styled-segment abc &underlined=$true)(styled abc bright-cyan); put $t[1][underlined]").Puts(false),
	)
}
