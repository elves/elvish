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
		eval.String(text),
		eval.String(strconv.Itoa(start)),
		eval.String(strconv.Itoa(end)),
		types.NewList(groups),
	}}
}

func newSubmatch(text string, start, end int) *eval.Struct {
	return &eval.Struct{submatchDescriptor, []types.Value{
		eval.String(text),
		eval.String(strconv.Itoa(start)),
		eval.String(strconv.Itoa(end))}}
}
