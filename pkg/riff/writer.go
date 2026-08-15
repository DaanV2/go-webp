package riff

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/daanv2/go-webp/pkg/extensions/numbers"
	"github.com/daanv2/go-webp/pkg/extensions/xio"
)

func WriteChunk(w io.Writer, c *Chunk) error {
	fourCC := c.ID
	payload := c.Payload
	payloadSize := uint32(len(c.Payload)) // nolint:gosec // integer overflow is acceptable

	// Write FourCC
	if err := xio.AssertWrite(w, fourCC[:]); err != nil {
		return fmt.Errorf("error writing chunk fourCC: %w", err)
	}

	// write Chunk Size
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], payloadSize)
	if err := xio.AssertWrite(w, buf[:]); err != nil {
		return fmt.Errorf("error writing chunk size: %w", err)
	}

	// write Chunk Payload
	if err := xio.AssertWrite(w, payload); err != nil {
		return fmt.Errorf("error writing chunk payload: %w", err)
	}

	// Is odd, pad with byte (0)
	if numbers.IsOdd(payloadSize) {
		if err := xio.AssertWrite(w, []byte{0}); err != nil {
			return fmt.Errorf("error writing chunk padding: %w", err)
		}
	}

	return nil
}
