package dsp

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.



const USE_TABLES_FOR_ALPHA_MULT =0  // ALTERNATE_CODE


// -----------------------------------------------------------------------------

const MFIX = 24  // 24bit fixed-point arithmetic
const HALF = ((uint(1) << MFIX) >> 1)
const KINV_255 =((uint(1) << MFIX) / uint(255))

func Mult(x uint8, mult uint32) uint32 {
  v := (x * mult + HALF) >> MFIX
  assert.Assert(v <= 255);  // <- 24bit precision is enough to ensure that.
  return v
}


func GetScale(a uint32, inverse bool) uint32 {
  return tenary.If(inverse, (uint(255) << MFIX) / a,  a * KINV_255)
}

func WebPMultARGBRow_C(/* const */ ptr *uint32, width int, inverse int) {
  var x int
  for x = 0; x < width; x++ {
    argb := ptr[x]
    if (argb < uint(0xff000000)) {     // alpha < 255
      if (argb <= uint(0x00ffffff)) {  // alpha == 0
        ptr[x] = 0
      } else {
        alpha := (argb >> 24) & 0xff
        scale := GetScale(alpha, inverse)
        out := argb & uint(0xff000000)
        out |= Mult(argb >> 0, scale) << 0
        out |= Mult(argb >> 8, scale) << 8
        out |= Mult(argb >> 16, scale) << 16
        ptr[x] = out
      }
    }
  }
}

func WebPMultRow_C(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int) {
  var x int
  for x = 0; x < width; x++ {
    a := alpha[x]
    if (a != 255) {
      if (a == 0) {
        ptr[x] = 0
      } else {
        scale := GetScale(a, inverse)
        ptr[x] = Mult(ptr[x], scale)
      }
    }
  }
}

type WebPMultARGBRow = func(/* const */ ptr *uint32, width int, inverse int)
type WebPMultRow = func(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int)

//------------------------------------------------------------------------------
// Generic per-plane calls

func WebPMultARGBRows(ptr *uint8, stride int, width int, num_rows int , inverse int) {
  var n int
  for n = 0; n < num_rows; n++ {
    WebPMultARGBRow((*uint32)(ptr), width, inverse)
    ptr += stride
  }
}

func WebPMultRows(ptr *uint8, stride int, /*const*/ alpha *uint8, alpha_stride int, width int, num_rows int , inverse int) {
  var n int
  for n = 0; n < num_rows; n++ {
    WebPMultRow(ptr, alpha, width, inverse)
    ptr += stride
    alpha += alpha_stride
  }
}

//------------------------------------------------------------------------------
// Premultiplied modes

// non dithered-modes

// (x * a * 32897) >> 23 is bit-wise equivalent to (int)(x * a / 255.)
// for all 8bit x or a. For bit-wise equivalence to (int)(x * a / 255. + 0.5),
// one can use instead: (x * a * 65793 + (1 << 23)) >> 24

func MULTIPLIER1(a uint) int {
	return int((a) * uint(32897))
}
func PREMULTIPLY(x, m uint) int {
	return (((x) * (m)) >> 23)
}

func ApplyAlphaMultiply_C(rgba *uint8, alpha_first int, w int, h int, stride int) {
  for ; h > 0; h-- {
    var rgb *uint8 = rgba + (tenary.If(alpha_first, 1, 0))
    var alpha *uint8 = rgba + (tenary.If(alpha_first, 0, 3))
    var i int
    for i = 0; i < w; i++ {
      a := alpha[4 * i]
      if (a != 0xff) {
        mult := MULTIPLIER1(a)
        rgb[4 * i + 0] = PREMULTIPLY(rgb[4 * i + 0], mult)
        rgb[4 * i + 1] = PREMULTIPLY(rgb[4 * i + 1], mult)
        rgb[4 * i + 2] = PREMULTIPLY(rgb[4 * i + 2], mult)
      }
    }
    rgba += stride
  }
}

func MULTIPLIER(a int) int {
	return ((a) * 0x1111)  // 0x1111 ~= (1 << 16) / 15
}

func dither_hi(x uint8) uint8 {
  return (x & 0xf0) | (x >> 4)
}

func dither_lo(x uint8) uint8 {
  return (x & 0x0f) | (x << 4)
}

func multiply(x uint8, m uint32) uint8 {
  return (x * m) >> 16
}

func ApplyAlphaMultiply4444_C(rgba *uint84444, w int, h int, stride int, rg_byte_pos int /* 0 or 1 */) {
  for ; h > 0; h-- {
    var i int
    for i = 0; i < w; i++ {
      rg := rgba4444[2 * i + rg_byte_pos]
      ba := rgba4444[2 * i + (rg_byte_pos ^ 1)]
      a := ba & 0x0f
      mult := MULTIPLIER(a)
      r := multiply(dither_hi(rg), mult)
      g := multiply(dither_lo(rg), mult)
      b := multiply(dither_hi(ba), mult)
      rgba4444[2 * i + rg_byte_pos] = (r & 0xf0) | ((g >> 4) & 0x0f)
      rgba4444[2 * i + (rg_byte_pos ^ 1)] = (b & 0xf0) | a
    }
    rgba4444 += stride
  }
}

func ApplyAlphaMultiply_16b_C(rgba *uint84444, w int, h int, stride int) {
  ApplyAlphaMultiply4444_C(rgba4444, w, h, stride, 1)
}

