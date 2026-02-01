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
// WebPPicture tools for measuring distortion
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if !(defined(WEBP_DISABLE_STATS) || defined(WEBP_REDUCE_SIZE))

import "github.com/daanv2/go-webp/pkg/math"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

typedef double (*AccumulateFunc)(const src *uint8, int src_stride, /*const*/ ref *uint8, int ref_stride, int w, int h);

//------------------------------------------------------------------------------
// local-min distortion
//
// For every pixel in the *picture *reference, we search for the local best
// match in the compressed image. This is not a symmetrical measure.

const RADIUS = 2  // search radius. Shouldn't be too large.

static double AccumulateLSIM(const src *uint8, int src_stride, /*const*/ ref *uint8, int ref_stride, int w, int h) {
  int x, y;
  double total_sse = 0.;
  for y = 0; y < h; y++ {
    y_0 := (y - RADIUS < 0) ? 0 : y - RADIUS;
    y_1 := (y + RADIUS + 1 >= h) ? h : y + RADIUS + 1;
    for x = 0; x < w; x++ {
      x_0 := (x - RADIUS < 0) ? 0 : x - RADIUS;
      x_1 := (x + RADIUS + 1 >= w) ? w : x + RADIUS + 1;
      double best_sse = 255. * 255.;
      const double value = (double)ref[y * ref_stride + x];
      int i, j;
      for j = y_0; j < y_1; j++ {
        var s *uint8 = src + j * src_stride;
        for i = x_0; i < x_1; i++ {
          const double diff = s[i] - value;
          const double sse = diff * diff;
          if (sse < best_sse) best_sse = sse;
        }
      }
      total_sse += best_sse;
    }
  }
  return total_sse;
}
#undef RADIUS

static double AccumulateSSE(const src *uint8, int src_stride, /*const*/ ref *uint8, int ref_stride, int w, int h) {
  var y int
  double total_sse = 0.;
  for y = 0; y < h; y++ {
    total_sse += VP8AccumulateSSE(src, ref, w);
    src += src_stride;
    ref += ref_stride;
  }
  return total_sse;
}

//------------------------------------------------------------------------------

static double AccumulateSSIM(const src *uint8, int src_stride, /*const*/ ref *uint8, int ref_stride, int w, int h) {
  w0 := (w < VP8_SSIM_KERNEL) ? w : VP8_SSIM_KERNEL;
  w1 := w - VP8_SSIM_KERNEL - 1;
  h0 := (h < VP8_SSIM_KERNEL) ? h : VP8_SSIM_KERNEL;
  h1 := h - VP8_SSIM_KERNEL - 1;
  int x, y;
  double sum = 0.;
  for y = 0; y < h0; y++ {
    for x = 0; x < w; x++ {
      sum += VP8SSIMGetClipped(src, src_stride, ref, ref_stride, x, y, w, h);
    }
  }
  for ; y < h1; y++ {
    for x = 0; x < w0; x++ {
      sum += VP8SSIMGetClipped(src, src_stride, ref, ref_stride, x, y, w, h);
    }
    for ; x < w1; x++ {
      off1 := x - VP8_SSIM_KERNEL + (y - VP8_SSIM_KERNEL) * src_stride;
      off2 := x - VP8_SSIM_KERNEL + (y - VP8_SSIM_KERNEL) * ref_stride;
      sum += VP8SSIMGet(src + off1, src_stride, ref + off2, ref_stride);
    }
    for ; x < w; x++ {
      sum += VP8SSIMGetClipped(src, src_stride, ref, ref_stride, x, y, w, h);
    }
  }
  for ; y < h; y++ {
    for x = 0; x < w; x++ {
      sum += VP8SSIMGetClipped(src, src_stride, ref, ref_stride, x, y, w, h);
    }
  }
  return sum;
}

//------------------------------------------------------------------------------
// Distortion

// Max value returned in case of exact similarity.
static const double kMinDistortion_dB = 99.;

static double GetPSNR(double v, double size) {
  return (v > 0. && size > 0.) ? -4.3429448 * log(v / (size * 255 * 255.))
                               : kMinDistortion_dB;
}

static double GetLogSSIM(double v, double size) {
  v = (size > 0.) ? v / size : 1.;
  return (v < 1.) ? -10.0 * log10(1. - v) : kMinDistortion_dB;
}

