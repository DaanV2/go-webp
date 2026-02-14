// Copyright 2023 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package utils


const(
	// Sorts by minimizing L1 deltas between consecutive colors, giving more
	// weight to RGB colors.
	kSortedDefault PaletteSorting = 0
	// Implements the modified Zeng method from "A Survey on Palette Reordering
	// Methods for Improving the Compression of Color-Indexed Images" by Armando
	// J. Pinho and Antonio J. R. Neves.
	kMinimizeDelta PaletteSorting = 1

	kModifiedZeng PaletteSorting = 2
	kUnusedPalette PaletteSorting = 3
	kPaletteSortingNum PaletteSorting = 4

	COLOR_HASH_SIZE =(MAX_PALETTE_SIZE * 4)
	COLOR_HASH_RIGHT_SHIFT =22  // 32 - log2(COLOR_HASH_SIZE).
)

// The different ways a palette can be sorted.
type PaletteSorting int

type Sum struct {
   index uint8
   sum uint32
}


func PaletteComponentDistance(v uint32) uint32 {
  return tenary.If(v <= 128, v, 256 - v)
}

// Computes a value that is related to the entropy created by the
// palette entry diff.
//
// Note that the last & 0xff is a no-operation in the next statement, but
// removed by most compilers and is here only for regularity of the code.
func PaletteColorDistance(col1, col2 uint32 ) uint32 {
  diff := VP8LSubPixels(col1, col2)
  kMoreWeightForRGBThanForAlpha := 9
  var score uint32
  score = PaletteComponentDistance((diff >> 0) & 0xff)
  score += PaletteComponentDistance((diff >> 8) & 0xff)
  score += PaletteComponentDistance((diff >> 16) & 0xff)
  score *= kMoreWeightForRGBThanForAlpha
  score += PaletteComponentDistance((diff >> 24) & 0xff)
  return score
}
  
 // Palette reordering for smaller sum of deltas (and for smaller storage).
func PaletteCompareColorsForQsort(/* const */ p1 , /*const*/ p2 *uint8) int {
  a := WebPMemToUint32(p1)
  b := WebPMemToUint32(p2)
  assert.Assert(a != b)
  return tenary.If(a < b, -1, 1)
}

func SwapColor(/* const */ col1 *uint32, /*const*/ col2 *uint32) {
  tmp := *col1
  *col1 = *col2
  *col2 = tmp
}

// Returns the index of 'color' in the sorted palette 'sorted' of size 'num_colors'.
func SearchColorNoIdx(  sorted []uint32, color uint32, num_colors int) int {
  low := 0
  hi = num_colors
  if sorted[low] == color {
    return low  // loop invariant: sorted[low] != color
}
  for {
    mid := (low + hi) >> 1
    if (sorted[mid] == color) {
      return mid
    } else if (sorted[mid] < color) {
      low = mid
    } else {
      hi = mid
    }
  }

  assert.Assert(0)
  return 0
}

// Sort palette in increasing order and prepare an inverse mapping array.
func PrepareMapToPalette(palette []uint32,  num_colors uint32, sorted []uint32, idx_map []uint32) {
  var i uint32
  memcpy(sorted, palette, num_colors * sizeof(*sorted))
  qsort(sorted, num_colors, sizeof(*sorted), PaletteCompareColorsForQsort)
  for i = 0; i < num_colors; i++ {
    idx_map[SearchColorNoIdx(sorted, palette[i], num_colors)] = i
  }
}


