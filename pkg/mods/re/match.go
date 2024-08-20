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

type submatchStruct struct {
	Text  string
	Start int
	End   int
}
