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
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

// Number of partitions for the three dominant (literal, red and blue) symbol
// costs.
const NUM_PARTITIONS =4
// The size of the bin-hash corresponding to the three dominant costs.
const BIN_SIZE =(NUM_PARTITIONS * NUM_PARTITIONS * NUM_PARTITIONS)
// Maximum number of histograms allowed in greedy combining algorithm.
const MAX_HISTO_GREEDY =100

// Enum to meaningfully access the elements of the Histogram arrays.
type HistogramIndex int

const ( 
	LITERAL HistogramIndex = iota
	RED
	BLUE
	ALPHA
	DISTANCE
)

// Return the size of the histogram for a given cache_bits.
GetHistogramSize(int cache_bits) int  {
  literal_size := VP8LHistogramNumCodes(cache_bits);
  total_size := sizeof(VP8LHistogram) + sizeof(int) * literal_size;
  assert.Assert(total_size <= (uint64)0x7fffffff);

  return (int)total_size;
}

func HistogramStatsClear(const h *VP8LHistogram) {
  var i int
  for i = 0; i < 5; i++ {
    h.trivial_symbol[i] = VP8L_NON_TRIVIAL_SYM;
    // By default, the histogram is assumed to be used.
    h.is_used[i] = 1;
  }
  h.bit_cost = 0;
  stdlib.Memset(h.costs, 0, sizeof(h.costs));
}

func HistogramClear(const h *VP8LHistogram) {
  var literal *uint32 = h.literal;
  cache_bits := h.palette_code_bits;
  histo_size := GetHistogramSize(cache_bits);
  stdlib.Memset(h, 0, histo_size);
  h.palette_code_bits = cache_bits;
  h.literal = literal;
  HistogramStatsClear(h);
}

// Swap two histogram pointers.
func HistogramSwap(*VP8LHistogram* const h1, *VP8LHistogram* const h2) {
  var tmp *VP8LHistogram = *h1;
  *h1 = *h2;
  *h2 = tmp;
}

func HistogramCopy(const src *VP8LHistogram, /*const*/ dst *VP8LHistogram) {
  var dst_literal *uint32 = dst.literal;
  dst_cache_bits := dst.palette_code_bits;
  literal_size := VP8LHistogramNumCodes(dst_cache_bits);
  histo_size := GetHistogramSize(dst_cache_bits);
  assert.Assert(src.palette_code_bits == dst_cache_bits);
  memcpy(dst, src, histo_size);
  dst.literal = dst_literal;
  memcpy(dst.literal, src.literal, literal_size * sizeof(*dst.literal));
}

func VP8LFreeHistogram(const h *VP8LHistogram) { WebPSafeFree(h); }

func VP8LFreeHistogramSet(const histograms *VP8LHistogramSet) {
  WebPSafeFree(histograms);
}

func VP8LHistogramCreate(const h *VP8LHistogram, /*const*/ refs *VP8LBackwardRefs, int palette_code_bits) {
  if (palette_code_bits >= 0) {
    h.palette_code_bits = palette_code_bits;
  }
  HistogramClear(h);
  VP8LHistogramStoreRefs(refs, /*distance_modifier=*/nil, /*distance_modifier_arg0=*/0, h);
}

func VP8LHistogramInit(const h *VP8LHistogram, int palette_code_bits, int init_arrays) {
  h.palette_code_bits = palette_code_bits;
  if (init_arrays) {
    HistogramClear(h);
  } else {
    HistogramStatsClear(h);
  }
}

VP *VP8LHistogram8LAllocateHistogram(int cache_bits) {
  histo *VP8LHistogram = nil;
  total_size := GetHistogramSize(cache_bits);
  var memory *uint8 = (*uint8)WebPSafeMalloc(total_size, sizeof(*memory));
  if (memory == nil) return nil;
  histo = (*VP8LHistogram)memory;
  // 'literal' won't necessary be aligned.
  histo.literal = (*uint32)(memory + sizeof(VP8LHistogram));
  VP8LHistogramInit(histo, cache_bits, /*init_arrays=*/0);
  return histo;
}

// Resets the pointers of the histograms to point to the bit buffer in the set.
func HistogramSetResetPointers(const set *VP8LHistogramSet, int cache_bits) {
  var i int
  histo_size := GetHistogramSize(cache_bits);
  memory *uint8 = (*uint8)(set.histograms);
  memory += set.max_size * sizeof(*set.histograms);
  for i = 0; i < set.max_size; i++ {
    memory = (*uint8)WEBP_ALIGN(memory);
    set.histograms[i] = (*VP8LHistogram)memory;
    // 'literal' won't necessary be aligned.
    set.histograms[i].literal = (*uint32)(memory + sizeof(VP8LHistogram));
    memory += histo_size;
  }
}

// Returns the total size of the VP8LHistogramSet.
static uint64 HistogramSetTotalSize(int size, int cache_bits) {
  histo_size := GetHistogramSize(cache_bits);
  return (sizeof(VP8LHistogramSet) +
          size * (sizeof(*VP8LHistogram) + histo_size + WEBP_ALIGN_CST));
}

VP *VP8LHistogramSet8LAllocateHistogramSet(int size, int cache_bits) {
  var i int
  set *VP8LHistogramSet;
  total_size := HistogramSetTotalSize(size, cache_bits);
  memory *uint8 = (*uint8)WebPSafeMalloc(total_size, sizeof(*memory));
  if (memory == nil) return nil;

  set = (*VP8LHistogramSet)memory;
  memory += sizeof(*set);
  set.histograms = (*VP8LHistogram*)memory;
  set.max_size = size;
  set.size = size;
  HistogramSetResetPointers(set, cache_bits);
  for i = 0; i < size; i++ {
    VP8LHistogramInit(set.histograms[i], cache_bits, /*init_arrays=*/0);
  }
  return set;
}

