package generics

import "unsafe"

func SizeOf[T any]() int {
	var t T
	return int(unsafe.Sizeof(t))
}
