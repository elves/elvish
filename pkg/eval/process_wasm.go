package eval

import "syscall"

func putSelfInFg() error { return nil }

func makeSysProcAttr(bg bool) *syscall.SysProcAttr { return nil }
