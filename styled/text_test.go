package styled

import (
	"testing"

	"github.com/elves/elvish/tt"
)

// TODO: Test other methods of Text.

var Args = tt.Args

var (
	text0 = Text{}
	text1 = Text{red("lorem")}
	text2 = Text{red("lorem"), blue("foobar")}
)

func red(s string) Segment  { return Segment{Style{Foreground: "red"}, s} }
func blue(s string) Segment { return Segment{Style{Foreground: "blue"}, s} }

var partitionTests = tt.Table{
	Args(text0).Rets([]Text{text0}),
	Args(text1).Rets([]Text{text1}),
	Args(text1, 0).Rets([]Text{text0, text1}),
	Args(text1, 1).Rets([]Text{Text{red("l")}, Text{red("orem")}}),
	Args(text1, 5).Rets([]Text{text1, text0}),
	Args(text2).Rets([]Text{text2}),
	Args(text2, 0).Rets([]Text{text0, text2}),
	Args(text2, 1).Rets([]Text{
		Text{red("l")}, Text{red("orem"), blue("foobar")}}),
	Args(text2, 2).Rets([]Text{
		Text{red("lo")}, Text{red("rem"), blue("foobar")}}),
	Args(text2, 5).Rets([]Text{Text{red("lorem")}, Text{blue("foobar")}}),
	Args(text2, 6).Rets([]Text{
		Text{red("lorem"), blue("f")}, Text{blue("oobar")}}),
	Args(text2, 11).Rets([]Text{text2, text0}),
}

func TestPartition(t *testing.T) {
	tt.Test(t, tt.Fn("Text.Parition", Text.Partition), partitionTests)
}
