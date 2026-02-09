package dsp

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Image transforms and color space conversion methods for lossless decoder.
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)


import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"



//------------------------------------------------------------------------------
// Decoding

typedef uint32 (*VP8LPredictorFunc)(/* const */ left *uint32, /*const*/ top *uint32);
extern VP8LPredictorFunc VP8LPredictors[16];

uint32 VP8LPredictor2_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor3_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor4_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor5_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor6_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor7_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor8_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor9_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor10_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor11_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor12_C(/* const */ left *uint32, /*const*/ top *uint32);
uint32 VP8LPredictor13_C(/* const */ left *uint32, /*const*/ top *uint32);

// These Add/Sub function expects upper[-1] and out[-1] to be readable.
typedef func (*VP8LPredictorAddSubFunc)(/* const */ in *uint32, /*const*/ upper *uint32, num_pixels int, out *uint32);
extern VP8LPredictorAddSubFunc VP8LPredictorsAdd[16];
extern VP8LPredictorAddSubFunc VP8LPredictorsAdd_C[16];
extern VP8LPredictorAddSubFunc VP8LPredictorsAdd_SSE[16];

typedef func (*VP8LProcessDecBlueAndRedFunc)(/* const */ src *uint32, num_pixels int, dst *uint32);
extern VP8LProcessDecBlueAndRedFunc VP8LAddGreenToBlueAndRed;
extern VP8LProcessDecBlueAndRedFunc VP8LAddGreenToBlueAndRed_SSE;

type VP8LMultipliers struct {
  // Note: the members are uint8, so that any negative values are
  // automatically converted to "mod 256" values.
  green_to_red uint8
  green_to_blue uint8
  red_to_blue uint8
}
typedef func (*VP8LTransformColorInverseFunc)(/* const */ m *VP8LMultipliers, /*const*/ src *uint32, num_pixels int, dst *uint32);
extern VP8LTransformColorInverseFunc VP8LTransformColorInverse;
extern VP8LTransformColorInverseFunc VP8LTransformColorInverse_SSE;

struct VP8LTransform;  // Defined in dec/vp8li.h.

// Performs inverse transform of data given transform information, start and end
// rows. Transform will be applied to rows [row_start, row_end[.
// The and pointers refer *in to *out source and destination data respectively
// corresponding to the intermediate row (row_start).
func VP8LInverseTransform(/* const */ struct const transform *VP8LTransform, row_start int, row_end int, /*const*/ in *uint32, /*const*/ out *uint32);

// Color space conversion.
typedef func (*VP8LConvertFunc)(/* const */ src *uint32, num_pixels int, dst *uint8);
extern VP8LConvertFunc VP8LConvertBGRAToRGB;
extern VP8LConvertFunc VP8LConvertBGRAToRGBA;
extern VP8LConvertFunc VP8LConvertBGRAToRGBA4444;
extern VP8LConvertFunc VP8LConvertBGRAToRGB565;
extern VP8LConvertFunc VP8LConvertBGRAToBGR;
extern VP8LConvertFunc VP8LConvertBGRAToRGB_SSE;
extern VP8LConvertFunc VP8LConvertBGRAToRGBA_SSE;

// Converts from BGRA to other color spaces.
func VP8LConvertFromBGRA(/* const */ in_data *uint32, num_pixels int, WEBP_CSP_MODE out_colorspace, /*const*/ rgba *uint8);

typedef func (*VP8LMapARGBFunc)(/* const */ src *uint32, /*const*/ color_map *uint32, dst *uint32, y_start int, y_end int, width int);
typedef func (*VP8LMapAlphaFunc)(/* const */ src *uint8, /*const*/ color_map *uint32, dst *uint8, y_start int, y_end int, width int);

extern VP8LMapARGBFunc VP8LMapColor32b;
extern VP8LMapAlphaFunc VP8LMapColor8b;

// Similar to the static method ColorIndexInverseTransform() that is part of
// lossless.c, but used only for alpha decoding. It takes uint8 (rather than
// uint32) arguments for 'src' and 'dst'.
func VP8LColorIndexInverseTransformAlpha(
     transform *VP8LTransform, y_start int, y_end int, /*const*/ src *uint8, dst *uint8);

// Expose some C-only fallback functions
func VP8LTransformColorInverse_C(/* const */ m *VP8LMultipliers, /*const*/ src *uint32, num_pixels int, dst *uint32);

func VP8LConvertBGRAToRGB_C(/* const */ src *uint32, num_pixels int, dst *uint8);
func VP8LConvertBGRAToRGBA_C(/* const */ src *uint32, num_pixels int, dst *uint8);
func VP8LConvertBGRAToRGBA4444_C(/* const */ src *uint32, num_pixels int, dst *uint8);
func VP8LConvertBGRAToRGB565_C(/* const */ src *uint32, num_pixels int, dst *uint8);
func VP8LConvertBGRAToBGR_C(/* const */ src *uint32, num_pixels int, dst *uint8);
func VP8LAddGreenToBlueAndRed_C(/* const */ src *uint32, num_pixels int, dst *uint32);

