package xio

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrWrittenUnexpectedAmount = errors.New("didn't write the expected amount of bytes")
	ErrReadUnexpectedAmount    = errors.New("didn't read the expected amount of bytes")
)

func AssertWrite(writer io.Writer, data []byte) error {
	n, err := writer.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("%w: written %d but wanted %d", ErrWrittenUnexpectedAmount, n, len(data))
	}

	return nil
}

// AssertRead expects that dst will be filled entirely, thus must n == len(dst)
func AssertRead(reader io.Reader, dst []byte) error {
	n, err := reader.Read(dst)
	if err != nil {
		return err
	}
	if n != len(dst) {
		return fmt.Errorf("%w: read %d but wanted %d", ErrReadUnexpectedAmount, n, len(dst))
	}

	return nil
}
