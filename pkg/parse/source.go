package parse

import (
	"fmt"
)

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

// IsStructMap marks that Source is a structmap.
func (src Source) IsStructMap() {}

// Repr returns the representation of Source as if it were a map, except that
// the code field is replaced by "...", since it is typically very large.
func (src Source) Repr(int) string {
	return fmt.Sprintf(
		"[&name=%s &code=<...> &is-file=$%v]", Quote(src.Name), src.IsFile)
}
