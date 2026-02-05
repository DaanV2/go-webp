package dsp

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Image transforms and color space conversion methods for lossless decoder.
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)
//          Urvang Joshi (urvang@google.com)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------
// Image transforms.

func Average2(uint32 a0, uint32 a1) uint32 {
  return (((a0 ^ a1) & uint(0xfefefefe)) >> 1) + (a0 & a1);
}

func Average3(uint32 a0, uint32 a1, uint32 a2) uint32 {
  return Average2(Average2(a0, a2), a1);
}

func Average4(uint32 a0, uint32 a1, uint32 a2, uint32 a3) uint32 {
  return Average2(Average2(a0, a1), Average2(a2, a3));
}

func Clip255(uint32 a) uint32 {
  if (a < 256) {
    return a;
  }
  // return 0, when a is a negative integer.
  // return 255, when a is positive.
  return ~a >> 24;
}

func AddSubtractComponentFull(int a, b int, c int) int {
  return Clip255((uint32)(a + b - c));
}

func ClampedAddSubtractFull(uint32 c0, uint32 c1, uint32 c2) uint32 {
  a := AddSubtractComponentFull(c0 >> 24, c1 >> 24, c2 >> 24);
  r := AddSubtractComponentFull((c0 >> 16) & 0xff, (c1 >> 16) & 0xff, (c2 >> 16) & 0xff);
  g := AddSubtractComponentFull((c0 >> 8) & 0xff, (c1 >> 8) & 0xff, (c2 >> 8) & 0xff);
  b := AddSubtractComponentFull(c0 & 0xff, c1 & 0xff, c2 & 0xff);
  return ((uint32)a << 24) | (r << 16) | (g << 8) | b;
}

func AddSubtractComponentHalf(int a, b int) int {
  return Clip255((uint32)(a + (a - b) / 2));
}

func ClampedAddSubtractHalf(uint32 c0, uint32 c1, uint32 c2) uint32 {
  ave := Average2(c0, c1);
  a := AddSubtractComponentHalf(ave >> 24, c2 >> 24);
  r := AddSubtractComponentHalf((ave >> 16) & 0xff, (c2 >> 16) & 0xff);
  g := AddSubtractComponentHalf((ave >> 8) & 0xff, (c2 >> 8) & 0xff);
  b := AddSubtractComponentHalf((ave >> 0) & 0xff, (c2 >> 0) & 0xff);
  return ((uint32)a << 24) | (r << 16) | (g << 8) | b;
}

// gcc <= 4.9 on ARM generates incorrect code in Select() when Sub3() is
// inlined.
#if defined(__arm__) && defined(__GNUC__) && LOCAL_GCC_VERSION <= 0x409
const LOCAL_INLINE =__attribute__((noinline))
#else
const LOCAL_INLINE =
#endif

static LOCAL_INLINE int Sub3(int a, b int, c int) {
  pb := b - c;
  pa := a - c;
  return abs(pb) - abs(pa);
}

#undef LOCAL_INLINE

func Select(uint32 a, uint32 b, uint32 c) uint32 {
  pa_minus_pb :=
      Sub3((a >> 24), (b >> 24), (c >> 24)) +
      Sub3((a >> 16) & 0xff, (b >> 16) & 0xff, (c >> 16) & 0xff) +
      Sub3((a >> 8) & 0xff, (b >> 8) & 0xff, (c >> 8) & 0xff) +
      Sub3((a) & 0xff, (b) & 0xff, (c) & 0xff);
  return (pa_minus_pb <= 0) ? a : b;
}

//------------------------------------------------------------------------------
// Predictors

