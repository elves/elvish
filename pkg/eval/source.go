package eval

import (
	"fmt"

	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
	"github.com/xiaq/persistent/hash"
)

// Source describes a piece of source code.
type Source struct {
	Name string
	Code string

	IsFile bool
}

// NewInteractiveSource returns a Source for a piece of code entered
// interactively.
func NewInteractiveSource(name, code string) *Source {
	return &Source{Name: "[tty]", Code: code}
}

// NewScriptSource returns a Source for a piece of code used as a script.
func NewScriptSource(path, code string) *Source {
	return &Source{Name: path, Code: code, IsFile: true}
}

// NewModuleSource returns a Source for a piece of code used as a module.
func NewModuleSource(path, code string) *Source {
	return &Source{Name: path, Code: code, IsFile: true}
}

// NewInternalGoSource returns a Source for use as a placeholder when calling Elvish
// functions from Go code. It has no associated code.
func NewInternalGoSource(name string) *Source {
	return &Source{Name: name}
}

func NewInternalElvishSource(root bool, name, code string) *Source {
	return &Source{Name: name, Code: code}
}

func (src *Source) Kind() string {
	return "map"
}

func (src *Source) Hash() uint32 {
	return hash.DJB(
		hash.String(src.Name), hash.String(src.Code), hashBool(src.IsFile))
}

func hashBool(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func (src *Source) Equal(other interface{}) bool {
	if src2, ok := other.(*Source); ok {
		return *src == *src2
	}
	return false
}

func (src *Source) Repr(int) string {
	return fmt.Sprintf(
		"<src name:%s code:... is-file:$%v>", parse.Quote(src.Name), src.IsFile)
}

func (src *Source) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "name":
		return src.Name, true
	case "code":
		return src.Code, true
	case "is-file":
		return src.IsFile, true
	default:
		return nil, false
	}
}

func (src *Source) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, "name", "code", "is-file")
}
