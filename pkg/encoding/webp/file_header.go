package webp

import (
	"encoding/binary"
	"errors"

	"github.com/daanv2/go-webp/pkg/riff"
)

const (
	WEBP_FILE_HEADER_SIZE = 12
)

type FileHeader struct {
	// The first starting ID of the file, should be "RIFF"
	FileID1 riff.FourCC
	// The size of the file, minus 10 bytes
	FileSize uint32
	// The second starting ID of the file, should be "WEBP"
	FileID2 riff.FourCC
}

func (f *FileHeader) Parse(data []byte) error {
	if len(data) < WEBP_FILE_HEADER_SIZE {
		return errors.New("file header is not 12 bytes")
	}
	copy(f.FileID1[:], data[0:4])
	f.FileSize = binary.LittleEndian.Uint32(data[4:8])
	copy(f.FileID2[:], data[8:12])

	return nil
}

func (f *FileHeader) ToBytes() [WEBP_FILE_HEADER_SIZE]byte {
	var result [WEBP_FILE_HEADER_SIZE]byte

	copy(result[0:4], f.FileID1[:])
	binary.LittleEndian.PutUint32(result[4:8], f.FileSize)
	copy(result[8:12], f.FileID2[:])

	return result
}
