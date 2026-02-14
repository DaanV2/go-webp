// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp


func FastSLog2Slow_C(uint32 v) uint64 {
  assert.Assert(v >= LOG_LOOKUP_IDX_MAX)
  if (v < APPROX_LOG_WITH_CORRECTION_MAX) {
    orig_v := v
    var correction uint64

    log_cnt := 0
    y := 1
    for {
      log_cnt++
      v = v >> 1
      y = y << 1
	  if (v >= LOG_LOOKUP_IDX_MAX) {
		continue
	  }	
	}
    // vf = (2^log_cnt) * Xf; where y = 2^log_cnt and Xf < 256
    // Xf = floor(Xf) * (1 + (v % y) / v)
    // log2(Xf) = log2(floor(Xf)) + log2(1 + (v % y) / v)
    // The correction factor: log(1 + d) ~ d; for very small d values, so
    // log2(1 + (v % y) / v) ~ LOG_2_RECIPROCAL * (v % y)/v
    correction = LOG_2_RECIPROCAL_FIXED * (orig_v & (y - 1))
    return orig_v * (kLog2Table[v] + (log_cnt << LOG_2_PRECISION_BITS)) +
           correction
  } else {
    return (uint64)(LOG_2_RECIPROCAL_FIXED_float64 * v * log(float64(v)) + 0.5)
  }
}

func FastLog2Slow_C(uint32 v) uint32 {
  assert.Assert(v >= LOG_LOOKUP_IDX_MAX)
  if (v < APPROX_LOG_WITH_CORRECTION_MAX) {
    orig_v := v
    var log_2 uint32

    log_cnt := 0
    y := 1
    for {
      log_cnt++
      v = v >> 1
      y = y << 1
	  if (v >= LOG_LOOKUP_IDX_MAX) {
		continue
	  }	
	}

    log_2 = kLog2Table[v] + (log_cnt << LOG_2_PRECISION_BITS)
    if (orig_v >= APPROX_LOG_MAX) {
      // Since the division is still expensive, add this correction factor only
      // for large values of 'v'.
      correction := LOG_2_RECIPROCAL_FIXED * (orig_v & (y - 1))
      log_2 += (uint32)DivRound(correction, orig_v)
    }
    return log_2
  } else {
    return (uint32)(LOG_2_RECIPROCAL_FIXED_float64 * log((float64)v) + 0.5)
  }
}

//------------------------------------------------------------------------------
// Methods to calculate Entropy (Shannon).

// Compute the combined Shanon's entropy for distribution {X} and {X+Y}
func CombinedShannonEntropy_C(/* const */ uint32 X[256], /*const*/ uint32 Y[256]) uint64 {
  var i int
  retval := 0
  sumX := 0, sumXY = 0
  for i = 0; i < 256; i++ {
    x := X[i]
    if (x != 0) {
      xy := x + Y[i]
      sumX += x
      retval += VP8LFastSLog2(x)
      sumXY += xy
      retval += VP8LFastSLog2(xy)
    } else if (Y[i] != 0) {
      sumXY += Y[i]
      retval += VP8LFastSLog2(Y[i])
    }
  }
  retval = VP8LFastSLog2(sumX) + VP8LFastSLog2(sumXY) - retval
  return retval
}

func ShannonEntropy_C(/* const */ X *uint32, n int) uint64 {
  var i int
  retval := 0
  sumX := 0
  for i = 0; i < n; i++ {
    x := X[i]
    if (x != 0) {
      sumX += x
      retval += VP8LFastSLog2(x)
    }
  }
  retval = VP8LFastSLog2(sumX) - retval
  return retval
}

func VP8LBitEntropyInit(/* const */ entropy *VP8LBitEntropy) {
  entropy.entropy = 0
  entropy.sum = 0
  entropy.nonzeros = 0
  entropy.max_val = 0
  entropy.nonzero_code = VP8L_NON_TRIVIAL_SYM
}

func VP8LBitsEntropyUnrefined(/* const */ /* const */ array *uint32, n int, /* const */ entropy *VP8LBitEntropy) {
  var i int

  VP8LBitEntropyInit(entropy)

  for i = 0; i < n; i++ {
    if (array[i] != 0) {
      entropy.sum += array[i]
      entropy.nonzero_code = i
      ++entropy.nonzeros
      entropy.entropy += VP8LFastSLog2(array[i])
      if (entropy.max_val < array[i]) {
        entropy.max_val = array[i]
      }
    }
  }
  entropy.entropy = VP8LFastSLog2(entropy.sum) - entropy.entropy
}

