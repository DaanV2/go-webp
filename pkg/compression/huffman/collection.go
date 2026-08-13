package huffman

import (
	"iter"
)

// ByteReader is an interface that defines a method for reading a single byte from an input stream.
// See [bytes.Buffer] or [bufio.Reader] for implementations of this interface.
type ByteReader interface {
	// ReadByte reads a single byte from the input stream and returns it.
	ReadByte() (byte, error)
}

// BufferIter returns an iterator that reads bytes from the provided bufio.Reader until EOF or an error occurs.
// The iterator yields each byte read from the buffer.
func BufferIter[T ByteReader](buf T) iter.Seq[byte] {
	return func(yield func(byte) bool) {
		for {
			b, err := buf.ReadByte()
			if err != nil { // EOF or other error
				return
			}
			if !yield(b) {
				return
			}
		}
	}
}

// BufferIterWithLimit returns an iterator that reads bytes from the provided bufio.Reader up to a specified limit.
// The iterator yields each byte read from the buffer until the limit is reached or an error occurs.
func BufferIterWithLimit[T ByteReader](buf T, limit int) iter.Seq[byte] {
	return func(yield func(byte) bool) {
		count := 0
		for count < limit {
			b, err := buf.ReadByte()
			if err != nil { // EOF or other error
				return
			}
			if !yield(b) {
				return
			}
			count++
		}
	}
}
