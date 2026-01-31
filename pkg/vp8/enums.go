package vp8

type VP8LEncoderARGBContent int

const (
	kEncoderNone VP8LEncoderARGBContent = iota
	kEncoderARGB
	kEncoderNearLossless
	kEncoderPalette
)

type VP8StatusCode int

const (
	VP8_STATUS_OK VP8StatusCode = iota
	VP8_STATUS_OUT_OF_MEMORY
	VP8_STATUS_INVALID_PARAM
	VP8_STATUS_BITSTREAM_ERROR
	VP8_STATUS_UNSUPPORTED_FEATURE
	VP8_STATUS_SUSPENDED
	VP8_STATUS_USER_ABORT
	VP8_STATUS_NOT_ENOUGH_DATA
)