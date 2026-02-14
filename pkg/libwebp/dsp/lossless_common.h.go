// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/dsp"
	"github.com/daanv2/go-webp/pkg/libwebp/utils"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/stddef"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

type VP8LPrefixCode struct {
  code int8
  extra_bits int8
} 


type VP8LFastLog2SlowFunc = func (v uint32) uint32
type VP8LFastSLog2SlowFunc = func (v uint32 ) uint64

// color mapping related functions.
func VP8GetARGBIndex(idx uint32) uint32 {
  return (idx >> 8) & 0xff
}

func VP8GetAlphaIndex(idx uint8) uint8 { return idx; }

func VP8GetARGBValue(val uint32) uint32 { return val; }

func VP8GetAlphaValue(val uint32) uint8 {
  return (val >> 8) & 0xff
}

// Computes sampled size of 'size' when sampling using 'sampling bits'.
func VP8LSubSampleSize(size uint32, uint32 sampling_bits) uint32 {
  return (size + (1 << sampling_bits) - 1) >> sampling_bits
}

// Converts near lossless quality into max number of bits shaved off.
func VP8LNearLosslessBits(distance intnear_lossless_quality) distance int{
  //    100 . 0
  // 80..99 . 1
  // 60..79 . 2
  // 40..59 . 3
  // 20..39 . 4
  //  0..19 . 5
  return 5 - near_lossless_quality / 20
}

func VP8LFastLog2(v uint32) uint32 {
  return tenary.If(v < LOG_LOOKUP_IDX_MAX, kLog2Table[v], VP8LFastLog2Slow(v))
}
// Fast calculation of v * log2(v) for integer input.
func VP8LFastSLog2(v uint32) uint64 {
  return tenary.If(v < LOG_LOOKUP_IDX_MAX, kSLog2Table[v], VP8LFastSLog2Slow(v))
}

func RightShiftRound(v uint64, uint32 shift) uint64 {
  return (v + (uint64(1) << shift >> 1)) >> shift
}

func DivRound(int64 a, int64 b) int64 {
  return tenary.If((a < 0) == (b < 0), (a + b / 2) / b, (a - b / 2) / b)
}



// -----------------------------------------------------------------------------
// PrefixEncode()

// Splitting of distance and length codes into prefixes and
// extra bits. The prefixes are encoded with an entropy code
// while the extra bits are stored just as normal bits.
func VP8LPrefixEncodeBitsNoLUT(distance intdistance, /*const*/ code *int, /*const*/ extra_bits *int) {
  highest_bit := BitsLog2Floor(--distance)
  second_highest_bit := (distance >> (highest_bit - 1)) & 1
  *extra_bits = highest_bit - 1
  *code = 2 * highest_bit + second_highest_bit
}

func VP8LPrefixEncodeNoLUT(distance intdistance, /*const*/ code *int, /*const*/ extra_bits *int, /*const*/ extra_bits_value *int) {
  highest_bit := BitsLog2Floor(--distance)
  second_highest_bit := (distance >> (highest_bit - 1)) & 1
  *extra_bits = highest_bit - 1
  *extra_bits_value = distance & ((1 << *extra_bits) - 1)
  *code = 2 * highest_bit + second_highest_bit
}

func VP8LPrefixEncodeBits(distance intdistance, /*const*/ code *int, /*const*/ extra_bits *int) {
  if (distance < PREFIX_LOOKUP_IDX_MAX) {
    var prefix_code VP8LPrefixCode = kPrefixEncodeCode[distance]
    *code = prefix_code.code
    *extra_bits = prefix_code.extra_bits
  } else {
    VP8LPrefixEncodeBitsNoLUT(distance, code, extra_bits)
  }
}

func VP8LPrefixEncode(distance int, /*const*/ code *int, /*const*/ extra_bits *int, /*const*/ extra_bits_value *int) {
  if (distance < PREFIX_LOOKUP_IDX_MAX) {
    var prefix_code VP8LPrefixCode = kPrefixEncodeCode[distance]
    *code = prefix_code.code
    *extra_bits = prefix_code.extra_bits
    *extra_bits_value = kPrefixEncodeExtraBitsValue[distance]
  } else {
    VP8LPrefixEncodeNoLUT(distance, code, extra_bits, extra_bits_value)
  }
}

// Sum of each component, mod 256.
func VP8LAddPixels(a uint32, b uint32) uint32 {
  alpha_and_green := (a & uint(0xff00ff00)) + (b & uint(0xff00ff00))
  red_and_blue := (a & uint(0x00ff00ff)) + (b & uint(0x00ff00ff))
  return (alpha_and_green & uint(0xff00ff00)) | (red_and_blue & uint(0x00ff00ff))
}

// Difference of each component, mod 256.
func VP8LSubPixels(a uint32, b uint32) uint32 {
  alpha_and_green := uint(0x00ff00ff) + (a & uint(0xff00ff00)) - (b & uint(0xff00ff00))
  red_and_blue := uint(0xff00ff00) + (a & uint(0x00ff00ff)) - (b & uint(0x00ff00ff))
  return (alpha_and_green & uint(0xff00ff00)) | (red_and_blue & uint(0x00ff00ff))
}


