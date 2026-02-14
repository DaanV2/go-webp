package enc

// Copyright 2017 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Improves a given set of backward references by analyzing its bit cost.
// The algorithm is similar to the Zopfli compression algorithm but tailored to
// images.
//
// Author: Vincent Rabaud (vrabaud@google.com)
//

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

const VALUES_IN_BYTE =256

extern func VP8LClearBackwardRefs(/* const */ refs *VP8LBackwardRefs)
extern int VP8LDistanceToPlaneCode(xsize int, dist int)
extern func VP8LBackwardRefsCursorAdd(/* const */ refs *VP8LBackwardRefs, /*const*/ PixOrCopy v)

type CostModel struct {
  uint32 alpha[VALUES_IN_BYTE]
  uint32 red[VALUES_IN_BYTE]
  uint32 blue[VALUES_IN_BYTE]
  uint32 distance[NUM_DISTANCE_CODES]
  literal *uint32
}

func ConvertPopulationCountTableToBitEstimates(
    int num_symbols, /*const*/ uint32 population_counts[], uint32 output[]) {
  sum := 0
  nonzeros := 0
  var i int
  for i = 0; i < num_symbols; i++ {
    sum += population_counts[i]
    if (population_counts[i] > 0) {
      nonzeros++
    }
  }
  if (nonzeros <= 1) {
    stdlib.Memset(output, 0, num_symbols * sizeof(*output))
  } else {
    logsum := VP8LFastLog2(sum)
    for i = 0; i < num_symbols; i++ {
      output[i] = logsum - VP8LFastLog2(population_counts[i])
    }
  }
}

func CostModelBuild(/* const */ m *CostModel,  xsize , cache_bits int, /*const*/ refs *VP8LBackwardRefs) int {
  ok := 0
  var histo *VP8LHistogram = VP8LAllocateHistogram(cache_bits)
  if histo == nil { goto Error }

  // The following code is similar to VP8LHistogramCreate but converts the
  // distance to plane code.
  VP8LHistogramInit(histo, cache_bits, /*init_arrays=*/1)
  VP8LHistogramStoreRefs(refs, VP8LDistanceToPlaneCode, xsize, histo)

  ConvertPopulationCountTableToBitEstimates(
      VP8LHistogramNumCodes(histo.palette_code_bits), histo.literal, m.literal)
  ConvertPopulationCountTableToBitEstimates(VALUES_IN_BYTE, histo.red, m.red)
  ConvertPopulationCountTableToBitEstimates(VALUES_IN_BYTE, histo.blue, m.blue)
  ConvertPopulationCountTableToBitEstimates(VALUES_IN_BYTE, histo.alpha, m.alpha)
  ConvertPopulationCountTableToBitEstimates(NUM_DISTANCE_CODES, histo.distance, m.distance)
  ok = 1

Error:
  histo = nil
  return ok
}

func GetLiteralCost(/* const */ m *CostModel, uint32 v) int64 {
  return (int64)m.alpha[v >> 24] + m.red[(v >> 16) & 0xff] +
         m.literal[(v >> 8) & 0xff] + m.blue[v & 0xff]
}

func GetCacheCost(/* const */ m *CostModel, uint32 idx) int64 {
  literal_idx := VALUES_IN_BYTE + NUM_LENGTH_CODES + idx
  return (int64)m.literal[literal_idx]
}

func GetLengthCost(/* const */ m *CostModel, uint32 length) int64 {
  int code, extra_bits
  VP8LPrefixEncodeBits(length, &code, &extra_bits)
  return (int64)m.literal[VALUES_IN_BYTE + code] +
         ((int64)extra_bits << LOG_2_PRECISION_BITS)
}

func GetDistanceCost(/* const */ m *CostModel, uint32 distance) int64 {
  int code, extra_bits
  VP8LPrefixEncodeBits(distance, &code, &extra_bits)
  return (int64)m.distance[code] +
         ((int64)extra_bits << LOG_2_PRECISION_BITS)
}

