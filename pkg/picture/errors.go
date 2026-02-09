package picture

import "errors"

// Encoding error conditions.
type WebPEncodingError error

var (
	ENC_OK                            error = nil
	ENC_ERROR_OUT_OF_MEMORY           error = errors.New("memory error allocating objects")
	ENC_ERROR_BITSTREAM_OUT_OF_MEMORY error = errors.New("memory error while flushing bits")
	ENC_ERROR_nil_PARAMETER           error = errors.New("a pointer parameter is nil")
	ENC_ERROR_INVALID_CONFIGURATION   error = errors.New("configuration is invalid")
	ENC_ERROR_BAD_DIMENSION           error = errors.New("picture has invalid width/height")
	ENC_ERROR_PARTITION0_OVERFLOW     error = errors.New("partition is bigger than 512k")
	ENC_ERROR_PARTITION_OVERFLOW      error = errors.New("partition is bigger than 16M")
	ENC_ERROR_BAD_WRITE               error = errors.New("error while flushing bytes")
	ENC_ERROR_FILE_TOO_BIG            error = errors.New("file is bigger than 4G")
	ENC_ERROR_USER_ABORT              error = errors.New("abort request by user")
	ENC_ERROR_LAST                    error = errors.New("list terminator. always last.")
)
