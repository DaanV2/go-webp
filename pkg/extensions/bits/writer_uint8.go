package xbits

import (
	"fmt"
	"io"
)

var (
	_ io.WriterTo = &BitUint8Writer{}
)

// BitUint8Writer packs bits into a slice of uint8 words in little-endian
// (LSB-first) bit order: within each word, bits are written starting at bit 0
// and advance toward bit 7, so the first bit written occupies the
// lowest-order bit of the word.
type BitUint8Writer struct {
	data   []uint8
	bitPos int
}

func NewBitUint8Writer() *BitUint8Writer {
	return &BitUint8Writer{
		data:   make([]uint8, 0),
		bitPos: 0,
	}
}

func (w *BitUint8Writer) BitLength() int {
	if len(w.data) == 0 {
		return 0
	}

	return (len(w.data)-1)*8 + w.bitPos
}

// WriteBit writes a single bit to the output stream. The bit is represented as a boolean value, where true represents 1 and false represents 0.
func (w *BitUint8Writer) WriteBit(bit bool) error {
	var value uint32
	if bit {
		value = 1
	}

	return w.WriteBits(value, 1)
}

// WriteBits writes the specified number of bits from the given value to the output stream. The bits are taken from the least significant bits of the value.
func (w *BitUint8Writer) WriteBits(bits uint32, count uint8) error {
	if count == 0 {
		return nil
	}
	if count > 32 {
		return fmt.Errorf("count must be between 1 and 32, got %d", count)
	}

	for count > 0 {
		if len(w.data) == 0 || w.bitPos == 8 {
			w.data = append(w.data, 0)
			w.bitPos = 0
		}

		// Write what we can to the current uint8
		remainingBitsInCurrent := 8 - w.bitPos
		n := count
		if int(n) > remainingBitsInCurrent {
			n = uint8(remainingBitsInCurrent) // nolint:gosec // Overflow is acceptable
		}

		mask := uint32(1)<<n - 1
		w.data[len(w.data)-1] |= uint8((bits & mask) << w.bitPos) // nolint:gosec // Overflow is acceptable
		bits >>= n
		w.bitPos += int(n)
		count -= n
	}

	return nil
}

func (w *BitUint8Writer) Bytes() (data []byte, bitPos int) {
	b := make([]byte, 0, len(w.data))
	b = append(b, w.data...)

	return b, w.bitPos % 8
}

func (w *BitUint8Writer) WriteTo(receiver io.Writer) (int64, error) {
	n, err := receiver.Write(w.data)

	return int64(n), err
}

// Reset clears the writer's internal state, allowing it to be reused for writing new bits.
func (w *BitUint8Writer) Reset() {
	w.data = w.data[:0]
	w.bitPos = 0
}
