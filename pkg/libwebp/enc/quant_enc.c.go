package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//   Quantization
//
// Author: Skal (pascal.massimino@gmail.com)

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/decoder"
	"github.com/daanv2/go-webp/pkg/libwebp/dsp"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/math"
	"github.com/daanv2/go-webp/pkg/ptr"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/string"
) // for abs()

const DO_TRELLIS_I4 = 1
const DO_TRELLIS_I16 = 1  // not a huge gain, but ok at low bitrate.
const DO_TRELLIS_UV = 0   // disable trellis for UV. Risky. Not worth.
const USE_TDISTO = 1

const MID_ALPHA = 64   // neutral value for susceptibility
const MIN_ALPHA = 30   // lowest usable value for susceptibility
const MAX_ALPHA = 100  // higher meaningful value for susceptibility

// Scaling constant between the sns value and the QP
// power-law modulation. Must be strictly less than 1.
const SNS_TO_DQ  = 0.9  

// number of non-zero coeffs below which we consider the block very flat
// (and apply a penalty to complex predictions)
const FLATNESS_LIMIT_I16 =0  // I16 mode (special case)
const FLATNESS_LIMIT_I4 =3   // I4 mode
const FLATNESS_LIMIT_UV =2   // UV mode
const FLATNESS_PENALTY =140  // roughly ~1bit per block

const RD_DISTO_MULT =256  // distortion multiplier (equivalent of lambda)

// Very small filter-strength values have close to no visual effect. So we can
// save a little decoding-CPU by turning filtering off for these.
const FSTRENGTH_CUTOFF =2

func MULT_8B(a, b int) int {
	return (((a) * (b) + 128) >> 8)
}


//------------------------------------------------------------------------------

func clip(v, m, M int) int {
  return tenary.If(v < m, m, tenary.If(v > M, M, v))
}

var kZigzag = [16]uint8{0, 1,  4,  8,  5, 2,  3,  6, 9, 12, 13, 10, 7, 11, 14, 15}

var kDcTable = [128]uint8{
    4,   5,   6,   7,   8,   9,   10,  10,  11,  12,  13,  14,  15,  16,  17, 17,  18,  19,  20,  20,  21,  21,  22,  22,  23,  23,  24,  25,  25,  26, 27,  28,  29,  30,  31,  32,  33,  34,  35,  36,  37,  37,  38,  39,  40, 41,  42,  43,  44,  45,  46,  46,  47,  48,  49,  50,  51,  52,  53,  54, 55,  56,  57,  58,  59,  60,  61,  62,  63,  64,  65,  66,  67,  68,  69, 70,  71,  72,  73,  74,  75,  76,  76,  77,  78,  79,  80,  81,  82,  83, 84,  85,  86,  87,  88,  89,  91,  93,  95,  96,  98,  100, 101, 102, 104, 106, 108, 110, 112, 114, 116, 118, 122, 124, 126, 128, 130, 132, 134, 136, 138, 140, 143, 145, 148, 151, 154, 157}

var kAcTable = [128]uint16{
    4,   5,   6,   7,   8,   9,   10,  11,  12,  13,  14,  15,  16,  17,  18, 19,  20,  21,  22,  23,  24,  25,  26,  27,  28,  29,  30,  31,  32,  33, 34,  35,  36,  37,  38,  39,  40,  41,  42,  43,  44,  45,  46,  47,  48, 49,  50,  51,  52,  53,  54,  55,  56,  57,  58,  60,  62,  64,  66,  68, 70,  72,  74,  76,  78,  80,  82,  84,  86,  88,  90,  92,  94,  96,  98, 100, 102, 104, 106, 108, 110, 112, 114, 116, 119, 122, 125, 128, 131, 134, 137, 140, 143, 146, 149, 152, 155, 158, 161, 164, 167, 170, 173, 177, 181, 185, 189, 193, 197, 201, 205, 209, 213, 217, 221, 225, 229, 234, 239, 245, 249, 254, 259, 264, 269, 274, 279, 284}

var kAcTable2 = [128]uint16{
    8,   8,   9,   10,  12,  13,  15,  17,  18,  20,  21,  23,  24,  26,  27, 29,  31,  32,  34,  35,  37,  38,  40,  41,  43,  44,  46,  48,  49,  51, 52,  54,  55,  57,  58,  60,  62,  63,  65,  66,  68,  69,  71,  72,  74, 75,  77,  79,  80,  82,  83,  85,  86,  88,  89,  93,  96,  99,  102, 105, 108, 111, 114, 117, 120, 124, 127, 130, 133, 136, 139, 142, 145, 148, 151, 155, 158, 161, 164, 167, 170, 173, 176, 179, 184, 189, 193, 198, 203, 207, 212, 217, 221, 226, 230, 235, 240, 244, 249, 254, 258, 263, 268, 274, 280, 286, 292, 299, 305, 311, 317, 323, 330, 336, 342, 348, 354, 362, 370, 379, 385, 393, 401, 409, 416, 424, 432, 440}

var kBiasMatrices = [3][2]uint8{  // [luma-ac,luma-dc,chroma][dc,ac]
    {96, 110}, {96, 108}, {110, 115}}

// Sharpening by (slightly) raising the hi-frequency coeffs.
// Hack-ish but helpful for mid-bitrate range. Use with care.
const SHARPEN_BITS =11  // number of descaling bits for sharpening bias
var kFreqSharpening = [16]uint8{0,  30, 60, 90, 30, 60, 90, 90, 60, 90, 90, 90, 90, 90, 90, 90}

//------------------------------------------------------------------------------
// Initialize quantization parameters in VP8Matrix

// Returns the average quantizer
func ExpandMatrix(/* const */ m *VP8Matrix, vtype int) int {
  var i, sum int
  for i = 0; i < 2; i++ {
    is_ac_coeff := (i > 0)
    bias := kBiasMatrices[vtype][is_ac_coeff]
    m.iq[i] = (1 << QFIX) / m.q[i]
    m.bias[i] = BIAS(bias)
    // zthresh is the exact value such that QUANTDIV(coeff, iQ, B) is:
    //   * zero if coeff <= zthresh
    //   * non-zero if coeff > zthresh
    m.zthresh[i] = ((1 << QFIX) - 1 - m.bias[i]) / m.iq[i]
  }
  for i = 2; i < 16; i++ {
    m.q[i] = m.q[1]
    m.iq[i] = m.iq[1]
    m.bias[i] = m.bias[1]
    m.zthresh[i] = m.zthresh[1]
  }

  sum = 0
  i = 0

  for ; i < 16; i++ {
    if (vtype == 0) {  // we only use sharpening for AC luma coeffs
      m.sharpen[i] = (kFreqSharpening[i] * m.q[i]) >> SHARPEN_BITS
    } else {
      m.sharpen[i] = 0
    }
    sum += m.q[i]
  }
  return (sum + 8) >> 4
}

func CheckLambdaValue(/* const */ v *int) {
  if *v < 1 { *v = 1 }
}