func AddSingleLiteralWithCostModel(
    const argb *uint32, /*const*/ hashers *VP8LColorCache, /*const*/ cost_model *CostModel, idx int, use_color_cache int, int64 prev_cost, /*const*/ cost *int64, /*const*/ dist_array *uint16) {
  cost_val := prev_cost
  color := argb[idx]
  ix := use_color_cache ? VP8LColorCacheContains(hashers, color) : -1
  if (ix >= 0) {
    // use_color_cache is true and hashers contains color
    cost_val += DivRound(GetCacheCost(cost_model, ix) * 68, 100)
  } else {
    if use_color_cache { VP8LColorCacheInsert(hashers, color) }
    cost_val += DivRound(GetLiteralCost(cost_model, color) * 82, 100)
  }
  if (cost[idx] > cost_val) {
    cost[idx] = cost_val
    dist_array[idx] = 1;  // only one is inserted.
  }
}

// -----------------------------------------------------------------------------
// CostManager and interval handling

// Empirical value to afunc high memory consumption but good for performance.
const COST_CACHE_INTERVAL_SIZE_MAX =500

// To perform backward reference every pixel at index 'index' is considered and
// the cost for the MAX_LENGTH following pixels computed. Those following pixels
// at index 'index' + k (k from 0 to MAX_LENGTH) have a cost of:
//     cost = distance cost at index + GetLengthCost(cost_model, k)
// and the minimum value is kept. GetLengthCost(cost_model, k) is cached in an
// array of size MAX_LENGTH.
// Instead of performing MAX_LENGTH comparisons per pixel, we keep track of the
// minimal values using intervals of constant cost.
// An interval is defined by the 'index' of the pixel that generated it and
// is only useful in a range of indices from 'start' to 'end' (exclusive), i.e.
// it contains the minimum value for pixels between start and end.
// Intervals are stored in a linked list and ordered by 'start'. When a new
// interval has a better value, old intervals are split or removed. There are
// therefore no overlapping intervals.
type CostInterval struct {
  cost int64
  start int
  end int
  index int
  previous *CostInterval
  next *CostInterval
}

// The GetLengthCost(cost_model, k) are cached in a CostCacheInterval.
type CostCacheInterval struct {
  cost int64
  start int
  end int  // Exclusive.
} 

// This structure is in charge of managing intervals and costs.
// It caches the different CostCacheInterval, caches the different
// GetLengthCost(cost_model, k) in cost_cache and the CostInterval's (whose
// 'count' is limited by COST_CACHE_INTERVAL_SIZE_MAX).
const COST_MANAGER_MAX_FREE_LIST =10
type CostManager struct {
  head *CostInterval
  count int  // The number of stored intervals.
  cache_intervals *CostCacheInterval
  cache_intervals_size uint64
  // Contains the GetLengthCost(cost_model, k).
  cost_cache [MAX_LENGTH]int64
  costs *int64
  dist_array *uint16
  // Most of the time, we only need few intervals . use a free-list, to avoid
  // fragmentation with small allocs in most common cases.
  intervals [COST_MANAGER_MAX_FREE_LIST]CostInterval
  free_intervals *CostInterval
  // These are regularly malloc'd remains. This list can't grow larger than than
  // size COST_CACHE_INTERVAL_SIZE_MAX - COST_MANAGER_MAX_FREE_LIST, note.
  recycled_intervals *CostInterval
}

func CostIntervalAddToFreeList(/* const */ manager *CostManager, /*const*/ interval *CostInterval) {
  interval.next = manager.free_intervals
  manager.free_intervals = interval
}

func CostIntervalIsInFreeList(/* const */ manager *CostManager, /*const*/ interval *CostInterval) int {
  return (interval >= &manager.intervals[0] &&
          interval <= &manager.intervals[COST_MANAGER_MAX_FREE_LIST - 1])
}

