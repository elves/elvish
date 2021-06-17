package eval

import "syscall"

// Error number 232 is what Windows returns when trying to write on a pipe who
// reader has gone. The syscall package defines an EPIPE on Windows, but that's
// not what Windows API actually returns.
//
// https://docs.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-
var epipe = syscall.Errno(232)
