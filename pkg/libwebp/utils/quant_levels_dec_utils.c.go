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


// const USE_DITHERING =  // uncomment to enable ordered dithering (not vital)

const FIX = 16                            // fix-point precision for averaging
const LFIX = 2                            // extra precision for look-up table
const LUT_SIZE =((1 << (8 + LFIX)) - 1)  // look-up table size
const CORRECTION_LUT_SIZE =(1 + 2 * LUT_SIZE)


const DFIX = 4   // extra precision for ordered dithering
const DSIZE = 4  // dithering size (must be a power of two)
// cf. https://en.wikipedia.org/wiki/Ordered_dithering
var  kOrderedDither = [DSIZE][DSIZE]uint8{
    {0, 8, 2, 10},  // coefficients are in DFIX fixed-point precision
    {12, 4, 14, 6}, 
	{3, 11, 1, 9}, 
	{15, 7, 13, 5},
}


type SmoothParams struct {
  width, height int ;            // dimension
  stride int;                   // stride in bytes
  row int;                      // current input row being processed
  src *uint8;  // input pointer
  dst []uint8;  // output pointer

   radius int;  // filter radius (=delay)
   scale int;   // normalization factor, in FIX bits precision

  mem *void;  // all memory

  // various scratch buffers
  start *uint16
  cur *uint16
  end *uint16
  top *uint16
    average *uint16

  // input levels distribution
   num_levels int      // number of quantized levels
   min, max int        // min and max level values
   min_level_dist int  // smallest distance between two consecutive levels

  // size = 1 + 2*LUT_SIZE  . ~4k memory
   correction []int16 /* (CORRECTION_LUT_SIZE) */
} 

//------------------------------------------------------------------------------

const CLIP_8b_MASK = (int)(~uint(0) << (8 + DFIX))
func clip_8b(v int) uint8 {
  return tenary.If(!(v & CLIP_8b_MASK), (uint8)(v >> DFIX), tenary.If(v < 0, uint(0), uint(255)))
}

// vertical accumulation
func VFilter(/* const */ p *SmoothParams) {
  var src *uint8 = p.src
  w := p.width
  var cur *uint16 = p.cur
  var top *uint16 = p.top
  var out *uint16 = p.end
  sum := 0;  // all arithmetic is modulo 16bit
  var x int

  for x = 0; x < w; x++ {
    var new_value uint16 
    sum += src[x]
    new_value = top[x] + sum
    out[x] = new_value - cur[x];  // vertical sum of 'r' pixels.
    cur[x] = new_value
  }
  // move input pointers one row down
  p.top = p.cur
  p.cur += w
  if p.cur == p.end { {p.cur = p.start }}  // roll-over
  // We replicate edges, as it's somewhat easier as a boundary condition.
  // That's why we don't update the 'src' pointer on top/bottom area:
  if (p.row >= 0 && p.row < p.height - 1) {
    p.src += p.stride
  }
}

// horizontal accumulation. We use mirror replication of missing pixels, as it's
// a little easier to implement (surprisingly).
func HFilter(/* const */ p *SmoothParams) {
  var in *uint16 = p.end
  var out *uint16 = p.average
  scale := p.scale
  w := p.width
  r := p.radius

  var x int
  for x = 0; x <= r; x++ {  // left mirroring
    delta := in[x + r - 1] + in[r - x]
    out[x] = (delta * scale) >> FIX
  }
  for ; x < w - r; x++ {  // bulk middle run
    delta := in[x + r] - in[x - r - 1]
    out[x] = (delta * scale) >> FIX
  }
  for ; x < w; x++ {  // right mirroring
    delta := 2 * in[w - 1] - in[2 * w - 2 - r - x] - in[x - r - 1]
    out[x] = (delta * scale) >> FIX
  }
}

// emit one filtered output row
func ApplyFilter(/* const */ p *SmoothParams) {
  var average *uint16 = p.average
  w := p.width
  // correction is , pointing to the start of the LUT.
  // We need the middle pointer for negative indexing.
  var correction *int16 = p.correction + LUT_SIZE
  var dither *uint8 = kOrderedDither[p.row % DSIZE]
  var dst []uint8 = p.dst
  var x int
  for x = 0; x < w; x++ {
    v := dst[x]
    if (v < p.max && v > p.min) {
      c := (v << DFIX) + correction[average[x] - (v << LFIX)]
      dst[x] = clip_8b(c + dither[x % DSIZE])
    }
  }
  p.dst += p.stride;  // advance output pointer
}

