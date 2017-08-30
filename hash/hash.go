// Package hash contains some common hash functions suitable for use in hash
// maps.
package hash

import "unsafe"

const DJBInit = 5381

func DJBCombine(acc, h uint32) uint32 {
	return mul33(acc) + h
}

func UInt32(u uint32) uint32 {
	return u
}

func UInt64(u uint64) uint32 {
	return mul33(uint32(u>>32)) + uint32(u&0xffffffff)
}

func Pointer(p unsafe.Pointer) uint32 {
	if unsafe.Sizeof(p) == 4 {
		return UInt32(uint32(uintptr(p)))
	} else {
		return UInt64(uint64(uintptr(p)))
	}
	// NOTE: We don't care about 128-bit archs yet.
}

func String(s string) uint32 {
	h := uint32(DJBInit)
	for i := 0; i < len(s); i++ {
		h = DJBCombine(h, uint32(s[i]))
	}
	return h
}

func mul33(u uint32) uint32 {
	return u<<5 + u
}