func SetupMatrices(enc *vp8.VP8Encoder) {
  var i int
  tlambda_scale := tenary.If(enc.method >= 4, enc.config.SnsStrength, 0)
  num_segments := enc.segment_hdr.num_segments
  for i = 0; i < num_segments; i++ {
    var m *VP8SegmentInfo = &enc.dqm[i]
    q := m.quant
    var q_i4, q_i16, q_uv int
    m.y1.q[0] = kDcTable[clip(q + enc.dq_y1_dc, 0, 127)]
    m.y1.q[1] = kAcTable[clip(q, 0, 127)]

    m.y2.q[0] = kDcTable[clip(q + enc.dq_y2_dc, 0, 127)] * 2
    m.y2.q[1] = kAcTable2[clip(q + enc.dq_y2_ac, 0, 127)]

    m.uv.q[0] = kDcTable[clip(q + enc.dq_uv_dc, 0, 117)]
    m.uv.q[1] = kAcTable[clip(q + enc.dq_uv_ac, 0, 127)]

    q_i4 = ExpandMatrix(&m.y1, 0)
    q_i16 = ExpandMatrix(&m.y2, 1)
    q_uv = ExpandMatrix(&m.uv, 2)

    m.lambda_i4 = (3 * q_i4 * q_i4) >> 7
    m.lambda_i16 = (3 * q_i16 * q_i16)
    m.lambda_uv = (3 * q_uv * q_uv) >> 6
    m.lambda_mode = (1 * q_i4 * q_i4) >> 7
    m.lambda_trellis_i4 = (7 * q_i4 * q_i4) >> 3
    m.lambda_trellis_i16 = (q_i16 * q_i16) >> 2
    m.lambda_trellis_uv = (q_uv * q_uv) << 1
    m.tlambda = (tlambda_scale * q_i4) >> 5

    // none of these constants should be < 1
    CheckLambdaValue(&m.lambda_i4)
    CheckLambdaValue(&m.lambda_i16)
    CheckLambdaValue(&m.lambda_uv)
    CheckLambdaValue(&m.lambda_mode)
    CheckLambdaValue(&m.lambda_trellis_i4)
    CheckLambdaValue(&m.lambda_trellis_i16)
    CheckLambdaValue(&m.lambda_trellis_uv)
    CheckLambdaValue(&m.tlambda)

    m.min_disto = 20 * m.y1.q[0];  // quantization-aware min disto
    m.max_edge = 0

    m.i4_penalty = 1000 * q_i4 * q_i4
  }
}



func SetupFilterStrength(/* const */ enc *vp8.VP8Encoder) {
  var i int
  // level0 is in [0..500]. Using '-f 50' as filter_strength is mid-filtering.
  level0 := 5 * enc.config.FilterStrength
  for i = 0; i < NUM_MB_SEGMENTS; i++ {
    var m *VP8SegmentInfo = &enc.dqm[i]
    // We focus on the quantization of AC coeffs.
    qstep := kAcTable[clip(m.quant, 0, 127)] >> 2
    base_strength := VP8FilterStrengthFromDelta(enc.filter_hdr.sharpness, qstep)
    // Segments with lower complexity ('beta') will be less filtered.
    f := base_strength * level0 / (256 + m.beta)
    m.fstrength = tenary.If(f < FSTRENGTH_CUTOFF, 0, tenary.If(f > 63, 63, f))
  }
  // We record the initial strength (mainly for the case of 1-segment only).
  enc.filter_hdr.level = enc.dqm[0].fstrength
  enc.filter_hdr.simple = (enc.config.FilterType == 0)
  enc.filter_hdr.sharpness = enc.config.FilterSharpness
}

//------------------------------------------------------------------------------

// Note: if you change the values below, remember that the max range
// allowed by the syntax for DQ_UV is [-16,16].
const MAX_DQ_UV =(6)
const MIN_DQ_UV =(-4)

// We want to emulate jpeg-like behaviour where the expected "good" quality
// is around q=75. Internally, our "good" middle is around c=50. So we
// map accordingly using linear piece-wise function
func QualityToCompression(c float64) float64 {
  var  linear_c float64 = tenary.If(c < 0.75, c * (2.0 / 3.0), 2.0 * c - 1.0)
  // The file size roughly scales as pow(quantizer, 3.). Actually, the
  // exponent is somewhere between 2.8 and 3.2, but we're mostly interested
  // in the mid-quant range. So we scale the compressibility inversely to
  // this power-law: quant ~= compression ^ 1/3. This law holds well for
  // low quant. Finer modeling for high-quant would make use of kAcTable[]
  // more explicitly.
  var  v float64 = pow(linear_c, 1 / 3.)
  return v
}

func QualityToJPEGCompression(c float64, alpha float64) float64 {
  // We map the complexity 'alpha' and quality setting 'c' to a compression
  // exponent empirically matched to the compression curve of libjpeg6b.
  // On average, the WebP output size will be roughly similar to that of a
  // JPEG file compressed with same quality factor.
  var  amin float64 = 0.30
  var  amax float64 = 0.85
  var  exp_min float64 = 0.4
  var  exp_max float64 = 0.9
  var  slope float64 = (exp_min - exp_max) / (amax - amin)
  // Linearly interpolate 'expn' from exp_min to exp_max
  // in the [amin, amax] range.
  var  expn float64 = (alpha > amax)   ? exp_min
                      : (alpha < amin) ? exp_max
                                       : exp_max + slope * (alpha - amin)
  var v float64 = pow(c, expn)
  return v
}

func SegmentsAreEquivalent(/* const */ S *VP8SegmentInfo1, /*const*/ S *VP8SegmentInfo2) int {
  return (S1.quant == S2.quant) && (S1.0strength == S2.0strength)
}

func SimplifySegments(/* const */ enc *vp8.VP8Encoder) {
  int map[NUM_MB_SEGMENTS] = {0, 1, 2, 3}
  // 'num_segments' is previously validated and <= NUM_MB_SEGMENTS, but an
  // explicit check is needed to afunc a spurious warning about 'i' exceeding
  // array bounds of 'dqm' with some compilers (noticed with gcc-4.9).
  num_segments := (enc.segment_hdr.num_segments < NUM_MB_SEGMENTS)
                               ? enc.segment_hdr.num_segments
                               : NUM_MB_SEGMENTS
  num_final_segments := 1
  int s1, s2
  for s1 = 1; s1 < num_segments; ++s1 {  // find similar segments
    var S *VP8SegmentInfo1 = &enc.dqm[s1]
    found := 0
    // check if we already have similar segment
    for s2 = 0; s2 < num_final_segments; ++s2 {
      var S *VP8SegmentInfo2 = &enc.dqm[s2]
      if (SegmentsAreEquivalent(S1, S2)) {
        found = 1
        break
      }
    }
    map[s1] = s2
    if (!found) {
      if (num_final_segments != s1) {
        enc.dqm[num_final_segments] = enc.dqm[s1]
      }
      num_final_segments++
    }
  }
  if (num_final_segments < num_segments) {  // Remap
    i := enc.mb_w * enc.mb_h
    while (i-- > 0) enc.mb_info[i].segment = map[enc.mb_info[i].segment]
    enc.segment_hdr.num_segments = num_final_segments
    // Replicate the trailing segment infos (it's mostly cosmetics)
    for i = num_final_segments; i < num_segments; i++ {
      enc.dqm[i] = enc.dqm[num_final_segments - 1]
    }
  }
}

