package assert

import "runtime/debug"

func Assert(condition bool) {
	if !condition {
		s := debug.Stack()

		panic("assertion failed:\n" + string(s))
	}
}