package sys

import "runtime"

func DumpStack() {
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	print(string(buf))
}
