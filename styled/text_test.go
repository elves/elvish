package styled

import (
	"testing"

	"github.com/elves/elvish/tt"
)

// TODO: Test other methods of Text.

var Args = tt.Args

func TestPlain(t *testing.T) {
	tt.Test(t, tt.Fn("Plain", Plain), tt.Table{
		Args("test").Rets(Text{&Segment{Text: "test"}}),
		Args("").Rets(Text{&Segment{}}),
	})
}

func TestMakeText(t *testing.T) {
	tt.Test(t, tt.Fn("MakeText", MakeText), tt.Table{
		Args("test").Rets(Text{&Segment{Text: "test"}}),
		Args("test red", "red").Rets(Text{&Segment{
			Text: "test red", Style: Style{Foreground: "red"}}}),
		Args("test red", "red", "bold").Rets(Text{&Segment{
			Text: "test red", Style: Style{Foreground: "red", Bold: true}}}),
	})
}

var (
	text0 = Text{}
	text1 = Text{red("lorem")}
	text2 = Text{red("lorem"), blue("foobar")}
)

func red(s string) *Segment  { return &Segment{Style{Foreground: "red"}, s} }
func blue(s string) *Segment { return &Segment{Style{Foreground: "blue"}, s} }

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

	Args(text1, 1, 2).Rets([]Text{Text{red("l")}, Text{red("o")}, Text{red("rem")}}),
	Args(text1, 1, 2, 3, 4).Rets([]Text{
		Text{red("l")}, Text{red("o")}, Text{red("r")}, Text{red("e")}, Text{red("m")}}),
	Args(text2, 2, 4, 6).Rets([]Text{
		Text{red("lo")}, Text{red("re")},
		Text{red("m"), blue("f")}, Text{blue("oobar")}}),
	Args(text2, 6, 8).Rets([]Text{
		Text{red("lorem"), blue("f")}, Text{blue("oo")}, Text{blue("bar")}}),
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
		Args(Text{red("lorem")}, '\n').Rets([]Text{Text{red("lorem")}}),
		Args(Text{red("lorem"), blue("ipsum"), red("dolar")}, '\n').Rets(
			[]Text{
				Text{red("lorem"), blue("ipsum"), red("dolar")},
			}),
		Args(Text{red("lo\nrem")}, '\n').Rets([]Text{
			Text{red("lo")}, Text{red("rem")},
		}),
		Args(Text{red("lo\nrem"), blue("ipsum")}, '\n').Rets(
			[]Text{
				Text{red("lo")},
				Text{red("rem"), blue("ipsum")},
			}),
		Args(Text{red("lo\nrem"), blue("ip\nsum")}, '\n').Rets(
			[]Text{
				Text{red("lo")},
				Text{red("rem"), blue("ip")},
				Text{blue("sum")},
			}),
		Args(Text{red("lo\nrem"), blue("ip\ns\num"), red("dolar")}, '\n').Rets(
			[]Text{
				Text{red("lo")},
				Text{red("rem"), blue("ip")},
				Text{blue("s")},
				Text{blue("um"), red("dolar")},
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
