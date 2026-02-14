package generics_test

import (
	"testing"

	"github.com/daanv2/go-webp/pkg/generics"
	"github.com/stretchr/testify/require"
)

func TestSizeOf(t *testing.T) {
	require.Equal(t, 4, generics.SizeOf[int32]())
	require.Equal(t, 8, generics.SizeOf[int64]())
	require.Equal(t, 4, generics.SizeOf[float32]())
	require.Equal(t, 8, generics.SizeOf[float64]())
}
