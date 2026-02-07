package picture

import "errors"

// Encoding error conditions.
type WebPEncodingError error

var (
	VP8_ENC_OK                            error = nil
	VP8_ENC_ERROR_OUT_OF_MEMORY           error = errors.New("memory error allocating objects")
	VP8_ENC_ERROR_BITSTREAM_OUT_OF_MEMORY error = errors.New("memory error while flushing bits")
	VP8_ENC_ERROR_nil_PARAMETER           error = errors.New("a pointer parameter is nil")
	VP8_ENC_ERROR_INVALID_CONFIGURATION   error = errors.New("configuration is invalid")
	VP8_ENC_ERROR_BAD_DIMENSION           error = errors.New("picture has invalid width/height")
	VP8_ENC_ERROR_PARTITION0_OVERFLOW     error = errors.New("partition is bigger than 512k")
	VP8_ENC_ERROR_PARTITION_OVERFLOW      error = errors.New("partition is bigger than 16M")
	VP8_ENC_ERROR_BAD_WRITE               error = errors.New("error while flushing bytes")
	VP8_ENC_ERROR_FILE_TOO_BIG            error = errors.New("file is bigger than 4G")
	VP8_ENC_ERROR_USER_ABORT              error = errors.New("abort request by user")
	VP8_ENC_ERROR_LAST                    error = errors.New("list terminator. always last.")
)
