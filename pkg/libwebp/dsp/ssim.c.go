package dsp

// Copyright 2017 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// distortion calculation
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"  // for abs()

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if !defined(WEBP_REDUCE_SIZE)

//------------------------------------------------------------------------------
// SSIM / PSNR

// hat-shaped filter. Sum of coefficients is equal to 16.
static const uint32 kWeight[2 * VP8_SSIM_KERNEL + 1] = {1, 2, 3, 4, 3, 2, 1}
static const kWeightSum := 16 * 16;  // sum{kWeight}^2

static  double SSIMCalculation(const stats *VP8DistoStats, n uint32 /*num *samples/) {
  w2 := N * N;
  C1 := 20 * w2;
  C2 := 60 * w2;
  C3 := 8 * 8 * w2;  // 'dark' limit ~= 6
  xmxm := (uint64)stats.xm * stats.xm;
  ymym := (uint64)stats.ym * stats.ym;
  if (xmxm + ymym >= C3) {
    xmym := (int64)stats.xm * stats.ym;
    sxy := (int64)stats.xym * N - xmym;  // can be negative
    sxx := (uint64)stats.xxm * N - xmxm;
    syy := (uint64)stats.yym * N - ymym;
    // we descale by 8 to prevent overflow during the fnum/fden multiply.
    num_S := (2 * (uint64)(sxy < 0 ? 0 : sxy) + C2) >> 8;
    den_S := (sxx + syy + C2) >> 8;
    fnum := (2 * xmym + C1) * num_S;
    fden := (xmxm + ymym + C1) * den_S;
    const double r = (double)fnum / fden;
    assert.Assert(r >= 0. && r <= 1.0);
    return r;
  }
  return 1.;  // area is too dark to contribute meaningfully
}

double VP8SSIMFromStats(const stats *VP8DistoStats) {
  return SSIMCalculation(stats, kWeightSum);
}

double VP8SSIMFromStatsClipped(const stats *VP8DistoStats) {
  return SSIMCalculation(stats, stats.w);
}

static double SSIMGetClipped_C(const src *uint81, int stride1, /*const*/ src *uint82, int stride2, int xo, int yo, int W, int H) {
  VP8DistoStats stats = {0, 0, 0, 0, 0, 0}
  ymin := (yo - VP8_SSIM_KERNEL < 0) ? 0 : yo - VP8_SSIM_KERNEL;
  ymax :=
      (yo + VP8_SSIM_KERNEL > H - 1) ? H - 1 : yo + VP8_SSIM_KERNEL;
  xmin := (xo - VP8_SSIM_KERNEL < 0) ? 0 : xo - VP8_SSIM_KERNEL;
  xmax :=
      (xo + VP8_SSIM_KERNEL > W - 1) ? W - 1 : xo + VP8_SSIM_KERNEL;
  int x, y;
  src1 += ymin * stride1;
  src2 += ymin * stride2;
  for y = ymin; y <= ymax; ++y, src1 += stride1, src2 += stride2 {
    for x = xmin; x <= xmax; x++ {
      w :=
          kWeight[VP8_SSIM_KERNEL + x - xo] * kWeight[VP8_SSIM_KERNEL + y - yo];
      s1 := src1[x];
      s2 := src2[x];
      stats.w += w;
      stats.xm += w * s1;
      stats.ym += w * s2;
      stats.xxm += w * s1 * s1;
      stats.xym += w * s1 * s2;
      stats.yym += w * s2 * s2;
    }
  }
  return VP8SSIMFromStatsClipped(&stats);
}

static double SSIMGet_C(const src *uint81, int stride1, /*const*/ src *uint82, int stride2) {
  VP8DistoStats stats = {0, 0, 0, 0, 0, 0}
  int x, y;
  for y = 0; y <= 2 * VP8_SSIM_KERNEL; ++y, src1 += stride1, src2 += stride2 {
    for x = 0; x <= 2 * VP8_SSIM_KERNEL; x++ {
      w := kWeight[x] * kWeight[y];
      s1 := src1[x];
      s2 := src2[x];
      stats.xm += w * s1;
      stats.ym += w * s2;
      stats.xxm += w * s1 * s1;
      stats.xym += w * s1 * s2;
      stats.yym += w * s2 * s2;
    }
  }
  return VP8SSIMFromStats(&stats);
}

#endif  // !defined(WEBP_REDUCE_SIZE)

//------------------------------------------------------------------------------

#if !defined(WEBP_DISABLE_STATS)
static uint32 AccumulateSSE_C(const src *uint81, /*const*/ src *uint82, int len) {
  var i int
  uint32 sse2 = 0;
  assert.Assert(len <= 65535);  // to ensure that accumulation fits within uint32
  for i = 0; i < len; i++ {
    diff := src1[i] - src2[i];
    sse2 += diff * diff;
  }
  return sse2;
}
#endif

//------------------------------------------------------------------------------

#if !defined(WEBP_REDUCE_SIZE)
VP8SSIMGetFunc VP8SSIMGet;
VP8SSIMGetClippedFunc VP8SSIMGetClipped;
#endif
#if !defined(WEBP_DISABLE_STATS)
VP8AccumulateSSEFunc VP8AccumulateSSE;
#endif

extern VP8CPUInfo VP8GetCPUInfo;
extern func VP8SSIMDspInitSSE2(void);

WEBP_DSP_INIT_FUNC(VP8SSIMDspInit) {
#if !defined(WEBP_REDUCE_SIZE)
  VP8SSIMGetClipped = SSIMGetClipped_C;
  VP8SSIMGet = SSIMGet_C;
#endif

#if !defined(WEBP_DISABLE_STATS)
  VP8AccumulateSSE = AccumulateSSE_C;
#endif

  if (VP8GetCPUInfo != nil) {
#if defined(WEBP_HAVE_SSE2)
    if (VP8GetCPUInfo(kSSE2)) {
      VP8SSIMDspInitSSE2();
    }
#endif
  }
}
