package eval

import (
	"errors"
	"testing"
)

func TestStyledString(t *testing.T) {
	Test(t,
		That("print (styled abc hopefully-never-exists)").ErrorsWith(errors.New("hopefully-never-exists is not a valid style transformer")),
		That("print (styled abc bold)").Prints("\033[1mabc\033[m"),
		That("print (styled abc red cyan)").Prints("\033[36mabc\033[m"),
		That("print (styled abc bg-green)").Prints("\033[42mabc\033[m"),
		That("print (styled abc no-dim)").Prints("abc"),
	)
}

func TestStyledSegment(t *testing.T) {
	Test(t,
		That("print (styled (styled-segment abc &fg-color=cyan) bold)").
			Prints("\033[1;36mabc\033[m"),
		That("print (styled (styled-segment (styled-segment abc &fg-color=magenta) &dim=$true) cyan)").
			Prints("\033[2;36mabc\033[m"),
		That("print (styled (styled-segment abc &inverse=$true) inverse)").
			Prints("\033[7mabc\033[m"),
		That("print (styled (styled-segment abc) toggle-inverse)").
			Prints("\033[7mabc\033[m"),
		That("print (styled (styled-segment abc &inverse=$true) no-inverse)").
			Prints("abc"),
		That("print (styled (styled-segment abc &inverse=$true) toggle-inverse)").
			Prints("abc"),
	)
}

func TestStyledText(t *testing.T) {
	Test(t,
		That("print (styled (styled abc red) blue)").
			Prints("\033[34mabc\033[m"),
		That("print (styled (styled abc italic) red)").
			Prints("\033[3;31mabc\033[m"),
		That("print (styled (styled abc inverse) inverse)").
			Prints("\033[7mabc\033[m"),
		That("print (styled (styled abc inverse) no-inverse)").Prints("abc"),
		That("print (styled (styled abc inverse) toggle-inverse)").Prints("abc"),
		That("print (styled (styled abc inverse) toggle-inverse toggle-inverse)").Prints("\033[7mabc\033[m"),
	)
}

func TestStyled_DoesNotModifyArgument(t *testing.T) {
	Test(t,
		That("x = (styled text); _ = (styled $x red); put $x[0][fg-color]").
			Puts("default"),
		That("x = (styled-segment text); _ = (styled $x red); put $x[fg-color]").
			Puts("default"),
	)
}

func TestStyledConcat(t *testing.T) {
	Test(t,
		// string+segment
		That("print abc(styled-segment abc &fg-color=red)").Prints("abc\033[31mabc\033[m"),
		// segment+string
		That("print (styled-segment abc &fg-color=red)abc").Prints("\033[31mabc\033[mabc"),
		// segment+segment
		That("print (styled-segment abc &bg-color=red)(styled-segment abc &fg-color=red)").Prints("\033[41mabc\033[m\033[31mabc\033[m"),
		// segment+text
		That("print (styled-segment abc &underlined=$true)(styled abc lightcyan)").Prints("\033[4mabc\033[m\033[96mabc\033[m"),
	)

	Test(t,
		// string+text
		That("print abc(styled abc blink)").Prints("abc\033[5mabc\033[m"),
		// text+string
		That("print (styled abc blink)abc").Prints("\033[5mabc\033[mabc"),
		// text+segment
		That("print (styled abc inverse)(styled-segment abc &bg-color=white)").Prints("\033[7mabc\033[m\033[107mabc\033[m"),
		// text+text
		That("print (styled abc bold)(styled abc dim)").Prints("\033[1mabc\033[m\033[2mabc\033[m"),
	)
}

func TestFunctionalStyleStylings(t *testing.T) {
	// lambda
	Test(t,
		That("print (styled abc [s]{ put $s })").Prints("abc"),
		That("print (styled abc [s]{ styled-segment $s &bold=$true &italic=$false })").Prints("\033[1mabc\033[m"),
		That("print (styled abc italic [s]{ styled-segment $s &bold=$true &italic=$false })").Prints("\033[1mabc\033[m"),
	)

	// fn
	Test(t,
		That("fn f [s]{ put $s }; print (styled abc $f~)").Prints("abc"),
		That("fn f [s]{ styled-segment $s &bold=$true &italic=$false }; print (styled abc $f~)").Prints("\033[1mabc\033[m"),
		That("fn f [s]{ styled-segment $s &bold=$true &italic=$false }; print (styled abc italic $f~)").Prints("\033[1mabc\033[m"),
	)

	// var
	Test(t,
		That("f = [s]{ put $s }; print (styled abc $f)").Prints("abc"),
		That("f = [s]{ styled-segment $s &bold=$true &italic=$false }; print (styled abc $f)").Prints("\033[1mabc\033[m"),
		That("f = [s]{ styled-segment $s &bold=$true &italic=$false }; print (styled abc italic $f)").Prints("\033[1mabc\033[m"),
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
		That("t = (styled-segment abc &underlined=$true)(styled abc lightcyan); put $t[1][fg-color]").Puts("lightcyan"),
		That("t = (styled-segment abc &underlined=$true)(styled abc lightcyan); put $t[1][underlined]").Puts(false),
	)
}
