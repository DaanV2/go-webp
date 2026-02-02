package stdlib

// MemCpy is a conversion of C's memcpy function for Go slices.
// Deprecated: Use [MemCpy2] instead.
func MemCpy[T any](dst, src []T, size int) {
	MemCpy2(dst[:size], src[:size])
}

// MemCpy2 is a conversion of C's memcpy function for Go slices.
func MemCpy2[T any](dst, src []T) {
	i := max(len(dst), len(src))

	copy(dst[:i], src[:i])
}