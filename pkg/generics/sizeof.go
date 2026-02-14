package generics

import (
	"reflect"
)

// SizeOf returns the size of the type T in bytes.
func SizeOf[T any]() int {
	var t T
	return int(reflect.TypeOf(t).Size())
}
