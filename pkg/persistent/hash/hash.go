// Package hash contains some common hash functions suitable for use in hash
// maps.
package hash

import "unsafe"

const DJBInit uint32 = 5381

func DJBCombine(acc, h uint32) uint32 {
	return mul33(acc) + h
}

func DJB(hs ...uint32) uint32 {
	acc := DJBInit
	for _, h := range hs {
		acc = DJBCombine(acc, h)
	}
	return acc
}

func UInt32(u uint32) uint32 {
	return u
}

func UInt64(u uint64) uint32 {
	return mul33(uint32(u>>32)) + uint32(u&0xffffffff)
}

func Pointer(p unsafe.Pointer) uint32 {
	switch unsafe.Sizeof(p) {
	case 4:
		return UInt32(uint32(uintptr(p)))
	case 8:
		return UInt64(uint64(uintptr(p)))
	default:
		panic("unhandled pointer size")
	}
}

func UIntPtr(p uintptr) uint32 {
	switch unsafe.Sizeof(p) {
	case 4:
		return UInt32(uint32(p))
	case 8:
		return UInt64(uint64(p))
	default:
		panic("unhandled pointer size")
	}
}

func String(s string) uint32 {
	h := DJBInit
	for i := 0; i < len(s); i++ {
		h = DJBCombine(h, uint32(s[i]))
	}
	return h
}

func mul33(u uint32) uint32 {
	return u<<5 + u
}
