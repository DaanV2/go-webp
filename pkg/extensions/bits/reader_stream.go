package xbits

import (
	"errors"
	"fmt"
	"io"
)

// BitStreamReader reads bits from an underlying io.Reader one byte at a time
// in little-endian (LSB-first) bit order, so memory usage stays constant
// regardless of how many bits are read.
type BitStreamReader struct {
	r    io.Reader
	buf  [1]byte
	cur  uint8
	left uint8
}

func NewBitStreamReader(r io.Reader) *BitStreamReader {
	return &BitStreamReader{
		r: r,
	}
}

// ReadBit reads a single bit from the input stream and returns it as a boolean value, where true represents 1 and false represents 0.
func (r *BitStreamReader) ReadBit() (bool, error) {
	bits, err := r.ReadBits(1)
	if err != nil {
		return false, err
	}

	return bits == 1, nil
}

// ReadBits reads the specified number of bits from the input stream and returns them as a uint32 value. The bits are read from the least significant bits of the returned value.
func (r *BitStreamReader) ReadBits(count uint8) (uint32, error) {
	if count == 0 {
		return 0, nil
	}
	if count > 32 {
		return 0, fmt.Errorf("count must be between 1 and 32, got %d", count)
	}

	var result uint32
	var shift uint8

	for count > 0 {
		if r.left == 0 {
			if _, err := io.ReadFull(r.r, r.buf[:]); err != nil {
				if errors.Is(err, io.EOF) && shift > 0 {
					err = io.ErrUnexpectedEOF
				}

				return 0, err
			}
			r.cur = r.buf[0]
			r.left = 8
		}

		// Read what we can from the current byte
		bitPos := 8 - r.left
		n := min(count, r.left)

		mask := uint32(1)<<n - 1
		bits := (uint32(r.cur) >> bitPos) & mask
		result |= bits << shift

		shift += n
		r.left -= n
		count -= n
	}

	return result, nil
}

// Reset discards any buffered bits and sets a new underlying reader, allowing the BitStreamReader to be reused.
func (r *BitStreamReader) Reset(src io.Reader) {
	r.r = src
	r.cur = 0
	r.left = 0
}
