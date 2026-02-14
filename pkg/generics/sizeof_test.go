package generics_test

import (
	"testing"

	"github.com/daanv2/go-webp/pkg/generics"
	"github.com/stretchr/testify/require"
)

func TestSizeOf(t *testing.T) {
	testSize[int32](t, 4)
	testSize[int64](t, 8)
	testSize[float32](t, 4)
	testSize[float64](t, 8)
}

func testSize[T any](t *testing.T, expected int) {
	t.Helper()

	var zero T
	require.Equal(t, expected, generics.SizeOf(zero))
}
