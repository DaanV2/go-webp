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

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

// This table gives, for a given sharpness, the filtering strength to be
// used (at least) in order to filter a given edge step delta.
// This is constructed by brute force inspection: for all delta, we iterate
// over all possible filtering strength / thresh until needs_filter() returns
// true.
const MAX_DELTA_SIZE =64
static const uint8 kLevelsFromDelta[8][MAX_DELTA_SIZE] = {
    {0,  1,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 14, 16, 17, 19, 20, 22, 23, 25, 26, 28, 29, 31, 32, 34, 35, 37, 38, 40, 41, 43, 44, 46, 47, 49, 50, 52, 53, 55, 56, 58, 59, 61, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 13, 15, 16, 18, 19, 21, 22, 24, 25, 27, 28, 30, 31, 33, 34, 36, 37, 39, 40, 42, 43, 45, 46, 48, 49, 51, 52, 54, 55, 57, 58, 60, 61, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  3,  5,  6,  7,  8,  9,  11, 12, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 13, 15, 16, 17, 19, 20, 22, 23, 25, 26, 28, 29, 31, 32, 34, 35, 37, 38, 40, 41, 43, 44, 46, 47, 49, 50, 52, 53, 55, 56, 58, 59, 61, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 13, 15, 16, 18, 19, 21, 22, 24, 25, 27, 28, 30, 31, 33, 34, 36, 37, 39, 40, 42, 43, 45, 46, 48, 49, 51, 52, 54, 55, 57, 58, 60, 61, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}, {0,  1,  2,  4,  5,  7,  8,  9,  11, 12, 14, 15, 17, 18, 20, 21, 23, 24, 26, 27, 29, 30, 32, 33, 35, 36, 38, 39, 41, 42, 44, 45, 47, 48, 50, 51, 53, 54, 56, 57, 59, 60, 62, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63}}

int VP8FilterStrengthFromDelta(int sharpness, int delta) {
  pos := (delta < MAX_DELTA_SIZE) ? delta : MAX_DELTA_SIZE - 1;
  assert.Assert(sharpness >= 0 && sharpness <= 7);
  return kLevelsFromDelta[sharpness][pos];
}

//------------------------------------------------------------------------------
// Paragraph 15.4: compute the inner-edge filtering strength

#if !defined(WEBP_REDUCE_SIZE)

static int GetILevel(int sharpness, int level) {
  if (sharpness > 0) {
    if (sharpness > 4) {
      level >>= 2;
    } else {
      level >>= 1;
    }
    if (level > 9 - sharpness) {
      level = 9 - sharpness;
    }
  }
  if (level < 1) level = 1;
  return level;
}

func DoFilter(const it *VP8EncIterator, int level) {
  var enc *VP8Encoder = it.enc;
  ilevel := GetILevel(enc.config.filter_sharpness, level);
  limit := 2 * level + ilevel;

  var y_dst *uint8 = it.yuv_out2 + Y_OFF_ENC;
  var u_dst *uint8 = it.yuv_out2 + U_OFF_ENC;
  var v_dst *uint8 = it.yuv_out2 + V_OFF_ENC;

  // copy current block to yuv_out2
  memcpy(y_dst, it.yuv_out, YUV_SIZE_ENC * sizeof(uint8));

  if (enc.filter_hdr.simple == 1) {  // simple
    VP8SimpleHFilter16i(y_dst, BPS, limit);
    VP8SimpleVFilter16i(y_dst, BPS, limit);
  } else {  // complex
    hev_thresh := (level >= 40) ? 2 : (level >= 15) ? 1 : 0;
    VP8HFilter16i(y_dst, BPS, limit, ilevel, hev_thresh);
    VP8HFilter8i(u_dst, v_dst, BPS, limit, ilevel, hev_thresh);
    VP8VFilter16i(y_dst, BPS, limit, ilevel, hev_thresh);
    VP8VFilter8i(u_dst, v_dst, BPS, limit, ilevel, hev_thresh);
  }
}

//------------------------------------------------------------------------------
// SSIM metric for one macroblock