func VP8LHistogramSetClear(const set *VP8LHistogramSet) {
  var i int
  cache_bits := set.histograms[0].palette_code_bits;
  size := set.max_size;
  total_size := HistogramSetTotalSize(size, cache_bits);
  memory *uint8 = (*uint8)set;

  stdlib.Memset(memory, 0, total_size);
  memory += sizeof(*set);
  set.histograms = (*VP8LHistogram*)memory;
  set.max_size = size;
  set.size = size;
  HistogramSetResetPointers(set, cache_bits);
  for i = 0; i < size; i++ {
    set.histograms[i].palette_code_bits = cache_bits;
  }
}

// Removes the histogram 'i' from 'set'.
func HistogramSetRemoveHistogram(const set *VP8LHistogramSet, int i) {
  set.histograms[i] = set.histograms[set.size - 1];
  --set.size;
  assert.Assert(set.size > 0);
}

// -----------------------------------------------------------------------------

func HistogramAddSinglePixOrCopy(
    const histo *VP8LHistogram, /*const*/ v *PixOrCopy, int (distance_modifier *const)(int, int), int distance_modifier_arg0) {
  if (PixOrCopyIsLiteral(v)) {
    ++histo.alpha[PixOrCopyLiteral(v, 3)];
    ++histo.red[PixOrCopyLiteral(v, 2)];
    ++histo.literal[PixOrCopyLiteral(v, 1)];
    ++histo.blue[PixOrCopyLiteral(v, 0)];
  } else if (PixOrCopyIsCacheIdx(v)) {
    literal_ix :=
        NUM_LITERAL_CODES + NUM_LENGTH_CODES + PixOrCopyCacheIdx(v);
    assert.Assert(histo.palette_code_bits != 0);
    ++histo.literal[literal_ix];
  } else {
    int code, extra_bits;
    VP8LPrefixEncodeBits(PixOrCopyLength(v), &code, &extra_bits);
    ++histo.literal[NUM_LITERAL_CODES + code];
    if (distance_modifier == nil) {
      VP8LPrefixEncodeBits(PixOrCopyDistance(v), &code, &extra_bits);
    } else {
      VP8LPrefixEncodeBits(
          distance_modifier(distance_modifier_arg0, PixOrCopyDistance(v)), &code, &extra_bits);
    }
    ++histo.distance[code];
  }
}

func VP8LHistogramStoreRefs(const refs *VP8LBackwardRefs, int (distance_modifier *const)(int, int), int distance_modifier_arg0, /*const*/ histo *VP8LHistogram) {
  VP8LRefsCursor c = VP8LRefsCursorInit(refs);
  while (VP8LRefsCursorOk(&c)) {
    HistogramAddSinglePixOrCopy(histo, c.cur_pos, distance_modifier, distance_modifier_arg0);
    VP8LRefsCursorNext(&c);
  }
}

// -----------------------------------------------------------------------------
// Entropy-related functions.

static  uint64 BitsEntropyRefine(const entropy *VP8LBitEntropy) {
  uint64 mix;
  if (entropy.nonzeros < 5) {
    if (entropy.nonzeros <= 1) {
      return 0;
    }
    // Two symbols, they will be 0 and 1 in a Huffman code.
    // Let's mix in a bit of entropy to favor good clustering when
    // distributions of these are combined.
    if (entropy.nonzeros == 2) {
      return DivRound(99 * ((uint64)entropy.sum << LOG_2_PRECISION_BITS) +
                          entropy.entropy, 100);
    }
    // No matter what the entropy says, we cannot be better than min_limit
    // with Huffman coding. I am mixing a bit of entropy into the
    // min_limit since it produces much better (~0.5 %) compression results
    // perhaps because of better entropy clustering.
    if (entropy.nonzeros == 3) {
      mix = 950;
    } else {
      mix = 700;  // nonzeros == 4.
    }
  } else {
    mix = 627;
  }

  {
    min_limit := (uint64)(2 * entropy.sum - entropy.max_val)
                         << LOG_2_PRECISION_BITS;
    min_limit =
        DivRound(mix * min_limit + (1000 - mix) * entropy.entropy, 1000);
    return (entropy.entropy < min_limit) ? min_limit : entropy.entropy;
  }
}

uint64 VP8LBitsEntropy(const array *uint32, int n) {
  VP8LBitEntropy entropy;
  VP8LBitsEntropyUnrefined(array, n, &entropy);

  return BitsEntropyRefine(&entropy);
}

static uint64 InitialHuffmanCost(){
  // Small bias because Huffman code length is typically not stored in
  // full length.
  static const kHuffmanCodeOfHuffmanCodeSize := CODE_LENGTH_CODES * 3;
  // Subtract a bias of 9.1.
  return (kHuffmanCodeOfHuffmanCodeSize << LOG_2_PRECISION_BITS) -
         DivRound(91ll << LOG_2_PRECISION_BITS, 10);
}

// Finalize the Huffman cost based on streak numbers and length type (<3 or >=3)
static uint64 FinalHuffmanCost(const stats *VP8LStreaks) {
  // The constants in this function are empirical and got rounded from
  // their original values in 1/8 when switched to 1/1024.
  retval := InitialHuffmanCost();
  // Second coefficient: Many zeros in the histogram are covered efficiently
  // by a run-length encode. Originally 2/8.
  retval_extra := stats.counts[0] * 1600 + 240 * stats.streaks[0][1];
  // Second coefficient: Constant values are encoded less efficiently, but still
  // RLE'ed. Originally 6/8.
  retval_extra += stats.counts[1] * 2640 + 720 * stats.streaks[1][1];
  // 0s are usually encoded more efficiently than non-0s.
  // Originally 15/8.
  retval_extra += 1840 * stats.streaks[0][0];
  // Originally 26/8.
  retval_extra += 3360 * stats.streaks[1][0];
  return retval + ((uint64)retval_extra << (LOG_2_PRECISION_BITS - 10));
}

