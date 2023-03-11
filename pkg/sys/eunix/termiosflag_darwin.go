package eunix

// Only Darwin uses 64-bit flags in Termios on 64-bit architectures. See:
// https://cs.opensource.google/search?q=%5BIOCL%5Dflag.*uint64&sq=&ss=go%2Fx%2Fsys
//
// Darwin uses 32-bit flags on 32-bit architectures, but Go no longer supports
// them since Go 1.15.

type termiosFlag = uint64