func CostManagerInitFreeList(/* const */ manager *CostManager) {
  var i int
  manager.free_intervals = nil
  for i = 0; i < COST_MANAGER_MAX_FREE_LIST; i++ {
    CostIntervalAddToFreeList(manager, &manager.intervals[i])
  }
}

func DeleteIntervalList(/* const */ manager *CostManager, /*const*/ interval *CostInterval) {
  while (interval != nil) {
    var next *CostInterval = interval.next

    interval = next
  }
}

func CostManagerClear(/* const */ manager *CostManager) {
  if (manager == nil){ return;}

  // Clear the interval lists.
  DeleteIntervalList(manager, manager.head)
  manager.head = nil
  DeleteIntervalList(manager, manager.recycled_intervals)
  manager.recycled_intervals = nil

  // Reset pointers, 'count' and 'cache_intervals_size'.
  stdlib.Memset(manager, 0, sizeof(*manager))
  CostManagerInitFreeList(manager)
}

func CostManagerInit(/* const */ manager *CostManager, /*const*/ dist_array *uint16, pix_count int, /*const*/ cost_model *CostModel) int {
  var i int
  cost_cache_size := tenary.If(pix_count > MAX_LENGTH, MAX_LENGTH, pix_count)

  manager.costs = nil
  manager.cache_intervals = nil
  manager.head = nil
  manager.recycled_intervals = nil
  manager.count = 0
  manager.dist_array = dist_array
  CostManagerInitFreeList(manager)

  // Fill in the 'cost_cache'.
  // Has to be done in two passes due to a GCC bug on i686
  // related to https://gcc.gnu.org/bugzilla/show_bug.cgi?id=323
  for i = 0; i < cost_cache_size; i++ {
    manager.cost_cache[i] = GetLengthCost(cost_model, i)
  }
  manager.cache_intervals_size = 1
  for i = 1; i < cost_cache_size; i++ {
    // Get the number of bound intervals.
    if (manager.cost_cache[i] != manager.cost_cache[i - 1]) {
      ++manager.cache_intervals_size
    }
  }

  // With the current cost model, we usually have below 20 intervals.
  // The worst case scenario with a cost model would be if every length has a
  // different cost, hence MAX_LENGTH but that is impossible with the current
  // implementation that spirals around a pixel.
  assert.Assert(manager.cache_intervals_size <= MAX_LENGTH)
//   manager.cache_intervals = (*CostCacheInterval)WebPSafeMalloc(manager.cache_intervals_size, sizeof(*manager.cache_intervals))
//   if (manager.cache_intervals == nil) {
//     CostManagerClear(manager)
//     return 0
//   }
  manager.cache_intervals = make([]CostCacheInterval, manager.cache_intervals_size)

  // Fill in the 'cache_intervals'.
  {
    cur *CostCacheInterval = manager.cache_intervals

    // Consecutive values in 'cost_cache' are compared and if a big enough
    // difference is found, a new interval is created and bounded.
    cur.start = 0
    cur.end = 1
    cur.cost = manager.cost_cache[0]
    for i = 1; i < cost_cache_size; i++ {
      cost_val := manager.cost_cache[i]
      if (cost_val != cur.cost) {
        cur++
        // Initialize an interval.
        cur.start = i
        cur.cost = cost_val
      }
      cur.end = i + 1
    }
    assert.Assert((uint64)(cur - manager.cache_intervals) + 1 == manager.cache_intervals_size)
  }

//   manager.costs = (*int64)WebPSafeMalloc(pix_count, sizeof(*manager.costs))
//   if (manager.costs == nil) {
//     CostManagerClear(manager)
//     return 0
//   }
  manager.costs = make([]int64, pix_count)
  // Set the initial 'costs' to INT64_MAX for every pixel as we will keep the
  // minimum.
  for (i = 0; i < pix_count; ++i) {manager.costs[i] = WEBP_INT64_MAX;}

  return 1
}