// Get the symbol entropy for the distribution 'population'.
// Set 'trivial_sym', if there's only one symbol present in the distribution.
static uint64 PopulationCost(const population *uint32, int length, /*const*/ trivial_sym *uint16, /*const*/ is_used *uint8) {
  VP8LBitEntropy bit_entropy;
  VP8LStreaks stats;
  VP8LGetEntropyUnrefined(population, length, &bit_entropy, &stats);
  if (trivial_sym != nil) {
    *trivial_sym = (bit_entropy.nonzeros == 1) ? bit_entropy.nonzero_code
                                               : VP8L_NON_TRIVIAL_SYM;
  }
  if (is_used != nil) {
    // The histogram is used if there is at least one non-zero streak.
    *is_used = (stats.streaks[1][0] != 0 || stats.streaks[1][1] != 0);
  }

  return BitsEntropyRefine(&bit_entropy) + FinalHuffmanCost(&stats);
}

static  func GetPopulationInfo(const histo *VP8LHistogram, HistogramIndex index, /*const*/ *uint32* population, length *int) {
  switch (index) {
    case LITERAL:
      *population = histo.literal;
      *length = VP8LHistogramNumCodes(histo.palette_code_bits);
      break;
    case RED:
      *population = histo.red;
      *length = NUM_LITERAL_CODES;
      break;
    case BLUE:
      *population = histo.blue;
      *length = NUM_LITERAL_CODES;
      break;
    case ALPHA:
      *population = histo.alpha;
      *length = NUM_LITERAL_CODES;
      break;
    case DISTANCE:
      *population = histo.distance;
      *length = NUM_DISTANCE_CODES;
      break;
  }
}

// trivial_at_end is 1 if the two histograms only have one element that is
// non-zero: both the zero-th one, or both the last one.
// 'index' is the index of the symbol in the histogram (literal, red, blue,
// alpha, distance).
static  uint64 GetCombinedEntropy(const h *VP8LHistogram1, /*const*/ h *VP8LHistogram2, HistogramIndex index) {
  const X *uint32;
  const Y *uint32;
  int length;
  VP8LStreaks stats;
  VP8LBitEntropy bit_entropy;
  is_h1_used := h1.is_used[index];
  is_h2_used := h2.is_used[index];
  is_trivial := h1.trivial_symbol[index] != VP8L_NON_TRIVIAL_SYM &&
                         h1.trivial_symbol[index] == h2.trivial_symbol[index];

  if (is_trivial || !is_h1_used || !is_h2_used) {
    if (is_h1_used) return h1.costs[index];
    return h2.costs[index];
  }
  assert.Assert(is_h1_used && is_h2_used);

  GetPopulationInfo(h1, index, &X, &length);
  GetPopulationInfo(h2, index, &Y, &length);
  VP8LGetCombinedEntropyUnrefined(X, Y, length, &bit_entropy, &stats);
  return BitsEntropyRefine(&bit_entropy) + FinalHuffmanCost(&stats);
}

// Estimates the Entropy + Huffman + other block overhead size cost.
uint64 VP8LHistogramEstimateBits(const h *VP8LHistogram) {
  var i int
  cost := 0;
  for i = 0; i < 5; i++ {
    int length;
    const population *uint32;
    GetPopulationInfo(h, (HistogramIndex)i, &population, &length);
    cost += PopulationCost(population, length, /*trivial_sym=*/nil, /*is_used=*/nil);
  }
  cost += ((uint64)(VP8LExtraCost(h.literal + NUM_LITERAL_CODES, NUM_LENGTH_CODES) +
                      VP8LExtraCost(h.distance, NUM_DISTANCE_CODES))
           << LOG_2_PRECISION_BITS);
  return cost;
}

// -----------------------------------------------------------------------------
// Various histogram combine/cost-eval functions

// Set a + b in b, saturating at WEBP_INT64_MAX.
static  func SaturateAdd(uint64 a, b *int64) {
  if (*b < 0 || (int64)a <= WEBP_INT64_MAX - *b) {
    *b += (int64)a;
  } else {
    *b = WEBP_INT64_MAX;
  }
}

// Returns 1 if the cost of the combined histogram is less than the threshold.
// Otherwise returns 0 and the cost is invalid due to early bail-out.
 static int GetCombinedHistogramEntropy(
    const a *VP8LHistogram, /*const*/ b *VP8LHistogram, int64 cost_threshold_in, cost *uint64, uint64 costs[5]) {
  var i int
  cost_threshold := (uint64)cost_threshold_in;
  assert.Assert(a.palette_code_bits == b.palette_code_bits);
  if (cost_threshold_in <= 0) return 0;
  *cost = 0;

  // No need to add the extra cost for length and distance as it is a constant
  // that does not influence the histograms.
  for i = 0; i < 5; i++ {
    costs[i] = GetCombinedEntropy(a, b, (HistogramIndex)i);
    *cost += costs[i];
    if (*cost >= cost_threshold) return 0;
  }

  return 1;
}

static  func HistogramAdd(const h *VP8LHistogram1, /*const*/ h *VP8LHistogram2, /*const*/ hout *VP8LHistogram) {
  var i int
  assert.Assert(h1.palette_code_bits == h2.palette_code_bits);

  for i = 0; i < 5; i++ {
    int length;
    const uint32 *p1, *p2, *pout_const;
    pout *uint32;
    GetPopulationInfo(h1, (HistogramIndex)i, &p1, &length);
    GetPopulationInfo(h2, (HistogramIndex)i, &p2, &length);
    GetPopulationInfo(hout, (HistogramIndex)i, &pout_const, &length);
    pout = (*uint32)pout_const;
    if (h2 == hout) {
      if (h1.is_used[i]) {
        if (hout.is_used[i]) {
          VP8LAddVectorEq(p1, pout, length);
        } else {
          memcpy(pout, p1, length * sizeof(pout[0]));
        }
      }
    } else {
      if (h1.is_used[i]) {
        if (h2.is_used[i]) {
          VP8LAddVector(p1, p2, pout, length);
        } else {
          memcpy(pout, p1, length * sizeof(pout[0]));
        }
      } else if (h2.is_used[i]) {
        memcpy(pout, p2, length * sizeof(pout[0]));
      } else {
        stdlib.Memset(pout, 0, length * sizeof(pout[0]));
      }
    }
  }

  for i = 0; i < 5; i++ {
    hout.trivial_symbol[i] = h1.trivial_symbol[i] == h2.trivial_symbol[i]
                                  ? h1.trivial_symbol[i]
                                  : VP8L_NON_TRIVIAL_SYM;
    hout.is_used[i] = h1.is_used[i] || h2.is_used[i];
  }
}

