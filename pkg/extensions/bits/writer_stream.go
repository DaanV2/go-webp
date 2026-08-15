package xbits

import (
	"fmt"
	"io"
)

// BitStreamWriter packs bits into bytes in little-endian (LSB-first) bit
// order and flushes each completed byte to the underlying io.Writer as soon
// as it fills up, so memory usage stays constant regardless of how many bits
// are written.
type BitStreamWriter struct {
	w      io.Writer
	cur    uint8
	bitPos int
}

func NewBitStreamWriter(w io.Writer) *BitStreamWriter {
	return &BitStreamWriter{
		w: w,
	}
}

// WriteBit writes a single bit to the output stream. The bit is represented as a boolean value, where true represents 1 and false represents 0.
func (w *BitStreamWriter) WriteBit(bit bool) error {
	var value uint32
	if bit {
		value = 1
	}

	return w.WriteBits(value, 1)
}

// WriteBits writes the specified number of bits from the given value to the output stream. The bits are taken from the least significant bits of the value.
func (w *BitStreamWriter) WriteBits(bits uint32, count uint8) error {
	if count == 0 {
		return nil
	}
	if count > 32 {
		return fmt.Errorf("count must be between 1 and 32, got %d", count)
	}

	for count > 0 {
		// Write what we can to the current byte
		remainingBitsInCurrent := 8 - w.bitPos
		n := count
		if int(n) > remainingBitsInCurrent {
			n = uint8(remainingBitsInCurrent) // nolint:gosec // Overflow is acceptable
		}

		mask := uint32(1)<<n - 1
		w.cur |= uint8((bits & mask) << w.bitPos) // nolint:gosec // Overflow is acceptable
		bits >>= n
		w.bitPos += int(n)
		count -= n

		if w.bitPos == 8 {
			if err := w.flush(); err != nil {
				return err
			}
		}
	}

	return nil
}

// flush writes the current byte to the underlying writer and resets it.
func (w *BitStreamWriter) flush() error {
	_, err := w.w.Write([]byte{w.cur})
	w.cur = 0
	w.bitPos = 0

	return err
}

// Flush pads any partially-written byte with zero bits and writes it to the
// underlying io.Writer. Call this once writing is complete to ensure all
// buffered bits reach the underlying writer.
func (w *BitStreamWriter) Flush() error {
	if w.bitPos == 0 {
		return nil
	}

	return w.flush()
}

// PendingBits returns the number of bits currently buffered but not yet flushed to the underlying writer.
func (w *BitStreamWriter) PendingBits() int {
	return w.bitPos
}

// Reset discards any buffered bits and sets a new underlying writer, allowing the BitStreamWriter to be reused.
func (w *BitStreamWriter) Reset(dst io.Writer) {
	w.w = dst
	w.cur = 0
	w.bitPos = 0
}