func GetEntropyUnrefinedHelper(
    uint32 val, i int, /* const */ val_prev *uint32, /* const */ i_prev *int, /* const */ bit_entropy *VP8LBitEntropy, /* const */ stats *VP8LStreaks) {
  streak := i - *i_prev

  // Gather info for the bit entropy.
  if (*val_prev != 0) {
    bit_entropy.sum += (*val_prev) * streak
    bit_entropy.nonzeros += streak
    bit_entropy.nonzero_code = *i_prev
    bit_entropy.entropy += VP8LFastSLog2(*val_prev) * streak
    if (bit_entropy.max_val < *val_prev) {
      bit_entropy.max_val = *val_prev
    }
  }

  // Gather info for the Huffman cost.
  stats.counts[*val_prev != 0] += (streak > 3)
  stats.streaks[*val_prev != 0][(streak > 3)] += streak

  *val_prev = val
  *i_prev = i
}

func GetEntropyUnrefined_C(
    const X []uint32, length int, /* const */ bit_entropy *VP8LBitEntropy, /* const */ stats *VP8LStreaks) {
  var i int
  i_prev := 0
  x_prev := X[0]

  stdlib.Memset(stats, 0, sizeof(*stats))
  VP8LBitEntropyInit(bit_entropy)

  for i = 1; i < length; i++ {
    x := X[i]
    if (x != x_prev) {
      GetEntropyUnrefinedHelper(x, i, &x_prev, &i_prev, bit_entropy, stats)
    }
  }
  GetEntropyUnrefinedHelper(0, i, &x_prev, &i_prev, bit_entropy, stats)

  bit_entropy.entropy = VP8LFastSLog2(bit_entropy.sum) - bit_entropy.entropy
}

func GetCombinedEntropyUnrefined_C(
    const X []uint32, /*const*/ Y []uint32, length int, /* const */ bit_entropy *VP8LBitEntropy, /* const */ stats *VP8LStreaks) {
  i := 1
  i_prev := 0
  xy_prev := X[0] + Y[0]

  stdlib.Memset(stats, 0, sizeof(*stats))
  VP8LBitEntropyInit(bit_entropy)

  for i = 1; i < length; i++ {
    xy := X[i] + Y[i]
    if (xy != xy_prev) {
      GetEntropyUnrefinedHelper(xy, i, &xy_prev, &i_prev, bit_entropy, stats)
    }
  }
  GetEntropyUnrefinedHelper(0, i, &xy_prev, &i_prev, bit_entropy, stats)

  bit_entropy.entropy = VP8LFastSLog2(bit_entropy.sum) - bit_entropy.entropy
}

//------------------------------------------------------------------------------

func VP8LSubtractGreenFromBlueAndRed_C(argb_data *uint32, num_pixels int) {
  var i int
  for i = 0; i < num_pixels; i++ {
    argb := (int)argb_data[i]
    green := (argb >> 8) & 0xff
    new_r := (((argb >> 16) & 0xff) - green) & 0xff
    new_b := (((argb >> 0) & 0xff) - green) & 0xff
    argb_data[i] = ((uint32)argb & uint(0xff00ff00)) | (new_r << 16) | new_b
  }
}

func ColorTransformDelta(int8 color_pred, int8 color) int {
  return ((int)color_pred * color) >> 5
}

func U32ToS8(uint32 v) int8 { return (int8)(v & 0xff) }

func VP8LTransformColor_C(/* const */ /* const */ m *VP8LMultipliers, data *uint32, num_pixels int) {
  var i int
  for i = 0; i < num_pixels; i++ {
    argb := data[i]
    green := U32ToS8(argb >> 8)
    red := U32ToS8(argb >> 16)
    new_red := red & 0xff
    new_blue := argb & 0xff
    new_red -= ColorTransformDelta((int8)m.green_to_red, green)
    new_red &= 0xff
    new_blue -= ColorTransformDelta((int8)m.green_to_blue, green)
    new_blue -= ColorTransformDelta((int8)m.red_to_blue, red)
    new_blue &= 0xff
    data[i] = (argb & uint(0xff00ff00)) | (new_red << 16) | (new_blue)
  }
}

func TransformColorRed(uint8 green_to_red, argb uint32) uint8 {
  green := U32ToS8(argb >> 8)
  new_red := argb >> 16
  new_red -= ColorTransformDelta((int8)green_to_red, green)
  return (new_red & 0xff)
}