// Returns count of unique colors in 'pic', assuming pic.use_argb is true.
// If the unique color count is more than MAX_PALETTE_SIZE, returns
// MAX_PALETTE_SIZE+1.
// If 'palette' is not nil and the number of unique colors is less than or
// equal to MAX_PALETTE_SIZE, also outputs the actual unique colors into
// 'palette' in a sorted order. Note: 'palette' is assumed to be an array
// already allocated with at least MAX_PALETTE_SIZE elements.
func GetColorPalette(/* const */ pic *picture.Picture, /*const*/   palette *uint32/* (MAX_PALETTE_SIZE) */ ) int {
  var i int
  var x, y int
  num_colors := 0
  var in_use [COLOR_HASH_SIZE]uint8
  var colors [COLOR_HASH_SIZE]uint32
  var argb *uint32 = pic.argb
  width := pic.width
  height := pic.height
  last_pix := ~argb[0];  // so we're sure that last_pix != argb[0]
  assert.Assert(pic != nil)
  assert.Assert(pic.use_argb)

  for y = 0; y < height; y++ {
    for x = 0; x < width; x++ {
      var key int 
      if (argb[x] == last_pix) {
        continue
      }
      last_pix = argb[x]
      key = VP8LHashPix(last_pix, COLOR_HASH_RIGHT_SHIFT)
      for {
        if (!in_use[key]) {
          colors[key] = last_pix
          in_use[key] = 1
          num_colors++
          if (num_colors > MAX_PALETTE_SIZE) {
            return MAX_PALETTE_SIZE + 1;  // Exact count not needed.
          }
          break
        } else if (colors[key] == last_pix) {
          break;  // The color is already there.
        } else {
          // Some other color sits here, so do linear conflict resolution.
          key++
          key &= (COLOR_HASH_SIZE - 1);  // Key mask.
        }
      }
    }
    argb += pic.argb_stride
  }

  if (palette != nil) {  // Fill the colors into palette.
    num_colors = 0
    for i = 0; i < COLOR_HASH_SIZE; i++ {
      if (in_use[i]) {
        palette[num_colors] = colors[i]
        num_colors++
      }
    }
    qsort(palette, num_colors, sizeof(*palette), PaletteCompareColorsForQsort)
  }
  return num_colors
}


// -----------------------------------------------------------------------------

// The palette has been sorted by alpha. This function checks if the other
// components of the palette have a monotonic development with regards to
// position in the palette. If all have monotonic development, there is
// no benefit to re-organize them greedily. A monotonic development
// would be spotted in green-only situations (like lossy alpha) or gray-scale
// images.
func PaletteHasNonMonotonousDeltas(palette *uint32 , num_colors int) int {
  predict := 0x000000
  var i int
  sign_found := 0x00
  for i = 0; i < num_colors; i++ {
    diff := VP8LSubPixels(palette[i], predict)
    rd := (diff >> 16) & 0xff
    gd := (diff >> 8) & 0xff
    bd := (diff >> 0) & 0xff
    if (rd != 0x00) {
      sign_found |= tenary.If(rd < 0x80, 1, 2)
    }
    if (gd != 0x00) {
      sign_found |= tenary.If(gd < 0x80, 8, 16)
    }
    if (bd != 0x00) {
      sign_found |= tenary.If(bd < 0x80, 64, 128)
    }
    predict = palette[i]
  }
  return (sign_found & (sign_found << 1)) != 0;  // two consequent signs.
}

func PaletteSortMinimizeDeltas(/* const */ palette_sorted *uint32 , num_colors int, /*const*/ palette *uint32) {
  predict := 0x00000000
  var i, k int
  memcpy(palette, palette_sorted, num_colors * sizeof(*palette))
  if !PaletteHasNonMonotonousDeltas(palette_sorted, num_colors) { {return }}
  // Find greedily always the closest color of the predicted color to minimize
  // deltas in the palette. This reduces storage needs since the
  // palette is stored with delta encoding.
  if (num_colors > 17) {
    if (palette[0] == 0) {
      num_colors--
      SwapColor(&palette[num_colors], &palette[0])
    }
  }
  for i = 0; i < num_colors; i++ {
    best_ix := i
    best_score := ~uint(0)
    for k = i; k < num_colors; k++ {
      cur_score := PaletteColorDistance(palette[k], predict)
      if (best_score > cur_score) {
        best_score = cur_score
        best_ix = k
      }
    }
    SwapColor(&palette[best_ix], &palette[i])
    predict = palette[i]
  }
}