// Must be called before calling any of the above methods.
func VP8LDspInit(void);

//------------------------------------------------------------------------------
// Encoding

typedef func (*VP8LProcessEncBlueAndRedFunc)(dst *uint32, num_pixels int);
extern VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed;
extern VP8LProcessEncBlueAndRedFunc VP8LSubtractGreenFromBlueAndRed_SSE;
typedef func (*VP8LTransformColorFunc)(
    const /* const */ m *VP8LMultipliers, dst *uint32, num_pixels int);
extern VP8LTransformColorFunc VP8LTransformColor;
extern VP8LTransformColorFunc VP8LTransformColor_SSE;
typedef func (*VP8LCollectColorBlueTransformsFunc)(
    const argb *uint32, stride int, tile_width int, tile_height int, green_to_blue int, red_to_blue int, uint32 histo[]);
extern VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms;
extern VP8LCollectColorBlueTransformsFunc VP8LCollectColorBlueTransforms_SSE;

typedef func (*VP8LCollectColorRedTransformsFunc)(
    const argb *uint32, stride int, tile_width int, tile_height int, green_to_red int, uint32 histo[]);
extern VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms;
extern VP8LCollectColorRedTransformsFunc VP8LCollectColorRedTransforms_SSE;

// Expose some C-only fallback functions
func VP8LTransformColor_C(/* const */ /* const */ m *VP8LMultipliers, data *uint32, num_pixels int);
func VP8LSubtractGreenFromBlueAndRed_C(argb_data *uint32, num_pixels int);
func VP8LCollectColorRedTransforms_C(/* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_red int, uint32 histo[]);
func VP8LCollectColorBlueTransforms_C(/* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_blue int, red_to_blue int, uint32 histo[]);

// -----------------------------------------------------------------------------
// Huffman-cost related functions.

type VP8LCostFunc = func(/* const */ population *uint32, length int)uint32
type VP8LCombinedShannonEntropyFunc = func(/* const */ X [256]uint32, /*const*/ Y [256]uint32)uint64
type VP8LShannonEntropyFunc = func(/* const */ X *uint32, length int)uint64

type VP8LStreaks struct {      // small struct to hold counters
   counts[2]int     // index: 0=zero streak, 1=non-zero streak
  streaks[2][2]int   // [zero/non-zero][streak<3 / streak>=3]
} ;

type VP8LBitEntropy struct {          // small struct to hold bit entropy results
  entropy uint64       // entropy
  sum uint32           // sum of the population
  nonzeros int           // number of non-zero elements in the population
  max_val uint32       // maximum value in the population
  nonzero_code uint32  // index of the last non-zero in the population
}

func VP8LBitEntropyInit(/* const */ entropy *VP8LBitEntropy);

// Get the combined symbol bit entropy and Huffman cost stats for the
// distributions 'X' and 'Y'. Those results can then be refined according to
// codec specific heuristics.
typedef func (*VP8LGetCombinedEntropyUnrefinedFunc)(
    const uint32 X[], /*const*/ uint32 Y[], length int, /* const */ bit_entropy *VP8LBitEntropy, /* const */ stats *VP8LStreaks);
extern VP8LGetCombinedEntropyUnrefinedFunc VP8LGetCombinedEntropyUnrefined;

// Get the entropy for the distribution 'X'.
typedef func (*VP8LGetEntropyUnrefinedFunc)(
    const uint32 X[], length int, /* const */ bit_entropy *VP8LBitEntropy, /* const */ stats *VP8LStreaks);
extern VP8LGetEntropyUnrefinedFunc VP8LGetEntropyUnrefined;

func VP8LBitsEntropyUnrefined(/* const */ /* const */ array *uint32, n int, /* const */ entropy *VP8LBitEntropy);

typedef func (*VP8LAddVectorFunc)(/* const */ a *uint32, /*const*/ b *uint32, out *uint32, size int);
extern VP8LAddVectorFunc VP8LAddVector;
typedef func (*VP8LAddVectorEqFunc)(/* const */ a *uint32, out *uint32, size int);
extern VP8LAddVectorEqFunc VP8LAddVectorEq;

// -----------------------------------------------------------------------------
// PrefixEncode()

typedef int (*VP8LVectorMismatchFunc)(/* const */ array *uint321, /*const*/ array *uint322, length int);
// Returns the first index where array1 and array2 are different.
extern VP8LVectorMismatchFunc VP8LVectorMismatch;

typedef func (*VP8LBundleColorMapFunc)(/* const */ /* const */ row *uint8, width int, xbits int, dst *uint32);
extern VP8LBundleColorMapFunc VP8LBundleColorMap;
extern VP8LBundleColorMapFunc VP8LBundleColorMap_SSE;
func VP8LBundleColorMap_C(/* const */ /* const */ row *uint8, width int, xbits int, dst *uint32);

// Must be called before calling any of the above methods.
func VP8LEncDspInit(void);

//------------------------------------------------------------------------------



#endif  // WEBP_DSP_LOSSLESS_H_
