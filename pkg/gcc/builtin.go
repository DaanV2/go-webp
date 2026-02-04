package gcc

import "math/bits"

//Returns the number of leading 0-bits in x, starting at the most significant bit position. If x is 0, the result is undefined.
func Builtin_CLZ(x uint32) int {
	return bits.LeadingZeros32(x)
}

//Returns the number of trailing 0-bits in x, starting at the least significant bit position. If x is 0, the result is undefined.
func Builtin_CTZ(x uint32) int {
	return bits.TrailingZeros32(x)
}

//Returns x with the order of the bytes reversed; for example, 0xabcd becomes 0xcdab. Byte here always means exactly 8 bits.
func Builtin_bswap16(x uint16) uint16 {
	return (x>>8)&0xff | (x&0xff)<<8
}

func Builtin_bswap32(x uint32) uint32 {
	return (x>>24)&0xff | (x>>8)&0xff00 | (x<<8)&0xff0000 | (x<<24)&0xff000000
}
func Builtin_bswap64(x uint64) uint64 {
	return (x>>56)&0xff |
		(x>>40)&0xff00 |
		(x>>24)&0xff0000 |
		(x>>8)&0xff000000 |
		(x<<8)&0xff00000000 |
		(x<<24)&0xff0000000000 |
		(x<<40)&0xff000000000000 |
		(x<<56)&0xff00000000000000
}
