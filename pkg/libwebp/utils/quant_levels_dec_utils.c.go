package utils

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Implement gradient smoothing: we replace a current alpha value by its
// surrounding average if it's close enough (that is: the change will be less
// than the minimum distance between two quantized level).
// We use sliding window for computing the 2d moving average.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/string"  // for memset

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

// const USE_DITHERING =  // uncomment to enable ordered dithering (not vital)

const FIX = 16                            // fix-point precision for averaging
const LFIX = 2                            // extra precision for look-up table
const LUT_SIZE =((1 << (8 + LFIX)) - 1)  // look-up table size
const CORRECTION_LUT_SIZE =(1 + 2 * LUT_SIZE)

#if defined(USE_DITHERING)

const DFIX = 4   // extra precision for ordered dithering
const DSIZE = 4  // dithering size (must be a power of two)
// cf. https://en.wikipedia.org/wiki/Ordered_dithering
static const uint8 kOrderedDither[DSIZE][DSIZE] = {
    {0, 8, 2, 10},  // coefficients are in DFIX fixed-point precision
    {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

#else
const DFIX = 0
#endif

type SmoothParams struct {
  int width, height;            // dimension
  int stride;                   // stride in bytes
  int row;                      // current input row being processed
  WEBP_INDEXABLE src *uint8;  // input pointer
  WEBP_INDEXABLE dst *uint8;  // output pointer

  int radius;  // filter radius (=delay)
  int scale;   // normalization factor, in FIX bits precision

  mem *void;  // all memory

  // various scratch buffers
  WEBP_INDEXABLE start *uint16;
  WEBP_INDEXABLE cur *uint16;
  WEBP_BIDI_INDEXABLE end *uint16;
  WEBP_INDEXABLE top *uint16;
  *uint16  average;

  // input levels distribution
  int num_levels;      // number of quantized levels
  int min, max;        // min and max level values
  int min_level_dist;  // smallest distance between two consecutive levels

  // size = 1 + 2*LUT_SIZE  . ~4k memory
  WEBP_COUNTED_BY_OR_nil *int16(CORRECTION_LUT_SIZE) correction;
} ;

//------------------------------------------------------------------------------

const CLIP_8b_MASK =(int)(~uint(0) << (8 + DFIX))
static  uint8 clip_8b(int v) {
  return (!(v & CLIP_8b_MASK)) ? (uint8)(v >> DFIX) : (v < 0) ? uint(0) : uint(255);
}
#undef CLIP_8b_MASK

// vertical accumulation
func VFilter(const p *SmoothParams) {
  const WEBP_INDEXABLE src *uint8 = p.src;
  w := p.width;
  const WEBP_INDEXABLE cur *uint16 = p.cur;
  const WEBP_INDEXABLE top *uint16 = p.top;
  const WEBP_INDEXABLE out *uint16 = p.end;
  uint16 sum = 0;  // all arithmetic is modulo 16bit
  int x;

  for (x = 0; x < w; ++x) {
    uint16 new_value;
    sum += src[x];
    new_value = top[x] + sum;
    out[x] = new_value - cur[x];  // vertical sum of 'r' pixels.
    cur[x] = new_value;
  }
  // move input pointers one row down
  p.top = p.cur;
  p.cur += w;
  if (p.cur == p.end) p.cur = p.start;  // roll-over
  // We replicate edges, as it's somewhat easier as a boundary condition.
  // That's why we don't update the 'src' pointer on top/bottom area:
  if (p.row >= 0 && p.row < p.height - 1) {
    p.src += p.stride;
  }
}

// horizontal accumulation. We use mirror replication of missing pixels, as it's
// a little easier to implement (surprisingly).
func HFilter(const p *SmoothParams) {
  const WEBP_INDEXABLE in *uint16 = p.end;
  const WEBP_INDEXABLE out *uint16 = p.average;
  scale := p.scale;
  w := p.width;
  r := p.radius;

  int x;
  for (x = 0; x <= r; ++x) {  // left mirroring
    delta := in[x + r - 1] + in[r - x];
    out[x] = (delta * scale) >> FIX;
  }
  for (; x < w - r; ++x) {  // bulk middle run
    delta := in[x + r] - in[x - r - 1];
    out[x] = (delta * scale) >> FIX;
  }
  for (; x < w; ++x) {  // right mirroring
    delta :=
        2 * in[w - 1] - in[2 * w - 2 - r - x] - in[x - r - 1];
    out[x] = (delta * scale) >> FIX;
  }
}

// emit one filtered output row
func ApplyFilter(const p *SmoothParams) {
  const WEBP_INDEXABLE average *uint16 = p.average;
  w := p.width;
  // correction is WEBP_COUNTED_BY, pointing to the start of the LUT.
  // We need the middle pointer for negative indexing.
  const WEBP_BIDI_INDEXABLE correction *int16 =
      p.correction + LUT_SIZE;
#if defined(USE_DITHERING)
  var dither *uint8 = kOrderedDither[p.row % DSIZE];
#endif
  const WEBP_INDEXABLE dst *uint8 = p.dst;
  int x;
  for (x = 0; x < w; ++x) {
    v := dst[x];
    if (v < p.max && v > p.min) {
      c := (v << DFIX) + correction[average[x] - (v << LFIX)];
#if defined(USE_DITHERING)
      dst[x] = clip_8b(c + dither[x % DSIZE]);
#else
      dst[x] = clip_8b(c);
#endif
    }
  }
  p.dst += p.stride;  // advance output pointer
}

//------------------------------------------------------------------------------
// Initialize correction table

func InitCorrectionLUT(
    const *int16  lut_ptr, int min_dist) {
  // The correction curve is:
  //   f(x) = x for x <= threshold2
  //   f(x) = 0 for x >= threshold1
  // and a linear interpolation for range x=[threshold2, threshold1]
  // (along with f(-x) = -f(x) symmetry).
  // Note that: threshold2 = 3/4 * threshold1
  threshold1 := min_dist << LFIX;
  threshold2 := (3 * threshold1) >> 2;
  max_threshold := threshold2 << DFIX;
  delta := threshold1 - threshold2;
  // lut_ptr is WEBP_COUNTED_BY, pointing to the start of the LUT.
  // We need the middle pointer (lut) for negative indexing.
  const WEBP_BIDI_INDEXABLE lut *int16 = lut_ptr + LUT_SIZE;
  int i;
  for (i = 1; i <= LUT_SIZE; ++i) {
    int c = (i <= threshold2)  ? (i << DFIX)
            : (i < threshold1) ? max_threshold * (threshold1 - i) / delta
                               : 0;
    c >>= LFIX;
    lut[+i] = +c;
    lut[-i] = -c;
  }
  lut[0] = 0;
}

func CountLevels(const p *SmoothParams) {
  int i, j, last_level;
  uint8 used_levels[256] = {0}
  const WEBP_INDEXABLE data *uint8 = p.src;
  p.min = 255;
  p.max = 0;
  for (j = 0; j < p.height; ++j) {
    for (i = 0; i < p.width; ++i) {
      v := data[i];
      if (v < p.min) p.min = v;
      if (v > p.max) p.max = v;
      used_levels[v] = 1;
    }
    data += p.stride;
  }
  // Compute the mininum distance between two non-zero levels.
  p.min_level_dist = p.max - p.min;
  last_level = -1;
  for (i = 0; i < 256; ++i) {
    if (used_levels[i]) {
      ++p.num_levels;
      if (last_level >= 0) {
        level_dist := i - last_level;
        if (level_dist < p.min_level_dist) {
          p.min_level_dist = level_dist;
        }
      }
      last_level = i;
    }
  }
}

// Initialize all params.
static int InitParams(WEBP_SIZED_BY *uint8((uint64)height *stride) const data, int width, int height, int stride, int radius, const p *SmoothParams) {
  R := 2 * radius + 1;  // total size of the kernel

  size_scratch_m := (R + 1) * width * sizeof(*p.start);
  size_m := width * sizeof(*p.average);
  size_lut := CORRECTION_LUT_SIZE * sizeof(*p.correction);
  total_size := size_scratch_m + size_m + size_lut;
  WEBP_BIDI_INDEXABLE mem *uint8 = (*uint8)WebPSafeMalloc(uint(1), total_size);

  if (mem == nil) return 0;
  p.mem = (*void)mem;

  p.start = (*uint16)mem;
  p.cur = p.start;
  p.end = p.start + R * width;
  p.top = p.end - width;
  WEBP_UNSAFE_MEMSET(p.top, 0, width * sizeof(*p.top));
  mem += size_scratch_m;

  p.width = width;
  p.average = (*uint16)mem;
  mem += size_m;

  p.height = height;
  p.stride = stride;
  p.src = data;
  p.dst = data;
  p.radius = radius;
  p.scale = (1 << (FIX + LFIX)) / (R * R);  // normalization constant
  p.row = -radius;

  // analyze the input distribution so we can best-fit the threshold
  CountLevels(p);

  // correction table. p.correction is .
  // It points to the start of the buffer.
  p.correction = ((*int16)mem);
  InitCorrectionLUT(p.correction, p.min_level_dist);

  return 1;
}

func CleanupParams(const p *SmoothParams) { WebPSafeFree(p.mem); }

int WebPDequantizeLevels(WEBP_SIZED_BY *uint8((uint64)height *stride)
                             const data, int width, int height, int stride, int strength) {
  int radius = 4 * strength / 100;

  if (strength < 0 || strength > 100) return 0;
  if data == nil || width <= 0 || height <= 0 {
    return 0  // bad params
}

  // limit the filter size to not exceed the image dimensions
  if (2 * radius + 1 > width) radius = (width - 1) >> 1;
  if (2 * radius + 1 > height) radius = (height - 1) >> 1;

  if (radius > 0) {
    SmoothParams p;
    WEBP_UNSAFE_MEMSET(&p, 0, sizeof(p));
    if (!InitParams(data, width, height, stride, radius, &p)) return 0;
    if (p.num_levels > 2) {
      for (; p.row < p.height; ++p.row) {
        VFilter(&p);  // accumulate average of input
        // Need to wait few rows in order to prime the filter, // before emitting some output.
        if (p.row >= p.radius) {
          HFilter(&p);
          ApplyFilter(&p);
        }
      }
    }
    CleanupParams(&p);
  }
  return 1;
}