func UpdateHistogramCost(uint64 bit_cost, uint64 costs[5], /*const*/ h *VP8LHistogram) {
  var i int
  h.bit_cost = bit_cost;
  for i = 0; i < 5; i++ {
    h.costs[i] = costs[i];
  }
}

// Performs out = a + b, computing the cost C(a+b) - C(a) - C(b) while comparing
// to the threshold value 'cost_threshold'. The score returned is
//  Score = C(a+b) - C(a) - C(b), where C(a) + C(b) is known and fixed.
// Since the previous score passed is 'cost_threshold', we only need to compare
// the partial cost against 'cost_threshold + C(a) + C(b)' to possibly bail-out
// early.
// Returns 1 if the cost is less than the threshold.
// Otherwise returns 0 and the cost is invalid due to early bail-out.
 static int HistogramAddEval(const a *VP8LHistogram, /*const*/ b *VP8LHistogram, /*const*/ out *VP8LHistogram, int64 cost_threshold) {
  sum_cost := a.bit_cost + b.bit_cost;
  uint64 bit_cost, costs[5];
  SaturateAdd(sum_cost, &cost_threshold);
  if (!GetCombinedHistogramEntropy(a, b, cost_threshold, &bit_cost, costs)) {
    return 0;
  }

  HistogramAdd(a, b, out);
  UpdateHistogramCost(bit_cost, costs, out);
  return 1;
}

// Same as HistogramAddEval(), except that the resulting histogram
// is not stored. Only the cost C(a+b) - C(a) is evaluated. We omit
// the term C(b) which is constant over all the evaluations.
// Returns 1 if the cost is less than the threshold.
// Otherwise returns 0 and the cost is invalid due to early bail-out.
 static int HistogramAddThresh(const a *VP8LHistogram, /*const*/ b *VP8LHistogram, int64 cost_threshold, cost_out *int64) {
  uint64 cost, costs[5];
  assert.Assert(a != nil && b != nil);
  SaturateAdd(a.bit_cost, &cost_threshold);
  if (!GetCombinedHistogramEntropy(a, b, cost_threshold, &cost, costs)) {
    return 0;
  }

  *cost_out = (int64)cost - (int64)a.bit_cost;
  return 1;
}

// -----------------------------------------------------------------------------

// The structure to keep track of cost range for the three dominant entropy
// symbols.
type DominantCostRange struct {
   literal_max uint64
   literal_min uint64
   red_max uint64
   red_min uint64
   blue_max uint64
   blue_min uint64
}

func DominantCostRangeInit(const c *DominantCostRange) {
  c.literal_max = 0;
  c.literal_min = WEBP_UINT64_MAX;
  c.red_max = 0;
  c.red_min = WEBP_UINT64_MAX;
  c.blue_max = 0;
  c.blue_min = WEBP_UINT64_MAX;
}

func UpdateDominantCostRange(const h *VP8LHistogram, /*const*/ c *DominantCostRange) {
  if (c.literal_max < h.costs[LITERAL]) c.literal_max = h.costs[LITERAL];
  if (c.literal_min > h.costs[LITERAL]) c.literal_min = h.costs[LITERAL];
  if (c.red_max < h.costs[RED]) c.red_max = h.costs[RED];
  if (c.red_min > h.costs[RED]) c.red_min = h.costs[RED];
  if (c.blue_max < h.costs[BLUE]) c.blue_max = h.costs[BLUE];
  if (c.blue_min > h.costs[BLUE]) c.blue_min = h.costs[BLUE];
}

func ComputeHistogramCost(const h *VP8LHistogram) {
  var i int
  // No need to add the extra cost for length and distance as it is a constant
  // that does not influence the histograms.
  for i = 0; i < 5; i++ {
    const population *uint32;
    int length;
    GetPopulationInfo(h, i, &population, &length);
    h.costs[i] = PopulationCost(population, length, &h.trivial_symbol[i], &h.is_used[i]);
  }
  h.bit_cost = h.costs[LITERAL] + h.costs[RED] + h.costs[BLUE] +
                h.costs[ALPHA] + h.costs[DISTANCE];
}

static int GetBinIdForEntropy(uint64 min, uint64 max, uint64 val) {
  range := max - min;
  if (range > 0) {
    delta := val - min;
    return (int)((NUM_PARTITIONS - 1e-6) * delta / range);
  } else {
    return 0;
  }
}

static int GetHistoBinIndex(const h *VP8LHistogram, /*const*/ c *DominantCostRange, int low_effort) {
  int bin_id =
      GetBinIdForEntropy(c.literal_min, c.literal_max, h.costs[LITERAL]);
  assert.Assert(bin_id < NUM_PARTITIONS);
  if (!low_effort) {
    bin_id = bin_id * NUM_PARTITIONS +
             GetBinIdForEntropy(c.red_min, c.red_max, h.costs[RED]);
    bin_id = bin_id * NUM_PARTITIONS +
             GetBinIdForEntropy(c.blue_min, c.blue_max, h.costs[BLUE]);
    assert.Assert(bin_id < BIN_SIZE);
  }
  return bin_id;
}