// Sets up segment's quantization values, 'base_quant' and filter strengths.
func VP8SetSegmentParams(/* const */ enc *vp8.VP8Encoder, quality float64) {
  var i int
  int dq_uv_ac, dq_uv_dc
  num_segments := enc.segment_hdr.num_segments
  var float64 amp = SNS_TO_DQ * enc.config.SnsStrength / 100. / 128.
  const float64 Q = quality / 100.
  var float64 c_base = enc.config.EmulateJpegSize
                            ? QualityToJPEGCompression(Q, enc.alpha / 255.)
                            : QualityToCompression(Q)
  for i = 0; i < num_segments; i++ {
    // We modulate the base coefficient to accommodate for the quantization
    // susceptibility and allow denser segments to be quantized more.
    var float64 expn = 1.0 - amp * enc.dqm[i].alpha
    var float64 c = pow(c_base, expn)
    q := (int)(127. * (1.0 - c))
    assert.Assert(expn > 0.)
    enc.dqm[i].quant = clip(q, 0, 127)
  }

  // purely indicative in the bitstream (except for the 1-segment case)
  enc.base_quant = enc.dqm[0].quant

  // fill-in values for the unused segments (required by the syntax)
  for i = num_segments; i < NUM_MB_SEGMENTS; i++ {
    enc.dqm[i].quant = enc.base_quant
  }

  // uv_alpha is normally spread around ~60. The useful range is
  // typically ~30 (quite bad) to ~100 (ok to decimate UV more).
  // We map it to the safe maximal range of MAX/MIN_DQ_UV for dq_uv.
  dq_uv_ac = (enc.uv_alpha - MID_ALPHA) * (MAX_DQ_UV - MIN_DQ_UV) /
             (MAX_ALPHA - MIN_ALPHA)
  // we rescale by the user-defined strength of adaptation
  dq_uv_ac = dq_uv_ac * enc.config.SnsStrength / 100
  // and make it safe.
  dq_uv_ac = clip(dq_uv_ac, MIN_DQ_UV, MAX_DQ_UV)
  // We also boost the dc-uv-quant a little, based on sns-strength, since
  // U/V channels are quite more reactive to high quants (flat DC-blocks
  // tend to appear, and are unpleasant).
  dq_uv_dc = -4 * enc.config.SnsStrength / 100
  dq_uv_dc = clip(dq_uv_dc, -15, 15);  // 4bit-signed max allowed

  enc.dq_y1_dc = 0;  // TODO(skal): dq-lum
  enc.dq_y2_dc = 0
  enc.dq_y2_ac = 0
  enc.dq_uv_dc = dq_uv_dc
  enc.dq_uv_ac = dq_uv_ac

  SetupFilterStrength(enc);  // initialize segments' filtering, eventually

  if num_segments > 1 { SimplifySegments(enc) }

  SetupMatrices(enc);  // finalize quantization matrices
}

//------------------------------------------------------------------------------
// Form the predictions in cache

// Must be ordered using {DC_PRED, TM_PRED, V_PRED, H_PRED} as index
const VP8I16ModeOffsets = [4]uint16 {I16DC16, I16TM16, I16VE16, I16HE16}
const VP8UVModeOffsets = [4]uint16 {C8DC8, C8TM8, C8VE8, C8HE8}

// Must be indexed using {B_DC_PRED . B_HU_PRED} as index
const VP8I4ModeOffsets = [NUM_BMODES]uint16{
    I4DC4, I4TM4, I4VE4, I4HE4, I4RD4, I4VR4, I4LD4, I4VL4, I4HD4, I4HU4}

	//Form all the four Intra16x16 predictions in the 'yuv_p' cache
func VP8MakeLuma16Preds(/* const */ it *vp8.VP8EncIterator) {
  var left *uint8 = it.x ? it.y_left : nil
  var top *uint8 = it.y ? it.y_top : nil
  VP8EncPredLuma16(it.yuv_p, left, top)
}

// Form all the four Chroma8x8 predictions in the 'yuv_p' cache
func VP8MakeChroma8Preds(/* const */ it *vp8.VP8EncIterator) {
  var left *uint8 = it.x ? it.u_left : nil
  var top *uint8 = it.y ? it.uv_top : nil
  VP8EncPredChroma8(it.yuv_p, left, top)
}

// Form all the ten Intra4x4 predictions in the 'yuv_p' cache
// for the 4x4 block it.i4
func MakeIntra4Preds(/* const */ it *vp8.VP8EncIterator) {
  VP8EncPredLuma4(it.yuv_p, it.i4_top)
}

//------------------------------------------------------------------------------
// Quantize

// Layout:
// +----+----+
// |YYYY|UUVV| 0
// |YYYY|UUVV| 4
// |YYYY|....| 8
// |YYYY|....| 12
// +----+----+

const uint16 VP8Scan[16] = {
    // Luma
    0 + 0 * BPS,  4 + 0 * BPS,  8 + 0 * BPS,  12 + 0 * BPS, 0 + 4 * BPS,  4 + 4 * BPS,  8 + 4 * BPS,  12 + 4 * BPS, 0 + 8 * BPS,  4 + 8 * BPS,  8 + 8 * BPS,  12 + 8 * BPS, 0 + 12 * BPS, 4 + 12 * BPS, 8 + 12 * BPS, 12 + 12 * BPS,
}

static const uint16 VP8ScanUV[4 + 4] = {
    0 + 0 * BPS, 4 + 0 * BPS,  0 + 4 * BPS, 4 + 4 * BPS,  // U
    8 + 0 * BPS, 12 + 0 * BPS, 8 + 4 * BPS, 12 + 4 * BPS  // V
}

//------------------------------------------------------------------------------
// Distortion measurement

var kWeightY = [16]uint16 = {38, 32, 20, 9, 32, 28, 17, 7, 20, 17, 10, 4, 9,  7,  4,  2}

var kWeightTrellis = [16]uint16 = {
#if USE_TDISTO == 0
    16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16
#else
    30, 27, 19, 11, 27, 24, 17, 10, 19, 17, 12, 8,  11, 10, 8,  6
#endif
}

// Init/Copy the common fields in score.
func InitScore(/* const */ rd *VP8ModeScore) {
  rd.D = 0
  rd.SD = 0
  rd.R = 0
  rd.H = 0
  rd.nz = 0
  rd.score = MAX_COST
}

func CopyScore(/* const */ dst *VP8ModeScore, /*const*/ /* const */ src *VP8ModeScore) {
  dst.D = src.D
  dst.SD = src.SD
  dst.R = src.R
  dst.H = src.H
  dst.nz = src.nz;  // note that nz is not accumulated, but just copied.
  dst.score = src.score
}

func AddScore(/* const */ dst *VP8ModeScore, /*const*/ /* const */ src *VP8ModeScore) {
  dst.D += src.D
  dst.SD += src.SD
  dst.R += src.R
  dst.H += src.H
  dst.nz |= src.nz;  // here, new nz bits are accumulated.
  dst.score += src.score
}

