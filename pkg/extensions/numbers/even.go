package numbers

import "golang.org/x/exp/constraints"

func IsEven[T constraints.Integer](v T) bool {
	return v%2 == 0
}

func IsOdd[T constraints.Integer](v T) bool {
	return v%2 == 1
}
