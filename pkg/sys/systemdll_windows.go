package sys

import "golang.org/x/sys/windows"

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")
