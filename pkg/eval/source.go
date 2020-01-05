package eval

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
	"github.com/xiaq/persistent/hash"
)

// Source describes a piece of source code.
type Source struct {
	typ  sourceType
	name string
	path string
	code string
}

// NewInteractiveSource returns a Source for a piece of code entered
// interactively.
func NewInteractiveSource(code string) *Source {
	return &Source{interactiveSource, "", "", code}
}

// NewScriptSource returns a Source for a piece of code used as a script.
func NewScriptSource(name, path, code string) *Source {
	return &Source{scriptSource, name, path, code}
}

// NewModuleSource returns a Source for a piece of code used as a module.
func NewModuleSource(name, path, code string) *Source {
	return &Source{moduleSource, name, path, code}
}

// NewInternalGoSource returns a Source for use as a placeholder when calling Elvish
// functions from Go code. It has no associated code.
func NewInternalGoSource(name string) *Source {
	return &Source{internalGoSource, name, name, ""}
}

func (src *Source) describePath() string {
	if src.typ == interactiveSource {
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
	util.Feed(f, "type", "name", "path", "code")
}

// sourceType records the type of a piece of source code.
type sourceType int

const (
	// A special value used for the Frame when calling Elvish functions from Go.
	// This is the only sourceType without associated code.
	internalGoSource sourceType = iota
	// Code entered interactively.
	interactiveSource
	// Code from the main script.
	scriptSource
	// Code from
	moduleSource
)

func (t sourceType) String() string {
	switch t {
	case internalGoSource:
		return "internal"
	case interactiveSource:
		return "interactive"
	case scriptSource:
		return "script"
	case moduleSource:
		return "module"
	default:
		return "bad type " + strconv.Itoa(int(t))
	}
}