// Given the cost and the position that define an interval, update the cost at
// pixel 'i' if it is smaller than the previously computed value.
func UpdateCost(/* const */ manager *CostManager, i int, position int, int64 cost) {
  k := i - position
  assert.Assert(k >= 0 && k < MAX_LENGTH)

  if (manager.costs[i] > cost) {
    manager.costs[i] = cost
    manager.dist_array[i] = k + 1
  }
}

// Given the cost and the position that define an interval, update the cost for
// all the pixels between 'start' and 'end' excluded.
func UpdateCostPerInterval(/* const */ manager *CostManager, start int, end int, position int, int64 cost) {
  var i int
  for (i = start; i < end; ++i) UpdateCost(manager, i, position, cost)
}

// Given two intervals, make 'prev' be the previous one of 'next' in 'manager'.
func ConnectIntervals(/* const */ manager *CostManager, /*const*/ prev *CostInterval, /*const*/ next *CostInterval) {
  if (prev != nil) {
    prev.next = next
  } else {
    manager.head = next
  }

  if next != nil { next.previous = prev }
}

// Pop an interval in the manager.
func PopInterval(/* const */ manager *CostManager, /*const*/ interval *CostInterval) {
  if interval == nil { return }

  ConnectIntervals(manager, interval.previous, interval.next)
  if (CostIntervalIsInFreeList(manager, interval)) {
    CostIntervalAddToFreeList(manager, interval)
  } else {  // recycle regularly malloc'd intervals too
    interval.next = manager.recycled_intervals
    manager.recycled_intervals = interval
  }
  --manager.count
  assert.Assert(manager.count >= 0)
}

// Update the cost at index i by going over all the stored intervals that
// overlap with i.
// If 'do_clean_intervals' is set to something different than 0, intervals that
// end before 'i' will be popped.
func UpdateCostAtIndex(/* const */ manager *CostManager, i int, do_clean_intervals int) {
  current *CostInterval = manager.head

  while (current != nil && current.start <= i) {
    var next *CostInterval = current.next
    if (current.end <= i) {
      if (do_clean_intervals) {
        // We have an outdated interval, remove it.
        PopInterval(manager, current)
      }
    } else {
      UpdateCost(manager, i, current.index, current.cost)
    }
    current = next
  }
}

// Given a current orphan interval and its previous interval, before
// it was orphaned (which can be nil), set it at the right place in the list
// of intervals using the 'start' ordering and the previous interval as a hint.
func PositionOrphanInterval(/* const */ manager *CostManager, /*const*/ current *CostInterval, previous *CostInterval) {
  assert.Assert(current != nil)

  if previous == nil { previous = manager.head }
  while (previous != nil && current.start < previous.start) {
    previous = previous.previous
  }
  while (previous != nil && previous.next != nil &&
         previous.next.start < current.start) {
    previous = previous.next
  }

  if (previous != nil) {
    ConnectIntervals(manager, current, previous.next)
  } else {
    ConnectIntervals(manager, current, manager.head)
  }
  ConnectIntervals(manager, previous, current)
}

// Insert an interval in the list contained in the manager by starting at
// 'interval_in' as a hint. The intervals are sorted by 'start' value.
func InsertInterval(/* const */ manager *CostManager, /*const*/ interval_in *CostInterval, int64 cost, position int, start int, end int) {
  interval_new *CostInterval

  if start >= end { return }
  if (manager.count >= COST_CACHE_INTERVAL_SIZE_MAX) {
    // Serialize the interval if we cannot store it.
    UpdateCostPerInterval(manager, start, end, position, cost)
    return
  }
  if (manager.free_intervals != nil) {
    interval_new = manager.free_intervals
    manager.free_intervals = interval_new.next
  } else if (manager.recycled_intervals != nil) {
    interval_new = manager.recycled_intervals
    manager.recycled_intervals = interval_new.next
  } else {  // malloc for good
    // interval_new = (*CostInterval)WebPSafeMalloc(1, sizeof(*interval_new))
    // if (interval_new == nil) {
    //   // Write down the interval if we cannot create it.
    //   UpdateCostPerInterval(manager, start, end, position, cost)
    //   return
    // }
	interval_new.next = new(CostInterval)
  }

  interval_new.cost = cost
  interval_new.index = position
  interval_new.start = start
  interval_new.end = end
  PositionOrphanInterval(manager, interval_new, interval_in)

  ++manager.count
}