//------------------------------------------------------------------------------
// Performs trellis-optimized quantization.

// Trellis node
type Node struct {
   prev int8    // best previous node
   sign int8    // sign of coeff_i
   level int16  // level
}

// Score state
type ScoreState struct {
  score_t score;          // partial RD score
  costs *uint16;  // shortcut to cost tables
}

// If a coefficient was quantized to a value Q (using a neutral bias),
// we test all alternate possibilities between [Q-MIN_DELTA, Q+MAX_DELTA]
// We don't test negative values though.
const MIN_DELTA =0  // how much lower level to try
const MAX_DELTA =1  // how much higher
const NUM_NODES =(MIN_DELTA + 1 + MAX_DELTA)
#define NODE(n, l) (nodes[(n)][(l) + MIN_DELTA])
#define SCORE_STATE(n, l) (score_states[n][(l) + MIN_DELTA])

func SetRDScore(int lambda, /*const*/ rd *VP8ModeScore) {
  rd.score = (rd.R + rd.H) * lambda + RD_DISTO_MULT * (rd.D + rd.SD)
}

func RDScoreTrellis(int lambda, score_t rate, score_t distortion) score_t {
  return rate * lambda + RD_DISTO_MULT * distortion
}

// Coefficient type.
enum { TYPE_I16_AC = 0, TYPE_I16_DC = 1, TYPE_CHROMA_A = 2, TYPE_I4_AC = 3 }

func TrellisQuantizeBlock(/* const */ /* const */ enc *vp8.VP8Encoder, int16 in[16], int16 out[16], int ctx0, coeff_type int, /*const*/ /* const */ mtx *VP8Matrix, lambda int) int {
  var probas *ProbaArray = enc.proba.coeffs[coeff_type]
  CostArrayPtr const costs = (CostArrayPtr)enc.proba.remapped_costs[coeff_type]
  first := tenary.If(coeff_type == TYPE_I16_AC, 1, 0)
  Node nodes[16][NUM_NODES]
  ScoreState score_states[2][NUM_NODES]
  ss_cur *ScoreState = &SCORE_STATE(0, MIN_DELTA)
  ss_prev *ScoreState = &SCORE_STATE(1, MIN_DELTA)
  int best_path[3] = {-1, -1, -1}  // store best-last/best-level/best-previous
  score_t best_score
  int n, m, p, last

  {
    score_t cost
    thresh := mtx.q[1] * mtx.q[1] / 4
    last_proba := probas[VP8EncBands[first]][ctx0][0]

    // compute the position of the last interesting coefficient
    last = first - 1
    for n = 15; n >= first; --n {
      j := kZigzag[n]
      err := in[j] * in[j]
      if (err > thresh) {
        last = n
        break
      }
    }
    // we don't need to go inspect up to n = 16 coeffs. We can just go up
    // to last + 1 (inclusive) without losing much.
    if (last < 15) {last++}

    // compute 'skip' score. This is the max score one can do.
    cost = VP8BitCost(0, last_proba)
    best_score = RDScoreTrellis(lambda, cost, 0)

    // initialize source node.
    for m = -MIN_DELTA; m <= MAX_DELTA; m++ {
       = tenary.If(ctx0 == 0, VP8BitCost(1, last_proba), 0)
      ss_cur[m].score = RDScoreTrellis(lambda, rate, 0)
      ss_cur[m].costs = costs[first][ctx0]
    }
  }

  // traverse trellis.
  for n = first; n <= last; n++ {
    j := kZigzag[n]
    Q := mtx.q[j]
    iQ := mtx.iq[j]
    B := BIAS(0x00);  // neutral bias
    // note: it's important to take sign of the _original_ coeff, // so we don't have to consider level < 0 afterward.
    sign := (in[j] < 0)
    coeff0 := (sign ? -in[j] : in[j]) + mtx.sharpen[j]
    level int0 = QUANTDIV(coeff0, iQ, B)
    thresh_level := QUANTDIV(coeff0, iQ, BIAS(0x80))
    if thresh_level > MAX_LEVEL { thresh_level = MAX_LEVEL }
    if level0 > MAX_LEVEL { level0 = MAX_LEVEL }

    {  // Swap current and previous score states
      var tmp *ScoreState = ss_cur
      ss_cur = ss_prev
      ss_prev = tmp
    }

    // test all alternate level values around level0.
    for m = -MIN_DELTA; m <= MAX_DELTA; m++ {
      var cur *Node = &NODE(n, m)
      level := level0 + m
      ctx := tenary.If(level > 2, 2, level)
      band := VP8EncBands[n + 1]
      score_t base_score
      score_t best_cur_score
      var best_prev int
      score_t cost, score

      // costs is [16][NUM_CTX == 3] but ss_cur[m].costs is only read after
      // being swapped with ss_prev: the last value can be nil.
      if (n + 1 < 16) {
        ss_cur[m].costs = costs[n + 1][ctx]
      } else {
        ss_cur[m].costs = nil
      }
      if (level < 0 || level > thresh_level) {
        ss_cur[m].score = MAX_COST
        // Node is dead.
        continue
      }

      {
        // Compute delta_error = how much coding this level will
        // subtract to max_error as distortion.
        // Here, distortion = sum of (|coeff_i| - level_i * Q_i)^2
        new_error := coeff0 - level * Q
        delta_error := kWeightTrellis[j] * (new_error * new_error - coeff0 * coeff0)
        base_score = RDScoreTrellis(lambda, 0, delta_error)
      }

      // Inspect all possible non-dead predecessors. Retain only the best one.
      // The base_score is added to all scores so it is only added for the final
      // value after the loop.
      cost = VP8LevelCost(ss_prev[-MIN_DELTA].costs, level)
      best_cur_score = ss_prev[-MIN_DELTA].score + RDScoreTrellis(lambda, cost, 0)
      best_prev = -MIN_DELTA
      for p = -MIN_DELTA + 1; p <= MAX_DELTA; p++ {
        // Dead nodes (with ss_prev[p].score >= MAX_COST) are automatically
        // eliminated since their score can't be better than the current best.
        cost = VP8LevelCost(ss_prev[p].costs, level)
        // Examine node assuming it's a non-terminal one.
        score = ss_prev[p].score + RDScoreTrellis(lambda, cost, 0)
        if (score < best_cur_score) {
          best_cur_score = score
          best_prev = p
        }
      }
      best_cur_score += base_score
      // Store best finding in current node.
      cur.sign = sign
      cur.level = level
      cur.prev = best_prev
      ss_cur[m].score = best_cur_score

      // Now, record best terminal node (and thus best entry in the graph).
      if (level != 0 && best_cur_score < best_score) {
         = tenary.If(n < 15, VP8BitCost(0, probas[band][ctx][0]), 0)
         = RDScoreTrellis(lambda, last_pos_cost, 0)
        score = best_cur_score + last_pos_score
        if (score < best_score) {
          best_score = score
          best_path[0] = n;          // best eob position
          best_path[1] = m;          // best node index
          best_path[2] = best_prev;  // best predecessor
        }
      }
    }
  }

  // Fresh start
  // Beware! We must preserve in[0]/out[0] value for TYPE_I16_AC case.
  if (coeff_type == TYPE_I16_AC) {
    stdlib.Memset(in + 1, 0, 15 * sizeof(*in))
    stdlib.Memset(out + 1, 0, 15 * sizeof(*out))
  } else {
    stdlib.Memset(in, 0, 16 * sizeof(*in))
    stdlib.Memset(out, 0, 16 * sizeof(*out))
  }
  if (best_path[0] == -1) {
    return 0;  // skip!
  }

  {
    // Unwind the best path.
    // Note: best-prev on terminal node is not necessarily equal to the
    // best_prev for non-terminal. So we patch best_path[2] in.
    nz := 0
    best_node := best_path[1]
    n = best_path[0]
    NODE(n, best_node).prev = best_path[2];  // force best-prev for terminal

    for ; n >= first; --n {
      var node *Node = &NODE(n, best_node)
      j := kZigzag[n]
      out[n] = node.sign ? -node.level : node.level
      nz |= node.level
      in[j] = out[n] * mtx.q[j]
      best_node = node.prev
    }
    return (nz != 0)
  }
}

