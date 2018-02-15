package re

import (
	"strconv"

	"github.com/elves/elvish/eval/vals"
	"github.com/xiaq/persistent/vector"
)

var (
	matchDescriptor    = vals.NewStructDescriptor("text", "start", "end", "groups")
	submatchDescriptor = vals.NewStructDescriptor("text", "start", "end")
)

func newMatch(text string, start, end int, groups vector.Vector) *vals.Struct {
	return vals.NewStruct(matchDescriptor, []interface{}{
		text, strconv.Itoa(start), strconv.Itoa(end), groups,
	})
}

func newSubmatch(text string, start, end int) *vals.Struct {
	return vals.NewStruct(submatchDescriptor, []interface{}{
		string(text),
		string(strconv.Itoa(start)),
		string(strconv.Itoa(end))})
}