// -----------------------------------------------------------------------------
// Modified Zeng method from "A Survey on Palette Reordering
// Methods for Improving the Compression of Color-Indexed Images" by Armando J.
// Pinho and Antonio J. R. Neves.

// Finds the biggest cooccurrence in the matrix.
func CoOccurrenceFindMax(cooccurrence []uint32/* (num_num_colors *colors) */ , num_colors uint32, /*const*/ c1 *uint8, /*const*/ c2 *uint8) {
  // Find the index that is most frequently located adjacent to other
  // (different) indexes.
  best_sum := uint(0)
  var i, j, best_cooccurrence uint32
  *c1 = uint(0)
  for i = 0; i < num_colors; i++ {
    sum := 0
    for (j = 0; j < num_colors; ++j) {sum += cooccurrence[i * num_colors + j];}
    if (sum > best_sum) {
      best_sum = sum
      *c1 = i
    }
  }
  // Find the index that is most frequently found adjacent to *c1.
  *c2 = uint(0)
  best_cooccurrence = uint(0)
  for i = 0; i < num_colors; i++ {
    if (cooccurrence[*c1 * num_colors + i] > best_cooccurrence) {
      best_cooccurrence = cooccurrence[*c1 * num_colors + i]
      *c2 = i
    }
  }
  assert.Assert(*c1 != *c2)
}

// Builds the cooccurrence matrix
func CoOccurrenceBuild(/* const */ pic *picture.Picture, /*const*/ palette *uint32 , num_colors uint32, cooccurrence []uint32/* (num_num_colors *colors) */) int {
  var lines, line_top, line_current, line_tmp *uint32
  var x, y int
  var src []uint32 = pic.argb
  prev_pix := ~src[0]
  prev_idx := uint(0)
  var idx_map [MAX_PALETTE_SIZE]uint32
  var palette_sorted [MAX_PALETTE_SIZE]uint32
//   lines = (*uint32)WebPSafeMalloc(2 * pic.width, sizeof(*lines))
//   if (lines == nil) {
//     return 0
//   }
  lines := make([]uint32, 2 * pic.width)

  line_top = &lines[0]
  line_current = &lines[pic.width]
  PrepareMapToPalette(palette, num_colors, palette_sorted, idx_map)
  for y = 0; y < pic.height; y++ {
    for x = 0; x < pic.width; x++ {
      pix := src[x]
      if (pix != prev_pix) {
        prev_idx = idx_map[SearchColorNoIdx(palette_sorted, pix, num_colors)]
        prev_pix = pix
      }
      line_current[x] = prev_idx
      // 4-connectivity is what works best as mentioned in "On the relation
      // between Memon's and the modified Zeng's palette reordering methods".
      if (x > 0 && prev_idx != line_current[x - 1]) {
        left_idx := line_current[x - 1]
        cooccurrence[prev_idx * num_colors + left_idx] = cooccurrence[prev_idx * num_colors + left_idx] + 1
        cooccurrence[left_idx * num_colors + prev_idx] = cooccurrence[left_idx * num_colors + prev_idx] + 1
      }
      if (y > 0 && prev_idx != line_top[x]) {
        top_idx := line_top[x]
        cooccurrence[prev_idx * num_colors + top_idx] = cooccurrence[prev_idx * num_colors + top_idx] + 1
        cooccurrence[top_idx * num_colors + prev_idx] = cooccurrence[top_idx * num_colors + prev_idx] + 1
      }
    }
    line_tmp = line_top
    line_top = line_current
    line_current = line_tmp
    src += pic.argb_stride
  }

  return 1
}