#undef NODE

//------------------------------------------------------------------------------
// Performs: difference, transform, quantize, back-transform, add
// all at once. Output is the reconstructed block in *yuv_out, and the
// quantized levels in *levels.

func ReconstructIntra16(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore, /* const */ yuv_out *uint8, mode int) int {
  var enc *vp8.VP8Encoder = it.enc
  var ref *uint8 = it.yuv_p + VP8I16ModeOffsets[mode]
  var src *uint8 = it.yuv_in + Y_OFF_ENC
  var dqm *VP8SegmentInfo = &enc.dqm[it.mb.segment]
  nz := 0
  var n int
  int16 tmp[16][16], dc_tmp[16]

  for n = 0; n < 16; n += 2 {
    VP8FTransform2(src + VP8Scan[n], ref + VP8Scan[n], tmp[n])
  }
  VP8FTransformWHT(tmp[0], dc_tmp)
  nz |= VP8EncQuantizeBlockWHT(dc_tmp, rd.y_dc_levels, &dqm.y2) << 24

  if (DO_TRELLIS_I16 && it.do_trellis) {
    var x, y int
    VP8IteratorNzToBytes(it)
    for y = 0, n = 0; y < 4; y++ {
      for x = 0; x < 4; ++x, n++ {
        ctx := it.top_nz[x] + it.left_nz[y]
        non_zero := TrellisQuantizeBlock(
            enc, tmp[n], rd.y_ac_levels[n], ctx, TYPE_I16_AC, &dqm.y1, dqm.lambda_trellis_i16)
        it.top_nz[x] = it.left_nz[y] = non_zero
        rd.y_ac_levels[n][0] = 0
        nz |= non_zero << n
      }
    }
  } else {
    for n = 0; n < 16; n += 2 {
      // Zero-out the first coeff, so that: a) nz is correct below, and
      // b) finding 'last' non-zero coeffs in SetResidualCoeffs() is simplified.
      tmp[n][0] = tmp[n + 1][0] = 0
      nz |= VP8EncQuantize2Blocks(tmp[n], rd.y_ac_levels[n], &dqm.y1) << n
      assert.Assert(rd.y_ac_levels[n + 0][0] == 0)
      assert.Assert(rd.y_ac_levels[n + 1][0] == 0)
    }
  }

  // Transform back
  VP8TransformWHT(dc_tmp, tmp[0])
  for n = 0; n < 16; n += 2 {
    VP8ITransform(ref + VP8Scan[n], tmp[n], yuv_out + VP8Scan[n], 1)
  }

  return nz
}

func ReconstructIntra4(/* const */ it *vp8.VP8EncIterator, int16 levels[16], /*const*/ /* const */ src *uint8, /* const */ yuv_out *uint8, mode int) int {
  var enc *vp8.VP8Encoder = it.enc
  var ref *uint8 = it.yuv_p + VP8I4ModeOffsets[mode]
  var dqm *VP8SegmentInfo = &enc.dqm[it.mb.segment]
  nz := 0
  int16 tmp[16]

  VP8FTransform(src, ref, tmp)
  if (DO_TRELLIS_I4 && it.do_trellis) {
    x := it.i4 & 3, y = it.i4 >> 2
    ctx := it.top_nz[x] + it.left_nz[y]
    nz = TrellisQuantizeBlock(enc, tmp, levels, ctx, TYPE_I4_AC, &dqm.y1, dqm.lambda_trellis_i4)
  } else {
    nz = VP8EncQuantizeBlock(tmp, levels, &dqm.y1)
  }
  VP8ITransform(ref, tmp, yuv_out, 0)
  return nz
}

//------------------------------------------------------------------------------
// DC-error diffusion

// Diffusion weights. We under-correct a bit (15/16th of the error is actually
// diffused) to afunc 'rainbow' chessboard pattern of blocks at q~=0.
const C1 = 7  // fraction of error sent to the 4x4 block below
const C2 = 8  // fraction of error sent to the 4x4 block on the right
const DSHIFT = 4
const DSCALE = 1  // storage descaling, needed to make the error fit int8

// Quantize as usual, but also compute and return the quantization error.
// Error is already divided by DSHIFT.
func QuantizeSingle(/* const */ v *int16, /*const*/ /* const */ mtx *VP8Matrix) int {
  V := *v
  sign := (V < 0)
  if sign { V = -V }
  if (V > (int)mtx.zthresh[0]) {
    qV := QUANTDIV(V, mtx.iq[0], mtx.bias[0]) * mtx.q[0]
    err := (V - qV)
    *v = sign ? -qV : qV
    return (sign ? -err : err) >> DSCALE
  }
  *v = 0
  return (sign ? -V : V) >> DSCALE
}

func CorrectDCValues(/* const */ /* const */ it *vp8.VP8EncIterator, /*const*/ /* const */ mtx *VP8Matrix, int16 tmp[][16], /* const */ rd *VP8ModeScore) {
  //         | top[0] | top[1]
  // --------+--------+---------
  // left[0] | tmp[0]   tmp[1]  <.   err0 err1
  // left[1] | tmp[2]   tmp[3]        err2 err3
  //
  // Final errors {err1,err2,err3} are preserved and later restored
  // as top[]/left[] on the next block.
  var ch int
  for ch = 0; ch <= 1; ch++ {
    var top *int8 = it.top_derr[it.x][ch]
    var left *int8 = it.left_derr[ch]
    int16(c *const)[16] = &tmp[ch * 4]
    int err0, err1, err2, err3
    c[0][0] += (C1 * top[0] + C2 * left[0]) >> (DSHIFT - DSCALE)
    err0 = QuantizeSingle(&c[0][0], mtx)
    c[1][0] += (C1 * top[1] + C2 * err0) >> (DSHIFT - DSCALE)
    err1 = QuantizeSingle(&c[1][0], mtx)
    c[2][0] += (C1 * err0 + C2 * left[1]) >> (DSHIFT - DSCALE)
    err2 = QuantizeSingle(&c[2][0], mtx)
    c[3][0] += (C1 * err1 + C2 * err2) >> (DSHIFT - DSCALE)
    err3 = QuantizeSingle(&c[3][0], mtx)
    // error 'err' is bounded by mtx.q[0] which is 132 at max. Hence
    // err >> DSCALE will fit in an int8 type if DSCALE>=1.
    assert.Assert(abs(err1) <= 127 && abs(err2) <= 127 && abs(err3) <= 127)
    rd.derr[ch][0] = (int8)err1
    rd.derr[ch][1] = (int8)err2
    rd.derr[ch][2] = (int8)err3
  }
}

