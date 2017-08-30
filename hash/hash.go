// Package hash contains some common hash functions suitable for use in hash
// maps.
package hash

func UInt32(u uint32) uint32 {
	return u
}

func UInt64(u uint64) uint32 {
	return mul33(uint32(u>>32)) + uint32(u&0xffffffff)
}

func String(s string) uint32 {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = mul33(h) + uint32(s[i])
	}
	return h
}

func mul33(u uint32) uint32 {
	return u<<5 + u
}
