package edit

import (
	"reflect"
	"testing"
)

var tokenizeTestCases = []struct {
	src  string
	want []tokenTest
}{
	{"", []tokenTest{}},
	{"ls", []tokenTest{{Bareword, "ls"}}},
	{` ls 'a' "b" [c] `, []tokenTest{
		{Sep, " "}, {Bareword, "ls"}, {Sep, " "}, {SingleQuoted, `'a'`},
		{Sep, " "}, {DoubleQuoted, `"b"`}, {Sep, " "}, {Sep, "["},
		{Bareword, "c"}, {Sep, "]"}, {Sep, " "},
	}},
	{`ls ~a $b *c`, []tokenTest{
		{Bareword, "ls"}, {Sep, " "}, {Tilde, "~"}, {Bareword, "a"},
		{Sep, " "}, {Variable, "$b"}, {Sep, " "}, {Wildcard, "*"},
		{Bareword, "c"},
	}},
	{"a | b \n c #d", []tokenTest{
		{Bareword, "a"}, {Sep, " "}, {Sep, "|"}, {Sep, " "},
		{Bareword, "b"}, {Sep, " "}, {Sep, "\n"}, {Sep, " "},
		{Bareword, "c"}, {Sep, " "}, {Sep, "#d"},
	}},
	{"a ]", []tokenTest{
		{Bareword, "a"}, {Sep, " "}, {ParserError, "]"},
	}},
}

type tokenTest struct {
	Type TokenKind
	Text string
}

func (test tokenTest) test(t *Token) bool {
	return test.Type == t.Type && test.Text == t.Text
}

func testTokens(want []tokenTest, tokens []Token) bool {
	if len(want) != len(tokens) {
		return false
	}
	for i, t := range want {
		if !t.test(&tokens[i]) {
			return false
		}
	}
	return true
}

func TestTokenize(t *testing.T) {
	for _, testcase := range tokenizeTestCases {
		tokens := tokenize(testcase.src)
		if !testTokens(testcase.want, tokens) {
			t.Errorf("tokenize(%q) => %v, want %v", testcase.src, tokens, testcase.want)
		}
	}
}

var wordifyTestCases = []struct {
	src  string
	want []string
}{
	{" ls | cat\n\n ", []string{"ls", "|", "cat"}},
	{"a;\nb", []string{"a", ";", "b"}},
	{`put ['ha ha'] "lala" `, []string{"put", "[", "'ha ha'", "]", `"lala"`}},
}

func TestWordify(t *testing.T) {
	for _, testcase := range wordifyTestCases {
		words := wordify(testcase.src)
		if !reflect.DeepEqual(words, testcase.want) {
			t.Errorf("wordify(%q) => %#v, want %#v", testcase.src, words, testcase.want)
		}
	}
}
