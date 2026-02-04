package enc

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Author: Jyrki Alakuijala (jyrki@google.com)
//
// Models the histograms of literal and distance codes.


import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// Not a trivial literal symbol.
const VP8L_NON_TRIVIAL_SYM =((uint16)(0xffff))

// A simple container for histograms of data.
type VP8LHistogram struct {
  // 'literal' contains green literal, palette-code and
  // copy-length-prefix histogram
  literal *uint32;  // Pointer to the allocated buffer for literal.
  uint32 red[NUM_LITERAL_CODES];
  uint32 blue[NUM_LITERAL_CODES];
  uint32 alpha[NUM_LITERAL_CODES];
  // Backward reference prefix-code histogram.
  uint32 distance[NUM_DISTANCE_CODES];
  int palette_code_bits;
  // The following members are only used within VP8LGetHistoImageSymbols.

  // Index of the unique value of a histogram if any, VP8L_NON_TRIVIAL_SYM
  // otherwise.
  uint16 trivial_symbol[5];
  uint64 bit_cost;  // Cached value of total bit cost.
  // Cached values of entropy costs: literal, red, blue, alpha, distance
  uint64 costs[5];
  uint8 is_used[5];  // 5 for literal, red, blue, alpha, distance
  uint16 bin_id;     // entropy bin index.
} ;

// Collection of histograms with fixed capacity, allocated as one
// big memory chunk.
type VP8LHistogramSet struct {
  int size;      // number of slots currently in use
  int max_size;  // maximum capacity
  *VP8LHistogram* histograms;
} ;

// Create the histogram.
//
// The input data is the PixOrCopy data, which models the literals, stop
// codes and backward references (both distances and lengths).  Also: if
// palette_code_bits is >= 0, initialize the histogram with this value.
func VP8LHistogramCreate(const h *VP8LHistogram, /*const*/ refs *VP8LBackwardRefs, int palette_code_bits);

// Set the palette_code_bits and reset the stats.
// If init_arrays is true, the arrays are also filled with 0's.
func VP8LHistogramInit(const h *VP8LHistogram, int palette_code_bits, int init_arrays);

// Collect all the references into a histogram (without reset)
// The distance modifier function is applied to the distance before
// the histogram is updated. It can be nil.
func VP8LHistogramStoreRefs(const refs *VP8LBackwardRefs, int (distance_modifier *const)(int, int), int distance_modifier_arg0, /*const*/ histo *VP8LHistogram);

// Free the memory allocated for the histogram.
func VP8LFreeHistogram(const histo *VP8LHistogram);

// Free the memory allocated for the histogram set.
func VP8LFreeHistogramSet(const histo *VP8LHistogramSet);

// Allocate an array of pointer to histograms, allocated and initialized
// using 'cache_bits'. Return nil in case of memory error.
VP *VP8LHistogramSet8LAllocateHistogramSet(int size, int cache_bits);

// Set the histograms in set to 0.
func VP8LHistogramSetClear(const set *VP8LHistogramSet);

// Allocate and initialize histogram object with specified 'cache_bits'.
// Returns nil in case of memory error.
// Special case of VP8LAllocateHistogramSet, with size equals 1.
VP *VP8LHistogram8LAllocateHistogram(int cache_bits);

static  int VP8LHistogramNumCodes(int palette_code_bits) {
  return NUM_LITERAL_CODES + NUM_LENGTH_CODES +
         ((palette_code_bits > 0) ? (1 << palette_code_bits) : 0);
}

// Builds the histogram image. pic and percent are for progress.
// Returns false in case of error (stored in pic.error_code).
int VP8LGetHistoImageSymbols(int xsize, int ysize, /*const*/ refs *VP8LBackwardRefs, int quality, int low_effort, int histogram_bits, int cache_bits, /*const*/ image_histo *VP8LHistogramSet, /*const*/ tmp_histo *VP8LHistogram, /*const*/ histogram_symbols *uint32, /*const*/ pic *WebPPicture, int percent_range, /*const*/ percent *int);

// Returns the entropy for the symbols in the input array.
uint64 VP8LBitsEntropy(const array *uint32, int n);

// Estimate how many bits the combined entropy of literals and distance
// approximately maps to.
uint64 VP8LHistogramEstimateBits(const h *VP8LHistogram);

#ifdef __cplusplus
}
#endif

#endif  // WEBP_ENC_HISTOGRAM_ENC_H_
