package xbits_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// bitOp is a single (value, count) write/read operation used by the
// round-trip fuzz tests.
type bitOp struct {
	value uint32
	count uint8
}

// genBitOps turns raw fuzzer bytes into a sequence of bit operations, each
// with a count in [1, 32] and a value already masked to that count.
func genBitOps(ops []byte) []bitOp {
	result := make([]bitOp, 0, len(ops)/2)

	for i := 0; i+1 < len(ops); i += 2 {
		count := ops[i+1]%32 + 1
		mask := uint32(1)<<count - 1
		value := uint32(ops[i]) & mask

		result = append(result, bitOp{value: value, count: count})
	}

	return result
}

// runBitsRoundTripFuzz writes ops through writeBits, then uses newReader to
// obtain a ReadBits func over whatever the writer produced, and asserts it
// reads back the exact same values in the same order.
func runBitsRoundTripFuzz(
	t *testing.T,
	ops []byte,
	writeBits func(value uint32, count uint8) error,
	newReader func() func(count uint8) (uint32, error),
) {
	t.Helper()

	expected := genBitOps(ops)
	for _, op := range expected {
		require.NoError(t, writeBits(op.value, op.count))
	}

	readBits := newReader()
	for _, op := range expected {
		got, err := readBits(op.count)
		require.NoError(t, err)
		require.Equal(t, op.value, got)
	}
}
