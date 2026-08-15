package riff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/daanv2/go-webp/pkg/extensions/numbers"
	"github.com/daanv2/go-webp/pkg/extensions/xio"
)

func ReadChunkBytes(data []byte) (*Chunk, int64, error) {
	if len(data) < (4 + 4) {
		return nil, 0, errors.New("need atleast 8 bytes for riff chunk")
	}

	var fourCC FourCC
	// Read FourCC
	copy(fourCC[:], data[:4])

	// Read Size
	payloadSize := binary.LittleEndian.Uint32(data[4:8])

	// Setup Chunk
	c := &Chunk{
		ID:      fourCC,
		Payload: make([]byte, payloadSize),
	}
	bytesRead := int64(8 + payloadSize)

	// Read Chunk Payload
	copy(c.Payload, data[12:])

	// Is odd, pad with byte (0)
	if numbers.IsOdd(payloadSize) {
		bytesRead += 1
	}

	return c, bytesRead, nil
}

func ReadChunk(reader io.Reader) (*Chunk, error) {
	var fourCC FourCC

	// Read FourCC
	if err := xio.AssertRead(reader, fourCC[:]); err != nil {
		return nil, fmt.Errorf("error reading chunk fourCC: %w", err)
	}

	// Read Chunk Size
	var buf [4]byte
	if err := xio.AssertRead(reader, buf[:]); err != nil {
		return nil, fmt.Errorf("error reading chunk size: %w", err)
	}
	payloadSize := binary.LittleEndian.Uint32(buf[:])

	// Read Chunk Payload
	payload := make([]byte, payloadSize)
	if err := xio.AssertRead(reader, payload); err != nil {
		return nil, fmt.Errorf("error reading chunk payload: %w", err)
	}

	// Is odd, pad with byte (0)
	if numbers.IsOdd(payloadSize) {
		if err := xio.AssertRead(reader, buf[:1]); err != nil {
			return nil, fmt.Errorf("error reading chunk padding: %w", err)
		}
	}

	return &Chunk{
		ID:      fourCC,
		Payload: payload,
	}, nil
}