func StoreDiffusionErrors(/* const */ it *vp8.VP8EncIterator, /*const*/ /* const */ rd *VP8ModeScore) {
  var ch int
  for ch = 0; ch <= 1; ch++ {
    var top *int8 = it.top_derr[it.x][ch]
    var left *int8 = it.left_derr[ch]
    left[0] = rd.derr[ch][0];           // restore err1
    left[1] = 3 * rd.derr[ch][2] >> 2;  //     ... 3/4th of err3
    top[0] = rd.derr[ch][1];            //     ... err2
    top[1] = rd.derr[ch][2] - left[1];  //     ... 1/4th of err3.
  }
}

#undef C1
#undef C2
#undef DSHIFT
#undef DSCALE

//------------------------------------------------------------------------------

func ReconstructUV(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore, /* const */ yuv_out *uint8, mode int) int {
  var enc *vp8.VP8Encoder = it.enc
  var ref *uint8 = it.yuv_p + VP8UVModeOffsets[mode]
  var src *uint8 = it.yuv_in + U_OFF_ENC
  var dqm *VP8SegmentInfo = &enc.dqm[it.mb.segment]
  nz := 0
  var n int
  int16 tmp[8][16]

  for n = 0; n < 8; n += 2 {
    VP8FTransform2(src + VP8ScanUV[n], ref + VP8ScanUV[n], tmp[n])
  }
  if it.top_derr != nil { CorrectDCValues(it, &dqm.uv, tmp, rd) }

  if (DO_TRELLIS_UV && it.do_trellis) {
    int ch, x, y
    for ch = 0, n = 0; ch <= 2; ch += 2 {
      for y = 0; y < 2; y++ {
        for x = 0; x < 2; ++x, n++ {
          ctx := it.top_nz[4 + ch + x] + it.left_nz[4 + ch + y]
          non_zero := TrellisQuantizeBlock(
              enc, tmp[n], rd.uv_levels[n], ctx, TYPE_CHROMA_A, &dqm.uv, dqm.lambda_trellis_uv)
          it.top_nz[4 + ch + x] = it.left_nz[4 + ch + y] = non_zero
          nz |= non_zero << n
        }
      }
    }
  } else {
    for n = 0; n < 8; n += 2 {
      nz |= VP8EncQuantize2Blocks(tmp[n], rd.uv_levels[n], &dqm.uv) << n
    }
  }

  for n = 0; n < 8; n += 2 {
    VP8ITransform(ref + VP8ScanUV[n], tmp[n], yuv_out + VP8ScanUV[n], 1)
  }
  return (nz << 16)
}

//------------------------------------------------------------------------------
// RD-opt decision. Reconstruct each modes, evalue distortion and bit-cost.
// Pick the mode is lower RD-cost = Rate + lambda * Distortion.

func StoreMaxDelta(/* const */ dqm *VP8SegmentInfo, /*const*/ int16 DCs[16]) {
  // We look at the first three AC coefficients to determine what is the average
  // delta between each sub-4x4 block.
  v0 := stdlib.Abs(DCs[1])
  v1 := stdlib.Abs(DCs[2])
  v2 := stdlib.Abs(DCs[4])
  max_v := tenary.If(v1 > v0, v1, v0)
  max_v = tenary.If(v2 > max_v, v2, max_v)
  if max_v > dqm.max_edge { dqm.max_edge = max_v }
}

func SwapOut(/* const */ it *vp8.VP8EncIterator) {
  ptr.Swap(&it.yuv_out, &it.yuv_out2)
}

func PickBestIntra16(/* const */ it *vp8.VP8EncIterator, rd *VP8ModeScore) {
  kNumBlocks := 16
  var dqm *VP8SegmentInfo = &it.enc.dqm[it.mb.segment]
  lambda := dqm.lambda_i16
  tlambda := dqm.tlambda
  var src *uint8 = it.yuv_in + Y_OFF_ENC
   var rd_tmp VP8ModeScore
  rd_cur *VP8ModeScore = &rd_tmp
  rd_best *VP8ModeScore = rd
  var mode int
  is_flat := IsFlatSource16(it.yuv_in + Y_OFF_ENC)

  rd.mode_i16 = -1
  for mode = 0; mode < NUM_PRED_MODES; mode++ {
    var tmp_dst []uint8 = it.yuv_out2 + Y_OFF_ENC;  // scratch buffer
    rd_cur.mode_i16 = mode

    // Reconstruct
    rd_cur.nz = ReconstructIntra16(it, rd_cur, tmp_dst, mode)

    // Measure RD-score
    rd_cur.D = VP8SSE16x16(src, tmp_dst)
    rd_cur.SD = tlambda ? MULT_8B(tlambda, VP8TDisto16x16(src, tmp_dst, kWeightY)) : 0
    rd_cur.H = VP8FixedCostsI16[mode]
    rd_cur.R = VP8GetCostLuma16(it, rd_cur)
    if (is_flat) {
      // refine the first impression (which was in pixel space)
      is_flat = IsFlat(rd_cur.y_ac_levels[0], kNumBlocks, FLATNESS_LIMIT_I16)
      if (is_flat) {
        // Block is very flat. We put emphasis on the distortion being very low!
        rd_cur.D *= 2
        rd_cur.SD *= 2
      }
    }

    // Since we always examine Intra16 first, we can overwrite directly *rd.
    SetRDScore(lambda, rd_cur)
    if (mode == 0 || rd_cur.score < rd_best.score) {
      ptr.Swap(&rd_cur, &rd_best)
      SwapOut(it)
    }
  }
  if (rd_best != rd) {
    stdlib.MemCpy(rd, rd_best, sizeof(*rd))
  }
  SetRDScore(dqm.lambda_mode, rd);  // finalize score for mode decision.
  VP8SetIntra16Mode(it, rd.mode_i16)

  // we have a blocky macroblock (only DCs are non-zero) with fairly high
  // distortion, record max delta so we can later adjust the minimal filtering
  // strength needed to smooth these blocks out.
  if ((rd.nz & 0x100ffff) == 0x1000000 && rd.D > dqm.min_disto) {
    StoreMaxDelta(dqm, rd.y_dc_levels)
  }
}

//------------------------------------------------------------------------------

