package webp

import (
	"errors"
	"io"
)

type Writer struct {
}

func Write(w io.Writer, fileSize uint32) error {
	baseHeader := (&FileHeader{
		FileID1:  FILE_HEADER_RIFF,
		FileID2:  FILE_HEADER_WEBP,
		FileSize: fileSize,
	}).ToBytes()

	n, err := w.Write(baseHeader[:])
	if err != nil {
		return err
	}
	if n != WEBP_FILE_HEADER_SIZE {
		return errors.New("didn't write the expect 12 bytes for the header")
	}

	return nil
}