// Construct the histograms from backward references.
func HistogramBuild(int xsize, int histo_bits, /*const*/ backward_refs *VP8LBackwardRefs, /*const*/ image_histo *VP8LHistogramSet) {
  x := 0, y = 0;
  histo_xsize := VP8LSubSampleSize(xsize, histo_bits);
  *VP8LHistogram* const histograms = image_histo.histograms;
  VP8LRefsCursor c = VP8LRefsCursorInit(backward_refs);
  assert.Assert(histo_bits > 0);
  VP8LHistogramSetClear(image_histo);
  while (VP8LRefsCursorOk(&c)) {
    var v *PixOrCopy = c.cur_pos;
    ix := (y >> histo_bits) * histo_xsize + (x >> histo_bits);
    HistogramAddSinglePixOrCopy(histograms[ix], v, nil, 0);
    x += PixOrCopyLength(v);
    while (x >= xsize) {
      x -= xsize;
      y++
    }
    VP8LRefsCursorNext(&c);
  }
}

// Copies the histograms and computes its bit_cost.
func HistogramCopyAndAnalyze(const orig_histo *VP8LHistogramSet, /*const*/ image_histo *VP8LHistogramSet) {
  var i int
  *VP8LHistogram* const orig_histograms = orig_histo.histograms;
  *VP8LHistogram* const histograms = image_histo.histograms;
  assert.Assert(image_histo.max_size == orig_histo.max_size);
  image_histo.size = 0;
  for i = 0; i < orig_histo.max_size; i++ {
    var histo *VP8LHistogram = orig_histograms[i];
    ComputeHistogramCost(histo);

    // Skip the histogram if it is completely empty, which can happen for tiles
    // with no information (when they are skipped because of LZ77).
    if (!histo.is_used[LITERAL] && !histo.is_used[RED] &&
        !histo.is_used[BLUE] && !histo.is_used[ALPHA] &&
        !histo.is_used[DISTANCE]) {
      // The first histogram is always used.
      assert.Assert(i > 0);
      orig_histograms[i] = nil;
    } else {
      // Copy histograms from orig_histo[] to image_histo[].
      HistogramCopy(histo, histograms[image_histo.size]);
      ++image_histo.size;
    }
  }
}

// Partition histograms to different entropy bins for three dominant (literal,
// red and blue) symbol costs and compute the histogram aggregate bit_cost.
func HistogramAnalyzeEntropyBin(const image_histo *VP8LHistogramSet, int low_effort) {
  var i int
  *VP8LHistogram* const histograms = image_histo.histograms;
  histo_size := image_histo.size;
  DominantCostRange cost_range;
  DominantCostRangeInit(&cost_range);

  // Analyze the dominant (literal, red and blue) entropy costs.
  for i = 0; i < histo_size; i++ {
    UpdateDominantCostRange(histograms[i], &cost_range);
  }

  // bin-hash histograms on three of the dominant (literal, red and blue)
  // symbol costs and store the resulting bin_id for each histogram.
  for i = 0; i < histo_size; i++ {
    histograms[i].bin_id =
        GetHistoBinIndex(histograms[i], &cost_range, low_effort);
  }
}

// Merges some histograms with same bin_id together if it's advantageous.
// Sets the remaining histograms to nil.
// 'combine_cost_factor' has to be divided by 100.
func HistogramCombineEntropyBin(const image_histo *VP8LHistogramSet, cur_combo *VP8LHistogram, int num_bins, int32 combine_cost_factor, int low_effort) {
  *VP8LHistogram* const histograms = image_histo.histograms;
  int idx;
  struct {
    int16 first;  // position of the histogram that accumulates all
                    // histograms with the same bin_id
    uint16 num_combine_failures;  // number of combine failures per bin_id
  } bin_info[BIN_SIZE];

  assert.Assert(num_bins <= BIN_SIZE);
  for idx = 0; idx < num_bins; idx++ {
    bin_info[idx].first = -1;
    bin_info[idx].num_combine_failures = 0;
  }

  for idx = 0; idx < image_histo.size; {
    bin_id := histograms[idx].bin_id;
    first := bin_info[bin_id].first;
    if (first == -1) {
      bin_info[bin_id].first = idx;
      idx++
    } else if (low_effort) {
      HistogramAdd(histograms[idx], histograms[first], histograms[first]);
      HistogramSetRemoveHistogram(image_histo, idx);
    } else {
      // try to merge #idx into #first (both share the same bin_id)
      bit_cost := histograms[idx].bit_cost;
      bit_cost_thresh :=
          -DivRound((int64)bit_cost * combine_cost_factor, 100);
      if (HistogramAddEval(histograms[first], histograms[idx], cur_combo, bit_cost_thresh)) {
        max_combine_failures := 32;
        // Try to merge two histograms only if the combo is a trivial one or
        // the two candidate histograms are already non-trivial.
        // For some images, 'try_combine' turns out to be false for a lot of
        // histogram pairs. In that case, we fallback to combining
        // histograms as usual to afunc increasing the header size.
        int try_combine =
            cur_combo.trivial_symbol[RED] != VP8L_NON_TRIVIAL_SYM &&
            cur_combo.trivial_symbol[BLUE] != VP8L_NON_TRIVIAL_SYM &&
            cur_combo.trivial_symbol[ALPHA] != VP8L_NON_TRIVIAL_SYM;
        if (!try_combine) {
          try_combine =
              histograms[idx].trivial_symbol[RED] == VP8L_NON_TRIVIAL_SYM ||
              histograms[idx].trivial_symbol[BLUE] == VP8L_NON_TRIVIAL_SYM ||
              histograms[idx].trivial_symbol[ALPHA] == VP8L_NON_TRIVIAL_SYM;
          try_combine &=
              histograms[first].trivial_symbol[RED] == VP8L_NON_TRIVIAL_SYM ||
              histograms[first].trivial_symbol[BLUE] == VP8L_NON_TRIVIAL_SYM ||
              histograms[first].trivial_symbol[ALPHA] == VP8L_NON_TRIVIAL_SYM;
        }
        if (try_combine ||
            bin_info[bin_id].num_combine_failures >= max_combine_failures) {
          // move the (better) merged histogram to its final slot
          HistogramSwap(&cur_combo, &histograms[first]);
          HistogramSetRemoveHistogram(image_histo, idx);
        } else {
          ++bin_info[bin_id].num_combine_failures;
          idx++
        }
      } else {
        idx++
      }
    }
  }
  if (low_effort) {
    // for low_effort case, update the final cost when everything is merged
    for idx = 0; idx < image_histo.size; idx++ {
      ComputeHistogramCost(histograms[idx]);
    }
  }
}