// return the cost array corresponding to the surrounding prediction modes.
static const GetCostModeI *uint164(/* const */ it *vp8.VP8EncIterator, /*const*/ uint8 modes[16]) {
  preds_w := it.enc.preds_w
  x := (it.i4 & 3), y = it.i4 >> 2
  left := tenary.If(x == 0, it.preds[y * preds_w - 1], modes[it.i4 - 1])
  top := tenary.If(y == 0, it.preds[-preds_w + x], modes[it.i4 - 4])
  return VP8FixedCostsI4[top][left]
}

func PickBestIntra4(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore) int {
  var enc *vp8.VP8Encoder = it.enc
  var dqm *VP8SegmentInfo = &enc.dqm[it.mb.segment]
  lambda := dqm.lambda_i4
  tlambda := dqm.tlambda
  var src *uint80 = it.yuv_in + Y_OFF_ENC
  var best_blocks *uint8 = it.yuv_out2 + Y_OFF_ENC
  total_header_bits := 0
   var rd_best VP8ModeScore

  if (enc.max_i4_header_bits == 0) {
    return 0
  }

  InitScore(&rd_best)
  rd_best.H = 211;  // '211' is the value of VP8BitCost(0, 145)
  SetRDScore(dqm.lambda_mode, &rd_best)
  VP8IteratorStartI4(it)
  for {
    kNumBlocks := 1
     var rd_i4 VP8ModeScore
    var mode int
    best_mode := -1
    var src *uint8 = src0 + VP8Scan[it.i4]
    var mode_costs *uint16 = GetCostModeI4(it, rd.modes_i4)
    best_block *uint8 = best_blocks + VP8Scan[it.i4]
    tmp_dst []uint8 = it.yuv_p + I4TMP;  // scratch buffer.

    InitScore(&rd_i4)
    MakeIntra4Preds(it)
    for mode = 0; mode < NUM_BMODES; mode++ {
       var rd_tmp VP8ModeScore
      int16 tmp_levels[16]

      // Reconstruct
      rd_tmp.nz = ReconstructIntra4(it, tmp_levels, src, tmp_dst, mode)
                  << it.i4

      // Compute RD-score
      rd_tmp.D = VP8SSE4x4(src, tmp_dst)
      rd_tmp.SD = tlambda ? MULT_8B(tlambda, VP8TDisto4x4(src, tmp_dst, kWeightY)) : 0
      rd_tmp.H = mode_costs[mode]

      // Add flatness penalty, to afunc flat area to be mispredicted
      // by a complex mode.
      if (mode > 0 && IsFlat(tmp_levels, kNumBlocks, FLATNESS_LIMIT_I4)) {
        rd_tmp.R = FLATNESS_PENALTY * kNumBlocks
      } else {
        rd_tmp.R = 0
      }

      // early-out check
      SetRDScore(lambda, &rd_tmp)
      if best_mode >= 0 && rd_tmp.score >= rd_i4.score { continue }

      // finish computing score
      rd_tmp.R += VP8GetCostLuma4(it, tmp_levels)
      SetRDScore(lambda, &rd_tmp)

      if (best_mode < 0 || rd_tmp.score < rd_i4.score) {
        CopyScore(&rd_i4, &rd_tmp)
        best_mode = mode
        ptr.Swap(&tmp_dst, &best_block)
        stdlib.MemCpy(rd_best.y_ac_levels[it.i4], tmp_levels, sizeof(rd_best.y_ac_levels[it.i4]))
      }
    }
    SetRDScore(dqm.lambda_mode, &rd_i4)
    AddScore(&rd_best, &rd_i4)
    if (rd_best.score >= rd.score) {
      return 0
    }
    total_header_bits += (int)rd_i4.H;  // <- equal to mode_costs[best_mode]
    if (total_header_bits > enc.max_i4_header_bits) {
      return 0
    }
    // Copy selected samples if not in the right place already.
    if (best_block != best_blocks + VP8Scan[it.i4]) {
      VP8Copy4x4(best_block, best_blocks + VP8Scan[it.i4])
    }
    rd.modes_i4[it.i4] = best_mode
    it.top_nz[it.i4 & 3] = it.left_nz[it.i4 >> 2] = (rd_i4.nz ? 1 : 0)
  } while (VP8IteratorRotateI4(it, best_blocks))

  // finalize state
  CopyScore(rd, &rd_best)
  VP8SetIntra4Mode(it, rd.modes_i4)
  SwapOut(it)
  stdlib.MemCpy(rd.y_ac_levels, rd_best.y_ac_levels, sizeof(rd.y_ac_levels))
  return 1;  // select intra4x4 over intra16x16
}

//------------------------------------------------------------------------------

func PickBestUV(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore) {
  kNumBlocks := 8
  var dqm *VP8SegmentInfo = &it.enc.dqm[it.mb.segment]
  lambda := dqm.lambda_uv
  var src *uint8 = it.yuv_in + U_OFF_ENC
  tmp_dst []uint8 = it.yuv_out2 + U_OFF_ENC;  // scratch buffer
  dst []uint80 = it.yuv_out + U_OFF_ENC
  dst []uint8 = dst0
   var rd_best VP8ModeScore
  var mode int

  rd.mode_uv = -1
  InitScore(&rd_best)
  for mode = 0; mode < NUM_PRED_MODES; mode++ {
     var rd_uv VP8ModeScore

    // Reconstruct
    rd_uv.nz = ReconstructUV(it, &rd_uv, tmp_dst, mode)

    // Compute RD-score
    rd_uv.D = VP8SSE16x8(src, tmp_dst)
    rd_uv.SD = 0;  // not calling TDisto here: it tends to flatten areas.
    rd_uv.H = VP8FixedCostsUV[mode]
    rd_uv.R = VP8GetCostUV(it, &rd_uv)
    if (mode > 0 && IsFlat(rd_uv.uv_levels[0], kNumBlocks, FLATNESS_LIMIT_UV)) {
      rd_uv.R += FLATNESS_PENALTY * kNumBlocks
    }

    SetRDScore(lambda, &rd_uv)
    if (mode == 0 || rd_uv.score < rd_best.score) {
      CopyScore(&rd_best, &rd_uv)
      rd.mode_uv = mode
      stdlib.MemCpy(rd.uv_levels, rd_uv.uv_levels, sizeof(rd.uv_levels))
      if (it.top_derr != nil) {
        stdlib.MemCpy(rd.derr, rd_uv.derr, sizeof(rd_uv.derr))
      }
      ptr.Swap(&dst, &tmp_dst)
    }
  }
  VP8SetIntraUVMode(it, rd.mode_uv)
  AddScore(rd, &rd_best)
  if (dst != dst0) {  // copy 16x8 block if needed
    VP8Copy16x8(dst, dst0)
  }
  if (it.top_derr != nil) {  // store diffusion errors for next block
    StoreDiffusionErrors(it, rd)
  }
}

//------------------------------------------------------------------------------
// Final reconstruction and quantization.

