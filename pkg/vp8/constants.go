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
const (
	YUV_SIZE_ENC  = (constants.BPS * 16)
	PRED_SIZE_ENC = (32*constants.BPS + 16*constants.BPS + 8*constants.BPS) // I16+Chroma+I4 preds
	Y_OFF_ENC     = (0)
	U_OFF_ENC     = (16)
	V_OFF_ENC     = (16 + 8)
)

// WebPGetDecoderVersion Return the decoder's version number, packed in hexadecimal using 8bits for
// each of major/minor/revision. E.g: v2.5.7 is 0x020507.
func WebPGetDecoderVersion() int {
	return (DEC_MAJ_VERSION << 16) | (DEC_MIN_VERSION << 8) | DEC_REV_VERSION
}

// The Boolean decoder needs to maintain infinite precision on the 'value'
// field. However, since 'range' is only 8bit, we only need an active window of
// 8 bits for 'value". Left bits (MSB) gets zeroed and shifted away when
// 'value' falls below 128, 'range' is updated, and fresh bits read from the
// bitstream are brought in as LSB. To afunc reading the fresh bits one by one
// (slow), we cache BITS of them ahead. The total of (BITS + 8) bits must fit
// into a natural register (with type bit_t). To fetch BITS bits from bitstream
// we use a type lbit_t.
//
// BITS can be any multiple of 8 from 8 to 56 (inclusive).
// Pick values that fit natural register size.

// #if defined(__i386__) || defined(_M_IX86)  // x86 32bit
// const BITS = 24
// #elif defined(__x86_64__) || defined(_M_X64)  // x86 64bit
// const BITS = 56
// #elif defined(__arm__) || defined(_M_ARM)  // ARM
// const BITS = 24
// #elif WEBP_AARCH64  // ARM 64bit
// const BITS = 56
// #elif defined(__mips__)  // MIPS
// const BITS = 24
// #elif defined(__wasm__)  // WASM
// const BITS = 56
// #else  // reasonable default

// #endif

const (
	BITS                  = 56
	VP8L_MAX_NUM_BIT_READ = 24

	VP8L_LBITS = 64 // Number of bits prefetched (= bit-size of vp8l_val_t).
	VP8L_WBITS = 32 // Minimum number of bytes ready after VP8LFillBitWindow.

	VP8L_LOG8_WBITS = 4 // Number of bytes needed to store VP8L_WBITS bits.

	// This is the minimum amount of size the memory buffer is guaranteed to grow
	// when extra space is needed.
	MIN_EXTRA_SIZE = (uint64(32768))

	// C: VP8L_WRITER_BYTES    = 4  // sizeof(vp8l_wtype_t)
	// C: VP8L_WRITER_BITS     = 32 // 8 * sizeof(vp8l_wtype_t)
	// C: VP8L_WRITER_MAX_BITS = 64 // 8 * sizeof(vp8l_atype_t)

	SYNC_EVERY_N_ROWS = 8 // minimum number of rows between check-points
)

var (
	kVP8Log2Range = [128]uint8{
		7, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 4, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0}

	// range = ((range - 1) << kVP8Log2Range[range]) + 1
	kVP8NewRange = [128]uint8{
		127, 127, 191, 127, 159, 191, 223, 127, 143, 159, 175, 191, 207, 223, 239, 127, 135, 143, 151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 127, 131, 135, 139, 143, 147, 151, 155, 159, 163, 167, 171, 175, 179, 183, 187, 191, 195, 199, 203, 207, 211, 215, 219, 223, 227, 231, 235, 239, 243, 247, 251, 127, 129, 131, 133, 135, 137, 139, 141, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167, 169, 171, 173, 175, 177, 179, 181, 183, 185, 187, 189, 191, 193, 195, 197, 199, 201, 203, 205, 207, 209, 211, 213, 215, 217, 219, 221, 223, 225, 227, 229, 231, 233, 235, 237, 239, 241, 243, 245, 247, 249, 251, 253, 127}

	kBitMask = [VP8L_MAX_NUM_BIT_READ + 1]uint32{
		0, 0x000001, 0x000003, 0x000007, 0x00000f, 0x00001f, 0x00003f, 0x00007f, 0x0000ff, 0x0001ff, 0x0003ff, 0x0007ff, 0x000fff, 0x001fff, 0x003fff, 0x007fff, 0x00ffff, 0x01ffff, 0x03ffff, 0x07ffff, 0x0fffff, 0x1fffff, 0x3fffff, 0x7fffff, 0xffffff}

	kNorm = [128]uint8{ // renorm_sizes[i] = 8 - log2(i)
		7, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 4, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0}

	// range = ((range + 1) << kVP8Log2Range[range]) - 1
	kNewRange = [128]uint8{
		127, 127, 191, 127, 159, 191, 223, 127, 143, 159, 175, 191, 207, 223, 239, 127, 135, 143, 151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 127, 131, 135, 139, 143, 147, 151, 155, 159, 163, 167, 171, 175, 179, 183, 187, 191, 195, 199, 203, 207, 211, 215, 219, 223, 227, 231, 235, 239, 243, 247, 251, 127, 129, 131, 133, 135, 137, 139, 141, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167, 169, 171, 173, 175, 177, 179, 181, 183, 185, 187, 189, 191, 193, 195, 197, 199, 201, 203, 205, 207, 209, 211, 213, 215, 217, 219, 221, 223, 225, 227, 229, 231, 233, 235, 237, 239, 241, 243, 245, 247, 249, 251, 253, 127}
)
