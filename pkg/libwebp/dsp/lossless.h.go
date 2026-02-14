// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

type VP8LMultipliers struct {
	// Note: the members are uint8, so that any negative values are
	// automatically converted to "mod 256" values.
	green_to_red  uint8
	green_to_blue uint8
	red_to_blue   uint8
}

type VP8LStreaks struct { // small struct to hold counters
	counts  [2]int    // index: 0=zero streak, 1=non-zero streak
	streaks [2][2]int // [zero/non-zero][streak<3 / streak>=3]
}

type VP8LBitEntropy struct { // small struct to hold bit entropy results
	entropy      uint64 // entropy
	sum          uint32 // sum of the population
	nonzeros     int    // number of non-zero elements in the population
	max_val      uint32 // maximum value in the population
	nonzero_code uint32 // index of the last non-zero in the population
}

// These Add/Sub function expects upper[-1] and out[-1] to be readable.
type VP8LPredictorAddSubFunc = func( /* const */ in *uint32 /*const*/, upper *uint32, num_pixels int, out *uint32)
type VP8LProcessDecBlueAndRedFunc = func( /* const */ src []uint32, num_pixels int, dst *uint32)
type VP8LTransformColorInverseFunc = func( /* const */ m *VP8LMultipliers /*const*/, src []uint32, num_pixels int, dst *uint32)
type VP8LPredictorFunc = func( /* const */ left *uint32 /*const*/, top *uint32)

// Color space conversion.
type VP8LConvertFunc = func( /* const */ src []uint32, num_pixels int, dst []uint8)
type VP8LMapARGBFunc = func( /* const */ src []uint32 /*const*/, color_map *uint32, dst *uint32, y_start int, y_end int, width int)
type VP8LMapAlphaFunc = func( /* const */ src *uint8 /*const*/, color_map *uint32, dst []uint8, y_start int, y_end int, width int)

type VP8LProcessEncBlueAndRedFunc = func(dst *uint32, num_pixels int)
type VP8LTransformColorFunc = func( /*const*/ /* const */ m *VP8LMultipliers, dst *uint32, num_pixels int)
type VP8LCollectColorBlueTransformsFunc = func( /*const*/ argb *uint32, stride int, tile_width int, tile_height int, green_to_blue int, red_to_blue int, histo []uint32)
type VP8LCollectColorRedTransformsFunc = func( /*const*/ argb *uint32, stride int, tile_width int, tile_height int, green_to_red int, histo []uint32)

// Expose some C-only fallback functions
func VP8LTransformColor_C( /* const */ /* const */ m *VP8LMultipliers, data *uint32, num_pixels int)
func VP8LSubtractGreenFromBlueAndRed_C(argb_data *uint32, num_pixels int)
func VP8LCollectColorRedTransforms_C( /* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_red int, histo []uint32)
func VP8LCollectColorBlueTransforms_C( /* const */ argb *uint32, stride int, tile_width int, tile_height int, green_to_blue int, red_to_blue int, histo []uint32)

// -----------------------------------------------------------------------------
// Huffman-cost related functions.

type VP8LCostFunc = func( /* const */ population *uint32, length int) uint32
type VP8LCombinedShannonEntropyFunc = func( /* const */ X [256]uint32 /*const*/, Y [256]uint32) uint64
type VP8LShannonEntropyFunc = func( /* const */ X *uint32, length int) uint64

// Get the combined symbol bit entropy and Huffman cost stats for the
// distributions 'X' and 'Y'. Those results can then be refined according to
// codec specific heuristics.
type VP8LGetCombinedEntropyUnrefinedFunc = func( /*const*/ X []uint32 /*const*/, Y []uint32, length int /* const */, bit_entropy *VP8LBitEntropy /* const */, stats *VP8LStreaks)

// Get the entropy for the distribution 'X'.
type VP8LGetEntropyUnrefinedFunc = func( /*const*/ X []uint32, length int /* const */, bit_entropy *VP8LBitEntropy /* const */, stats *VP8LStreaks)

type VP8LAddVectorFunc = func( /* const */ a *uint32 /*const*/, b *uint32, out *uint32, size int)
type VP8LAddVectorEqFunc = func( /* const */ a *uint32, out *uint32, size int)
type VP8LVectorMismatchFunc = func( /* const */ array1 []uint32 /*const*/, array2 []uint32, length int) int
