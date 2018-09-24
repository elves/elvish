package core

import (
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/newedit/utils"
)

// Returns the first non-nil value. If all are nil, return utils.BasicMode{}
func getMode(modes ...types.Mode) types.Mode {
	for _, mode := range modes {
		if mode != nil {
			return mode
		}
	}
	return utils.BasicMode{}
}
