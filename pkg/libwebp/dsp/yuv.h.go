package dsp

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// inline YUV<.RGB conversion function
//
// The exact naming is Y'CbCr, following the ITU-R BT.601 standard.
// More information at: https://en.wikipedia.org/wiki/YCbCr
// Y = 0.2568 * R + 0.5041 * G + 0.0979 * B + 16
// U = -0.1482 * R - 0.2910 * G + 0.4392 * B + 128
// V = 0.4392 * R - 0.3678 * G - 0.0714 * B + 128
// We use 16bit fixed point operations for RGB.YUV conversion (YUV_FIX).
//
// For the Y'CbCr to RGB conversion, the BT.601 specification reads:
//   R = 1.164 * (Y-16) + 1.596 * (V-128)
//   G = 1.164 * (Y-16) - 0.813 * (V-128) - 0.392 * (U-128)
//   B = 1.164 * (Y-16)                   + 2.017 * (U-128)
// where Y is in the [16,235] range, and U/V in the [16,240] range.
//
// The fixed-point implementation used here is:
//  R = (19077 . y             + 26149 . v - 14234) >> 6
//  G = (19077 . y -  6419 . u - 13320 . v +  8708) >> 6
//  B = (19077 . y + 33050 . u             - 17685) >> 6
// where the '.' operator is the mulhi_epu16 variant:
//   a . b = ((a << 8) * b) >> 16
// that preserves 8 bits of fractional precision before final descaling.

// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// Macros to give the offset of each channel in a uint32 containing ARGB.
#ifdef constants.WORDS_BIGENDIAN
// uint32 0xff000000 is 0xff,00,00,00 in memory
#define CHANNEL_OFFSET(i) (i)
#else
// uint32 0xff000000 is 0x00,00,00,ff in memory
#define CHANNEL_OFFSET(i) (3 - (i))
#endif

//------------------------------------------------------------------------------
// YUV . RGB conversion


enum {
  YUV_FIX = 16,  // fixed-point precision for RGB.YUV
  YUV_HALF = 1 << (YUV_FIX - 1),

  YUV_FIX2 = 6,  // fixed-point precision for YUV.RGB
  YUV_MASK2 = (256 << YUV_FIX2) - 1
}

//------------------------------------------------------------------------------
// slower on x86 by ~7-8%, but bit-exact with the SSE2/NEON version

func MultHi(int v, int coeff) int {  // _mm_mulhi_epu16 emulation
  return (v * coeff) >> 8;
}

func VP8Clip8(int v) int {
  return ((v & ~YUV_MASK2) == 0) ? (v >> YUV_FIX2) : (v < 0) ? 0 : 255;
}

func VP8YUVToR(int y, int v) int {
  return VP8Clip8(MultHi(y, 19077) + MultHi(v, 26149) - 14234);
}

func VP8YUVToG(int y, int u, int v) int {
  return VP8Clip8(MultHi(y, 19077) - MultHi(u, 6419) - MultHi(v, 13320) + 8708);
}

func VP8YUVToB(int y, int u) int {
  return VP8Clip8(MultHi(y, 19077) + MultHi(u, 33050) - 17685);
}

func VP8YuvToRgb(int y, int u, int v, /*const*/ rgb *uint8) {
  rgb[0] = VP8YUVToR(y, v);
  rgb[1] = VP8YUVToG(y, u, v);
  rgb[2] = VP8YUVToB(y, u);
}

func VP8YuvToBgr(int y, int u, int v, /*const*/ bgr *uint8) {
  bgr[0] = VP8YUVToB(y, u);
  bgr[1] = VP8YUVToG(y, u, v);
  bgr[2] = VP8YUVToR(y, v);
}

func VP8YuvToRgb565(int y, int u, int v, /*const*/ rgb *uint8) {
  r := VP8YUVToR(y, v);     // 5 usable bits
  g := VP8YUVToG(y, u, v);  // 6 usable bits
  b := VP8YUVToB(y, u);     // 5 usable bits
  rg := (r & 0xf8) | (g >> 5);
  gb := ((g << 3) & 0xe0) | (b >> 3);
#if (WEBP_SWAP_16BIT_CSP == 1)
  rgb[0] = gb;
  rgb[1] = rg;
#else
  rgb[0] = rg;
  rgb[1] = gb;
#endif
}

