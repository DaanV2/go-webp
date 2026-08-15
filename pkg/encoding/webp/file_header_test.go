package webp_test

import (
	"testing"

	"github.com/daanv2/go-webp/pkg/encoding/webp"
	"github.com/stretchr/testify/require"
)

func Test_FileHeader(t *testing.T) {
	header := webp.FileHeader{
		FileID1:  webp.FILE_HEADER_RIFF,
		FileSize: 123456789,
		FileID2:  webp.FILE_HEADER_WEBP,
	}

	b := header.ToBytes()

	var otherHeader webp.FileHeader
	err := otherHeader.Parse(b[:])
	require.NoError(t, err)

	require.Equal(t, header, otherHeader)
}
