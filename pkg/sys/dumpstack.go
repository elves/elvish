package sys

import "runtime"

const dumpStackBufSizeInit = 8192

func DumpStack() string {
	buf := make([]byte, dumpStackBufSizeInit)
	for {
		n := runtime.Stack(buf, true)
		if n < cap(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, cap(buf)*2)
	}
}
