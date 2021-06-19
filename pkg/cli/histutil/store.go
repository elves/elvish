package histutil

import (
	"errors"

	"src.elv.sh/pkg/store/storedefs"
)

// Store is an abstract interface for history store.
type Store interface {
	// AddCmd adds a new command history entry and returns its sequence number.
	// Depending on the implementation, the Store might respect cmd.Seq and
	// return it as is, or allocate another sequence number.
	AddCmd(cmd storedefs.Cmd) (int, error)
	// AllCmds returns all commands kept in the store.
	AllCmds() ([]storedefs.Cmd, error)
	// Cursor returns a cursor that iterating through commands with the given
	// prefix. The cursor is initially placed just after the last command in the
	// store.
	Cursor(prefix string) Cursor
}

// Cursor is used to navigate a Store.
type Cursor interface {
	// Prev moves the cursor to the previous command.
	Prev()
	// Next moves the cursor to the next command.
	Next()
	// Get returns the command the cursor is currently at, or any error if the
	// cursor is in an invalid state. If the cursor is "over the edge", the
	// error is ErrEndOfHistory.
	Get() (storedefs.Cmd, error)
}

// ErrEndOfHistory is returned by Cursor.Get if the cursor is currently over the
// edge.
var ErrEndOfHistory = errors.New("end of history")