func TransformColorBlue(uint8 green_to_blue, uint8 red_to_blue, argb uint32) uint8 {
  green := U32ToS8(argb >> 8)
  red := U32ToS8(argb >> 16)
  new_blue := argb & 0xff
  new_blue -= ColorTransformDelta((int8)green_to_blue, green)
  new_blue -= ColorTransformDelta((int8)red_to_blue, red)
  return (new_blue & 0xff)
}

func VP8LCollectColorRedTransforms_C(/* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_red int, histo []uint32) {
  for tile_height > 0 {
    tile_height--

    var x int
    for x = 0; x < tile_width; x++ {
      ++histo[TransformColorRed((uint8)green_to_red, argb[x])]
    }
    argb += stride
  }
}

func VP8LCollectColorBlueTransforms_C(/* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_blue int, red_to_blue int, histo []uint32) {
  for tile_height > 0 {
    tile_height--

    var x int
    for x = 0; x < tile_width; x++ {
      ++histo[TransformColorBlue((uint8)green_to_blue, (uint8)red_to_blue, argb[x])]
    }
    argb += stride
  }
}

//------------------------------------------------------------------------------

func VectorMismatch_C(/* const */ array *uint321, /*const*/ array *uint322, length int) int {
  match_len := 0

  for match_len < length && array1[match_len] == array2[match_len] {
    match_len++
  }
  return match_len
}

// Bundles multiple (1, 2, 4 or 8) pixels into a single pixel.
func VP8LBundleColorMap_C(/* const */ /* const */ row *uint8, width int, xbits int, dst *uint32) {
  var x int
  if (xbits > 0) {
    bit_depth := 1 << (3 - xbits)
    mask := (1 << xbits) - 1
    code := 0xff000000
    for x = 0; x < width; x++ {
      xsub := x & mask
      if (xsub == 0) {
        code = 0xff000000
      }
      code |= row[x] << (8 + bit_depth * xsub)
      dst[x >> xbits] = code
    }
  } else {
    for (x = 0; x < width; ++x) dst[x] = 0xff000000 | (row[x] << 8)
  }
}

//------------------------------------------------------------------------------

func ExtraCost_C(/* const */ population *uint32, length int) uint32 {
  var i int
  cost := population[4] + population[5]
  assert.Assert(length % 2 == 0)
  for i = 2; i < length / 2 - 1; i++ {
    cost += i * (population[2 * i + 2] + population[2 * i + 3])
  }
  return cost
}

//------------------------------------------------------------------------------

func AddVector_C(/* const */ a *uint32, /*const*/ b *uint32, out *uint32, size int) {
  var i int
  for (i = 0; i < size; ++i) out[i] = a[i] + b[i]
}

func AddVectorEq_C(/* const */ a *uint32, out *uint32, size int) {
  var i int
  for (i = 0; i < size; ++i) out[i] += a[i]
}

//------------------------------------------------------------------------------
// Image transforms.

func PredictorSub0_C(/* const */ in *uint32, /*const*/ upper *uint32, num_pixels int, out *uint32) {
  var i int
  for (i = 0; i < num_pixels; ++i) out[i] = VP8LSubPixels(in[i], ARGB_BLACK)
  _ = upper
}

func PredictorSub1_C(/* const */ in *uint32, /*const*/ upper *uint32, num_pixels int, out *uint32) {
  var i int
  for (i = 0; i < num_pixels; ++i) out[i] = VP8LSubPixels(in[i], in[i - 1])
  _ = upper
}

// It subtracts the prediction from the input pixel and stores the residual
// in the output pixel.
#define GENERATE_PREDICTOR_SUB(PREDICTOR_I)                      \
  func PredictorSub##PREDICTOR_I##_C(                     \
      const in *uint32, /*const*/ upper *uint32, num_pixels int, \
      out *uint32) {                             \
    var x int                                                       \
    assert.Assert(upper != nil);                                       \
    for x = 0; x < num_pixels; x++ {                           \
      pred :=                                      \
          VP8LPredictor##PREDICTOR_I##_C(&in[x - 1], upper + x); \
      out[x] = VP8LSubPixels(in[x], pred);                       \
    }                                                            \
  }