// Given a new cost interval defined by its start at position, its length value
// and distance_cost, add its contributions to the previous intervals and costs.
// If handling the interval or one of its subintervals becomes to heavy, its
// contribution is added to the costs right away.
func PushInterval(/* const */ manager *CostManager, int64 distance_cost, position int, len int) {
  var i uint64
  interval *CostInterval = manager.head
  interval_next *CostInterval
  const cost_cache_intervals *CostCacheInterval = manager.cache_intervals
  // If the interval is small enough, no need to deal with the heavy
  // interval logic, just serialize it right away. This constant is empirical.
  kSkipDistance := 10

  if (len < kSkipDistance) {
    var j int
    for j = position; j < position + len; j++ {
      k := j - position
      var cost_tmp int64
      assert.Assert(k >= 0 && k < MAX_LENGTH)
      cost_tmp = distance_cost + manager.cost_cache[k]

      if (manager.costs[j] > cost_tmp) {
        manager.costs[j] = cost_tmp
        manager.dist_array[j] = k + 1
      }
    }
    return
  }

  for (i = 0
       i < manager.cache_intervals_size && cost_cache_intervals[i].start < len
       ++i) {
    // Define the intersection of the ith interval with the new one.
    start := position + cost_cache_intervals[i].start
    end := position +
        (cost_cache_intervals[i].end > len ? len : cost_cache_intervals[i].end)
    cost := distance_cost + cost_cache_intervals[i].cost

    for (; interval != nil && interval.start < end
         interval = interval_next) {
      interval_next = interval.next

      // Make sure we have some overlap
      if start >= interval.end { continue }

      if (cost >= interval.cost) {
        // When intervals are represented, the lower, the better.
        // [**********************************************************[
        // start                                                    end
        //                   [----------------------------------[
        //                   interval.start        interval.end
        // If we are worse than what we already have, add whatever we have so
        // far up to interval.
        start_new := interval.end
        InsertInterval(manager, interval, cost, position, start, interval.start)
        start = start_new
        if start >= end { break }
        continue
      }

      if (start <= interval.start) {
        if (interval.end <= end) {
          //                   [----------------------------------[
          //                   interval.start        interval.end
          // [**************************************************************[
          // start                                                        end
          // We can safely remove the old interval as it is fully included.
          PopInterval(manager, interval)
        } else {
          //              [------------------------------------[
          //              interval.start          interval.end
          // [*****************************[
          // start                       end
          interval.start = end
          break
        }
      } else {
        if (end < interval.end) {
          // [--------------------------------------------------------------[
          // interval.start                                    interval.end
          //                     [*****************************[
          //                     start                       end
          // We have to split the old interval as it fully contains the new one.
          end_original := interval.end
          interval.end = start
          InsertInterval(manager, interval, interval.cost, interval.index, end, end_original)
          interval = interval.next
          break
        } else {
          // [------------------------------------[
          // interval.start          interval.end
          //                     [*****************************[
          //                     start                       end
          interval.end = start
        }
      }
    }
    // Insert the remaining interval from start to end.
    InsertInterval(manager, interval, cost, position, start, end)
  }
}

