package stdlib

import (
	"cmp"
	"slices"
)

// NOTE: needs to check with C's memcmp, which returns 0 if equal, <0 if ptr1<ptr2, >0 if ptr1>ptr2
//go:fix inline
func MemCmp[S ~[]E, E cmp.Ordered](ptr1, ptr2 S, num int) int {
	return slices.Compare(ptr1[:num], ptr2[:num])
}