//------------------------------------------------------------------------------
// Initialize correction table

func InitCorrectionLUT(/* const */ lut_ptr *int16 , min_dist int) {
  // The correction curve is:
  //   f(x) = x for x <= threshold2
  //   f(x) = 0 for x >= threshold1
  // and a linear interpolation for range x=[threshold2, threshold1]
  // (along with f(-x) = -f(x) symmetry).
  // Note that: threshold2 = 3/4 * threshold1
  threshold1 := min_dist << LFIX
  threshold2 := (3 * threshold1) >> 2
  max_threshold := threshold2 << DFIX
  delta := threshold1 - threshold2
  // lut_ptr is , pointing to the start of the LUT.
  // We need the middle pointer (lut) for negative indexing.
  var lut *int16 = lut_ptr + LUT_SIZE
  var i int
  for i = 1; i <= LUT_SIZE; i++ {
    c := tenary.If(i <= threshold2, (i << DFIX), tenary.If(i < threshold1, max_threshold * (threshold1 - i) / delta, 0))
    c >>= LFIX
    lut[+i] = +c
    lut[-i] = -c
  }
  lut[0] = 0
}

func CountLevels(/* const */ p *SmoothParams) {
  var i, j, last_level int
  used_levels = [256]uint8{}
  var data *uint8 = p.src
  p.min = 255
  p.max = 0
  for j = 0; j < p.height; j++ {
    for i = 0; i < p.width; i++ {
      v := data[i]
      if v < p.min { {p.min = v }}
      if v > p.max { {p.max = v }}
      used_levels[v] = 1
    }
    data += p.stride
  }
  // Compute the mininum distance between two non-zero levels.
  p.min_level_dist = p.max - p.min
  last_level = -1
  for i = 0; i < 256; i++ {
    if (used_levels[i]) {
      p.num_levels++
      if (last_level >= 0) {
        level_dist := i - last_level
        if (level_dist < p.min_level_dist) {
          p.min_level_dist = level_dist
        }
      }
      last_level = i
    }
  }
}

// Initialize all params.
func InitParams(data []uint8/* (uint64(height)*stride) */, width, height, stride, radius int, p *SmoothParams) int {
  R := 2 * radius + 1;  // total size of the kernel

  size_scratch_m := (R + 1) * width * sizeof(*p.start)
  size_m := width * sizeof(*p.average)
  size_lut := CORRECTION_LUT_SIZE * sizeof(*p.correction)
  total_size := size_scratch_m + size_m + size_lut

//   mem *uint8 = (*uint8)WebPSafeMalloc(uint(1), total_size)
//   if mem == nil { return 0  }
  mem := make([]byte, total_size)

  p.mem = (*void)mem

  p.start = (*uint16)mem
  p.cur = p.start
  p.end = p.start + R * width
  p.top = p.end - width
  stdlib.Memset(p.top, 0, width * sizeof(*p.top))
  mem += size_scratch_m

  p.width = width
  p.average = (*uint16)mem
  mem += size_m

  p.height = height
  p.stride = stride
  p.src = data
  p.dst = data
  p.radius = radius
  p.scale = (1 << (FIX + LFIX)) / (R * R);  // normalization constant
  p.row = -radius

  // analyze the input distribution so we can best-fit the threshold
  CountLevels(p)

  // correction table. p.correction is .
  // It points to the start of the buffer.
  p.correction = ((*int16)mem)
  InitCorrectionLUT(p.correction, p.min_level_dist)

  return 1
}

func WebPDequantizeLevels(data []uint8/* (uint64(height)*stride) */ , width, height, stride , strength int ) int {
  radius := 4 * strength / 100

  if strength < 0 || strength > 100 { return 0  }
  if data == nil || width <= 0 || height <= 0 {
    return 0  // bad params
}

  // limit the filter size to not exceed the image dimensions
  if 2 * radius + 1 > width){ radius = (width - 1 { >> 1 }}
  if 2 * radius + 1 > height) {radius = (height - 1 { >> 1 }}

  if (radius > 0) {
    var p SmoothParams 
    stdlib.Memset(&p, 0, sizeof(p))
    if !InitParams(data, width, height, stride, radius, &p) { return 0  }
    if (p.num_levels > 2) {
      for ; p.row < p.height; ++p.row {
        VFilter(&p);  // accumulate average of input
        // Need to wait few rows in order to prime the filter, // before emitting some output.
        if (p.row >= p.radius) {
          HFilter(&p)
          ApplyFilter(&p)
        }
      }
    }
  }
  return 1
}