func VP8LPredictor0_C(/* const */ left *uint32, /*const*/ top *uint32) uint32 {
  (void)top;
  (void)left;
  return ARGB_BLACK;
}
func VP8LPredictor1_C(/* const */ left *uint32, /*const*/ top *uint32) uint32 {
  (void)top;
  return *left;
}
uint32 VP8LPredictor2_C(/* const */ left *uint32, /*const*/ top *uint32) {
  (void)left;
  return top[0];
}
uint32 VP8LPredictor3_C(/* const */ left *uint32, /*const*/ top *uint32) {
  (void)left;
  return top[1];
}
uint32 VP8LPredictor4_C(/* const */ left *uint32, /*const*/ top *uint32) {
  (void)left;
  return top[-1];
}
uint32 VP8LPredictor5_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average3(*left, top[0], top[1]);
  return pred;
}
uint32 VP8LPredictor6_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average2(*left, top[-1]);
  return pred;
}
uint32 VP8LPredictor7_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average2(*left, top[0]);
  return pred;
}
uint32 VP8LPredictor8_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average2(top[-1], top[0]);
  (void)left;
  return pred;
}
uint32 VP8LPredictor9_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average2(top[0], top[1]);
  (void)left;
  return pred;
}
uint32 VP8LPredictor10_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Average4(*left, top[-1], top[0], top[1]);
  return pred;
}
uint32 VP8LPredictor11_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := Select(top[0], *left, top[-1]);
  return pred;
}
uint32 VP8LPredictor12_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := ClampedAddSubtractFull(*left, top[0], top[-1]);
  return pred;
}
uint32 VP8LPredictor13_C(/* const */ left *uint32, /*const*/ top *uint32) {
  pred := ClampedAddSubtractHalf(*left, top[0], top[-1]);
  return pred;
}

func PredictorAdd0_C(/* const */ in *uint32, /*const*/ upper *uint32, num_pixels int, out *uint32) {
  var x int
  (void)upper;
  for (x = 0; x < num_pixels; ++x) out[x] = VP8LAddPixels(in[x], ARGB_BLACK);
}
func PredictorAdd1_C(/* const */ in *uint32, /*const*/ upper *uint32, num_pixels int, out *uint32) {
  var i int
  left := out[-1];
  (void)upper;
  for i = 0; i < num_pixels; i++ {
    out[i] = left = VP8LAddPixels(in[i], left);
  }
}
GENERATE_PREDICTOR_ADD(VP8LPredictor2_C, PredictorAdd2_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor3_C, PredictorAdd3_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor4_C, PredictorAdd4_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor5_C, PredictorAdd5_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor6_C, PredictorAdd6_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor7_C, PredictorAdd7_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor8_C, PredictorAdd8_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor9_C, PredictorAdd9_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor10_C, PredictorAdd10_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor11_C, PredictorAdd11_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor12_C, PredictorAdd12_C)
GENERATE_PREDICTOR_ADD(VP8LPredictor13_C, PredictorAdd13_C)

//------------------------------------------------------------------------------

// Inverse prediction.
func PredictorInverseTransform_C(/* const */ transform *VP8LTransform, y_start int, y_end int, /*const*/ in *uint32, out *uint32) {
  width := transform.xsize;
  if (y_start == 0) {  // First Row follows the L (mode=1) mode.
    PredictorAdd0_C(in, nil, 1, out);
    PredictorAdd1_C(in + 1, nil, width - 1, out + 1);
    in += width;
    out += width;
    y_start++
  }

  {
    y := y_start;
    tile_width := 1 << transform.bits;
    mask := tile_width - 1;
    tiles_per_row := VP8LSubSampleSize(width, transform.bits);
    const pred_mode_base *uint32 =
        transform.data + (y >> transform.bits) * tiles_per_row;

    while (y < y_end) {
      var pred_mode_src *uint32 = pred_mode_base;
      x := 1;
      // First pixel follows the T (mode=2) mode.
      PredictorAdd2_C(in, out - width, 1, out);
      // .. the rest:
      while (x < width) {
        var pred_func VP8LPredictorAddSubFunc =
            VP8LPredictorsAdd[((*pred_mode_src++) >> 8) & 0xf];
        x_end := (x & ~mask) + tile_width;
        if x_end > width { x_end = width }
        pred_func(in + x, out + x - width, x_end - x, out + x);
        x = x_end;
      }
      in += width;
      out += width;
      y++
      if ((y & mask) == 0) {  // Use the same mask, since tiles are squares.
        pred_mode_base += tiles_per_row;
      }
    }
  }
}

// Add green to blue and red channels (i.e. perform the inverse transform of
// 'subtract green').
func VP8LAddGreenToBlueAndRed_C(/* const */ src *uint32, num_pixels int, dst *uint32) {
  var i int
  for i = 0; i < num_pixels; i++ {
    argb := src[i];
    green := ((argb >> 8) & 0xff);
    red_blue := (argb & uint(0x00ff00ff));
    red_blue += (green << 16) | green;
    red_blue &= uint(0x00ff00ff);
    dst[i] = (argb & uint(0xff00ff00)) | red_blue;
  }
}

