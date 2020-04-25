package parse

import (
	"fmt"

	"github.com/elves/elvish/pkg/util"
	"github.com/xiaq/persistent/hash"
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

func (src Source) Kind() string {
	return "map"
}

func (src Source) Hash() uint32 {
	return hash.DJB(
		hash.String(src.Name), hash.String(src.Code), hashBool(src.IsFile))
}

func hashBool(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func (src Source) Equal(other interface{}) bool {
	if src2, ok := other.(Source); ok {
		return src == src2
	}
	return false
}

func (src Source) Repr(int) string {
	return fmt.Sprintf(
		"<src name:%s code:... is-file:$%v>", Quote(src.Name), src.IsFile)
}

func (src Source) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "name":
		return src.Name, true
	case "code":
		return src.Code, true
	case "path":
		// For backward compatibility; remove after 0.14 release.
		if src.IsFile {
			return src.Name, true
		}
		return "", true
	case "is-file":
		return src.IsFile, true
	default:
		return nil, false
	}
}

func (src Source) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, "name", "code", "is-file")
}