func BackwardReferencesHashChainDistanceOnly(xsize, ysize int, /*const*/ argb *uint32, cache_bits int, /*const*/ hash_chain *VP8LHashChain, /*const*/ refs *VP8LBackwardRefs, /*const*/ dist_array *uint16) int {
  var i int
  ok := 0
  cc_init := 0
  pix_count := xsize * ysize
  use_color_cache := (cache_bits > 0)
  literal_array_size := sizeof(*((*CostModel)nil).literal) * VP8LHistogramNumCodes(cache_bits)
  cost_model_size := sizeof(CostModel) + literal_array_size
  var hashers VP8LColorCache 
  
//   var cost_model *CostModel = (*CostModel)WebPSafeCalloc(uint64(1), cost_model_size)
//   var cost_manager *CostManager = (*CostManager)WebPSafeCalloc(uint64(1), sizeof(*cost_manager))
  cost_model := &CostModel{}
  cost_manager := &CostManager{}

  offset_prev := -1, len_prev = -1
  offset_cost := -1
  first_offset_is_constant := -1;  // initialized with 'impossible' value
  reach := 0

//   if cost_model == nil || cost_manager == nil { goto Error }

  cost_model.literal = (*uint32)(cost_model + 1)
  if (use_color_cache) {
    cc_init = VP8LColorCacheInit(&hashers, cache_bits)
    if !cc_init { goto Error }
  }

  if (!CostModelBuild(cost_model, xsize, cache_bits, refs)) {
    goto Error
  }

  if (!CostManagerInit(cost_manager, dist_array, pix_count, cost_model)) {
    goto Error
  }

  // We loop one pixel at a time, but store all currently best points to
  // non-processed locations from this point.
  dist_array[0] = 0
  // Add first pixel as literal.
  AddSingleLiteralWithCostModel(argb, &hashers, cost_model, /*idx=*/0, use_color_cache, /*prev_cost=*/0, cost_manager.costs, dist_array)

  for i = 1; i < pix_count; i++ {
    prev_cost := cost_manager.costs[i - 1]
    int offset, len
    VP8LHashChainFindCopy(hash_chain, i, &offset, &len)

    // Try adding the pixel as a literal.
    AddSingleLiteralWithCostModel(argb, &hashers, cost_model, i, use_color_cache, prev_cost, cost_manager.costs, dist_array)

    // If we are dealing with a non-literal.
    if (len >= 2) {
      if (offset != offset_prev) {
        code := VP8LDistanceToPlaneCode(xsize, offset)
        offset_cost = GetDistanceCost(cost_model, code)
        first_offset_is_constant = 1
        PushInterval(cost_manager, prev_cost + offset_cost, i, len)
      } else {
        assert.Assert(offset_cost >= 0)
        assert.Assert(len_prev >= 0)
        assert.Assert(first_offset_is_constant == 0 || first_offset_is_constant == 1)
        // Instead of considering all contributions from a pixel i by calling:
        //         PushInterval(cost_manager, prev_cost + offset_cost, i, len)
        // we optimize these contributions in case offset_cost stays the same
        // for consecutive pixels. This describes a set of pixels similar to a
        // previous set (e.g. constant color regions).
        if (first_offset_is_constant) {
          reach = i - 1 + len_prev - 1
          first_offset_is_constant = 0
        }

        if (i + len - 1 > reach) {
          // We can only be go further with the same offset if the previous
          // length was maxed, hence len_prev == len == MAX_LENGTH.
          // TODO(vrabaud), bump i to the end right away (insert cache and
          // update cost).
          // TODO(vrabaud), check if one of the points in between does not have
          // a lower cost.
          // Already consider the pixel at "reach" to add intervals that are
          // better than whatever we add.
          int offset_j, len_j = 0
          var j int
          assert.Assert(len == MAX_LENGTH || len == pix_count - i)
          // Figure out the last consecutive pixel within [i, reach + 1] with
          // the same offset.
          for j = i; j <= reach; j++ {
            VP8LHashChainFindCopy(hash_chain, j + 1, &offset_j, &len_j)
            if (offset_j != offset) {
              VP8LHashChainFindCopy(hash_chain, j, &offset_j, &len_j)
              break
            }
          }
          // Update the cost at j - 1 and j.
          UpdateCostAtIndex(cost_manager, j - 1, 0)
          UpdateCostAtIndex(cost_manager, j, 0)

          PushInterval(cost_manager, cost_manager.costs[j - 1] + offset_cost, j, len_j)
          reach = j + len_j - 1
        }
      }
    }

    UpdateCostAtIndex(cost_manager, i, 1)
    offset_prev = offset
    len_prev = len
  }

  ok = !refs.error
Error:
  if cc_init { VP8LColorCacheClear(&hashers) }
  CostManagerClear(cost_manager)

  return ok
}

