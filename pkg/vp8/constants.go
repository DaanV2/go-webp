// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import "github.com/daanv2/go-webp/pkg/constants"

const (
	// maximum value of 'transform_bits' in VP8LEncoder.
	MAX_TRANSFORM_BITS = (constants.MIN_TRANSFORM_BITS + (1 << constants.NUM_TRANSFORM_BITS) - 1)

	// version numbers
	DEC_MAJ_VERSION = 1
	DEC_MIN_VERSION = 6
	DEC_REV_VERSION = 0

	ENC_MAJ_VERSION = 1
	ENC_MIN_VERSION = 6
	ENC_REV_VERSION = 0

	// minimal width under which lossy multi-threading is always disabled
	MIN_WIDTH_FOR_THREADS = 512

	MB_FEATURE_TREE_PROBS = 3
	NUM_MB_SEGMENTS       = 4
	NUM_REF_LF_DELTAS     = 4
	NUM_MODE_LF_DELTAS    = 4 // I4x4, ZERO, *, SPLIT
	MAX_NUM_PARTITIONS    = 8 // Probabilities
	NUM_TYPES             = 4 // 0: i16-AC,  1: i16-DC,  2:chroma-AC,  3:i4-AC
	NUM_BANDS             = 8
	NUM_CTX               = 3
	NUM_PROBAS            = 11

	MAX_LF_LEVELS      = 64   // Maximum loop filter level
	MAX_VARIABLE_LEVEL = 67   // last (inclusive) level with variable cost
	MAX_LEVEL          = 2047 // max level (note: max codable is 2047 + 67)
)

// YUV-cache parameters. Cache is 32-bytes wide (= one cacheline).
// Constraints are: We need to store one 16x16 block of luma samples (y),
// and two 8x8 chroma blocks (u/v). These are better be 16-bytes aligned,
// in order to be SIMD-friendly. We also need to store the top, left and
// top-left samples (from previously decoded blocks), along with four
// extra top-right samples for luma (intra4x4 prediction only).
// One possible layout is, using 32 * (17 + 9) bytes:
//
//   .+------   <- only 1 pixel high
//   .|yyyyt.
//   .|yyyyt.
//   .|yyyyt.
//   .|yyyy..
//   .+--.+--   <- only 1 pixel high
//   .|uu.|vv
//   .|uu.|vv
//
// Every character is a 4x4 block, with legend:
//  '.' = unused
//  'y' = y-samples   'u' = u-samples     'v' = u-samples
//  '|' = left sample,   '-' = top sample,    '+' = top-left sample
//  't' = extra top-right sample for 4x4 modes
const (
	YUV_SIZE = (constants.BPS*17 + constants.BPS*9)
	Y_OFF    = (constants.BPS*1 + 8)
	U_OFF    = (Y_OFF + constants.BPS*16 + constants.BPS)
	V_OFF    = (U_OFF + 16)
)

// YUV-cache parameters. Cache is 32-bytes wide (= one cacheline).
// The original or reconstructed samples can be accessed using VP8Scan[].
// The predicted blocks can be accessed using offsets to 'yuv_p' and
// the arrays *VP8ModeOffsets[].
// * YUV Samples area ('yuv_in'/'yuv_out'/'yuv_out2')
//   (see VP8Scan[] for accessing the blocks, along with
//   Y_OFF_ENC/U_OFF_ENC/V_OFF_ENC):
//             +----+----+
//  Y_OFF_ENC  |YYYY|UUVV|
//  U_OFF_ENC  |YYYY|UUVV|
//  V_OFF_ENC  |YYYY|....| <- 25% wasted U/V area
//             |YYYY|....|
//             +----+----+
// * Prediction area ('yuv_p', size = PRED_SIZE_ENC)
//   Intra16 predictions (16x16 block each, two per row):
//         |I16DC16|I16TM16|
//         |I16VE16|I16HE16|
//   Chroma U/V predictions (16x8 block each, two per row):
//         |C8DC8|C8TM8|
//         |C8VE8|C8HE8|
//   Intra 4x4 predictions (4x4 block each)
//         |I4DC4 I4TM4 I4VE4 I4HE4|I4RD4 I4VR4 I4LD4 I4VL4|
//         |I4HD4 I4HU4 I4TMP .....|.......................| <- ~31% wasted
const YUV_SIZE_ENC =(constants.BPS * 16)
const PRED_SIZE_ENC =(32 * constants.BPS + 16 * constants.BPS + 8 * constants.BPS)  // I16+Chroma+I4 preds
const Y_OFF_ENC =(0)
const U_OFF_ENC =(16)
const V_OFF_ENC =(16 + 8)

// WebPGetDecoderVersion Return the decoder's version number, packed in hexadecimal using 8bits for
// each of major/minor/revision. E.g: v2.5.7 is 0x020507.
func WebPGetDecoderVersion() int {
	return (DEC_MAJ_VERSION << 16) | (DEC_MIN_VERSION << 8) | DEC_REV_VERSION
}