func ColorTransformDelta(int8 color_pred, int8 color) int {
  return ((int)color_pred * color) >> 5;
}

func ColorCodeToMultipliers(uint32 color_code, /*const*/ m *VP8LMultipliers) {
  m.green_to_red = (color_code >> 0) & 0xff;
  m.green_to_blue = (color_code >> 8) & 0xff;
  m.red_to_blue = (color_code >> 16) & 0xff;
}

func VP8LTransformColorInverse_C(/* const */ m *VP8LMultipliers, /*const*/ src *uint32, num_pixels int, dst *uint32) {
  var i int
  for i = 0; i < num_pixels; i++ {
    argb := src[i];
    green := (int8)(argb >> 8);
    red := argb >> 16;
    new_red := red & 0xff;
    new_blue := argb & 0xff;
    new_red += ColorTransformDelta((int8)m.green_to_red, green);
    new_red &= 0xff;
    new_blue += ColorTransformDelta((int8)m.green_to_blue, green);
    new_blue += ColorTransformDelta((int8)m.red_to_blue, (int8)new_red);
    new_blue &= 0xff;
    dst[i] = (argb & uint(0xff00ff00)) | (new_red << 16) | (new_blue);
  }
}

// Color space inverse transform.
func ColorSpaceInverseTransform_C(/* const */ transform *VP8LTransform, y_start int, y_end int, /*const*/ src *uint32, dst *uint32) {
  width := transform.xsize;
  tile_width := 1 << transform.bits;
  mask := tile_width - 1;
  safe_width := width & ~mask;
  remaining_width := width - safe_width;
  tiles_per_row := VP8LSubSampleSize(width, transform.bits);
  y := y_start;
  const pred_row *uint32 =
      transform.data + (y >> transform.bits) * tiles_per_row;

  while (y < y_end) {
    var pred *uint32 = pred_row;
    VP8LMultipliers m = {0, 0, 0}
    var src_safe_end *uint32 = src + safe_width;
    var src_end *uint32 = src + width;
    while (src < src_safe_end) {
      ColorCodeToMultipliers(*pred++, &m);
      VP8LTransformColorInverse(&m, src, tile_width, dst);
      src += tile_width;
      dst += tile_width;
    }
    if (src < src_end) {  // Left-overs using C-version.
      ColorCodeToMultipliers(*pred++, &m);
      VP8LTransformColorInverse(&m, src, remaining_width, dst);
      src += remaining_width;
      dst += remaining_width;
    }
    y++
    if (y & mask) == 0 { pred_row += tiles_per_row }
  }
}

// Separate out pixels packed together using pixel-bundling.
// We define two methods for ARGB data (uint32) and alpha-only data (uint8).
// clang-format off
#define COLOR_INDEX_INVERSE(FUNC_NAME, F_NAME, STATIC_DECL, TYPE, BIT_SUFFIX,  \
                            GET_INDEX, GET_VALUE)                              \
func F_NAME(/* const */ src *TYPE, /*const*/ color_map *uint32,           \
                   dst *TYPE, y_start int, y_end int, width int) {             \
  var y int                                                                       \
  for y = y_start; y < y_end; y++ {                                          \
    var x int                                                                     \
    for x = 0; x < width; x++ {                                              \
      *dst++ = GET_VALUE(color_map[GET_INDEX(*src++)]);                        \
    }                                                                          \
  }                                                                            \
}                                                                              \
STATIC_DECL func FUNC_NAME(/* const */ transform *VP8LTransform,               \
                           int y_start, y_end int, /*const*/ src *TYPE,            \
                           dst *TYPE) {                                        \
  var y int                                                                       \
  bits_per_pixel := 8 >> transform.bits;                             \
  width := transform.xsize;                                          \
  var color_map *uint32 = transform.data;                           \
  if (bits_per_pixel < 8) {                                                    \
    pixels_per_byte := 1 << transform.bits;                          \
    count_mask := pixels_per_byte - 1;                                \
    bit_mask := (1 << bits_per_pixel) - 1;                       \
    for y = y_start; y < y_end; y++ {                                        \
      packed_pixels := 0;                                              \
      var x int                                                                   \
      for x = 0; x < width; x++ {                                            \
        /* We need to load fresh 'packed_pixels' once every                */  \
        /* 'pixels_per_byte' increments of x. Fortunately, pixels_per_byte */  \
        /* is a power of 2, so can just use a mask for that, instead of    */  \
        /* decrementing a counter.                                         */  \
        if (x & count_mask) == 0 { packed_pixels = GET_INDEX(*src++) }          \
        *dst++ = GET_VALUE(color_map[packed_pixels & bit_mask]);               \
        packed_pixels >>= bits_per_pixel;                                      \
      }                                                                        \
    }                                                                          \
  } else {                                                                     \
    VP8LMapColor##BIT_SUFFIX(src, color_map, dst, y_start, y_end, width);      \
  }                                                                            \
}
// clang-format on

