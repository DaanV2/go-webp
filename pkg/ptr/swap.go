package ptr

func Swap[T any](a, b *T) {
	var tmp = *a
	*a = *b
	*b = tmp
}