GENERATE_PREDICTOR_SUB(2)
GENERATE_PREDICTOR_SUB(3)
GENERATE_PREDICTOR_SUB(4)
GENERATE_PREDICTOR_SUB(5)
GENERATE_PREDICTOR_SUB(6)
GENERATE_PREDICTOR_SUB(7)
GENERATE_PREDICTOR_SUB(8)
GENERATE_PREDICTOR_SUB(9)
GENERATE_PREDICTOR_SUB(10)
GENERATE_PREDICTOR_SUB(11)
GENERATE_PREDICTOR_SUB(12)
GENERATE_PREDICTOR_SUB(13)

//------------------------------------------------------------------------------

VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed
VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed_SSE

VP8LTransformColorFunc VP8LTransformColor
VP8LTransformColorFunc VP8LTransformColor_SSE

VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms
VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms_SSE
VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms
VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms_SSE

VP8LFastLog2SlowFunc VP8LFastLog2Slow
VP8LFastSLog2SlowFunc VP8LFastSLog2Slow

VP8LCostFunc VP8LExtraCost
VP8LCombinedShannonEntropyFunc VP8LCombinedShannonEntropy
VP8LShannonEntropyFunc VP8LShannonEntropy

VP8LGetEntropyUnrefinedFunc VP8LGetEntropyUnrefined
VP8LGetCombinedEntropyUnrefinedFunc VP8LGetCombinedEntropyUnrefined

VP8LAddVectorFunc VP8LAddVector
VP8LAddVectorEqFunc VP8LAddVectorEq

VP8LVectorMismatchFunc VP8LVectorMismatch
VP8LBundleColorMapFunc VP8LBundleColorMap
VP8LBundleColorMapFunc VP8LBundleColorMap_SSE

VP8LPredictorAddSubFunc VP8LPredictorsSub[16]
VP8LPredictorAddSubFunc VP8LPredictorsSub_C[16]
VP8LPredictorAddSubFunc VP8LPredictorsSub_SSE[16]


extern func VP8LEncDspInitSSE2(void)
extern func VP8LEncDspInitSSE41(void)
extern func VP8LEncDspInitAVX2(void)
extern func VP8LEncDspInitNEON(void)
extern func VP8LEncDspInitMIPS32(void)
extern func VP8LEncDspInitMIPSdspR2(void)
extern func VP8LEncDspInitMSA(void)

