package eval

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

// Source describes a piece of source code.
type Source struct {
	typ  SrcType
	name string
	path string
	code string
}

// NewInteractiveSource returns a Source for a piece of code entered
// interactively.
func NewInteractiveSource(code string) *Source {
	return &Source{SrcInteractive, "", "", code}
}

// NewScriptSource returns a Source for a piece of code used as a script.
func NewScriptSource(name, path, code string) *Source {
	return &Source{SrcScript, name, path, code}
}

// NewModuleSource returns a Source for a piece of code used as a module.
func NewModuleSource(name, path, code string) *Source {
	return &Source{SrcModule, name, path, code}
}

func NewInternalSource(name string) *Source {
	return &Source{SrcInternal, name, name, ""}
}

func (src *Source) describePath() string {
	if src.typ == SrcInteractive {
		return "[tty]"
	}
	return src.path
}

var (
	_ vals.Indexer = (*Source)(nil)
)

func (src *Source) Kind() string {
	return "map"
}

func (src *Source) Hash() uint32 {
	return hash.DJB(uint32(src.typ),
		hash.String(src.name), hash.String(src.path), hash.String(src.code))
}

func (src *Source) Equal(other interface{}) bool {
	if src2, ok := other.(*Source); ok {
		return *src == *src2
	}
	return false
}

func (src *Source) Repr(int) string {
	return fmt.Sprintf("<src type:%s name:%s path:%s code:...>",
		src.typ, parse.Quote(src.name), parse.Quote(src.path))
}

func (src *Source) Index(k interface{}) (interface{}, bool) {
	ret := ""
	switch k {
	case "type":
		ret = src.typ.String()
	case "name":
		ret = src.name
	case "path":
		ret = src.path
	case "code":
		ret = src.code
	default:
		return nil, false
	}
	return ret, true
}

func (src *Source) IterateKeys(f func(interface{}) bool) {
	vals.Feed(f, "type", "name", "path", "code")
}

// SrcType records the type of a piece of source code.
type SrcType int

const (
	// SrcInternal is a special SrcType for internal operations.
	SrcInternal SrcType = iota
	// SrcInteractive is the type of source code entered interactively.
	SrcInteractive
	// SrcScript is the type of source code used as a script.
	SrcScript
	// SrcModule is the type of source code used as a module.
	SrcModule
)

func (t SrcType) String() string {
	switch t {
	case SrcInternal:
		return "internal"
	case SrcInteractive:
		return "interactive"
	case SrcScript:
		return "script"
	case SrcModule:
		return "module"
	default:
		return "bad type " + strconv.Itoa(int(t))
	}
}
