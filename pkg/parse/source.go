package parse

// TODO(xiaq): Move this into the diag package after implementing phantom types.

// Source describes a piece of source code.
type Source struct {
	Name   string
	Code   string
	IsFile bool
}

// SourceForTest returns a Source used for testing.
func SourceForTest(code string) Source {
	return Source{Name: "[test]", Code: code}
}
