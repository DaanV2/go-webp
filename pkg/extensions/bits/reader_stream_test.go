package xbits_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func TestBitStreamReader_ReadBit(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0b0000_0101}))

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

func TestBitStreamReader_ReadBits_ZeroCountIsNoop(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF}))

	value, err := r.ReadBits(0)
	require.NoError(t, err)
	require.Equal(t, uint32(0), value)
}

func TestBitStreamReader_ReadBits_CountTooLarge(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF}))

	_, err := r.ReadBits(33)
	require.Error(t, err)
}

func TestBitStreamReader_ReadBits_CrossesByteBoundary(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF, 0x0F}))

	value, err := r.ReadBits(12)
	require.NoError(t, err)
	require.Equal(t, uint32(0xFFF), value)
}

func TestBitStreamReader_ReadBits_EOFOnEmptyStream(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader(nil))

	_, err := r.ReadBits(1)
	require.ErrorIs(t, err, io.EOF)
}

func TestBitStreamReader_ReadBits_UnexpectedEOFMidRead(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF}))

	// One byte is available, but 9 bits requires spilling into a second
	// byte that the stream doesn't have.
	_, err := r.ReadBits(9)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestBitStreamReader_ReadBits_ExactEOFAfterConsumingAllData(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF}))

	_, err := r.ReadBits(8)
	require.NoError(t, err)

	_, err = r.ReadBits(1)
	require.ErrorIs(t, err, io.EOF)
}

func TestBitStreamReader_PropagatesUnderlyingReadError(t *testing.T) {
	wantErr := errors.New("boom")
	r := xbits.NewBitStreamReader(iotest.ErrReader(wantErr))

	_, err := r.ReadBits(1)
	require.ErrorIs(t, err, wantErr)
}

func TestBitStreamReader_Reset(t *testing.T) {
	r := xbits.NewBitStreamReader(bytes.NewReader([]byte{0xFF}))

	_, err := r.ReadBits(8)
	require.NoError(t, err)

	r.Reset(bytes.NewReader([]byte{0xAA}))

	value, err := r.ReadBits(8)
	require.NoError(t, err)
	require.Equal(t, uint32(0xAA), value)
}

// Fuzz_BitStreamRoundTrip writes randomized (value, count) pairs through a
// BitStreamWriter and verifies a BitStreamReader reads back the exact same
// masked values in the same order.
func Fuzz_BitStreamRoundTrip(f *testing.F) {
	f.Add([]byte{0xFF, 5, 0x00, 3})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, ops []byte) {
		var buf bytes.Buffer
		writer := xbits.NewBitStreamWriter(&buf)

		runBitsRoundTripFuzz(t, ops, writer.WriteBits, func() func(uint8) (uint32, error) {
			require.NoError(t, writer.Flush())

			return xbits.NewBitStreamReader(bytes.NewReader(buf.Bytes())).ReadBits
		})
	})
}
