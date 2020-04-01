package ui

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/tt"
)

var Args = tt.Args

func TestT(t *testing.T) {
	tt.Test(t, tt.Fn("T", T), tt.Table{
		Args("test").Rets(Text{&Segment{Text: "test"}}),
		Args("test red", FgRed).Rets(Text{&Segment{
			Text: "test red", Style: Style{Foreground: Red}}}),
		Args("test red", FgRed, Bold).Rets(Text{&Segment{
			Text: "test red", Style: Style{Foreground: Red, Bold: true}}}),
	})
}

func TestTextAsElvishValue(t *testing.T) {
	vals.TestValue(t, T("text")).
		Kind("ui:text").
		Repr("(ui:text text)").
		AllKeys("0").
		Index("0", &Segment{Text: "text"})

	vals.TestValue(t, T("text", FgRed)).
		Repr("(ui:text (ui:text-segment text &fg-color=red))")
	vals.TestValue(t, T("text", Bold)).
		Repr("(ui:text (ui:text-segment text &bold=$true))")
}

var (
	text0 = Text{}
	text1 = Text{red("lorem")}
	text2 = Text{red("lorem"), blue("foobar")}
)

func red(s string) *Segment  { return &Segment{Style{Foreground: Red}, s} }
func blue(s string) *Segment { return &Segment{Style{Foreground: Blue}, s} }

var partitionTests = tt.Table{
	Args(text0).Rets([]Text{nil}),
	Args(text1).Rets([]Text{text1}),
	Args(text1, 0).Rets([]Text{nil, text1}),
	Args(text1, 1).Rets([]Text{{red("l")}, {red("orem")}}),
	Args(text1, 5).Rets([]Text{text1, nil}),
	Args(text2).Rets([]Text{text2}),
	Args(text2, 0).Rets([]Text{nil, text2}),
	Args(text2, 1).Rets([]Text{
		{red("l")}, {red("orem"), blue("foobar")}}),
	Args(text2, 2).Rets([]Text{
		{red("lo")}, {red("rem"), blue("foobar")}}),
	Args(text2, 5).Rets([]Text{{red("lorem")}, {blue("foobar")}}),
	Args(text2, 6).Rets([]Text{
		{red("lorem"), blue("f")}, {blue("oobar")}}),
	Args(text2, 11).Rets([]Text{text2, nil}),

	Args(text1, 1, 2).Rets([]Text{{red("l")}, {red("o")}, {red("rem")}}),
	Args(text1, 1, 2, 3, 4).Rets([]Text{
		{red("l")}, {red("o")}, {red("r")}, {red("e")}, {red("m")}}),
	Args(text2, 2, 4, 6).Rets([]Text{
		{red("lo")}, {red("re")},
		{red("m"), blue("f")}, {blue("oobar")}}),
	Args(text2, 6, 8).Rets([]Text{
		{red("lorem"), blue("f")}, {blue("oo")}, {blue("bar")}}),
}

func TestPartition(t *testing.T) {
	tt.Test(t, tt.Fn("Text.Parition", Text.Partition), partitionTests)
}

func TestCountRune(t *testing.T) {
	text := Text{red("lorem"), blue("ipsum")}
	tt.Test(t, tt.Fn("Text.CountRune", Text.CountRune), tt.Table{
		Args(text, 'l').Rets(1),
		Args(text, 'i').Rets(1),
		Args(text, 'm').Rets(2),
		Args(text, '\n').Rets(0),
	})
}

func TestCountLines(t *testing.T) {
	tt.Test(t, tt.Fn("Text.CountLines", Text.CountLines), tt.Table{
		Args(Text{red("lorem")}).Rets(1),
		Args(Text{red("lorem"), blue("ipsum")}).Rets(1),
		Args(Text{red("lor\nem"), blue("ipsum")}).Rets(2),
		Args(Text{red("lor\nem"), blue("ip\nsum")}).Rets(3),
	})
}

func TestSplitByRune(t *testing.T) {
	tt.Test(t, tt.Fn("Text.SplitByRune", Text.SplitByRune), tt.Table{
		Args(Text{}, '\n').Rets([]Text(nil)),
		Args(Text{red("lorem")}, '\n').Rets([]Text{{red("lorem")}}),
		Args(Text{red("lorem"), blue("ipsum"), red("dolar")}, '\n').Rets(
			[]Text{
				{red("lorem"), blue("ipsum"), red("dolar")},
			}),
		Args(Text{red("lo\nrem")}, '\n').Rets([]Text{
			{red("lo")}, {red("rem")},
		}),
		Args(Text{red("lo\nrem"), blue("ipsum")}, '\n').Rets(
			[]Text{
				{red("lo")},
				{red("rem"), blue("ipsum")},
			}),
		Args(Text{red("lo\nrem"), blue("ip\nsum")}, '\n').Rets(
			[]Text{
				{red("lo")},
				{red("rem"), blue("ip")},
				{blue("sum")},
			}),
		Args(Text{red("lo\nrem"), blue("ip\ns\num"), red("dolar")}, '\n').Rets(
			[]Text{
				{red("lo")},
				{red("rem"), blue("ip")},
				{blue("s")},
				{blue("um"), red("dolar")},
			}),
	})
}

func TestTrimWcwidth(t *testing.T) {
	tt.Test(t, tt.Fn("Text.TrimWcwidth", Text.TrimWcwidth), tt.Table{
		Args(Text{}, 1).Rets(Text(nil)),
		Args(Text{red("lorem")}, 3).Rets(Text{red("lor")}),
		Args(Text{red("lorem"), blue("ipsum")}, 6).Rets(
			Text{red("lorem"), blue("i")}),
		Args(Text{red("你好")}, 3).Rets(Text{red("你")}),
		Args(Text{red("你好"), blue("精灵语"), red("x")}, 7).Rets(
			Text{red("你好"), blue("精")}),
	})
}

type textVTStringTest struct {
	text         Text
	wantVTString string
}

func testTextVTString(t *testing.T, tests []textVTStringTest) {
	for _, test := range tests {
		vtString := test.text.VTString()
		if vtString != test.wantVTString {
			t.Errorf("got %q, want %q", vtString, test.wantVTString)
		}
	}
}