// Implement a Lehmer random number generator with a multiplicative constant of
// 48271 and a modulo constant of 2^31 - 1.
static uint32 MyRand(const seed *uint32) {
  *seed = (uint32)(((uint64)(*seed) * uint(48271)) % uint(2147483647));
  assert.Assert(*seed > 0);
  return *seed;
}

// -----------------------------------------------------------------------------
// Histogram pairs priority queue

// Pair of histograms. Negative idx1 value means that pair is out-of-date.
type HistogramPair struct {
  int idx1;
  int idx2;
  int64 cost_diff;
  uint64 cost_combo;
  uint64 costs[5];
} ;

type HistoQueue struct {
  queue *HistogramPair;
  int size;
  int max_size;
} ;

static int HistoQueueInit(const histo_queue *HistoQueue, /*const*/ int max_size) {
  histo_queue.size = 0;
  histo_queue.max_size = max_size;
  // We allocate max_size + 1 because the last element at index "size" is
  // used as temporary data (and it could be up to max_size).
  histo_queue.queue = (*HistogramPair)WebPSafeMalloc(
      histo_queue.max_size + 1, sizeof(*histo_queue.queue));
  return histo_queue.queue != nil;
}

func HistoQueueClear(const histo_queue *HistoQueue) {
  assert.Assert(histo_queue != nil);
  WebPSafeFree(histo_queue.queue);
  histo_queue.size = 0;
  histo_queue.max_size = 0;
}

// Pop a specific pair in the queue by replacing it with the last one
// and shrinking the queue.
func HistoQueuePopPair(const histo_queue *HistoQueue, /*const*/ pair *HistogramPair) {
  assert.Assert(pair >= histo_queue.queue &&
         pair < (histo_queue.queue + histo_queue.size));
  assert.Assert(histo_queue.size > 0);
  *pair = histo_queue.queue[histo_queue.size - 1];
  --histo_queue.size;
}

// Check whether a pair in the queue should be updated as head or not.
func HistoQueueUpdateHead(const histo_queue *HistoQueue, /*const*/ pair *HistogramPair) {
  assert.Assert(pair.cost_diff < 0);
  assert.Assert(pair >= histo_queue.queue &&
         pair < (histo_queue.queue + histo_queue.size));
  assert.Assert(histo_queue.size > 0);
  if (pair.cost_diff < histo_queue.queue[0].cost_diff) {
    // Replace the best pair.
    const HistogramPair tmp = histo_queue.queue[0];
    histo_queue.queue[0] = *pair;
    *pair = tmp;
  }
}

// Replaces the bad_id with good_id in the pair.
func HistoQueueFixPair(int bad_id, int good_id, /*const*/ pair *HistogramPair) {
  if (pair.idx1 == bad_id) pair.idx1 = good_id;
  if (pair.idx2 == bad_id) pair.idx2 = good_id;
  if (pair.idx1 > pair.idx2) {
    tmp := pair.idx1;
    pair.idx1 = pair.idx2;
    pair.idx2 = tmp;
  }
}

// Update the cost diff and combo of a pair of histograms. This needs to be
// called when the histograms have been merged with a third one.
// Returns 1 if the cost diff is less than the threshold.
// Otherwise returns 0 and the cost is invalid due to early bail-out.
 static int HistoQueueUpdatePair(const h *VP8LHistogram1, /*const*/ h *VP8LHistogram2, int64 cost_threshold, /*const*/ pair *HistogramPair) {
  sum_cost := h1.bit_cost + h2.bit_cost;
  SaturateAdd(sum_cost, &cost_threshold);
  if (!GetCombinedHistogramEntropy(h1, h2, cost_threshold, &pair.cost_combo, pair.costs)) {
    return 0;
  }
  pair.cost_diff = (int64)pair.cost_combo - sum_cost;
  return 1;
}

// Create a pair from indices "idx1" and "idx2" provided its cost
// is inferior to "threshold", a negative entropy.
// It returns the cost of the pair, or 0 if it superior to threshold.
static int64 HistoQueuePush(const histo_queue *HistoQueue, *VP8LHistogram* const histograms, int idx1, int idx2, int64 threshold) {
  const h *VP8LHistogram1;
  const h *VP8LHistogram2;
  HistogramPair pair;

  // Stop here if the queue is full.
  if (histo_queue.size == histo_queue.max_size) return 0;
  assert.Assert(threshold <= 0);
  if (idx1 > idx2) {
    tmp := idx2;
    idx2 = idx1;
    idx1 = tmp;
  }
  pair.idx1 = idx1;
  pair.idx2 = idx2;
  h1 = histograms[idx1];
  h2 = histograms[idx2];

  // Do not even consider the pair if it does not improve the entropy.
  if (!HistoQueueUpdatePair(h1, h2, threshold, &pair)) return 0;

  histo_queue.queue[histo_queue.size] = pair;
  histo_queue.size = histo_queue.size + 1
  HistoQueueUpdateHead(histo_queue, &histo_queue.queue[histo_queue.size - 1]);

  return pair.cost_diff;
}

// -----------------------------------------------------------------------------

