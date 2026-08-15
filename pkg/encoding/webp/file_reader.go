package webp

import (
	"errors"
	"fmt"
	"io"

	"github.com/daanv2/go-webp/pkg/extensions/xio"
	"github.com/daanv2/go-webp/pkg/riff"
)

var (
	FILE_HEADER_RIFF = riff.FourCC{'R', 'I', 'F', 'F'} // RIFF
	FILE_HEADER_WEBP = riff.FourCC{'W', 'E', 'B', 'P'} // WEBP
)

type foo int

type Reader struct {
}

func Read(reader io.Reader) (foo, error) {
	var buf [12]byte

	// Read file header
	if err := xio.AssertRead(reader, buf[:]); err != nil {
		return foo(0), fmt.Errorf("failed to read file header: %w", err)
	}

	// Parse header
	var header FileHeader
	if err := header.Parse(buf[:]); err != nil {
		return foo(0), err
	}

	// Validate
	if header.FileID1 != FILE_HEADER_RIFF || header.FileID2 != FILE_HEADER_WEBP {
		return foo(0), errors.New("invalid webp header")
	}

	// Read Payload
	payloadSize := header.FileSize
	remainingBytes := int64(payloadSize) - 8 // size starts at offset 8
	var chunks []*riff.Chunk

	for remainingBytes > 0 {
		c, err := riff.ReadChunk(reader)
		if err != nil {
			return foo(0), fmt.Errorf("error reading chunk: %w", err)
		}

		chunks = append(chunks, c)
		remainingBytes -= int64(c.ByteSize())
	}

	return foo(0), nil
}
