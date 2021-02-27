package query_test

import (
	"testing"

	"src.elv.sh/pkg/edit/query"
	"src.elv.sh/pkg/parse"
)

func TestCompile(t *testing.T) {
	test(t,
		That("empty query matches anything").
			Query("").Matches("foo", "bar", " ", ""),

		That("bareword matches any string containing it").
			Query("foo").Matches("foobar", "afoo").DoesNotMatch("", "faoo"),
		That("bareword is case-insensitive is query is all lower case").
			Query("foo").Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),
		That("bareword is case-sensitive is query is not all lower case").
			Query("Foo").Matches("Foobar").DoesNotMatch("foo", "FOO"),

		That("double quoted string works like bareword").
			Query(`"foo"`).Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),

		That("single quoted string works like bareword").
			Query(`'foo'`).Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),

		That("space-separated words work like an AND query").
			Query("foo bar").
			Matches("foobar", "bar foo", "foo lorem ipsum bar").
			DoesNotMatch("foo", "bar", ""),

		That("quoted string can be used when string contains spaces").
			Query(`"foo bar"`).
			Matches("__foo bar xyz").
			DoesNotMatch("foobar"),

		That("AND query matches if all components match").
			Query("[and foo bar]").Matches("foobar", "bar foo").DoesNotMatch("foo"),
		That("OR query matches if any component matches").
			Query("[or foo bar]").Matches("foo", "bar", "foobar").DoesNotMatch(""),
		That("RE query uses component as regular expression to match").
			Query("[re f..]").Matches("foo", "f..").DoesNotMatch("fo", ""),

		// Invalid queries
		That("empty list is invalid").
			Query("[]").DoesNotCompile("empty subquery"),
		That("starting list with non-literal is invalid").
			Query("[[foo] bar]").
			DoesNotCompile("non-literal subquery head not supported"),
		That("RE query with no argument is invalid").
			Query("[re]").
			DoesNotCompile("re subquery with no argument not supported"),
		That("RE query with two or more arguments is invalid").
			Query("[re foo bar]").
			DoesNotCompile("re subquery with two or more arguments not supported"),
		That("RE query with invalid regular expression is invalid").
			Query("[re '[']").
			DoesNotCompile("error parsing regexp: missing closing ]: `[`"),
		That("invalid syntax results in parse error").
			Query("[and").DoesNotParse("parse error: 4-4 in query: should be ']'"),

		// Unsupported for now, but may be in future
		That("options are not supported yet").
			Query("foo &k=v").DoesNotCompile("option not supported"),
		That("compound expressions are not supported yet").
			Query(`a"foo"`).DoesNotCompile("compound expression not supported"),
		That("indexing expressions are not supported yet").
			Query("foo[0]").DoesNotCompile("indexing expression not supported"),
		That("variable references are not supported yet").
			Query("$a").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("variable references in RE subquery are not supported yet").
			Query("[re $a]").
			DoesNotCompile("re subquery with primary expression of type Variable not supported"),
		That("variable references in AND subquery are not supported yet").
			Query("[and $a]").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("variable references in OR subquery are not supported yet").
			Query("[or $a]").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("other subqueries are not supported yet").
			Query("[other foo bar]").
			DoesNotCompile("head other not supported"),
	)
}

func test(t *testing.T, tests ...testCase) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := query.Compile(test.query)
			if errType := getErrorType(err); errType != test.errorType {
				t.Errorf("%q should have %s, but has %s",
					test.query, test.errorType, errType)
			}
			if err != nil {
				if err.Error() != test.errorMessage {
					t.Errorf("%q should have error message %q, but is %q",
						test.query, test.errorMessage, err)
				}
				return
			}
			for _, s := range test.matches {
				ok := q.Match(s)
				if !ok {
					t.Errorf("%q should match %q, but doesn't", test.query, s)
				}
			}
			for _, s := range test.doesntMatch {
				ok := q.Match(s)
				if ok {
					t.Errorf("%q shouldn't match %q, but does", test.query, s)
				}
			}
		})
	}
}

type testCase struct {
	name         string
	query        string
	matches      []string
	doesntMatch  []string
	errorType    errorType
	errorMessage string
}

func That(name string) testCase {
	return testCase{name: name}
}

func (t testCase) Query(q string) testCase {
	t.query = q
	return t
}

func (t testCase) DoesNotParse(message string) testCase {
	t.errorType = parseError
	t.errorMessage = message
	return t
}

func (t testCase) DoesNotCompile(message string) testCase {
	t.errorType = compileError
	t.errorMessage = message
	return t
}

func (t testCase) Matches(s ...string) testCase {
	t.matches = s
	return t
}

func (t testCase) DoesNotMatch(s ...string) testCase {
	t.doesntMatch = s
	return t
}

type errorType uint

const (
	noError errorType = iota
	parseError
	compileError
)

func getErrorType(err error) errorType {
	switch err.(type) {
	case nil:
		return noError
	case *parse.Error:
		return parseError
	default:
		return compileError
	}
}

func (et errorType) String() string {
	switch et {
	case noError:
		return "no error"
	case parseError:
		return "parse error"
	case compileError:
		return "compile error"
	default:
		panic("unreachable")
	}
}