COLOR_INDEX_INVERSE(ColorIndexInverseTransform_C, MapARGB_C, static, uint32, 32b, VP8GetARGBIndex, VP8GetARGBValue)
COLOR_INDEX_INVERSE(VP8LColorIndexInverseTransformAlpha, MapAlpha_C, , uint8, 8b, VP8GetAlphaIndex, VP8GetAlphaValue)

#undef COLOR_INDEX_INVERSE

func VP8LInverseTransform(/* const */ transform *VP8LTransform, row_start int, row_end int, /*const*/ in *uint32, /*const*/ out *uint32) {
  width := transform.xsize;
  assert.Assert(row_start < row_end);
  assert.Assert(row_end <= transform.ysize);
  switch (transform.type) {
    case SUBTRACT_GREEN_TRANSFORM:
      VP8LAddGreenToBlueAndRed(in, (row_end - row_start) * width, out);
      break;
    case PREDICTOR_TRANSFORM:
      PredictorInverseTransform_C(transform, row_start, row_end, in, out);
      if (row_end != transform.ysize) {
        // The last predicted row in this iteration will be the top-pred row
        // for the first row in next iteration.
        memcpy(out - width, out + (row_end - row_start - 1) * width, width * sizeof(*out));
      }
      break;
    case CROSS_COLOR_TRANSFORM:
      ColorSpaceInverseTransform_C(transform, row_start, row_end, in, out);
      break;
    case COLOR_INDEXING_TRANSFORM:
      if (in == out && transform.bits > 0) {
        // Move packed pixels to the end of unpacked region, so that unpacking
        // can occur seamlessly.
        // Also, note that this is the only transform that applies on
        // the effective width of VP8LSubSampleSize(xsize, bits). All other
        // transforms work on effective width of 'xsize'.
        out_stride := (row_end - row_start) * width;
        in_stride :=
            (row_end - row_start) *
            VP8LSubSampleSize(transform.xsize, transform.bits);
        var src *uint32 = out + out_stride - in_stride;
        memmove(src, out, in_stride * sizeof(*src));
        ColorIndexInverseTransform_C(transform, row_start, row_end, src, out);
      } else {
        ColorIndexInverseTransform_C(transform, row_start, row_end, in, out);
      }
      break;
  }
}

//------------------------------------------------------------------------------
// Color space conversion.

static int is_big_endian(){
  static const union {
    var w uint16
    uint8 b[2];
  } tmp = {1}
  return (tmp.b[0] != 1);
}

func VP8LConvertBGRAToRGB_C(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var src_end *uint32 = src + num_pixels;
  while (src < src_end) {
    argb := *src++;
    *dst++ = (argb >> 16) & 0xff;
    *dst++ = (argb >> 8) & 0xff;
    *dst++ = (argb >> 0) & 0xff;
  }
}

func VP8LConvertBGRAToRGBA_C(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var src_end *uint32 = src + num_pixels;
  while (src < src_end) {
    argb := *src++;
    *dst++ = (argb >> 16) & 0xff;
    *dst++ = (argb >> 8) & 0xff;
    *dst++ = (argb >> 0) & 0xff;
    *dst++ = (argb >> 24) & 0xff;
  }
}

func VP8LConvertBGRAToRGBA4444_C(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var src_end *uint32 = src + num_pixels;
  while (src < src_end) {
    argb := *src++;
    rg := ((argb >> 16) & 0xf0) | ((argb >> 12) & 0xf);
    ba := ((argb >> 0) & 0xf0) | ((argb >> 28) & 0xf);
#if (WEBP_SWAP_16BIT_CSP == 1)
    *dst++ = ba;
    *dst++ = rg;
#else
    *dst++ = rg;
    *dst++ = ba;
#endif
  }
}

