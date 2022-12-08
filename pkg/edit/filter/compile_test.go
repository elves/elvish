package filter_test

import (
	"testing"

	"src.elv.sh/pkg/edit/filter"
	"src.elv.sh/pkg/parse"
)

func TestCompile(t *testing.T) {
	test(t,
		That("empty filter matches anything").
			Filter("").Matches("foo", "bar", " ", ""),

		That("bareword matches any string containing it").
			Filter("foo").Matches("foobar", "afoo").DoesNotMatch("", "faoo"),
		That("bareword is case-insensitive is filter is all lower case").
			Filter("foo").Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),
		That("bareword is case-sensitive is filter is not all lower case").
			Filter("Foo").Matches("Foobar").DoesNotMatch("foo", "FOO"),

		That("double quoted string works like bareword").
			Filter(`"foo"`).Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),

		That("single quoted string works like bareword").
			Filter(`'foo'`).Matches("FOO", "Foo", "FOObar").DoesNotMatch("", "faoo"),

		That("space-separated words work like an AND filter").
			Filter("foo bar").
			Matches("foobar", "bar foo", "foo lorem ipsum bar").
			DoesNotMatch("foo", "bar", ""),

		That("quoted string can be used when string contains spaces").
			Filter(`"foo bar"`).
			Matches("__foo bar xyz").
			DoesNotMatch("foobar"),

		That("AND filter matches if all components match").
			Filter("[and foo bar]").Matches("foobar", "bar foo").DoesNotMatch("foo"),
		That("OR filter matches if any component matches").
			Filter("[or foo bar]").Matches("foo", "bar", "foobar").DoesNotMatch(""),
		That("RE filter uses component as regular expression to match").
			Filter("[re f..]").Matches("foo", "f..").DoesNotMatch("fo", ""),

		// Invalid queries
		That("empty list is invalid").
			Filter("[]").DoesNotCompile("empty subfilter"),
		That("starting list with non-literal is invalid").
			Filter("[[foo] bar]").
			DoesNotCompile("non-literal subfilter head not supported"),
		That("RE filter with no argument is invalid").
			Filter("[re]").
			DoesNotCompile("re subfilter with no argument not supported"),
		That("RE filter with two or more arguments is invalid").
			Filter("[re foo bar]").
			DoesNotCompile("re subfilter with two or more arguments not supported"),
		That("RE filter with invalid regular expression is invalid").
			Filter("[re '[']").
			DoesNotCompile("error parsing regexp: missing closing ]: `[`"),
		That("invalid syntax results in parse error").
			Filter("[and").DoesNotParse("parse error: [filter]:1:5: should be ']'"),

		// Unsupported for now, but may be in future
		That("options are not supported yet").
			Filter("foo &k=v").DoesNotCompile("option not supported"),
		That("compound expressions are not supported yet").
			Filter(`a"foo"`).DoesNotCompile("compound expression not supported"),
		That("indexing expressions are not supported yet").
			Filter("foo[0]").DoesNotCompile("indexing expression not supported"),
		That("variable references are not supported yet").
			Filter("$a").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("variable references in RE subfilter are not supported yet").
			Filter("[re $a]").
			DoesNotCompile("re subfilter with primary expression of type Variable not supported"),
		That("variable references in AND subfilter are not supported yet").
			Filter("[and $a]").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("variable references in OR subfilter are not supported yet").
			Filter("[or $a]").
			DoesNotCompile("primary expression of type Variable not supported"),
		That("other subqueries are not supported yet").
			Filter("[other foo bar]").
			DoesNotCompile("head other not supported"),
	)
}

func test(t *testing.T, tests ...testCase) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := filter.Compile(test.filter)
			if errType := getErrorType(err); errType != test.errorType {
				t.Errorf("%q should have %s, but has %s",
					test.filter, test.errorType, errType)
			}
			if err != nil {
				if err.Error() != test.errorMessage {
					t.Errorf("%q should have error message %q, but is %q",
						test.filter, test.errorMessage, err)
				}
				return
			}
			for _, s := range test.matches {
				ok := q.Match(s)
				if !ok {
					t.Errorf("%q should match %q, but doesn't", test.filter, s)
				}
			}
			for _, s := range test.doesntMatch {
				ok := q.Match(s)
				if ok {
					t.Errorf("%q shouldn't match %q, but does", test.filter, s)
				}
			}
		})
	}
}

type testCase struct {
	name         string
	filter       string
	matches      []string
	doesntMatch  []string
	errorType    errorType
	errorMessage string
}

func That(name string) testCase {
	return testCase{name: name}
}

func (t testCase) Filter(q string) testCase {
	t.filter = q
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
	if err == nil {
		return noError
	} else if parse.UnpackErrors(err) != nil {
		return parseError
	} else {
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