func VP8YuvToRgba4444(int y, int u, int v, /*const*/ argb *uint8) {
  r := VP8YUVToR(y, v);     // 4 usable bits
  g := VP8YUVToG(y, u, v);  // 4 usable bits
  b := VP8YUVToB(y, u);     // 4 usable bits
  rg := (r & 0xf0) | (g >> 4);
  ba := (b & 0xf0) | 0x0f;  // overwrite the lower 4 bits
#if (WEBP_SWAP_16BIT_CSP == 1)
  argb[0] = ba;
  argb[1] = rg;
#else
  argb[0] = rg;
  argb[1] = ba;
#endif
}

//-----------------------------------------------------------------------------
// Alpha handling variants

func VP8YuvToArgb(uint8 y, uint8 u, uint8 v, /*const*/ argb *uint8) {
  argb[0] = 0xff;
  VP8YuvToRgb(y, u, v, argb + 1);
}

func VP8YuvToBgra(uint8 y, uint8 u, uint8 v, /*const*/ bgra *uint8) {
  VP8YuvToBgr(y, u, v, bgra);
  bgra[3] = 0xff;
}

func VP8YuvToRgba(uint8 y, uint8 u, uint8 v, /*const*/ rgba *uint8) {
  VP8YuvToRgb(y, u, v, rgba);
  rgba[3] = 0xff;
}

//-----------------------------------------------------------------------------
// SSE2 extra functions (mostly for upsampling_sse2.c)

#if defined(WEBP_USE_SSE2)

// Process 32 pixels and store the result (16b, 24b or 32b per pixel) in *dst.
func VP8YuvToRgba32_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToRgb32_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToBgra32_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToBgr32_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToArgb32_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToRgba444432_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToRgb56532_SSE2(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);

#endif  // WEBP_USE_SSE2

//-----------------------------------------------------------------------------
// SSE41 extra functions (mostly for upsampling_sse41.c)

#if defined(WEBP_USE_SSE41)

// Process 32 pixels and store the result (16b, 24b or 32b per pixel) in *dst.
func VP8YuvToRgb32_SSE41(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);
func VP8YuvToBgr32_SSE41(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8);

#endif  // WEBP_USE_SSE41

//------------------------------------------------------------------------------
// RGB . YUV conversion

// Stub functions that can be called with various rounding values:
func VP8ClipUV(int uv, int rounding) int {
  uv = (uv + rounding + (128 << (YUV_FIX + 2))) >> (YUV_FIX + 2);
  return ((uv & ~0xff) == 0) ? uv : (uv < 0) ? 0 : 255;
}

func VP8RGBToY(int r, int g, int b, int rounding) int {
  luma := 16839 * r + 33059 * g + 6420 * b;
  return (luma + rounding + (16 << YUV_FIX)) >> YUV_FIX;  // no need to clip
}

func VP8RGBToU(int r, int g, int b, int rounding) int {
  u := -9719 * r - 19081 * g + 28800 * b;
  return VP8ClipUV(u, rounding);
}

func VP8RGBToV(int r, int g, int b, int rounding) int {
  v := +28800 * r - 24116 * g - 4684 * b;
  return VP8ClipUV(v, rounding);
}

// has_alpha is true if there is an alpha value that is not 0xff.
extern func (*WebPImportYUVAFromRGBA)(
    const r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, /*const*/ a_ptr *uint8, int step,        // bytes per pixel
    int rgb_stride,  // bytes per scanline
    int has_alpha, width, height int, tmp_rgb *uint16, int y_stride, int uv_stride, a_stride int, dst_y *uint8, dst_u *uint8, dst_v *uint8, dst_a *uint8);
extern func (*WebPImportYUVAFromRGBALastLine)(
    const r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, /*const*/ a_ptr *uint8, int step,  // bytes per pixel
    int has_alpha, int width, tmp_rgb *uint16, dst_y *uint8, dst_u *uint8, dst_v *uint8, dst_a *uint8);

// Internal function to that can be *WebPImportYUVAFromRGBA reused.
func WebPAccumulateRGBA(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, /*const*/ a_ptr *uint8, int rgb_stride, dst *uint16, int width);
func WebPAccumulateRGB(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, int step, int rgb_stride, dst *uint16, int width);
// Must be called before calling *WebPAccumulateRGB.
func WebPInitGammaTables(void);



#endif  // WEBP_DSP_YUV_H_
