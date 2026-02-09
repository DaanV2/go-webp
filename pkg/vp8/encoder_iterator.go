package vp8

import "github.com/daanv2/go-webp/pkg/constants"

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

// Layout of prediction blocks
// intra 16x16
const I16DC16 = (0 * 16 * constants.BPS)
const I16TM16 = (I16DC16 + 16)
const I16VE16 = (1 * 16 * constants.BPS)
const I16HE16 = (I16VE16 + 16)

// chroma 8x8, two U/V blocks side by side (hence: 16x8 each)
const C8DC8 = (2 * 16 * constants.BPS)
const C8TM8 = (C8DC8 + 1*16)
const C8VE8 = (2*16*constants.BPS + 8*constants.BPS)
const C8HE8 = (C8VE8 + 1*16)

// intra 4x4
const I4DC4 = (3*16*constants.BPS + 0)
const I4TM4 = (I4DC4 + 4)
const I4VE4 = (I4DC4 + 8)
const I4HE4 = (I4DC4 + 12)
const I4RD4 = (I4DC4 + 16)
const I4VR4 = (I4DC4 + 20)
const I4LD4 = (I4DC4 + 24)
const I4VL4 = (I4DC4 + 28)
const I4HD4 = (3*16*constants.BPS + 4*constants.BPS)
const I4HU4 = (I4HD4 + 4)
const I4TMP = (I4HD4 + 8)

const MAX_COST = 0x7fffffffffffff

const QFIX = 17

func BIAS(b int) int {
	return ((b) << (QFIX - 8))
}

// Fun fact: this is the _only_ line where we're actually being lossy and
// discarding bits.
func QUANTDIV(n, iQ, B uint32) int {
	return int((n*iQ + B) >> QFIX)
}

// Uncomment the following to remove token-buffer code:
// #define DISABLE_TOKEN_BUFFER

// quality below which error-diffusion is enabled
const ERROR_DIFFUSION_QUALITY = 98

//------------------------------------------------------------------------------
// Headers

type proba_t uint32 // 16b + 16b
type ProbaArray [NUM_CTX][NUM_PROBAS]uint8
type StatsArray [NUM_CTX][NUM_PROBAS]proba_t
type CostArray [NUM_CTX][MAX_VARIABLE_LEVEL + 1]uint16
type CostArrayPtr [NUM_CTX]*uint16 // for easy casting
type CostArrayMap [16][NUM_CTX]*uint16
type LFStats [NUM_MB_SEGMENTS][MAX_LF_LEVELS]float64 // filter stats

// segment features
type VP8EncSegmentHeader struct {
	num_segments int // Actual number of segments. 1 segment only = unused.
	// whether to update the segment map or not.
	// must be 0 if there's only 1 segment.
	update_map int
	size       int // bit-cost for transmitting the segment map
}

// Struct collecting all frame-persistent probabilities.
type VP8EncProba struct {
	segments       [3]uint8                         // probabilities for segment tree
	skip_proba     uint8                            // final probability of being skipped.
	coeffs         [NUM_TYPES][NUM_BANDS]ProbaArray // 1056 bytes
	stats          [NUM_TYPES][NUM_BANDS]StatsArray // 4224 bytes
	level_cost     [NUM_TYPES][NUM_BANDS]CostArray  // 13056 bytes
	remapped_costs [NUM_TYPES]CostArrayMap          // 1536 bytes
	dirty          int                              // if true, need to call VP8CalculateLevelCosts()
	use_skip_proba int                              // Note: we always use skip_proba for now.
	nb_skip        int                              // number of skipped blocks
}

// Filter parameters. Not actually used in the code (we don't perform
// the in-loop filtering), but filled from user's config
type VP8EncFilterHeader struct {
	simple        int // filtering type: 0=complex, 1=simple
	level         int // base filter level [0..63]
	sharpness     int // [0..7]
	i4x4_lf_delta int // delta filter level for i4x4 relative to i16x16
}

//------------------------------------------------------------------------------
// Informations about the macroblocks.

type VP8MBInfo struct {
	// block type
	vtype   uint  //: 2;  // 0=i4x4, 1=i16x16
	uv_mode uint  //: 2
	skip    uint  //: 1
	segment uint  //: 2
	alpha   uint8 // quantization-susceptibility
}

