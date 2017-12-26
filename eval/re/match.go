package re

import (
	"strconv"

	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/vector"
)

var (
	matchDescriptor    = eval.NewStructDescriptor("text", "start", "end", "groups")
	submatchDescriptor = eval.NewStructDescriptor("text", "start", "end")
)

func newMatch(text string, start, end int, groups vector.Vector) *eval.Struct {
	return &eval.Struct{matchDescriptor, []eval.Value{
		eval.String(text),
		eval.String(strconv.Itoa(start)),
		eval.String(strconv.Itoa(end)),
		eval.NewListFromVector(groups),
	}}
}

func newSubmatch(text string, start, end int) *eval.Struct {
	return &eval.Struct{submatchDescriptor, []eval.Value{
		eval.String(text),
		eval.String(strconv.Itoa(start)),
		eval.String(strconv.Itoa(end))}}
}
