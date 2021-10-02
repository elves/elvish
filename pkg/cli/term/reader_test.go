package term

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestIsReadErrorRecoverable(t *testing.T) {
	tt.Test(t, tt.Fn("IsReadErrorRecoverable", IsReadErrorRecoverable), tt.Table{
		Args(seqError{}).Rets(true),
		Args(ErrStopped).Rets(true),
		Args(errTimeout).Rets(true),

		Args(errors.New("other error")).Rets(false),
	})
}