type VP8Matrix struct {
	q       [16]uint16 // quantizer steps
	iq      [16]uint16 // reciprocals, fixed point.
	bias    [16]uint32 // rounding bias
	zthresh [16]uint32 // value below which a coefficient is zeroed
	sharpen [16]uint16 // frequency boosters for slight sharpening
}

type VP8SegmentInfo struct {
	y1, y2, uv VP8Matrix // quantization matrices
	// quant-susceptibility, range [-127,127]. Zero is neutral.
	// Lower values indicate a lower risk of blurriness.
	alpha     int
	beta      int // filter-susceptibility, range [0,255].
	quant     int // final segment quantizer.
	fstrength int // final in-loop filtering strength
	max_edge  int // max edge delta (for filtering strength)
	min_disto int // minimum distortion required to trigger filtering record
	// reactivities
	lambda_i16, lambda_i4, lambda_uv                         int
	lambda_mode, lambda_trellis, tlambda                     int
	lambda_trellis_i16, lambda_trellis_i4, lambda_trellis_uv int

	// lambda values for distortion-based evaluation
	i4_penalty score_t // penalty for using Intra4
}

type DError [2] /* u/v */ [2] /* top or left */ int8

// Handy transient struct to accumulate score and info during RD-optimization
// and mode evaluation.
type VP8ModeScore struct {
	D, SD       score_t   // Distortion, spectral distortion
	H, R, score score_t   // header bits, rate, score.
	y_dc_levels [16]int16 // Quantized levels for luma-DC, luma-AC, chroma.
	y_ac_levels [16][16]int16
	uv_levels   [4 + 4][16]int16
	mode_i16    int        // mode number for intra16 prediction
	modes_i4    [16]uint8  // mode numbers for intra4 predictions
	mode_uv     int        // mode number of chroma prediction
	nz          uint32     // non-zero blocks
	derr        [2][3]int8 // DC diffusion errors for U/V for blocks #1/2/3
}

// Iterator structure to iterate through macroblocks, pointing to the
// right neighbouring data (samples, predictions, contexts, ...)
type VP8EncIterator struct {
	x, y        int           // current macroblock
	yuv_in      *uint8        // input samples
	yuv_out     *uint8        // output samples
	yuv_out2    *uint8        // secondary buffer swapped with yuv_out.
	yuv_p       *uint8        // scratch buffer for prediction
	enc         *VP8Encoder   // back-pointer
	mb          *VP8MBInfo    // current macroblock
	bw          *VP8BitWriter // current bit-writer
	preds       *uint8        // intra mode predictors (4x4 blocks)
	nz          *uint32       // non-zero pattern
	i4_boundary [40]uint8     // 32+8 boundary samples needed by intra4x4
	i4_top      *uint8        // pointer to the current top boundary sample
	i4          int           // current intra4x4 mode being tested
	top_nz      [9]int        // top-non-zero context.
	left_nz     [9]int        // left-non-zero. left_nz[8] is independent.
	bit_count   [4][3]uint64  // bit counters for coded levels.
	luma_bits   uint64        // macroblock bit-cost for luma
	uv_bits     uint64        // macroblock bit-cost for chroma
	lf_stats    *LFStats      // filter stats (borrowed from enc)
	do_trellis  int           // if true, perform extra level optimisation
	count_down  int           // number of mb still to be processed
	count_down0 int           // starting counter value (for progress)
	percent0    int           // saved initial progress percent

	left_derr DError  // left error diffusion (u/v)
	top_derr  *DError // top diffusion error - nil if disabled

	y_left *uint8 // left luma samples (addressable from index -1 to 15).
	u_left *uint8 // left u samples (addressable from index -1 to 7)
	v_left *uint8 // left v samples (addressable from index -1 to 7)

	y_top  *uint8 // top luma samples at position 'x'
	uv_top *uint8 // top u/v samples at position 'x', packed as 16 bytes

	// memory for storing y/u/v_left
	yuv_left_mem [17 + 16 + 16 + 8 + constants.WEBP_ALIGN_CST]uint8
	// memory for *yuv
	yuv_mem [3*YUV_SIZE_ENC + PRED_SIZE_ENC + constants.WEBP_ALIGN_CST]uint8
}

type VP8TBuffer struct {
	pages     *VP8Tokens // first page
	last_page *VP8Tokens // last page
	tokens    *uint16    // set to (*last_page).tokens
	left      int        // how many free tokens left before the page is full
	page_size int        // number of tokens per page
	error     int        // true in case of malloc error
}