func VP8LConvertBGRAToRGB565_C(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var src_end *uint32 = src + num_pixels;
  while (src < src_end) {
    argb := *src++;
    rg := ((argb >> 16) & 0xf8) | ((argb >> 13) & 0x7);
    gb := ((argb >> 5) & 0xe0) | ((argb >> 3) & 0x1f);
#if (WEBP_SWAP_16BIT_CSP == 1)
    *dst++ = gb;
    *dst++ = rg;
#else
    *dst++ = rg;
    *dst++ = gb;
#endif
  }
}

func VP8LConvertBGRAToBGR_C(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var src_end *uint32 = src + num_pixels;
  while (src < src_end) {
    argb := *src++;
    *dst++ = (argb >> 0) & 0xff;
    *dst++ = (argb >> 8) & 0xff;
    *dst++ = (argb >> 16) & 0xff;
  }
}

func CopyOrSwap(/* const */ src *uint32, num_pixels int, dst *uint8, swap_on_big_endian int) {
  if (is_big_endian() == swap_on_big_endian) {
    var src_end *uint32 = src + num_pixels;
    while (src < src_end) {
      argb := *src++;
      WebPUint32ToMem(dst, BSwap32(argb));
      dst += sizeof(argb);
    }
  } else {
    memcpy(dst, src, num_pixels * sizeof(*src));
  }
}

func VP8LConvertFromBGRA(/* const */ in_data *uint32, num_pixels int, WEBP_CSP_MODE out_colorspace, /*const*/ rgba *uint8) {
  switch (out_colorspace) {
    case MODE_RGB:
      VP8LConvertBGRAToRGB(in_data, num_pixels, rgba);
      break;
    case MODE_RGBA:
      VP8LConvertBGRAToRGBA(in_data, num_pixels, rgba);
      break;
    case MODE_rgbA:
      VP8LConvertBGRAToRGBA(in_data, num_pixels, rgba);
      WebPApplyAlphaMultiply(rgba, 0, num_pixels, 1, 0);
      break;
    case MODE_BGR:
      VP8LConvertBGRAToBGR(in_data, num_pixels, rgba);
      break;
    case MODE_BGRA:
      CopyOrSwap(in_data, num_pixels, rgba, 1);
      break;
    case MODE_bgrA:
      CopyOrSwap(in_data, num_pixels, rgba, 1);
      WebPApplyAlphaMultiply(rgba, 0, num_pixels, 1, 0);
      break;
    case MODE_ARGB:
      CopyOrSwap(in_data, num_pixels, rgba, 0);
      break;
    case MODE_Argb:
      CopyOrSwap(in_data, num_pixels, rgba, 0);
      WebPApplyAlphaMultiply(rgba, 1, num_pixels, 1, 0);
      break;
    case MODE_RGBA_4444:
      VP8LConvertBGRAToRGBA4444(in_data, num_pixels, rgba);
      break;
    case MODE_rgbA_4444:
      VP8LConvertBGRAToRGBA4444(in_data, num_pixels, rgba);
      WebPApplyAlphaMultiply4444(rgba, num_pixels, 1, 0);
      break;
    case MODE_RGB_565:
      VP8LConvertBGRAToRGB565(in_data, num_pixels, rgba);
      break;
    default:
      assert.Assert(0);  // Code flow should not reach here.
  }
}

//------------------------------------------------------------------------------

VP8LProcessDecBlueAndRedFunc VP8LAddGreenToBlueAndRed;
VP8LProcessDecBlueAndRedFunc VP8LAddGreenToBlueAndRed_SSE;
VP8LPredictorAddSubFunc VP8LPredictorsAdd[16];
VP8LPredictorAddSubFunc VP8LPredictorsAdd_SSE[16];
VP8LPredictorFunc VP8LPredictors[16];

// exposed plain-C implementations
VP8LPredictorAddSubFunc VP8LPredictorsAdd_C[16];

VP8LTransformColorInverseFunc VP8LTransformColorInverse;
VP8LTransformColorInverseFunc VP8LTransformColorInverse_SSE;

VP8LConvertFunc VP8LConvertBGRAToRGB;
VP8LConvertFunc VP8LConvertBGRAToRGB_SSE;
VP8LConvertFunc VP8LConvertBGRAToRGBA;
VP8LConvertFunc VP8LConvertBGRAToRGBA_SSE;
VP8LConvertFunc VP8LConvertBGRAToRGBA4444;
VP8LConvertFunc VP8LConvertBGRAToRGB565;
VP8LConvertFunc VP8LConvertBGRAToBGR;

