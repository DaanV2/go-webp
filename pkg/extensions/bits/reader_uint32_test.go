package xbits_test

import (
	"io"
	"testing"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func TestBitUint32Reader_ReadBit(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0b101})

	bit, err := r.ReadBit()
	require.NoError(t, err)
	require.True(t, bit)

	bit, err = r.ReadBit()
	require.NoError(t, err)
	require.False(t, bit)

	bit, err = r.ReadBit()
	require.NoError(t, err)
	require.True(t, bit)
}

func TestBitUint32Reader_ReadBits_ZeroCountIsNoop(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0xFFFF_FFFF})

	value, err := r.ReadBits(0)
	require.NoError(t, err)
	require.Equal(t, uint32(0), value)
	require.Equal(t, 32, r.BitsRemaining())
}

func TestBitUint32Reader_ReadBits_CountTooLarge(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0xFFFF_FFFF})

	_, err := r.ReadBits(33)
	require.Error(t, err)
}

func TestBitUint32Reader_ReadBits_CrossesWordBoundary(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0xFFFF_FFFF, 0b1010})

	value, err := r.ReadBits(32)
	require.NoError(t, err)
	require.Equal(t, uint32(0xFFFF_FFFF), value)

	value, err = r.ReadBits(4)
	require.NoError(t, err)
	require.Equal(t, uint32(0b1010), value)
	require.Equal(t, 28, r.BitsRemaining())
}

func TestBitUint32Reader_ReadBits_InsufficientDataReturnsUnexpectedEOF(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0xFFFF_FFFF})

	_, err := r.ReadBits(32)
	require.NoError(t, err)

	_, err = r.ReadBits(1)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestBitUint32Reader_Reset(t *testing.T) {
	r := xbits.NewBitUint32Reader([]uint32{0xFFFF_FFFF})

	_, err := r.ReadBits(32)
	require.NoError(t, err)
	require.Equal(t, 0, r.BitsRemaining())

	r.Reset([]uint32{0xFFFF_FFFF, 0xFFFF_FFFF})
	require.Equal(t, 64, r.BitsRemaining())
}

// Fuzz_BitUint32RoundTrip writes randomized (value, count) pairs through a
// BitUint32Writer and verifies a BitUint32Reader reads back the exact same
// masked values in the same order.
func Fuzz_BitUint32RoundTrip(f *testing.F) {
	f.Add([]byte{0xFF, 5, 0x00, 3})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, ops []byte) {
		writer := xbits.NewBitUint32Writer()

		runBitsRoundTripFuzz(t, ops, writer.WriteBits, func() func(uint8) (uint32, error) {
			words, _ := writer.Uint32s()

			return xbits.NewBitUint32Reader(words).ReadBits
		})
	})
}
