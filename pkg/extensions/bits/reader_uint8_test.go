package xbits_test

import (
	"io"
	"testing"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func TestBitUint8Reader_ReadBit(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0b0000_0101})

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

func TestBitUint8Reader_ReadBits_ZeroCountIsNoop(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF})

	value, err := r.ReadBits(0)
	require.NoError(t, err)
	require.Equal(t, uint32(0), value)
	require.Equal(t, 8, r.BitsRemaining())
}

func TestBitUint8Reader_ReadBits_CountTooLarge(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF})

	_, err := r.ReadBits(33)
	require.Error(t, err)
}

func TestBitUint8Reader_ReadBits_CrossesByteBoundary(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF, 0x0F})

	value, err := r.ReadBits(12)
	require.NoError(t, err)
	require.Equal(t, uint32(0xFFF), value)
	require.Equal(t, 4, r.BitsRemaining())
}

func TestBitUint8Reader_ReadBits_InsufficientDataReturnsUnexpectedEOF(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF})

	_, err := r.ReadBits(9)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestBitUint8Reader_BitsRemaining(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF, 0xFF})
	require.Equal(t, 16, r.BitLength())
	require.Equal(t, 16, r.BitsRemaining())

	_, err := r.ReadBits(5)
	require.NoError(t, err)
	require.Equal(t, 11, r.BitsRemaining())
}

func TestBitUint8Reader_Reset(t *testing.T) {
	r := xbits.NewBitUint8Reader([]uint8{0xFF})

	_, err := r.ReadBits(8)
	require.NoError(t, err)
	require.Equal(t, 0, r.BitsRemaining())

	r.Reset([]uint8{0xFF, 0xFF})
	require.Equal(t, 16, r.BitsRemaining())
}

// Fuzz_BitUint8RoundTrip writes randomized (value, count) pairs through a
// BitUint8Writer and verifies a BitUint8Reader reads back the exact same
// masked values in the same order.
func Fuzz_BitUint8RoundTrip(f *testing.F) {
	f.Add([]byte{0xFF, 5, 0x00, 3})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, ops []byte) {
		writer := xbits.NewBitUint8Writer()

		runBitsRoundTripFuzz(t, ops, writer.WriteBits, func() func(uint8) (uint32, error) {
			data, _ := writer.Bytes()

			return xbits.NewBitUint8Reader(data).ReadBits
		})
	})
}
