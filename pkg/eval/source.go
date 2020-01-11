package eval

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
	"github.com/xiaq/persistent/hash"
)

// Source describes a piece of source code.
type Source struct {
	Type SourceType
	Name string
	Root bool
	Code string
}

// NewInteractiveSource returns a Source for a piece of code entered
// interactively.
func NewInteractiveSource(code string) *Source {
	return &Source{InteractiveSource, "[tty]", true, code}
}

// NewScriptSource returns a Source for a piece of code used as a script.
func NewScriptSource(path, code string) *Source {
	return &Source{FileSource, path, true, code}
}

// NewModuleSource returns a Source for a piece of code used as a module.
func NewModuleSource(path, code string) *Source {
	return &Source{FileSource, path, false, code}
}

// NewInternalGoSource returns a Source for use as a placeholder when calling Elvish
// functions from Go code. It has no associated code.
func NewInternalGoSource(name string) *Source {
	return &Source{InternalGoSource, name, true, ""}
}

func NewInternalElvishSource(root bool, name, code string) *Source {
	return &Source{InternalElvishSource, name, root, code}
}

func (src *Source) Kind() string {
	return "map"
}

func (src *Source) Hash() uint32 {
	var root uint32
	if src.Root {
		root = 1
	}
	return hash.DJB(uint32(src.Type),
		hash.String(src.Name), root, hash.String(src.Code))
}

func (src *Source) Equal(other interface{}) bool {
	if src2, ok := other.(*Source); ok {
		return *src == *src2
	}
	return false
}

func (src *Source) Repr(int) string {
	return fmt.Sprintf(
		"<src type:%s name:%s root:$%v code:...>",
		src.Type, parse.Quote(src.Name), src.Root)
}

func (src *Source) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "type":
		return src.Type.String(), true
	case "name":
		return src.Name, true
	case "root":
		return src.Root, true
	case "code":
		return src.Code, true
	default:
		return nil, false
	}
}

func (src *Source) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, "type", "name", "root", "code")
}

// SourceType records the type of a piece of source code.
type SourceType int

const (
	InvalidSource SourceType = iota
	// A special value used for the Frame when calling Elvish functions from Go.
	// This is the only sourceType without associated code.
	InternalGoSource
	// Code from an internal buffer.
	InternalElvishSource
	// Code entered interactively.
	InteractiveSource
	// Code from a file.
	FileSource
)

func (t SourceType) String() string {
	switch t {
	case InternalGoSource:
		return "internal"
	case InteractiveSource:
		return "interactive"
	case FileSource:
		return "file"
	default:
		return "bad type " + strconv.Itoa(int(t))
	}
}