static double GetMBSSIM(const yuv *uint81, const yuv *uint82) {
  int x, y;
  double sum = 0.;

  // compute SSIM in a 10 x 10 window
  for (y = VP8_SSIM_KERNEL; y < 16 - VP8_SSIM_KERNEL; y++) {
    for (x = VP8_SSIM_KERNEL; x < 16 - VP8_SSIM_KERNEL; x++) {
      sum += VP8SSIMGetClipped(yuv1 + Y_OFF_ENC, BPS, yuv2 + Y_OFF_ENC, BPS, x, y, 16, 16);
    }
  }
  for (x = 1; x < 7; x++) {
    for (y = 1; y < 7; y++) {
      sum += VP8SSIMGetClipped(yuv1 + U_OFF_ENC, BPS, yuv2 + U_OFF_ENC, BPS, x, y, 8, 8);
      sum += VP8SSIMGetClipped(yuv1 + V_OFF_ENC, BPS, yuv2 + V_OFF_ENC, BPS, x, y, 8, 8);
    }
  }
  return sum;
}

#endif  // !defined(WEBP_REDUCE_SIZE)

//------------------------------------------------------------------------------
// Exposed APIs: Encoder should call the following 3 functions to adjust
// loop filter strength

func VP8InitFilter(const it *VP8EncIterator) {
#if !defined(WEBP_REDUCE_SIZE)
  if (it.lf_stats != nil) {
    int s, i;
    for (s = 0; s < NUM_MB_SEGMENTS; s++) {
      for (i = 0; i < MAX_LF_LEVELS; i++) {
        (*it.lf_stats)[s][i] = 0;
      }
    }
    VP8SSIMDspInit();
  }
#else
  (void)it;
#endif
}

func VP8StoreFilterStats(const it *VP8EncIterator) {
#if !defined(WEBP_REDUCE_SIZE)
  int d;
  var enc *VP8Encoder = it.enc;
  s := it.mb.segment;
  level0 := enc.dqm[s].fstrength;

  // explore +/-quant range of values around level0
  delta_min := -enc.dqm[s].quant;
  delta_max := enc.dqm[s].quant;
  step_size := (delta_max - delta_min >= 4) ? 4 : 1;

  if (it.lf_stats == nil) return;

  // NOTE: Currently we are applying filter only across the sublock edges
  // There are two reasons for that.
  // 1. Applying filter on macro block edges will change the pixels in
  // the left and top macro blocks. That will be hard to restore
  // 2. Macro Blocks on the bottom and right are not yet compressed. So we
  // cannot apply filter on the right and bottom macro block edges.
  if (it.mb.type == 1 && it.mb.skip) return;

  // Always try filter level  zero
  (*it.lf_stats)[s][0] += GetMBSSIM(it.yuv_in, it.yuv_out);

  for (d = delta_min; d <= delta_max; d += step_size) {
    level := level0 + d;
    if (level <= 0 || level >= MAX_LF_LEVELS) {
      continue;
    }
    DoFilter(it, level);
    (*it.lf_stats)[s][level] += GetMBSSIM(it.yuv_in, it.yuv_out2);
  }
#else   // defined(WEBP_REDUCE_SIZE)
  (void)it;
#endif  // !defined(WEBP_REDUCE_SIZE)
}

func VP8AdjustFilterStrength(const it *VP8EncIterator) {
  var enc *VP8Encoder = it.enc;
#if !defined(WEBP_REDUCE_SIZE)
  if (it.lf_stats != nil) {
    int s;
    for (s = 0; s < NUM_MB_SEGMENTS; s++) {
      int i, best_level = 0;
      // Improvement over filter level 0 should be at least 1e-5 (relatively)
      double best_v = 1.00001 * (*it.lf_stats)[s][0];
      for (i = 1; i < MAX_LF_LEVELS; i++) {
        const double v = (*it.lf_stats)[s][i];
        if (v > best_v) {
          best_v = v;
          best_level = i;
        }
      }
      enc.dqm[s].fstrength = best_level;
    }
    return;
  }
#endif  // !defined(WEBP_REDUCE_SIZE)
  if (enc.config.filter_strength > 0) {
    max_level := 0;
    int s;
    for (s = 0; s < NUM_MB_SEGMENTS; s++) {
      var dqm *VP8SegmentInfo = &enc.dqm[s];
      // this '>> 3' accounts for some inverse WHT scaling
      delta := (dqm.max_edge * dqm.y2.q[1]) >> 3;
      level :=
          VP8FilterStrengthFromDelta(enc.filter_hdr.sharpness, delta);
      if (level > dqm.fstrength) {
        dqm.fstrength = level;
      }
      if (max_level < dqm.fstrength) {
        max_level = dqm.fstrength;
      }
    }
    enc.filter_hdr.level = max_level;
  }
}

// -----------------------------------------------------------------------------