func PaletteSortModifiedZeng(/* const */ pic *picture.Picture, /*const*/ palette_in *uint32 , num_colors uint32, /*const*/ palette *uint32) int {
  var i, j, ind uint32
  var remapping [MAX_PALETTE_SIZE]uint8
  var  sums[MAX_PALETTE_SIZE]Sum
   var first, last uint32
  var num_sums uint32
  // TODO(vrabaud) check whether one color images should use palette or not.
  if num_colors <= 1 { return 1  }
  // Build the co-occurrence matrix.
//   cooccurrence = (*uint32)WebPSafeCalloc(num_colors * num_colors, sizeof(*cooccurrence))
//   if (cooccurrence == nil) {
//     return 0
//   }
  cooccurrence := make([]uint32, num_colors * num_colors)

  if (!CoOccurrenceBuild(pic, palette_in, num_colors, WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
                             *uint32, cooccurrence, num_num_colors *colors * sizeof(*cooccurrence)))) {
    return 0
  }

  // Initialize the mapping list with the two best indices.
  CoOccurrenceFindMax(WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
                          const *uint32, cooccurrence, num_num_colors *colors * sizeof(*cooccurrence)), num_colors, &remapping[0], &remapping[1])

  // We need to append and prepend to the list of remapping. To this end, we
  // actually define the next start/end of the list as indices in a vector (with
  // a wrap around when the end is reached).
  first = 0
  last = 1
  num_sums = num_colors - 2;  // -2 because we know the first two values
  if (num_sums > 0) {
    // Initialize the sums with the first two remappings and find the best one
    var best_sum *Sum = &sums[0]
    best_sum.index = uint(0)
    best_sum.sum = uint(0)
    for i = 0, j = 0; i < num_colors; i++ {
      if i == remapping[0] || i == remapping[1] { continue }
      sums[j].index = i
      sums[j].sum = cooccurrence[i * num_colors + remapping[0]] +
                    cooccurrence[i * num_colors + remapping[1]]
      if sums[j].sum > best_sum.sum { best_sum = &sums[j] }
      j++
    }

    for num_sums > 0 {
      best_index := best_sum.index
      // Compute delta to know if we need to prepend or append the best index.
      delta := 0
      n := num_colors - num_sums
      for ind = first, j = 0; (ind + j) % num_colors != last + 1; j++ {
        l_j := remapping[(ind + j) % num_colors]
        delta += (n - 1 - 2 * (int32)j) *
                 (int32)cooccurrence[best_index * num_colors + l_j]
      }
      if (delta > 0) {
        first = tenary.If(first == 0, num_colors - 1, first - 1)
        remapping[first] = best_index
      } else {
        last++
        remapping[last] = best_index
      }
      // Remove best_sum from sums.
      *best_sum = sums[num_sums - 1]
      num_sums--
      // Update all the sums and find the best one.
      best_sum = &sums[0]
      for i = 0; i < num_sums; i++ {
        sums[i].sum += cooccurrence[best_index * num_colors + sums[i].index]
        if sums[i].sum > best_sum.sum { best_sum = &sums[i] }
      }
    }
  }
  assert.Assert((last + 1) % num_colors == first)

  // Re-map the palette.
  for i = 0; i < num_colors; i++ {
    palette[i] = palette_in[remapping[(first + i) % num_colors]]
  }
  return 1
}

// Sorts the palette according to the criterion defined by 'method'.
// 'palette_sorted' is the input palette sorted lexicographically, as done in
// PrepareMapToPalette. Returns 0 on memory allocation error.
// For kSortedDefault and kMinimizeDelta methods, 0 (if present) is set as the
// last element to optimize later storage.
func PaletteSort(method PaletteSorting, /*const*/ pic *picture.Picture, /*const*/ palette_sorted *uint32 , num_colors uint32, /*const*/ palette *uint32  ) int {
  switch (method) {
    case kSortedDefault:
      if (palette_sorted[0] == 0 && num_colors > 17) {
        memcpy(palette, palette_sorted + 1, (num_colors - 1) * sizeof(*palette_sorted))
        palette[num_colors - 1] = 0
      } else {
        memcpy(palette, palette_sorted, num_colors * sizeof(*palette))
      }
      return 1
    case kMinimizeDelta:
      PaletteSortMinimizeDeltas(palette_sorted, num_colors, palette)
      return 1
    case kModifiedZeng:
      return PaletteSortModifiedZeng(pic, palette_sorted, num_colors, palette)
    case kUnusedPalette:
    case kPaletteSortingNum:
      break
  }

  assert.Assert(0)
  return 0
}
