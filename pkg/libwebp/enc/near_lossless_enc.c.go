package enc

// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Near-lossless image preprocessing adjusts pixel values to help
// compressibility with a guarantee of maximum deviation between original and
// resulting pixel values.
//
// Author: Jyrki Alakuijala (jyrki@google.com)
// Converted to C by Aleksander Kramarz (akramarz@google.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if (WEBP_NEAR_LOSSLESS == 1)

const MIN_DIM_FOR_NEAR_LOSSLESS =64
const MAX_LIMIT_BITS =5

// Quantizes the value up or down to a multiple of 1<<bits (or to 255),
// choosing the closer one, resolving ties using bankers' rounding.
func FindClosestDiscretized(uint32 a, bits int) uint32 {
  mask := (uint(1) << bits) - 1
  biased := a + (mask >> 1) + ((a >> bits) & 1)
  assert.Assert(bits > 0)
  if biased > 0xff { return 0xff  }
  return biased & ~mask
}

// Applies FindClosestDiscretized to all channels of pixel.
func ClosestDiscretizedArgb(uint32 a, bits int) uint32 {
  return (FindClosestDiscretized(a >> 24, bits) << 24) |
         (FindClosestDiscretized((a >> 16) & 0xff, bits) << 16) |
         (FindClosestDiscretized((a >> 8) & 0xff, bits) << 8) |
         (FindClosestDiscretized(a & 0xff, bits))
}

// Checks if distance between corresponding channel values of pixels a and b
// is within the given limit.
func IsNear(uint32 a, uint32 b, limit int) int {
  var k int
  for k = 0; k < 4; k++ {
    delta := (int)((a >> (k * 8)) & 0xff) - (int)((b >> (k * 8)) & 0xff)
    if (delta >= limit || delta <= -limit) {
      return 0
    }
  }
  return 1
}

func IsSmooth(/* const */ prev_row *uint32, /*const*/ curr_row *uint32, /*const*/ next_row *uint32, ix int, limit int) int {
  // Check that all pixels in 4-connected neighborhood are smooth.
  return (IsNear(curr_row[ix], curr_row[ix - 1], limit) &&
          IsNear(curr_row[ix], curr_row[ix + 1], limit) &&
          IsNear(curr_row[ix], prev_row[ix], limit) &&
          IsNear(curr_row[ix], next_row[ix], limit))
}

// Adjusts pixel values of image with given maximum error.
func NearLossless(xsize int, ysize int, /*const*/ argb_src []uint32, stride int, limit_bits int, copy_buffer *uint32, argb_dst *uint32) {
  var x, y int
  limit := 1 << limit_bits
  prev_row *uint32 = copy_buffer
  curr_row *uint32 = prev_row + xsize
  next_row *uint32 = curr_row + xsize
  stdlib.MemCpy(curr_row, argb_src, xsize * sizeof(argb_src[0]))
  stdlib.MemCpy(next_row, argb_src + stride, xsize * sizeof(argb_src[0]))

  for y = 0; y < ysize; ++y, argb_src += stride, argb_dst += xsize {
    if (y == 0 || y == ysize - 1) {
      stdlib.MemCpy(argb_dst, argb_src, xsize * sizeof(argb_src[0]))
    } else {
      stdlib.MemCpy(next_row, argb_src + stride, xsize * sizeof(argb_src[0]))
      argb_dst[0] = argb_src[0]
      argb_dst[xsize - 1] = argb_src[xsize - 1]
      for x = 1; x < xsize - 1; x++ {
        if (IsSmooth(prev_row, curr_row, next_row, x, limit)) {
          argb_dst[x] = curr_row[x]
        } else {
          argb_dst[x] = ClosestDiscretizedArgb(curr_row[x], limit_bits)
        }
      }
    }
    {
      // Three-way swap.
      var temp *uint32 = prev_row
      prev_row = curr_row
      curr_row = next_row
      next_row = temp
    }
  }
}

// in near_lossless.c
// Near lossless preprocessing in RGB color-space.
func VP8ApplyNearLossless(/* const */ picture *picture.Picture, quality int, /*const*/ argb_dst *uint32) int {
  var i int
  copy_buffer *uint32
  xsize := picture.Width
  ysize := picture.Height
  stride := picture.ARGBStride
  limit_bits := VP8LNearLosslessBits(quality)
  assert.Assert(argb_dst != nil)
  assert.Assert(limit_bits > 0)
  assert.Assert(limit_bits <= MAX_LIMIT_BITS)

  // For small icon images, don't attempt to apply near-lossless compression.
  if ((xsize < MIN_DIM_FOR_NEAR_LOSSLESS &&
       ysize < MIN_DIM_FOR_NEAR_LOSSLESS) ||
      ysize < 3) {
    for i = 0; i < ysize; i++ {
      stdlib.MemCpy(argb_dst + i * xsize, picture.ARGB + i * picture.ARGBStride, xsize * sizeof(*argb_dst))
    }
    return 1
  }

//   copy_buffer = (*uint32)WebPSafeMalloc(xsize * 3, sizeof(*copy_buffer))
//   if (copy_buffer == nil) {
//     return 0
//   }
  copy_buffer := make([]uint32, xsize * 3)

  NearLossless(xsize, ysize, picture.ARGB, stride, limit_bits, copy_buffer, argb_dst)
  for i = limit_bits - 1; i != 0; --i {
    NearLossless(xsize, ysize, argb_dst, xsize, i, copy_buffer, argb_dst)
  }
  
  return 1
}
#else  // (WEBP_NEAR_LOSSLESS == 1)

// Define a stub to suppress compiler warnings.
extern func VP8LNearLosslessStub(void)
func VP8LNearLosslessStub(){}

#endif  // (WEBP_NEAR_LOSSLESS == 1)