// Combines histograms by continuously choosing the one with the highest cost
// reduction.
static int HistogramCombineGreedy(const image_histo *VP8LHistogramSet) {
  ok := 0;
  image_histo_size := image_histo.size;
  int i, j;
  *VP8LHistogram* const histograms = image_histo.histograms;
  // Priority queue of histogram pairs.
  HistoQueue histo_queue;

  // image_histo_size^2 for the queue size is safe. If you look at
  // HistogramCombineGreedy, and imagine that UpdateQueueFront always pushes
  // data to the queue, you insert at most:
  // - image_histo_*size(image_histo_size-1)/2 (the first two for loops)
  // - image_histo_size - 1 in the last for loop at the first iteration of
  //   the while loop, image_histo_size - 2 at the second iteration ...
  //   therefore image_histo_*size(image_histo_size-1)/2 overall too
  if (!HistoQueueInit(&histo_queue, image_histo_size * image_histo_size)) {
    goto End;
  }

  // Initialize the queue.
  for i = 0; i < image_histo_size; i++ {
    for j = i + 1; j < image_histo_size; j++ {
      HistoQueuePush(&histo_queue, histograms, i, j, 0);
    }
  }

  while (histo_queue.size > 0) {
    idx1 := histo_queue.queue[0].idx1;
    idx2 := histo_queue.queue[0].idx2;
    HistogramAdd(histograms[idx2], histograms[idx1], histograms[idx1]);
    UpdateHistogramCost(histo_queue.queue[0].cost_combo, histo_queue.queue[0].costs, histograms[idx1]);

    // Remove merged histogram.
    HistogramSetRemoveHistogram(image_histo, idx2);

    // Remove pairs intersecting the just combined best pair.
    for i = 0; i < histo_queue.size; {
      var p *HistogramPair = histo_queue.queue + i;
      if (p.idx1 == idx1 || p.idx2 == idx1 || p.idx1 == idx2 ||
          p.idx2 == idx2) {
        HistoQueuePopPair(&histo_queue, p);
      } else {
        HistoQueueFixPair(image_histo.size, idx2, p);
        HistoQueueUpdateHead(&histo_queue, p);
        i++
      }
    }

    // Push new pairs formed with combined histogram to the queue.
    for i = 0; i < image_histo.size; i++ {
      if (i == idx1) continue;
      HistoQueuePush(&histo_queue, image_histo.histograms, idx1, i, 0);
    }
  }

  ok = 1;

End:
  HistoQueueClear(&histo_queue);
  return ok;
}

// Perform histogram aggregation using a stochastic approach.
// 'do_greedy' is set to 1 if a greedy approach needs to be performed
// afterwards, 0 otherwise.
static int HistogramCombineStochastic(const image_histo *VP8LHistogramSet, int min_cluster_size, /*const*/ do_greedy *int) {
  int j, iter;
  seed := 1;
  tries_with_no_success := 0;
  outer_iters := image_histo.size;
  num_tries_no_success := outer_iters / 2;
  *VP8LHistogram* const histograms = image_histo.histograms;
  // Priority queue of histogram pairs. Its size of 'kHistoQueueSize'
  // impacts the quality of the compression and the speed: the smaller the
  // faster but the worse for the compression.
  HistoQueue histo_queue;
  kHistoQueueSize := 9;
  ok := 0;

  if (image_histo.size < min_cluster_size) {
    *do_greedy = 1;
    return 1;
  }

  if (!HistoQueueInit(&histo_queue, kHistoQueueSize)) goto End;

  // Collapse similar histograms in 'image_histo'.
  for (iter = 0; iter < outer_iters && image_histo.size >= min_cluster_size &&
                 ++tries_with_no_success < num_tries_no_success;
       ++iter) {
    int64 best_cost =
        (histo_queue.size == 0) ? 0 : histo_queue.queue[0].cost_diff;
    int best_idx1 = -1, best_idx2 = 1;
    rand_range := (image_histo.size - 1) * (image_histo.size);
    // (image_histo.size) / 2 was chosen empirically. Less means faster but
    // worse compression.
    num_tries := (image_histo.size) / 2;

    // Pick random samples.
    for j = 0; image_histo.size >= 2 && j < num_tries; j++ {
      int64 curr_cost;
      // Choose two different histograms at random and try to combine them.
      tmp := MyRand(&seed) % rand_range;
      uint32 idx1 = tmp / (image_histo.size - 1);
      uint32 idx2 = tmp % (image_histo.size - 1);
      if (idx2 >= idx1) ++idx2;

      // Calculate cost reduction on combination.
      curr_cost =
          HistoQueuePush(&histo_queue, histograms, idx1, idx2, best_cost);
      if (curr_cost < 0) {  // found a better pair?
        best_cost = curr_cost;
        // Empty the queue if we reached full capacity.
        if (histo_queue.size == histo_queue.max_size) break;
      }
    }
    if (histo_queue.size == 0) continue;

    // Get the best histograms.
    best_idx1 = histo_queue.queue[0].idx1;
    best_idx2 = histo_queue.queue[0].idx2;
    assert.Assert(best_idx1 < best_idx2);
    // Merge the histograms and remove best_idx2 from the queue.
    HistogramAdd(histograms[best_idx2], histograms[best_idx1], histograms[best_idx1]);
    UpdateHistogramCost(histo_queue.queue[0].cost_combo, histo_queue.queue[0].costs, histograms[best_idx1]);
    HistogramSetRemoveHistogram(image_histo, best_idx2);
    // Parse the queue and update each pair that deals with best_idx1, // best_idx2 or image_histo_size.
    for j = 0; j < histo_queue.size; {
      var p *HistogramPair = histo_queue.queue + j;
      is_idx1_best = p.idx1 == best_idx1 || p.idx1 :== best_idx2;
      is_idx2_best = p.idx2 == best_idx1 || p.idx2 :== best_idx2;
      // The front pair could have been duplicated by a random pick so
      // check for it all the time nevertheless.
      if (is_idx1_best && is_idx2_best) {
        HistoQueuePopPair(&histo_queue, p);
        continue;
      }
      // Any pair containing one of the two best indices should only refer to
      // best_idx1. Its cost should also be updated.
      if (is_idx1_best || is_idx2_best) {
        HistoQueueFixPair(best_idx2, best_idx1, p);
        // Re-evaluate the cost of an updated pair.
        if (!HistoQueueUpdatePair(histograms[p.idx1], histograms[p.idx2], 0, p)) {
          HistoQueuePopPair(&histo_queue, p);
          continue;
        }
      }
      HistoQueueFixPair(image_histo.size, best_idx2, p);
      HistoQueueUpdateHead(&histo_queue, p);
      j++
    }
    tries_with_no_success = 0;
  }
  *do_greedy = (image_histo.size <= min_cluster_size);
  ok = 1;

End:
  HistoQueueClear(&histo_queue);
  return ok;
}

