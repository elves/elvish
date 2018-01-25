package re

import (
	"strconv"

	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/vector"
)

var (
	matchDescriptor    = types.NewStructDescriptor("text", "start", "end", "groups")
	submatchDescriptor = types.NewStructDescriptor("text", "start", "end")
)

func newMatch(text string, start, end int, groups vector.Vector) *types.Struct {
	return types.NewStruct(matchDescriptor, []types.Value{
		string(text),
		string(strconv.Itoa(start)),
		string(strconv.Itoa(end)),
		types.NewList(groups),
	})
}

func newSubmatch(text string, start, end int) *types.Struct {
	return types.NewStruct(submatchDescriptor, []types.Value{
		string(text),
		string(strconv.Itoa(start)),
		string(strconv.Itoa(end))})
}
