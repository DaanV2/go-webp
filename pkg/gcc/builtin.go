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