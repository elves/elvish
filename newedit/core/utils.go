package core

import (
	"github.com/elves/elvish/newedit/editutil"
	"github.com/elves/elvish/newedit/types"
)

// Returns the first non-nil value. If all are nil, return utils.BasicMode{}
func getMode(modes ...types.Mode) types.Mode {
	for _, mode := range modes {
		if mode != nil {
			return mode
		}
	}
	return editutil.BasicMode{}
}
