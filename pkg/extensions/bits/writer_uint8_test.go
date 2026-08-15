package xbits_test

import (
	"testing"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func TestBitUint8Writer_WriteBit(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	require.NoError(t, w.WriteBit(true))
	require.NoError(t, w.WriteBit(false))
	require.NoError(t, w.WriteBit(true))

	data, bitPos := w.Bytes()
	require.Equal(t, []byte{0b0000_0101}, data)
	require.Equal(t, 3, bitPos)
	require.Equal(t, 3, w.BitLength())
}

func TestBitUint8Writer_WriteBits_ZeroCountIsNoop(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	require.NoError(t, w.WriteBits(0xFF, 0))

	data, bitPos := w.Bytes()
	require.Empty(t, data)
	require.Equal(t, 0, bitPos)
}

func TestBitUint8Writer_WriteBits_CountTooLarge(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	err := w.WriteBits(0, 33)
	require.Error(t, err)
}

func TestBitUint8Writer_WriteBits_CrossesByteBoundary(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	// 12 bits: fills the first byte and spills 4 bits into the second.
	require.NoError(t, w.WriteBits(0xFFF, 12))

	data, bitPos := w.Bytes()
	require.Equal(t, []byte{0xFF, 0x0F}, data)
	require.Equal(t, 4, bitPos)
	require.Equal(t, 12, w.BitLength())
}

func TestBitUint8Writer_WriteBits_OnlyMasksRelevantBits(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	require.NoError(t, w.WriteBits(0xFFFF_FFFF, 4))

	data, bitPos := w.Bytes()
	require.Equal(t, []byte{0x0F}, data)
	require.Equal(t, 4, bitPos)
}

func TestBitUint8Writer_Reset(t *testing.T) {
	w := xbits.NewBitUint8Writer()

	require.NoError(t, w.WriteBits(0xFF, 8))
	w.Reset()

	data, bitPos := w.Bytes()
	require.Empty(t, data)
	require.Equal(t, 0, bitPos)
	require.Equal(t, 0, w.BitLength())
}
