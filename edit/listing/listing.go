// Package listing contains definitions useful for listing modes. It does not
// contain the actual implementation, which is in the edit package.
package listing

import (
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
)

type Provider interface {
	Len() int
	Show(i int) (string, ui.Styled)
	Filter(filter string) int
	Accept(i int, ed eddefs.Editor)
	ModeTitle(int) string
}