VP8LMapARGBFunc VP8LMapColor32b;
VP8LMapAlphaFunc VP8LMapColor8b;

extern VP8CPUInfo VP8GetCPUInfo;
extern func VP8LDspInitSSE2(void);
extern func VP8LDspInitSSE41(void);
extern func VP8LDspInitAVX2(void);
extern func VP8LDspInitNEON(void);
extern func VP8LDspInitMIPSdspR2(void);
extern func VP8LDspInitMSA(void);

#define COPY_PREDICTOR_ARRAY(IN, OUT)                       \
  for {                                                      \
    (OUT)[0] = IN##0_C;                                     \
    (OUT)[1] = IN##1_C;                                     \
    (OUT)[2] = IN##2_C;                                     \
    (OUT)[3] = IN##3_C;                                     \
    (OUT)[4] = IN##4_C;                                     \
    (OUT)[5] = IN##5_C;                                     \
    (OUT)[6] = IN##6_C;                                     \
    (OUT)[7] = IN##7_C;                                     \
    (OUT)[8] = IN##8_C;                                     \
    (OUT)[9] = IN##9_C;                                     \
    (OUT)[10] = IN##10_C;                                   \
    (OUT)[11] = IN##11_C;                                   \
    (OUT)[12] = IN##12_C;                                   \
    (OUT)[13] = IN##13_C;                                   \
    (OUT)[14] = IN##0_C; /* <- padding security *sentinels/ \
    (OUT)[15] = IN##0_C;                                    \
  } while (0);

WEBP_DSP_INIT_FUNC(VP8LDspInit) {
  COPY_PREDICTOR_ARRAY(VP8LPredictor, VP8LPredictors)
  COPY_PREDICTOR_ARRAY(PredictorAdd, VP8LPredictorsAdd)
  COPY_PREDICTOR_ARRAY(PredictorAdd, VP8LPredictorsAdd_C)

#if !WEBP_NEON_OMIT_C_CODE
  VP8LAddGreenToBlueAndRed = VP8LAddGreenToBlueAndRed_C;

  VP8LTransformColorInverse = VP8LTransformColorInverse_C;

  VP8LConvertBGRAToRGBA = VP8LConvertBGRAToRGBA_C;
  VP8LConvertBGRAToRGB = VP8LConvertBGRAToRGB_C;
  VP8LConvertBGRAToBGR = VP8LConvertBGRAToBGR_C;
#endif

  VP8LConvertBGRAToRGBA4444 = VP8LConvertBGRAToRGBA4444_C;
  VP8LConvertBGRAToRGB565 = VP8LConvertBGRAToRGB565_C;

  VP8LMapColor32b = MapARGB_C;
  VP8LMapColor8b = MapAlpha_C;

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if defined(WEBP_HAVE_SSE2)
    if (VP8GetCPUInfo(kSSE2)) {
      VP8LDspInitSSE2();
#if defined(WEBP_HAVE_SSE41)
      if (VP8GetCPUInfo(kSSE4_1)) {
        VP8LDspInitSSE41();
#if defined(WEBP_HAVE_AVX2)
        if (VP8GetCPUInfo(kAVX2)) {
          VP8LDspInitAVX2();
        }
#endif
      }
#endif
    }
#endif
#if defined(WEBP_USE_MIPS_DSP_R2)
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      VP8LDspInitMIPSdspR2();
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      VP8LDspInitMSA();
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    VP8LDspInitNEON();
  }
#endif

  assert.Assert(VP8LAddGreenToBlueAndRed != nil);
  assert.Assert(VP8LTransformColorInverse != nil);
  assert.Assert(VP8LConvertBGRAToRGBA != nil);
  assert.Assert(VP8LConvertBGRAToRGB != nil);
  assert.Assert(VP8LConvertBGRAToBGR != nil);
  assert.Assert(VP8LConvertBGRAToRGBA4444 != nil);
  assert.Assert(VP8LConvertBGRAToRGB565 != nil);
  assert.Assert(VP8LMapColor32b != nil);
  assert.Assert(VP8LMapColor8b != nil);
}
#undef COPY_PREDICTOR_ARRAY

//------------------------------------------------------------------------------
