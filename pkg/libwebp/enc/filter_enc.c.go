package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Selecting filter level
//
// Author: somnath@google.com (Somnath Banerjee)

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/util/tenary"
	"github.com/daanv2/go-webp/pkg/vp8"
)

// This table gives, for a given sharpness, the filtering strength to be
// used (at least) in order to filter a given edge step delta.
// This is constructed by brute force inspection: for all delta, we iterate
// over all possible filtering strength / thresh until needs_filter() returns
// true.
const MAX_DELTA_SIZE =64

var kLevelsFromDelta = [8][MAX_DELTA_SIZE]uint8{
    {0,  1,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 14, 16, 17, 19, 20, 22, 23, 25, 26, 28, 29, 31, 32, 34, 35, 37, 38, 40, 41, 43, 44, 46, 47, 49, 50, 52, 53, 55, 56, 58, 59, 61, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 15, 16, 18, 19, 21, 22, 24, 25, 27, 28, 30, 31, 33, 34, 36, 37, 39, 40, 42, 43, 45, 46, 48, 49, 51, 52, 54, 55, 57, 58, 60, 61, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 13, 15, 16, 17, 19, 20, 22, 23, 25, 26, 28, 29, 31, 32, 34, 35, 37, 38, 40, 41, 43, 44, 46, 47, 49, 50, 52, 53, 55, 56, 58, 59, 61, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 13, 15, 16, 18, 19, 21, 22, 24, 25, 27, 28, 30, 31, 33, 34, 36, 37, 39, 40, 42, 43, 45, 46, 48, 49, 51, 52, 54, 55, 57, 58, 60, 61, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}}

// returns the approximate filtering strength needed to smooth a edge
// step of 'delta', given a sharpness parameter 'sharpness'.
func VP8FilterStrengthFromDelta( sharpness int , delta int) int {
  pos := tenary.If(delta < MAX_DELTA_SIZE, delta, MAX_DELTA_SIZE - 1)
  assert.Assert(sharpness >= 0 && sharpness <= 7)
  return int(kLevelsFromDelta[sharpness][pos])
}

//------------------------------------------------------------------------------
// Exposed APIs: Encoder should call the following 3 functions to adjust
// loop filter strength

func VP8InitFilter(/* const */ it *vp8.VP8EncIterator) {
  _ = it
}

func VP8StoreFilterStats(/* const */ it *vp8.VP8EncIterator) {
  _ = it
}

func VP8AdjustFilterStrength(/* const */ it *vp8.VP8EncIterator) {
  var enc *vp8.VP8Encoder = it.enc
  if (enc.config.FilterStrength > 0) {
    max_level := 0
    var s int
    for s = 0; s < vp8.NUM_MB_SEGMENTS; s++ {
      var dqm *vp8.VP8SegmentInfo = &enc.dqm[s]
      // this '>> 3' accounts for some inverse WHT scaling
      delta := (dqm.max_edge * dqm.y2.q[1]) >> 3
      level := VP8FilterStrengthFromDelta(enc.filter_hdr.sharpness, delta)
      if (level > dqm.fstrength) {
        dqm.fstrength = level
      }
      if (max_level < dqm.fstrength) {
        max_level = dqm.fstrength
      }
    }
    enc.filter_hdr.level = max_level
  }
}

// -----------------------------------------------------------------------------
