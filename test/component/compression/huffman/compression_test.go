package huffman_test

import (
	"testing"

	"github.com/daanv2/go-webp/pkg/compression/huffman"
	xbits "github.com/daanv2/go-webp/pkg/extensions/bits"
	"github.com/stretchr/testify/require"
)

func roundTrip[T comparable](t *testing.T, input []T) {
	t.Helper()

	builder := huffman.NewTreeBuilder[T]()
	builder.Analyse(input)
	tree := builder.BuildTree()
	require.NotNil(t, tree)

	encoder, err := huffman.BuildEncoder(tree)
	require.NoError(t, err)

	writer := xbits.NewBitUint8Writer()
	err = encoder.EncodeTo(input, writer)
	require.NoError(t, err)

	wdata, _ := writer.Bytes()

	if len(input) == 0 {
		require.Empty(t, wdata)

		return
	}

	decoder, err := huffman.NewDecoder(tree)
	require.NoError(t, err)

	reader := xbits.NewBitUint8Reader(wdata)
	result, err := decoder.DecodeN(reader, len(input))
	require.NoError(t, err)

	require.Equal(t, input, result)
}

func Test_HuffmanRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"single symbol", []byte{42}},
		{"two symbols", []byte{1, 2, 1, 2}},
		// This case exposes the bit-ordering bug: multi-bit codes were decoded
		// in reversed bit order because WithAppendedBit stored the first decision
		// in the MSB but the writer/reader are LSB-first.
		{"fuzz seed 1", []byte{1, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 4, 4, 5}},
		{"fuzz seed 2", []byte{1, 1, 1, 1, 2, 2, 2, 3, 3, 4}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roundTrip(t, tc.input)
		})
	}
}

func Fuzz_HuffmanCompression_Bytes(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 4, 4, 5})
	f.Add([]byte{1, 1, 1, 1, 2, 2, 2, 3, 3, 4})
	f.Add([]byte("Example text string"))
	f.Add([]byte("One Ring to rule them all, One Ring to find them, One Ring to bring them all and in the darkness bind them"))

	f.Fuzz(roundTrip[byte])
}

func Fuzz_HuffmanCompression_Runes(f *testing.F) {
	f.Add("Example text string")
	f.Add("One Ring to rule them all, One Ring to find them, One Ring to bring them all and in the darkness bind them")

	f.Fuzz(func(t *testing.T, input string) {
		roundTrip(t, []rune(input))
	})
}
