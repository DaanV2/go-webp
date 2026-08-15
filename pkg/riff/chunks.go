package riff

import "github.com/daanv2/go-webp/pkg/extensions/numbers"

type Chunk struct {
	ID      FourCC
	Payload []byte
}

func (c *Chunk) ByteSize() int {
	amount := c.ID.Len()
	amount += len(c.Payload)

	if numbers.IsOdd(len(c.Payload)) {
		amount += 1
	}

	return amount
}