// We pack the path at the end of and return *dist_array
// a pointer to this part of the array. Example:
// dist_array = [1x2xx3x2] => packed [1x2x1232], chosen_path = [1232]
func TraceBackwards(/* const */ dist_array *uint16, dist_array_size int, *uint16* const chosen_path, /*const*/ chosen_path_size *int) {
  path *uint16 = dist_array + dist_array_size
  cur *uint16 = dist_array + dist_array_size - 1
  while (cur >= dist_array) {
    k := *cur
    --path
    *path = k
    cur -= k
  }
  *chosen_path = path
  *chosen_path_size = (int)(dist_array + dist_array_size - path)
}

static int BackwardReferencesHashChainFollowChosenPath(
    const argb *uint32, cache_bits int, /*const*/ chosen_path *uint16, chosen_path_size int, /*const*/ hash_chain *VP8LHashChain, /*const*/ refs *VP8LBackwardRefs) {
  use_color_cache := (cache_bits > 0)
  var ix int
  i := 0
  ok := 0
  cc_init := 0
   var hashers VP8LColorCache

  if (use_color_cache) {
    cc_init = VP8LColorCacheInit(&hashers, cache_bits)
    if !cc_init { goto Error }
  }

  VP8LClearBackwardRefs(refs)
  for ix = 0; ix < chosen_path_size; ix++ {
    len := chosen_path[ix]
    if (len != 1) {
      var k int
      offset := VP8LHashChainFindOffset(hash_chain, i)
      VP8LBackwardRefsCursorAdd(refs, PixOrCopyCreateCopy(offset, len))
      if (use_color_cache) {
        for k = 0; k < len; k++ {
          VP8LColorCacheInsert(&hashers, argb[i + k])
        }
      }
      i += len
    } else {
       var v PixOrCopy
      idx := use_color_cache ? VP8LColorCacheContains(&hashers, argb[i]) : -1
      if (idx >= 0) {
        // use_color_cache is true and hashers contains argb[i]
        // push pixel as a color cache index
        v = PixOrCopyCreateCacheIdx(idx)
      } else {
        if use_color_cache { VP8LColorCacheInsert(&hashers, argb[i]) }
        v = PixOrCopyCreateLiteral(argb[i])
      }
      VP8LBackwardRefsCursorAdd(refs, v)
      i++
    }
  }
  ok = !refs.error
Error:
  if cc_init { VP8LColorCacheClear(&hashers) }
  return ok
}

func VP8LBackwardReferencesTraceBackwards(xsize int, ysize int, /*const*/ argb *uint32, cache_bits int, /*const*/ hash_chain *VP8LHashChain, /*const*/ refs_src *VP8LBackwardRefs, /*const*/ refs_dst *VP8LBackwardRefs) int {
  ok := 0
  dist_array_size := xsize * ysize
  var chosen_path *uint16 = nil
  chosen_path_size := 0
//   var dist_array *uint16 = (*uint16)WebPSafeMalloc(dist_array_size, sizeof(*dist_array))
//   if dist_array == nil { goto Error }
  dist_array = make([]uint16, dist_array_size)

  if (!BackwardReferencesHashChainDistanceOnly(
          xsize, ysize, argb, cache_bits, hash_chain, refs_src, dist_array)) {
    goto Error
  }
  TraceBackwards(dist_array, dist_array_size, &chosen_path, &chosen_path_size)
  if (!BackwardReferencesHashChainFollowChosenPath(
          argb, cache_bits, chosen_path, chosen_path_size, hash_chain, refs_dst)) {
    goto Error
  }
  ok = 1
Error:
  return ok
}
