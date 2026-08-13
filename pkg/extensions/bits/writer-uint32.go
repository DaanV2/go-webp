package xbits

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	_ io.WriterTo = &BitUint32Writer{}
)

// BitUint32Writer packs bits into a slice of uint32 words in little-endian
// (LSB-first) bit order: within each word, bits are written starting at bit 0
// and advance toward bit 31, so the first bit written occupies the
// lowest-order bit of the word.
type BitUint32Writer struct {
	data   []uint32
	bitPos int
}

func NewBitUint32Writer() *BitUint32Writer {
	return &BitUint32Writer{
		data:   make([]uint32, 0),
		bitPos: 0,
	}
}

func (w *BitUint32Writer) BitLength() int {
	return len(w.data)*32 + w.bitPos
}

// WriteBit writes a single bit to the output stream. The bit is represented as a boolean value, where true represents 1 and false represents 0.
func (w *BitUint32Writer) WriteBit(bit bool) error {
	var value uint32
	if bit {
		value = 1
	}
	return w.WriteBits(value, 1)
}

// WriteBits writes the specified number of bits from the given value to the output stream. The bits are taken from the least significant bits of the value.
func (w *BitUint32Writer) WriteBits(bits uint32, count uint8) error {
	if count == 0 {
		return nil
	}
	if count > 32 {
		return fmt.Errorf("count must be between 1 and 32, got %d", count)
	}

	for count > 0 {
		if len(w.data) == 0 || w.bitPos == 32 {
			w.data = append(w.data, 0)
			w.bitPos = 0
		}

		// Write what we can to the current uint32
		remainingBitsInCurrent := 32 - w.bitPos
		n := count
		if int(n) > remainingBitsInCurrent {
			n = uint8(remainingBitsInCurrent)
		}

		mask := uint32(1)<<n - 1
		w.data[len(w.data)-1] |= (bits & mask) << w.bitPos
		bits >>= n
		w.bitPos += int(n)
		count -= n
	}

	return nil
}

// Uint32s returns the underlying slice of uint32 values and the current bit position within the last uint32.
func (w *BitUint32Writer) Uint32s() ([]uint32, int) {
	return w.data, w.bitPos
}

func (w *BitUint32Writer) Bytes() ([]byte, int) {
	b := make([]byte, 0, len(w.data)*4)
	for _, v := range w.data {
		b = binary.LittleEndian.AppendUint32(b, v)
	}
	c := w.bitPos % 8

	return b, c
}

func (w *BitUint32Writer) WriteTo(receiver io.Writer) (int64, error) {
	buf := [4]byte{}
	var total int64

	for _, v := range w.data {
		binary.LittleEndian.PutUint32(buf[:], v)
		n, err := receiver.Write(buf[:])
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (w *BitUint32Writer) Reset() {
	w.data = w.data[:0]
	w.bitPos = 0
}
