// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

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

type VP8LImageTransformType int

const (
	PREDICTOR_TRANSFORM      VP8LImageTransformType = 0
	CROSS_COLOR_TRANSFORM    VP8LImageTransformType = 1
	SUBTRACT_GREEN_TRANSFORM VP8LImageTransformType = 2
	COLOR_INDEXING_TRANSFORM VP8LImageTransformType = 3
)

type VP8LDecodeState int

const (
	READ_DATA VP8LDecodeState = 0
	READ_HDR  VP8LDecodeState = 1
	READ_DIM  VP8LDecodeState = 2
)

// Rate-distortion optimization levels
type VP8RDLevel int

const (
	RD_OPT_NONE        VP8RDLevel = 0 // no rd-opt
	RD_OPT_BASIC       VP8RDLevel = 1 // basic scoring (no trellis)
	RD_OPT_TRELLIS     VP8RDLevel = 2 // perform trellis-quant on the final decision only
	RD_OPT_TRELLIS_ALL VP8RDLevel = 3 // trellis-quant for every scoring (much slower)
)
