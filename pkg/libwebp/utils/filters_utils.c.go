package utils

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// filter estimation
//
// Author: Urvang (urvang@google.com)

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// -----------------------------------------------------------------------------
// Quick estimate of a potentially interesting filter mode to try.

const SMAX = 16
#define SDIFF(a, b) (abs((a) - (b)) >> 4)  // Scoring diff, in [0..SMAX)

static  int GradientPredictor(uint8 a, uint8 b, uint8 c) {
  g := a + b - c;
  return ((g & ~0xff) == 0) ? g : (g < 0) ? 0 : 255;  // clip to 8bit
}

WEBP_FILTER_TYPE WebPEstimateBestFilter(
    const  *uint8((uint64)height *width) data, int width, int height) {
  int i, j;
  int bins[WEBP_FILTER_LAST][SMAX];
  stdlib.Memset(bins, 0, sizeof(bins));

  // We only sample every other pixels. That's enough.
  for j = 2; j < height - 1; j += 2 {
    var p *uint8 = data + j * width;
    mean := p[0];
    for i = 2; i < width - 1; i += 2 {
      diff0 := SDIFF(p[i], mean);
      diff1 := SDIFF(p[i], p[i - 1]);
      diff2 := SDIFF(p[i], p[i - width]);
      grad_pred :=
          GradientPredictor(p[i - 1], p[i - width], p[i - width - 1]);
      diff3 := SDIFF(p[i], grad_pred);
      bins[WEBP_FILTER_NONE][diff0] = 1;
      bins[WEBP_FILTER_HORIZONTAL][diff1] = 1;
      bins[WEBP_FILTER_VERTICAL][diff2] = 1;
      bins[WEBP_FILTER_GRADIENT][diff3] = 1;
      mean = (3 * mean + p[i] + 2) >> 2;
    }
  }
  {
    int filter;
    WEBP_FILTER_TYPE best_filter = WEBP_FILTER_NONE;
    best_score := 0x7fffffff;
    for filter = WEBP_FILTER_NONE; filter < WEBP_FILTER_LAST; filter++ {
      score := 0;
      for i = 0; i < SMAX; i++ {
        if (bins[filter][i] > 0) {
          score += i;
        }
      }
      if (score < best_score) {
        best_score = score;
        best_filter = (WEBP_FILTER_TYPE)filter;
      }
    }
    return best_filter;
  }
}

#undef SMAX
#undef SDIFF

//------------------------------------------------------------------------------
