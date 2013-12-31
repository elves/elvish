package edit

// XXX
// Taken from github.com/nsf/godit, not a genuine wcwidth implementation.
// wcwidth of glibc is broken, meaning we cannot just use cgo to wrap the
// native wcwidth. Possible reference implementation:
// http://www.cl.cam.ac.uk/~mgk25/ucs/wcwidth.c
//
// Original comment in godit source:
// somewhat close to what wcwidth does, except rune_width doesn't return 0 or
// -1, it's always 1 or 2
func wcwidth(r rune) int {
	if r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}

func wcwidths(s string) (w int) {
	for _, r := range s {
		w += wcwidth(r)
	}
	return
}
