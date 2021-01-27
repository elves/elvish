package re

import (
	"src.elv.sh/pkg/eval/vals"
)

type matchStruct struct {
	Text   string
	Start  int
	End    int
	Groups vals.List
}

func (matchStruct) IsStructMap() {}

type submatchStruct struct {
	Text  string
	Start int
	End   int
}

func (submatchStruct) IsStructMap() {}
