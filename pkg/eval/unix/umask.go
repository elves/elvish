// +build !windows,!plan9,!js

package unix

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/elves/elvish/pkg/eval/vars"
)

//elvdoc:var umask
//
// The file mode creation mask. Its value is a string in Elvish octal
// representation; e.g. 0o027. This makes it possible to use it in any context
// that expects a `$number`.
//
// When assigning a new value a string is implicitly treated as an octal
// number. If that fails the usual rules for interpreting
// [numbers](./language.html#number) are used. The following arer equivalent:
// `unix:umask = 027` and `unix:umask = 0o27`. You can also assign to it a
// `float64` data type that has no fractional component.
// The assigned value must be within the range [0 ... 0o777], otherwise the
// assignment will throw an exception.
//
// You can do a temporary assignment to affect a single command; e.g.
// `umask=077 touch a_file`. After the command completes the old umask will be
// restored. **Warning**: Since the umask applies to the entire process, not
// individual threads, changing it temporarily in this manner is dangerous if
// you are doing anything in parallel. Such as via the
// [`peach`](ref/builtin.html#peach) command.

// UmaskVariable is a variable whose value always reflects the current file
// creation permission mask. Setting it changes the current file creation
// permission mask for the process (not an individual thread).
type UmaskVariable struct{}

var _ vars.Var = UmaskVariable{}
var umaskMutex sync.Mutex

// Get returns the current file creation umask as a string.
func (UmaskVariable) Get() interface{} {
	// Note: The seemingly redundant syscall is because the unix.Umask() API
	// doesn't allow querying the current value without changing it. So ensure
	// we reinstate the curent value.
	umaskMutex.Lock()
	defer umaskMutex.Unlock()
	umask := unix.Umask(0)
	unix.Umask(umask)
	return fmt.Sprintf("0o%03o", umask)
}

// Set changes the current file creation umask. It can be called with a string
// (the usual case) or a float64.
func (UmaskVariable) Set(v interface{}) error {
	var umask int

	umaskMutex.Lock()
	defer umaskMutex.Unlock()

	switch v := v.(type) {
	case string:
		i, err := strconv.ParseInt(v, 8, 0)
		if err != nil {
			i, err = strconv.ParseInt(v, 0, 0)
			if err != nil {
				return errors.New("umask value not a valid number")
			}
		}
		umask = int(i)
	case float64:
		intPart, fracPart := math.Modf(v)
		if fracPart != 0 {
			return errors.New("umask value must be an integer")
		}
		umask = int(intPart)
	default:
		return errors.New("umask value must be a string or float64")
	}

	if umask < 0 || umask > 0o777 {
		return errors.New("umask value outside the range [0..0o777]")
	}

	unix.Umask(umask)
	return nil
}
