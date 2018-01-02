package re

import (
	"strconv"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/vector"
)

var (
	matchDescriptor    = eval.NewStructDescriptor("text", "start", "end", "groups")
	submatchDescriptor = eval.NewStructDescriptor("text", "start", "end")
)

func newMatch(text string, start, end int, groups vector.Vector) *eval.Struct {
	return &eval.Struct{matchDescriptor, []types.Value{
		types.String(text),
		types.String(strconv.Itoa(start)),
		types.String(strconv.Itoa(end)),
		types.NewList(groups),
	}}
}

func newSubmatch(text string, start, end int) *eval.Struct {
	return &eval.Struct{submatchDescriptor, []types.Value{
		types.String(text),
		types.String(strconv.Itoa(start)),
		types.String(strconv.Itoa(end))}}
}
