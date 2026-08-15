package xbits_test

import (
	"testing"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func TestBitUint32Writer_WriteBit(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	require.NoError(t, w.WriteBit(true))
	require.NoError(t, w.WriteBit(false))
	require.NoError(t, w.WriteBit(true))

	words, bitPos := w.Uint32s()
	require.Equal(t, []uint32{0b101}, words)
	require.Equal(t, 3, bitPos)
	require.Equal(t, 3, w.BitLength())
}

func TestBitUint32Writer_WriteBits_ZeroCountIsNoop(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	require.NoError(t, w.WriteBits(0xFFFF_FFFF, 0))

	words, bitPos := w.Uint32s()
	require.Empty(t, words)
	require.Equal(t, 0, bitPos)
}

func TestBitUint32Writer_WriteBits_CountTooLarge(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	err := w.WriteBits(0, 33)
	require.Error(t, err)
}

func TestBitUint32Writer_WriteBits_CrossesWordBoundary(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	require.NoError(t, w.WriteBits(0xFFFF_FFFF, 32))
	require.NoError(t, w.WriteBits(0b1010, 4))

	words, bitPos := w.Uint32s()
	require.Equal(t, []uint32{0xFFFF_FFFF, 0b1010}, words)
	require.Equal(t, 4, bitPos)
	require.Equal(t, 36, w.BitLength())
}

func TestBitUint32Writer_Bytes_IsLittleEndian(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	require.NoError(t, w.WriteBits(0x0403_0201, 32))

	data, bitPos := w.Bytes()
	require.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, data)
	require.Equal(t, 0, bitPos)
}

func TestBitUint32Writer_Reset(t *testing.T) {
	w := xbits.NewBitUint32Writer()

	require.NoError(t, w.WriteBits(0xFFFF_FFFF, 32))
	w.Reset()

	words, bitPos := w.Uint32s()
	require.Empty(t, words)
	require.Equal(t, 0, bitPos)
	require.Equal(t, 0, w.BitLength())
}
