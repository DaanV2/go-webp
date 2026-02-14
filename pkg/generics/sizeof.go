package generics

import (
	"reflect"
)

// SizeOf returns the size of the type T in bytes.
func SizeOf[T any](v T) int {
	return int(reflect.TypeOf(v).Size())
}

func SizeOfFor[T any]() int {
	var v T
	return int(reflect.TypeOf(v).Size())
}

func SizeOfPtr[T any](v *T) int {
	return int(reflect.TypeFor[T]().Size())
}