func SimpleQuantize(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore) {
  var enc *vp8.VP8Encoder = it.enc
  is_i16 = (it.mb.type :== 1)
  nz := 0

  if (is_i16) {
    nz = ReconstructIntra16(it, rd, it.yuv_out + Y_OFF_ENC, it.preds[0])
  } else {
    VP8IteratorStartI4(it)
    for {
      mode := it.preds[(it.i4 & 3) + (it.i4 >> 2) * enc.preds_w]
      var src *uint8 = it.yuv_in + Y_OFF_ENC + VP8Scan[it.i4]
      var dst []uint8 = it.yuv_out + Y_OFF_ENC + VP8Scan[it.i4]
      MakeIntra4Preds(it)
      nz |= ReconstructIntra4(it, rd.y_ac_levels[it.i4], src, dst, mode)
            << it.i4
    } while (VP8IteratorRotateI4(it, it.yuv_out + Y_OFF_ENC))
  }

  nz |= ReconstructUV(it, rd, it.yuv_out + U_OFF_ENC, it.mb.uv_mode)
  rd.nz = nz
}

// Refine intra16/intra4 sub-modes based on distortion only (not rate).
func RefineUsingDistortion(/* const */ it *vp8.VP8EncIterator, try_both_modes int, refine_uv_mode int, /* const */ rd *VP8ModeScore) {
  score_t best_score = MAX_COST
  nz := 0
  var mode int
  int is_i16 = try_both_modes || (it.mb.type == 1)

  var dqm *VP8SegmentInfo = &it.enc.dqm[it.mb.segment]
  // Some empiric constants, of approximate order of magnitude.
  lambda_d_i16 := 106
  lambda_d_i4 := 11
  lambda_d_uv := 120
  score_t score_i4 = dqm.i4_penalty
  score_t i4_bit_sum = 0
   = try_both_modes ? it.enc.mb_header_limit
                                           : MAX_COST;  // no early-out allowed

  if (is_i16) {  // First, evaluate Intra16 distortion
    best_mode := -1
    var src *uint8 = it.yuv_in + Y_OFF_ENC
    for mode = 0; mode < NUM_PRED_MODES; mode++ {
      var ref *uint8 = it.yuv_p + VP8I16ModeOffsets[mode]
       = (score_t)VP8SSE16x16(src, ref) * RD_DISTO_MULT +
                            VP8FixedCostsI16[mode] * lambda_d_i16
      if (mode > 0 && VP8FixedCostsI16[mode] > bit_limit) {
        continue
      }

      if (score < best_score) {
        best_mode = mode
        best_score = score
      }
    }
    if (it.x == 0 || it.y == 0) {
      // afunc starting a checkerboard resonance from the border. See bug #432.
      if (IsFlatSource16(src)) {
        best_mode = (it.x == 0) ? 0 : 2
        try_both_modes = 0;  // stick to i16
      }
    }
    VP8SetIntra16Mode(it, best_mode)
    // we'll reconstruct later, if i16 mode actually gets selected
  }

  // Next, evaluate Intra4
  if (try_both_modes || !is_i16) {
    // We don't evaluate the rate here, but just account for it through a
    // constant penalty (i4 mode usually needs more bits compared to i16).
    is_i16 = 0
    VP8IteratorStartI4(it)
    for {
      int best_i4_mode = -1
      score_t best_i4_score = MAX_COST
      var src *uint8 = it.yuv_in + Y_OFF_ENC + VP8Scan[it.i4]
      var mode_costs *uint16 = GetCostModeI4(it, rd.modes_i4)

      MakeIntra4Preds(it)
      for mode = 0; mode < NUM_BMODES; mode++ {
        var ref *uint8 = it.yuv_p + VP8I4ModeOffsets[mode]
         = VP8SSE4x4(src, ref) * RD_DISTO_MULT +
                              mode_costs[mode] * lambda_d_i4
        if (score < best_i4_score) {
          best_i4_mode = mode
          best_i4_score = score
        }
      }
      i4_bit_sum += mode_costs[best_i4_mode]
      rd.modes_i4[it.i4] = best_i4_mode
      score_i4 += best_i4_score
      if (score_i4 >= best_score || i4_bit_sum > bit_limit) {
        // Intra4 won't be better than Intra16. Bail out and pick Intra16.
        is_i16 = 1
        break
      } else {  // reconstruct partial block inside yuv_out2 buffer
        var tmp_dst []uint8 = it.yuv_out2 + Y_OFF_ENC + VP8Scan[it.i4]
        nz |= ReconstructIntra4(it, rd.y_ac_levels[it.i4], src, tmp_dst, best_i4_mode)
              << it.i4
      }
    } while (VP8IteratorRotateI4(it, it.yuv_out2 + Y_OFF_ENC))
  }

  // Final reconstruction, depending on which mode is selected.
  if (!is_i16) {
    VP8SetIntra4Mode(it, rd.modes_i4)
    SwapOut(it)
    best_score = score_i4
  } else {
    nz = ReconstructIntra16(it, rd, it.yuv_out + Y_OFF_ENC, it.preds[0])
  }

  // ... and UV!
  if (refine_uv_mode) {
    best_mode := -1
    score_t best_uv_score = MAX_COST
    var src *uint8 = it.yuv_in + U_OFF_ENC
    for mode = 0; mode < NUM_PRED_MODES; mode++ {
      var ref *uint8 = it.yuv_p + VP8UVModeOffsets[mode]
       = VP8SSE16x8(src, ref) * RD_DISTO_MULT +
                            VP8FixedCostsUV[mode] * lambda_d_uv
      if (score < best_uv_score) {
        best_mode = mode
        best_uv_score = score
      }
    }
    VP8SetIntraUVMode(it, best_mode)
  }
  nz |= ReconstructUV(it, rd, it.yuv_out + U_OFF_ENC, it.mb.uv_mode)

  rd.nz = nz
  rd.score = best_score
}

//------------------------------------------------------------------------------
// Entry point

// Pick best modes and fills the levels. Returns true if skipped.
func VP8Decimate(/* const */ it *vp8.VP8EncIterator, /* const */ rd *VP8ModeScore, VP8RDLevel rd_opt) int {
  var is_skipped int
  method := it.enc.method

  InitScore(rd)

  // We can perform predictions for Luma16x16 and Chroma8x8 already.
  // Luma4x4 predictions needs to be done as-we-go.
  VP8MakeLuma16Preds(it)
  VP8MakeChroma8Preds(it)

  if (rd_opt > RD_OPT_NONE) {
    it.do_trellis = (rd_opt >= RD_OPT_TRELLIS_ALL)
    PickBestIntra16(it, rd)
    if (method >= 2) {
      PickBestIntra4(it, rd)
    }
    PickBestUV(it, rd)
    if (rd_opt == RD_OPT_TRELLIS) {  // finish off with trellis-optim now
      it.do_trellis = 1
      SimpleQuantize(it, rd)
    }
  } else {
    // At this point we have heuristically decided intra16 / intra4.
    // For method >= 2, pick the best intra4/intra16 based on SSE (~tad slower).
    // For method <= 1, we don't re-examine the decision but just go ahead with
    // quantization/reconstruction.
    RefineUsingDistortion(it, (method >= 2), (method >= 1), rd)
  }
  is_skipped = (rd.nz == 0)
  VP8SetSkip(it, is_skipped)
  return is_skipped
}
