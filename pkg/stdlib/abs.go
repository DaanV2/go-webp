package stdlib

import (
	"math"
)

type Numbers interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func Abs[T Numbers](a T) T {
	return T(math.Abs(float64(a)))
}