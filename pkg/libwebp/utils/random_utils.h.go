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
// Pseudo-random utilities
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"


const VP8_RANDOM_DITHER_FIX = 8  // fixed-point precision for dithering
const VP8_RANDOM_TABLE_SIZE = 55

type VP8Random struct {
  index1, index2 int
  tab [VP8_RANDOM_TABLE_SIZE]uint32
  amp int
}


// Returns a centered pseudo-random number with 'num_bits' amplitude.
// (uses D.Knuth's Difference-based random generator).
// 'amp' is in VP8_RANDOM_DITHER_FIX fixed-point precision.
func VP8RandomBits2(/* const  */rg *VP8Random, num_bits, amp int) int {
  var diff int
  assert.Assert(num_bits + VP8_RANDOM_DITHER_FIX <= 31)

  diff = rg.tab[rg.index1] - rg.tab[rg.index2]
  if (diff < 0) {diff += (uint(1) << 31)}
  rg.tab[rg.index1] = diff
  if (rg.index1 == VP8_RANDOM_TABLE_SIZE) {rg.index1 = 0} else {
	rg.index1++
  }
  if (rg.index2 == VP8_RANDOM_TABLE_SIZE) {rg.index2 = 0} else {
	rg.index2++
  }
  
 
  // sign-extend, 0-center
  diff = (diff << 1) >> (32 - num_bits)
  diff = (diff * amp) >> VP8_RANDOM_DITHER_FIX  // restrict range
  diff += 1 << (num_bits - 1)                   // shift back to 0.5-center
  return diff
}

func VP8RandomBits(/* const  */rg *VP8Random, num_bits int) int {
  return VP8RandomBits2(rg, num_bits, rg.amp)
}