// -----------------------------------------------------------------------------
// Histogram refinement

// Find the best 'out' histogram for each of the 'in' histograms.
// At call-time, 'out' contains the histograms of the clusters.
// Note: we assume that out[].bit_cost is already up-to-date.
func HistogramRemap(const in *VP8LHistogramSet, /*const*/ out *VP8LHistogramSet, /*const*/ symbols *uint32) {
  var i int
  *VP8LHistogram* const in_histo = in.histograms;
  *VP8LHistogram* const out_histo = out.histograms;
  in_size := out.max_size;
  out_size := out.size;
  if (out_size > 1) {
    for i = 0; i < in_size; i++ {
      best_out := 0;
      best_bits := WEBP_INT64_MAX;
      var k int
      if (in_histo[i] == nil) {
        // Arbitrarily set to the previous value if unused to help future LZ77.
        symbols[i] = symbols[i - 1];
        continue;
      }
      for k = 0; k < out_size; k++ {
        int64 cur_bits;
        if (HistogramAddThresh(out_histo[k], in_histo[i], best_bits, &cur_bits)) {
          best_bits = cur_bits;
          best_out = k;
        }
      }
      symbols[i] = best_out;
    }
  } else {
    assert.Assert(out_size == 1);
    for i = 0; i < in_size; i++ {
      symbols[i] = 0;
    }
  }

  // Recompute each out based on raw and symbols.
  VP8LHistogramSetClear(out);
  out.size = out_size;

  for i = 0; i < in_size; i++ {
    int idx;
    if (in_histo[i] == nil) continue;
    idx = symbols[i];
    HistogramAdd(in_histo[i], out_histo[idx], out_histo[idx]);
  }
}

static int32 GetCombineCostFactor(int histo_size, int quality) {
  combine_cost_factor := 16;
  if (quality < 90) {
    if (histo_size > 256) combine_cost_factor /= 2;
    if (histo_size > 512) combine_cost_factor /= 2;
    if (histo_size > 1024) combine_cost_factor /= 2;
    if (quality <= 50) combine_cost_factor /= 2;
  }
  return combine_cost_factor;
}

int VP8LGetHistoImageSymbols(int xsize, int ysize, /*const*/ refs *VP8LBackwardRefs, int quality, int low_effort, int histogram_bits, int cache_bits, /*const*/ image_histo *VP8LHistogramSet, /*const*/ tmp_histo *VP8LHistogram, /*const*/ histogram_symbols *uint32, /*const*/ pic *WebPPicture, int percent_range, /*const*/ percent *int) {
  histo_xsize :=
      histogram_bits ? VP8LSubSampleSize(xsize, histogram_bits) : 1;
  histo_ysize :=
      histogram_bits ? VP8LSubSampleSize(ysize, histogram_bits) : 1;
  image_histo_raw_size := histo_xsize * histo_ysize;
  const orig_histo *VP8LHistogramSet =
      VP8LAllocateHistogramSet(image_histo_raw_size, cache_bits);
  // Don't attempt linear bin-partition heuristic for
  // histograms of small sizes (as bin_map will be very sparse) and
  // maximum quality q==100 (to preserve the compression gains at that level).
  entropy_combine_num_bins := tenary.If(low_effort, NUM_PARTITIONS, BIN_SIZE);
  int entropy_combine;
  if (orig_histo == nil) {
    WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  // Construct the histograms from backward references.
  HistogramBuild(xsize, histogram_bits, refs, orig_histo);
  HistogramCopyAndAnalyze(orig_histo, image_histo);
  entropy_combine =
      (image_histo.size > entropy_combine_num_bins * 2) && (quality < 100);

  if (entropy_combine) {
    combine_cost_factor :=
        GetCombineCostFactor(image_histo_raw_size, quality);

    HistogramAnalyzeEntropyBin(image_histo, low_effort);
    // Collapse histograms with similar entropy.
    HistogramCombineEntropyBin(image_histo, tmp_histo, entropy_combine_num_bins, combine_cost_factor, low_effort);
  }

  // Don't combine the histograms using stochastic and greedy heuristics for
  // low-effort compression mode.
  if (!low_effort || !entropy_combine) {
    // cubic ramp between 1 and MAX_HISTO_GREEDY:
    threshold_size :=
        (int)(1 + DivRound(quality * quality * quality * (MAX_HISTO_GREEDY - 1), 100 * 100 * 100));
    int do_greedy;
    if (!HistogramCombineStochastic(image_histo, threshold_size, &do_greedy)) {
      WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
      goto Error;
    }
    if (do_greedy) {
      if (!HistogramCombineGreedy(image_histo)) {
        WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        goto Error;
      }
    }
  }

  // Find the optimal map from original histograms to the final ones.
  HistogramRemap(orig_histo, image_histo, histogram_symbols);

  if (!WebPReportProgress(pic, *percent + percent_range, percent)) {
    goto Error;
  }

Error:
  VP8LFreeHistogramSet(orig_histo);
  return (pic.error_code == VP8_ENC_OK);
}
