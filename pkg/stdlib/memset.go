package stdlib

// Memset is a conversion of C's memset function for Go slices.
// Deprecated: Use built-in memory management or [stdlib.Memset2] instead.
//go:fix inline
func Memset[T any](data []T, value T, num int) {
	Memset2(data[:num], value)
}

// Memset2 is a conversion of C's memset function for Go slices.
func Memset2[T any](data []T, value T) {
	for i := range data {
		data[i] = value
	}
}