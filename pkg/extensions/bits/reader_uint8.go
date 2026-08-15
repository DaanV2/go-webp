package xbits

import (
	"fmt"
	"io"
)

// BitUint8Reader reads bits from a slice of uint8 words in little-endian
// (LSB-first) bit order: within each word, bits are read starting at bit 0
// and advance toward bit 7, matching the order written by [BitUint8Writer].
type BitUint8Reader struct {
	data   []uint8
	wordAt int
	bitPos int
}

func NewBitUint8Reader(data []uint8) *BitUint8Reader {
	return &BitUint8Reader{
		data: data,
	}
}

// BitLength returns the total number of bits available to read from the underlying data.
func (r *BitUint8Reader) BitLength() int {
	return len(r.data) * 8
}

// BitsRemaining returns the number of unread bits left in the stream.
func (r *BitUint8Reader) BitsRemaining() int {
	return r.BitLength() - (r.wordAt*8 + r.bitPos)
}

// ReadBit reads a single bit from the input stream and returns it as a boolean value, where true represents 1 and false represents 0.
func (r *BitUint8Reader) ReadBit() (bool, error) {
	bits, err := r.ReadBits(1)
	if err != nil {
		return false, err
	}

	return bits == 1, nil
}

// ReadBits reads the specified number of bits from the input stream and returns them as a uint32 value. The bits are read from the least significant bits of the returned value.
func (r *BitUint8Reader) ReadBits(count uint8) (uint32, error) {
	if count == 0 {
		return 0, nil
	}
	if count > 32 {
		return 0, fmt.Errorf("count must be between 1 and 32, got %d", count)
	}
	if int(count) > r.BitsRemaining() {
		return 0, io.ErrUnexpectedEOF
	}

	var result uint32
	var shift uint8

	for count > 0 {
		if r.bitPos == 8 {
			r.wordAt++
			r.bitPos = 0
		}

		// Read what we can from the current uint8
		remainingBitsInCurrent := 8 - r.bitPos
		n := count
		if int(n) > remainingBitsInCurrent {
			n = uint8(remainingBitsInCurrent) // nolint:gosec // Overflow is acceptable
		}

		mask := uint32(1)<<n - 1
		bits := (uint32(r.data[r.wordAt]) >> r.bitPos) & mask
		result |= bits << shift

		shift += n
		r.bitPos += int(n)
		count -= n
	}

	return result, nil
}

// Reset clears the reader's internal state and sets new data to read from.
func (r *BitUint8Reader) Reset(data []uint8) {
	r.data = data
	r.wordAt = 0
	r.bitPos = 0
}
