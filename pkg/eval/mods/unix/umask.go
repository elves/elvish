// +build !windows,!plan9,!js

package unix

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

const (
	validUmaskMsg = "integer in the range [0..0o777]"
)

//elvdoc:var umask
//
// The file mode creation mask. Its value is a string in Elvish octal
// representation; e.g. 0o027. This makes it possible to use it in any context
// that expects a `$number`.
//
// When assigning a new value a string is implicitly treated as an octal
// number. If that fails the usual rules for interpreting
// [numbers](./language.html#number) are used. The following are equivalent:
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

// Guard against concurrent fetch and assignment of $unix:umask. This assumes
// no other part of the elvish code base will call unix.Umask() as it only
// protects against races involving the aforementioned Elvish var.
var umaskMutex sync.Mutex

// Get returns the current file creation umask as a string.
func (UmaskVariable) Get() interface{} {
	// Note: The seemingly redundant syscall is because the unix.Umask() API
	// doesn't allow querying the current value without changing it. So ensure
	// we reinstate the current value.
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

	switch v := v.(type) {
	case string:
		i, err := strconv.ParseInt(v, 8, 0)
		if err != nil {
			i, err = strconv.ParseInt(v, 0, 0)
			if err != nil {
				return errs.BadValue{
					What: "umask", Valid: validUmaskMsg, Actual: vals.ToString(v)}
			}
		}
		umask = int(i)
	case float64:
		intPart, fracPart := math.Modf(v)
		if fracPart != 0 {
			return errs.BadValue{
				What: "umask", Valid: validUmaskMsg, Actual: vals.ToString(v)}
		}
		umask = int(intPart)
	default:
		return errs.BadValue{
			What: "umask", Valid: validUmaskMsg, Actual: vals.ToString(v)}
	}

	if umask < 0 || umask > 0o777 {
		// TODO: Switch to `%O` when Go 1.15 is the minimum acceptable version.
		// Until then the formatting of negative numbers will be weird.
		return errs.OutOfRange{
			What: "umask", ValidLow: "0", ValidHigh: "0o777",
			Actual: fmt.Sprintf("0o%o", umask)}
	}

	umaskMutex.Lock()
	defer umaskMutex.Unlock()
	unix.Umask(umask)
	return nil
}