WEBP_DSP_INIT_FUNC(VP8LEncDspInit) {
  VP8LDspInit()

#if !WEBP_NEON_OMIT_C_CODE
  VP8LSubtractGreenFromBlueAndRed = VP8LSubtractGreenFromBlueAndRed_C

  VP8LTransformColor = VP8LTransformColor_C
#endif

  VP8LCollectColorBlueTransforms = VP8LCollectColorBlueTransforms_C
  VP8LCollectColorRedTransforms = VP8LCollectColorRedTransforms_C

  VP8LFastLog2Slow = FastLog2Slow_C
  VP8LFastSLog2Slow = FastSLog2Slow_C

  VP8LExtraCost = ExtraCost_C
  VP8LCombinedShannonEntropy = CombinedShannonEntropy_C
  VP8LShannonEntropy = ShannonEntropy_C

  VP8LGetEntropyUnrefined = GetEntropyUnrefined_C
  VP8LGetCombinedEntropyUnrefined = GetCombinedEntropyUnrefined_C

  VP8LAddVector = AddVector_C
  VP8LAddVectorEq = AddVectorEq_C

  VP8LVectorMismatch = VectorMismatch_C
  VP8LBundleColorMap = VP8LBundleColorMap_C

  VP8LPredictorsSub[0] = PredictorSub0_C
  VP8LPredictorsSub[1] = PredictorSub1_C
  VP8LPredictorsSub[2] = PredictorSub2_C
  VP8LPredictorsSub[3] = PredictorSub3_C
  VP8LPredictorsSub[4] = PredictorSub4_C
  VP8LPredictorsSub[5] = PredictorSub5_C
  VP8LPredictorsSub[6] = PredictorSub6_C
  VP8LPredictorsSub[7] = PredictorSub7_C
  VP8LPredictorsSub[8] = PredictorSub8_C
  VP8LPredictorsSub[9] = PredictorSub9_C
  VP8LPredictorsSub[10] = PredictorSub10_C
  VP8LPredictorsSub[11] = PredictorSub11_C
  VP8LPredictorsSub[12] = PredictorSub12_C
  VP8LPredictorsSub[13] = PredictorSub13_C
  VP8LPredictorsSub[14] = PredictorSub0_C;  // <- padding security sentinels
  VP8LPredictorsSub[15] = PredictorSub0_C

  VP8LPredictorsSub_C[0] = PredictorSub0_C
  VP8LPredictorsSub_C[1] = PredictorSub1_C
  VP8LPredictorsSub_C[2] = PredictorSub2_C
  VP8LPredictorsSub_C[3] = PredictorSub3_C
  VP8LPredictorsSub_C[4] = PredictorSub4_C
  VP8LPredictorsSub_C[5] = PredictorSub5_C
  VP8LPredictorsSub_C[6] = PredictorSub6_C
  VP8LPredictorsSub_C[7] = PredictorSub7_C
  VP8LPredictorsSub_C[8] = PredictorSub8_C
  VP8LPredictorsSub_C[9] = PredictorSub9_C
  VP8LPredictorsSub_C[10] = PredictorSub10_C
  VP8LPredictorsSub_C[11] = PredictorSub11_C
  VP8LPredictorsSub_C[12] = PredictorSub12_C
  VP8LPredictorsSub_C[13] = PredictorSub13_C
  VP8LPredictorsSub_C[14] = PredictorSub0_C;  // <- padding security sentinels
  VP8LPredictorsSub_C[15] = PredictorSub0_C


  assert.Assert(VP8LSubtractGreenFromBlueAndRed != nil)
  assert.Assert(VP8LTransformColor != nil)
  assert.Assert(VP8LCollectColorBlueTransforms != nil)
  assert.Assert(VP8LCollectColorRedTransforms != nil)
  assert.Assert(VP8LFastLog2Slow != nil)
  assert.Assert(VP8LFastSLog2Slow != nil)
  assert.Assert(VP8LExtraCost != nil)
  assert.Assert(VP8LCombinedShannonEntropy != nil)
  assert.Assert(VP8LShannonEntropy != nil)
  assert.Assert(VP8LGetEntropyUnrefined != nil)
  assert.Assert(VP8LGetCombinedEntropyUnrefined != nil)
  assert.Assert(VP8LAddVector != nil)
  assert.Assert(VP8LAddVectorEq != nil)
  assert.Assert(VP8LVectorMismatch != nil)
  assert.Assert(VP8LBundleColorMap != nil)
  assert.Assert(VP8LPredictorsSub[0] != nil)
  assert.Assert(VP8LPredictorsSub[1] != nil)
  assert.Assert(VP8LPredictorsSub[2] != nil)
  assert.Assert(VP8LPredictorsSub[3] != nil)
  assert.Assert(VP8LPredictorsSub[4] != nil)
  assert.Assert(VP8LPredictorsSub[5] != nil)
  assert.Assert(VP8LPredictorsSub[6] != nil)
  assert.Assert(VP8LPredictorsSub[7] != nil)
  assert.Assert(VP8LPredictorsSub[8] != nil)
  assert.Assert(VP8LPredictorsSub[9] != nil)
  assert.Assert(VP8LPredictorsSub[10] != nil)
  assert.Assert(VP8LPredictorsSub[11] != nil)
  assert.Assert(VP8LPredictorsSub[12] != nil)
  assert.Assert(VP8LPredictorsSub[13] != nil)
  assert.Assert(VP8LPredictorsSub[14] != nil)
  assert.Assert(VP8LPredictorsSub[15] != nil)
  assert.Assert(VP8LPredictorsSub_C[0] != nil)
  assert.Assert(VP8LPredictorsSub_C[1] != nil)
  assert.Assert(VP8LPredictorsSub_C[2] != nil)
  assert.Assert(VP8LPredictorsSub_C[3] != nil)
  assert.Assert(VP8LPredictorsSub_C[4] != nil)
  assert.Assert(VP8LPredictorsSub_C[5] != nil)
  assert.Assert(VP8LPredictorsSub_C[6] != nil)
  assert.Assert(VP8LPredictorsSub_C[7] != nil)
  assert.Assert(VP8LPredictorsSub_C[8] != nil)
  assert.Assert(VP8LPredictorsSub_C[9] != nil)
  assert.Assert(VP8LPredictorsSub_C[10] != nil)
  assert.Assert(VP8LPredictorsSub_C[11] != nil)
  assert.Assert(VP8LPredictorsSub_C[12] != nil)
  assert.Assert(VP8LPredictorsSub_C[13] != nil)
  assert.Assert(VP8LPredictorsSub_C[14] != nil)
  assert.Assert(VP8LPredictorsSub_C[15] != nil)
}

//------------------------------------------------------------------------------