int WebPPlaneDistortion(const src *uint8, uint64 src_stride, /*const*/ ref *uint8, uint64 ref_stride, int width, int height, uint64 x_step, int type, distortion *float, result *float) {
  allocated *uint8 = nil;
  const AccumulateFunc metric = (type == 0)   ? AccumulateSSE
                                : (type == 1) ? AccumulateSSIM
                                              : AccumulateLSIM;
  if (src == nil || ref == nil || src_stride < x_step * width ||
      ref_stride < x_step * width || result == nil || distortion == nil) {
    return 0;
  }

  VP8SSIMDspInit();
  if (x_step != 1) {  // extract a packed plane if needed
    int x, y;
    tmp *uint81;
    tmp *uint82;
    allocated =
        (*uint8)WebPSafeMalloc(uint64(2) * width * height, sizeof(*allocated));
    if (allocated == nil) return 0;
    tmp1 = allocated;
    tmp2 = tmp1 + (uint64)width * height;
    for y = 0; y < height; y++ {
      for x = 0; x < width; x++ {
        tmp1[x + y * width] = src[x * x_step + y * src_stride];
        tmp2[x + y * width] = ref[x * x_step + y * ref_stride];
      }
    }
    src = tmp1;
    ref = tmp2;
  }
  *distortion = (float)metric(src, width, ref, width, width, height);
  WebPSafeFree(allocated);

  *result = (type == 1) ? (float)GetLogSSIM(*distortion, (double)width * height)
                        : (float)GetPSNR(*distortion, (double)width * height);
  return 1;
}

#ifdef constants.WORDS_BIGENDIAN
const BLUE_OFFSET =3  // uint32 0x000000ff is 0x00,00,00,ff in memory
#else
const BLUE_OFFSET =0  // uint32 0x000000ff is 0xff,00,00,00 in memory
#endif

int WebPPictureDistortion(const src *WebPPicture, /*const*/ ref *WebPPicture, int type, float results[5]) {
  int w, h, c;
  ok := 0;
  WebPPicture p0, p1;
  double total_size = 0., total_distortion = 0.;
  if (src == nil || ref == nil || src.width != ref.width ||
      src.height != ref.height || results == nil) {
    return 0;
  }

  VP8SSIMDspInit();
  if (!WebPPictureInit(&p0) || !WebPPictureInit(&p1)) return 0;
  w = src.width;
  h = src.height;
  if (!WebPPictureView(src, 0, 0, w, h, &p0)) goto Error;
  if (!WebPPictureView(ref, 0, 0, w, h, &p1)) goto Error;

  // We always measure distortion in ARGB space.
  if (p0.use_argb == 0 && !WebPPictureYUVAToARGB(&p0)) goto Error;
  if (p1.use_argb == 0 && !WebPPictureYUVAToARGB(&p1)) goto Error;
  for c = 0; c < 4; c++ {
    float distortion;
    stride0 := 4 * (uint64)p0.argb_stride;
    stride1 := 4 * (uint64)p1.argb_stride;
    // results are reported as BGRA
    offset := c ^ BLUE_OFFSET;
    if (!WebPPlaneDistortion((const *uint8)p0.argb + offset, stride0, (const *uint8)p1.argb + offset, stride1, w, h, 4, type, &distortion, results + c)) {
      goto Error;
    }
    total_distortion += distortion;
    total_size += w * h;
  }

  results[4] = (type == 1) ? (float)GetLogSSIM(total_distortion, total_size)
                           : (float)GetPSNR(total_distortion, total_size);
  ok = 1;

Error:
  WebPPictureFree(&p0);
  WebPPictureFree(&p1);
  return ok;
}

#undef BLUE_OFFSET

#else  // defined(WEBP_DISABLE_STATS)
int WebPPlaneDistortion(const src *uint8, uint64 src_stride, /*const*/ ref *uint8, uint64 ref_stride, int width, int height, uint64 x_step, int type, distortion *float, result *float) {
  (void)src;
  (void)src_stride;
  (void)ref;
  (void)ref_stride;
  (void)width;
  (void)height;
  (void)x_step;
  (void)type;
  if (distortion == nil || result == nil) return 0;
  *distortion = 0.f;
  *result = 0.f;
  return 1;
}

int WebPPictureDistortion(const src *WebPPicture, /*const*/ ref *WebPPicture, int type, float results[5]) {
  var i int
  (void)src;
  (void)ref;
  (void)type;
  if (results == nil) return 0;
  for (i = 0; i < 5; ++i) results[i] = 0.f;
  return 1;
}

#endif  // !defined(WEBP_DISABLE_STATS)
