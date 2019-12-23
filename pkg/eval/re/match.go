package re

import (
	"github.com/elves/elvish/pkg/eval/vals"
)

type matchStruct struct {
	Text   string    `json:"text"`
	Start  int       `json:"start"`
	End    int       `json:"end"`
	Groups vals.List `json:"groups"`
}

func (matchStruct) IsStructMap(vals.StructMapMarker) {}

type submatchStruct struct {
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

func (submatchStruct) IsStructMap(vals.StructMapMarker) {}
