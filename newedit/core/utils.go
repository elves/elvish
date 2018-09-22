package core

import (
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/newedit/utils"
)

func getMode(m types.Mode) types.Mode {
	if m == nil {
		return utils.BasicMode{}
	}
	return m
}
