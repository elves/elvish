package term

import (
	"testing"

	"src.elv.sh/pkg/sys/ewindows"
	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestConvertEvent(t *testing.T) {
	tt.Test(t, tt.Fn("convertEvent", convertEvent), tt.Table{
		// Only convert KeyEvent
		Args(&ewindows.MouseEvent{}).Rets(nil),
		// Only convert KeyDown events
		Args(&ewindows.KeyEvent{BKeyDown: 0}).Rets(nil),

		Args(&ewindows.KeyEvent{BKeyDown: 1, UChar: [2]byte{'a', 0}}).Rets(K('a')),
	})
}
