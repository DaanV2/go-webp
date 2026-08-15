package xbits_test

import (
	"bytes"
	"errors"
	"testing"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

type errWriter struct {
	err error
}

func (w *errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestBitStreamWriter_FlushesFullByteImmediately(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.NoError(t, w.WriteBits(0b1010_1010, 8))
	require.Equal(t, []byte{0b1010_1010}, buf.Bytes())
	require.Equal(t, 0, w.PendingBits())
}

func TestBitStreamWriter_BuffersPartialByte(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.NoError(t, w.WriteBits(0b101, 3))
	require.Empty(t, buf.Bytes())
	require.Equal(t, 3, w.PendingBits())
}

func TestBitStreamWriter_WriteBits_CrossesByteBoundary(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.NoError(t, w.WriteBits(0xFFF, 12))
	require.Equal(t, []byte{0xFF}, buf.Bytes())
	require.Equal(t, 4, w.PendingBits())
}

func TestBitStreamWriter_Flush_PadsPartialByte(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.NoError(t, w.WriteBits(0b101, 3))
	require.NoError(t, w.Flush())

	require.Equal(t, []byte{0b0000_0101}, buf.Bytes())
	require.Equal(t, 0, w.PendingBits())
}

func TestBitStreamWriter_Flush_NoopWhenNothingBuffered(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.NoError(t, w.Flush())
	require.Empty(t, buf.Bytes())
}

func TestBitStreamWriter_WriteBits_CountTooLarge(t *testing.T) {
	var buf bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf)

	require.Error(t, w.WriteBits(0, 33))
}

func TestBitStreamWriter_PropagatesUnderlyingWriteError(t *testing.T) {
	wantErr := errors.New("boom")
	w := xbits.NewBitStreamWriter(&errWriter{err: wantErr})

	err := w.WriteBits(0xFF, 8)
	require.ErrorIs(t, err, wantErr)
}

func TestBitStreamWriter_Reset(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	w := xbits.NewBitStreamWriter(&buf1)

	require.NoError(t, w.WriteBits(0b101, 3))
	w.Reset(&buf2)
	require.Equal(t, 0, w.PendingBits())

	require.NoError(t, w.WriteBits(0xFF, 8))
	require.Equal(t, []byte{0xFF}, buf2.Bytes())
	require.Empty(t, buf1.Bytes())
}

// TestBitStreamWriter_MatchesBitUint8Writer verifies both writers pack an
// identical bit sequence into byte-identical output.
func TestBitStreamWriter_MatchesBitUint8Writer(t *testing.T) {
	ops := []struct {
		value uint32
		count uint8
	}{
		{value: 0b1011, count: 4},
		{value: 0xFFFF, count: 16},
		{value: 0b1, count: 1},
		{value: 0x7F, count: 7},
	}

	bufWriter := xbits.NewBitUint8Writer()

	var buf bytes.Buffer
	streamWriter := xbits.NewBitStreamWriter(&buf)

	for _, o := range ops {
		require.NoError(t, bufWriter.WriteBits(o.value, o.count))
		require.NoError(t, streamWriter.WriteBits(o.value, o.count))
	}
	require.NoError(t, streamWriter.Flush())

	wantData, _ := bufWriter.Bytes()
	require.Equal(t, wantData, buf.Bytes())
}