func DispatchAlpha_C(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint8, dst_stride int) int {
  alpha_mask := 0xff
  var i, j int

  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      alpha_value := alpha[i]
      dst[4 * i] = alpha_value
      alpha_mask &= alpha_value
    }
    alpha += alpha_stride
    dst += dst_stride
  }

  return (alpha_mask != 0xff)
}

func DispatchAlphaToGreen_C(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint32, dst_stride int) {
  var i, j int
  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      dst[i] = alpha[i] << 8;  // leave A/R/B channels zero'd.
    }
    alpha += alpha_stride
    dst += dst_stride
  }
}

func ExtractAlpha_C(/* const */ argb *uint8, argb_stride int, width, height int, alpha *uint8, alpha_stride int) int {
  alpha_mask := 0xff
  var i, j int

  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      alpha_value := argb[4 * i]
      alpha[i] = alpha_value
      alpha_mask &= alpha_value
    }
    argb += argb_stride
    alpha += alpha_stride
  }
  return (alpha_mask == 0xff)
}

func ExtractGreen_C(/* const */ argb *uint32, alpha *uint8, size int) {
  var i int
  for (i = 0; i < size; ++i) {
	alpha[i] = argb[i] >> 8
}
}

func MakeARGB32(a, r, g, b uint32) uint32 {
  return uint32((a << 24) | (r << 16) | (g << 8) | b)
}

func PackARGB_C(/* const */ a *uint8, /*const*/ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, out *uint32) {
  var i int
  for i = 0; i < len; i++ {
    out[i] = MakeARGB32(a[4 * i], r[4 * i], g[4 * i], b[4 * i])
  }
}

func PackRGB_C(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, step int, out *uint32) {
  int i, offset = 0
  for i = 0; i < len; i++ {
    out[i] = MakeARGB32(0xff, r[offset], g[offset], b[offset])
    offset += step
  }
}

type WebPApplyAlphaMultiply = func(*uint8, int, int, int, int)
func WebPApplyAlphaMultiply4444 = func(*uint8, int, int, int)
type WebPDispatchAlpha = func(/* const */ *uint8, int, int, int, *uint8, int)
type WebPDispatchAlphaToGreen = func(/* const */ *uint8, int, int, int, *uint32, int)
type WebPExtractAlpha = func(/* const */ *uint8, int, int, int, *uint8, int)
type WebPExtractGreen = func(/* const */ argb *uint32, alpha *uint8, size int)
type WebPPackARGB = func(/* const */ a *uint8, /*const*/ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, int, *uint32)
type WebPPackRGB = func(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, step int, out *uint32)

int (*WebPHasAlpha8b)(/* const */ src *uint8, length int)
int (*WebPHasAlpha32b)(/* const */ src *uint8, length int)
type WebPAlphaReplace = func(src *uint32, length int, color uint32)

//------------------------------------------------------------------------------
// Init function


extern func WebPInitAlphaProcessingMIPSdspR2(void)
extern func WebPInitAlphaProcessingSSE2(void)
extern func WebPInitAlphaProcessingSSE41(void)
extern func WebPInitAlphaProcessingNEON(void)

WEBP_DSP_INIT_FUNC(WebPInitAlphaProcessing) {
  WebPMultARGBRow = WebPMultARGBRow_C
  WebPMultRow = WebPMultRow_C
  WebPApplyAlphaMultiply4444 = ApplyAlphaMultiply_16b_C

#ifdef constants.WORDS_BIGENDIAN
  WebPPackARGB = PackARGB_C
#endif
  WebPPackRGB = PackRGB_C
#if !WEBP_NEON_OMIT_C_CODE
  WebPApplyAlphaMultiply = ApplyAlphaMultiply_C
  WebPDispatchAlpha = DispatchAlpha_C
  WebPDispatchAlphaToGreen = DispatchAlphaToGreen_C
  WebPExtractAlpha = ExtractAlpha_C
  WebPExtractGreen = ExtractGreen_C
#endif

  WebPHasAlpha8b = HasAlpha8b_C
  WebPHasAlpha32b = HasAlpha32b_C
  WebPAlphaReplace = AlphaReplace_C

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if false
    if (VP8GetCPUInfo(kSSE2)) {
      WebPInitAlphaProcessingSSE2()
#if false
      if (VP8GetCPUInfo(kSSE4_1)) {
        WebPInitAlphaProcessingSSE41()
      }
#endif
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      WebPInitAlphaProcessingMIPSdspR2()
    }
#endif
  }

#if FALSE
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    WebPInitAlphaProcessingNEON()
  }
#endif

  assert.Assert(WebPMultARGBRow != nil)
  assert.Assert(WebPMultRow != nil)
  assert.Assert(WebPApplyAlphaMultiply != nil)
  assert.Assert(WebPApplyAlphaMultiply4444 != nil)
  assert.Assert(WebPDispatchAlpha != nil)
  assert.Assert(WebPDispatchAlphaToGreen != nil)
  assert.Assert(WebPExtractAlpha != nil)
  assert.Assert(WebPExtractGreen != nil)
#ifdef constants.WORDS_BIGENDIAN
  assert.Assert(WebPPackARGB != nil)
#endif
  assert.Assert(WebPPackRGB != nil)
  assert.Assert(WebPHasAlpha8b != nil)
  assert.Assert(WebPHasAlpha32b != nil)
  assert.Assert(WebPAlphaReplace != nil)
}
