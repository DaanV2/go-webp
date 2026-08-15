package riff_test

import (
	"bytes"
	"testing"

	"github.com/daanv2/go-webp/pkg/riff"
	"github.com/stretchr/testify/require"
)

func Fuzz_Riff_Write_Read(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3, 4, 5})

	f.Fuzz(func(t *testing.T, payload []byte) {
		var buf bytes.Buffer
		org := riff.Chunk{
			ID:      riff.FourCC{'T', 'E', 'S', 'T'},
			Payload: payload,
		}

		err := riff.WriteChunk(&buf, &org)
		require.NoError(t, err)

		reader := bytes.NewBuffer(buf.Bytes())
		reconstuct, err := riff.ReadChunk(reader)
		require.NoError(t, err)
		require.NotNil(t, reconstuct)

		require.Equal(t, org, *reconstuct)
	})
}
