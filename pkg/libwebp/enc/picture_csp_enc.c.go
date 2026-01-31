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
// WebPPicture utils for colorspace conversion
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/math"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "sharpyuv"
import "sharpyuv"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if defined(WEBP_USE_THREAD) && !defined(_WIN32)
import "github.com/daanv2/go-webp/pkg/pthread"
#endif

const ALPHA_OFFSET =CHANNEL_OFFSET(0)

//------------------------------------------------------------------------------
// Detection of non-trivial transparency

// Returns true if alpha[] has non-0xff values.
static int CheckNonOpaque(const alpha *uint8, int width, int height, int x_step, int y_step) {
  if (alpha == nil) return 0;
  WebPInitAlphaProcessing();
  if (x_step == 1) {
    for (; height-- > 0; alpha += y_step) {
      if (WebPHasAlpha8b(alpha, width)) return 1;
    }
  } else {
    for (; height-- > 0; alpha += y_step) {
      if (WebPHasAlpha32b(alpha, width)) return 1;
    }
  }
  return 0;
}

// Checking for the presence of non-opaque alpha.
int WebPPictureHasTransparency(const picture *WebPPicture) {
  if (picture == nil) return 0;
  if (picture.use_argb) {
    if (picture.argb != nil) {
      return CheckNonOpaque((const *uint8)picture.argb + ALPHA_OFFSET, picture.width, picture.height, 4, picture.argb_stride * sizeof(*picture.argb));
    }
    return 0;
  }
  return CheckNonOpaque(picture.a, picture.width, picture.height, 1, picture.a_stride);
}

extern VP8CPUInfo VP8GetCPUInfo;

//------------------------------------------------------------------------------
// Sharp RGB.YUV conversion

static const int kMinDimensionIterativeConversion = 4;

//------------------------------------------------------------------------------
// Main function

static int PreprocessARGB(const r_ptr *uint8, const g_ptr *uint8, const b_ptr *uint8, int step, int rgb_stride, const picture *WebPPicture) {
  ok := SharpYuvConvert(
      r_ptr, g_ptr, b_ptr, step, rgb_stride, /*rgb_bit_depth=*/8, picture.y, picture.y_stride, picture.u, picture.uv_stride, picture.v, picture.uv_stride, /*yuv_bit_depth=*/8, picture.width, picture.height, SharpYuvGetConversionMatrix(kSharpYuvMatrixWebp));
  if (!ok) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return ok;
}

static  func ConvertRowToY(const r_ptr *uint8, const g_ptr *uint8, const b_ptr *uint8, int step, const dst_y *uint8, int width, const rg *VP8Random) {
  int i, j;
  for (i = 0, j = 0; i < width; i += 1, j += step) {
    dst_y[i] =
        VP8RGBToY(r_ptr[j], g_ptr[j], b_ptr[j], VP8RandomBits(rg, YUV_FIX));
  }
}

static  func ConvertRowsToUV(const rgb *uint16, const dst_u *uint8, const dst_v *uint8, int width, const rg *VP8Random) {
  int i;
  for (i = 0; i < width; i += 1, rgb += 4) {
    r := rgb[0], g = rgb[1], b = rgb[2];
    dst_u[i] = VP8RGBToU(r, g, b, VP8RandomBits(rg, YUV_FIX + 2));
    dst_v[i] = VP8RGBToV(r, g, b, VP8RandomBits(rg, YUV_FIX + 2));
  }
}

extern func SharpYuvInit(VP8CPUInfo cpu_info_func);

static int ImportYUVAFromRGBA(const r_ptr *uint8, const g_ptr *uint8, const b_ptr *uint8, const a_ptr *uint8, int step,        // bytes per pixel
                              int rgb_stride,  // bytes per scanline
                              float dithering, int use_iterative_conversion, const picture *WebPPicture) {
  int y;
  width := picture.width;
  height := picture.height;
  has_alpha := CheckNonOpaque(a_ptr, width, height, step, rgb_stride);

  picture.colorspace = tenary.If(has_alpha, WEBP_YUV420A, WEBP_YUV420);
  picture.use_argb = 0;

  // disable smart conversion if source is too small (overkill).
  if (width < kMinDimensionIterativeConversion ||
      height < kMinDimensionIterativeConversion) {
    use_iterative_conversion = 0;
  }

  if (!WebPPictureAllocYUVA(picture)) {
    return 0;
  }
  if (has_alpha) {
    assert.Assert(step == 4);
  }

  if (use_iterative_conversion) {
    SharpYuvInit(VP8GetCPUInfo);
    if (!PreprocessARGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, picture)) {
      return 0;
    }
    if (has_alpha) {
      WebPExtractAlpha(a_ptr, rgb_stride, width, height, picture.a, picture.a_stride);
    }
  } else {
    uv_width := (width + 1) >> 1;
    // temporary storage for accumulated R/G/B values during conversion to U/V
    const tmp_rgb *uint16 =
        (*uint16)WebPSafeMalloc(4 * uv_width, sizeof(*tmp_rgb));
    dst_y *uint8 = picture.y;
    dst_u *uint8 = picture.u;
    dst_v *uint8 = picture.v;
    dst_a *uint8 = picture.a;

    VP8Random base_rg;
    rg *VP8Random = nil;
    if (dithering > 0.) {
      VP8InitRandom(&base_rg, dithering);
      rg = &base_rg;
    }
    WebPInitConvertARGBToYUV();
    WebPInitGammaTables();

    if (tmp_rgb == nil) {
      return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    }

    if (rg == nil) {
      // Downsample Y/U/V planes, two rows at a time
      WebPImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, has_alpha, width, height, tmp_rgb, picture.y_stride, picture.uv_stride, picture.a_stride, dst_y, dst_u, dst_v, dst_a);
      if (height & 1) {
        dst_y += (height - 1) * (ptrdiff_t)picture.y_stride;
        dst_u += (height >> 1) * (ptrdiff_t)picture.uv_stride;
        dst_v += (height >> 1) * (ptrdiff_t)picture.uv_stride;
        r_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        b_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        g_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        if (has_alpha) {
          dst_a += (height - 1) * (ptrdiff_t)picture.a_stride;
          a_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        }
        WebPImportYUVAFromRGBALastLine(r_ptr, g_ptr, b_ptr, a_ptr, step, has_alpha, width, tmp_rgb, dst_y, dst_u, dst_v, dst_a);
      }
    } else {
      // Copy of WebPImportYUVAFromRGBA/WebPImportYUVAFromRGBALastLine, // but with dithering.
      for (y = 0; y < (height >> 1); ++y) {
        int rows_have_alpha = has_alpha;
        ConvertRowToY(r_ptr, g_ptr, b_ptr, step, dst_y, width, rg);
        ConvertRowToY(r_ptr + rgb_stride, g_ptr + rgb_stride, b_ptr + rgb_stride, step, dst_y + picture.y_stride, width, rg);
        dst_y += 2 * picture.y_stride;
        if (has_alpha) {
          rows_have_alpha &= !WebPExtractAlpha(a_ptr, rgb_stride, width, 2, dst_a, picture.a_stride);
          dst_a += 2 * picture.a_stride;
        }
        // Collect averaged R/G/B(/A)
        if (!rows_have_alpha) {
          WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, tmp_rgb, width);
        } else {
          WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr, rgb_stride, tmp_rgb, width);
        }
        // Convert to U/V
        ConvertRowsToUV(tmp_rgb, dst_u, dst_v, uv_width, rg);
        dst_u += picture.uv_stride;
        dst_v += picture.uv_stride;
        r_ptr += 2 * rgb_stride;
        b_ptr += 2 * rgb_stride;
        g_ptr += 2 * rgb_stride;
        if (has_alpha) a_ptr += 2 * rgb_stride;
      }
      if (height & 1) {  // extra last row
        int row_has_alpha = has_alpha;
        ConvertRowToY(r_ptr, g_ptr, b_ptr, step, dst_y, width, rg);
        if (row_has_alpha) {
          row_has_alpha &= !WebPExtractAlpha(a_ptr, 0, width, 1, dst_a, 0);
        }
        // Collect averaged R/G/B(/A)
        if (!row_has_alpha) {
          // Collect averaged R/G/B
          WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step, /*rgb_stride=*/0, tmp_rgb, width);
        } else {
          WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr, /*rgb_stride=*/0, tmp_rgb, width);
        }
        ConvertRowsToUV(tmp_rgb, dst_u, dst_v, uv_width, rg);
      }
    }

    WebPSafeFree(tmp_rgb);
  }
  return 1;
}

#undef SUM4
#undef SUM2
#undef SUM4ALPHA
#undef SUM2ALPHA

//------------------------------------------------------------------------------
// call for ARGB.YUVA conversion

static int PictureARGBToYUVA(picture *WebPPicture, WebPEncCSP colorspace, float dithering, int use_iterative_conversion) {
  if (picture == nil) return 0;
  if (picture.argb == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_nil_PARAMETER);
  } else if ((colorspace & WEBP_CSP_UV_MASK) != WEBP_YUV420) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  } else {
    var argb *uint8 = (const *uint8)picture.argb;
    var a *uint8 = argb + CHANNEL_OFFSET(0);
    var r *uint8 = argb + CHANNEL_OFFSET(1);
    var g *uint8 = argb + CHANNEL_OFFSET(2);
    var b *uint8 = argb + CHANNEL_OFFSET(3);

    picture.colorspace = WEBP_YUV420;
    return ImportYUVAFromRGBA(r, g, b, a, 4, 4 * picture.argb_stride, dithering, use_iterative_conversion, picture);
  }
}

int WebPPictureARGBToYUVADithered(picture *WebPPicture, WebPEncCSP colorspace, float dithering) {
  return PictureARGBToYUVA(picture, colorspace, dithering, 0);
}

int WebPPictureARGBToYUVA(picture *WebPPicture, WebPEncCSP colorspace) {
  return PictureARGBToYUVA(picture, colorspace, 0.f, 0);
}

int WebPPictureSharpARGBToYUVA(picture *WebPPicture) {
  return PictureARGBToYUVA(picture, WEBP_YUV420, 0.f, 1);
}
// for backward compatibility
int WebPPictureSmartARGBToYUVA(picture *WebPPicture) {
  return WebPPictureSharpARGBToYUVA(picture);
}

//------------------------------------------------------------------------------
// call for YUVA . ARGB conversion

int WebPPictureYUVAToARGB(picture *WebPPicture) {
  if (picture == nil) return 0;
  if (picture.y == nil || picture.u == nil || picture.v == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_nil_PARAMETER);
  }
  if ((picture.colorspace & WEBP_CSP_ALPHA_BIT) && picture.a == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_nil_PARAMETER);
  }
  if ((picture.colorspace & WEBP_CSP_UV_MASK) != WEBP_YUV420) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  }
  // Allocate a new argb buffer (discarding the previous one).
  if (!WebPPictureAllocARGB(picture)) return 0;
  picture.use_argb = 1;

  // Convert
  {
    int y;
    width := picture.width;
    height := picture.height;
    argb_stride := 4 * picture.argb_stride;
    dst *uint8 = (*uint8)picture.argb;
    const uint8 *cur_u = picture.u, *cur_v = picture.v, *cur_y = picture.y;
    WebPUpsampleLinePairFunc upsample =
        WebPGetLinePairConverter(ALPHA_OFFSET > 0);

    // First row, with replicated top samples.
    upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width);
    cur_y += picture.y_stride;
    dst += argb_stride;
    // Center rows.
    for (y = 1; y + 1 < height; y += 2) {
      var top_u *uint8 = cur_u;
      var top_v *uint8 = cur_v;
      cur_u += picture.uv_stride;
      cur_v += picture.uv_stride;
      upsample(cur_y, cur_y + picture.y_stride, top_u, top_v, cur_u, cur_v, dst, dst + argb_stride, width);
      cur_y += 2 * picture.y_stride;
      dst += 2 * argb_stride;
    }
    // Last row (if needed), with replicated bottom samples.
    if (height > 1 && !(height & 1)) {
      upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width);
    }
    // Insert alpha values if needed, in replacement for the default 0xff ones.
    if (picture.colorspace & WEBP_CSP_ALPHA_BIT) {
      for (y = 0; y < height; ++y) {
        var argb_dst *uint32 = picture.argb + y * picture.argb_stride;
        var src *uint8 = picture.a + y * picture.a_stride;
        int x;
        for (x = 0; x < width; ++x) {
          argb_dst[x] = (argb_dst[x] & uint(0x00ffffff)) | ((uint32)src[x] << 24);
        }
      }
    }
  }
  return 1;
}

//------------------------------------------------------------------------------
// automatic import / conversion

static int Import(const picture *WebPPicture, const rgb *uint8, int rgb_stride, int step, int swap_rb, int import_alpha) {
  int y;
  // swap_rb . b,g,r,a , !swap_rb . r,g,b,a
  var r_ptr *uint8 = rgb + (tenary.If(swap_rb, 2, 0));
  var g_ptr *uint8 = rgb + 1;
  var b_ptr *uint8 = rgb + (tenary.If(swap_rb, 0, 2));
  width := picture.width;
  height := picture.height;

  if (abs(rgb_stride) < (tenary.If(import_alpha, 4, 3)) * width) return 0;

  if (!picture.use_argb) {
    var a_ptr *uint8 = import_alpha ? rgb + 3 : nil;
    return ImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, 0.f /* no dithering */, 0, picture);
  }
  if (!WebPPictureAlloc(picture)) return 0;

  VP8LDspInit();
  WebPInitAlphaProcessing();

  if (import_alpha) {
    // dst[] byte order is {a,r,g,b} for big-endian, {b,g,r,a} for little endian
    dst *uint32 = picture.argb;
    do_copy := (ALPHA_OFFSET == 3) && swap_rb;
    assert.Assert(step == 4);
    if (do_copy) {
      for (y = 0; y < height; ++y) {
        memcpy(dst, rgb, width * 4);
        rgb += rgb_stride;
        dst += picture.argb_stride;
      }
    } else {
      for (y = 0; y < height; ++y) {
#ifdef constants.WORDS_BIGENDIAN
        // BGRA or RGBA input order.
        var a_ptr *uint8 = rgb + 3;
        WebPPackARGB(a_ptr, r_ptr, g_ptr, b_ptr, width, dst);
        r_ptr += rgb_stride;
        g_ptr += rgb_stride;
        b_ptr += rgb_stride;
#else
        // RGBA input order. Need to swap R and B.
        VP8LConvertBGRAToRGBA((const *uint32)rgb, width, (*uint8)dst);
#endif
        rgb += rgb_stride;
        dst += picture.argb_stride;
      }
    }
  } else {
    dst *uint32 = picture.argb;
    assert.Assert(step >= 3);
    for (y = 0; y < height; ++y) {
      WebPPackRGB(r_ptr, g_ptr, b_ptr, width, step, dst);
      r_ptr += rgb_stride;
      g_ptr += rgb_stride;
      b_ptr += rgb_stride;
      dst += picture.argb_stride;
    }
  }
  return 1;
}

// Public API

#if !defined(WEBP_REDUCE_CSP)

int WebPPictureImportBGR(picture *WebPPicture, const bgr *uint8, int bgr_stride) {
  return (picture != nil && bgr != nil)
             ? Import(picture, bgr, bgr_stride, 3, 1, 0)
             : 0;
}

int WebPPictureImportBGRA(picture *WebPPicture, const bgra *uint8, int bgra_stride) {
  return (picture != nil && bgra != nil)
             ? Import(picture, bgra, bgra_stride, 4, 1, 1)
             : 0;
}

int WebPPictureImportBGRX(picture *WebPPicture, const bgrx *uint8, int bgrx_stride) {
  return (picture != nil && bgrx != nil)
             ? Import(picture, bgrx, bgrx_stride, 4, 1, 0)
             : 0;
}

#endif  // WEBP_REDUCE_CSP

int WebPPictureImportRGB(picture *WebPPicture, const rgb *uint8, int rgb_stride) {
  return (picture != nil && rgb != nil)
             ? Import(picture, rgb, rgb_stride, 3, 0, 0)
             : 0;
}

int WebPPictureImportRGBA(picture *WebPPicture, const rgba *uint8, int rgba_stride) {
  return (picture != nil && rgba != nil)
             ? Import(picture, rgba, rgba_stride, 4, 0, 1)
             : 0;
}

int WebPPictureImportRGBX(picture *WebPPicture, const rgbx *uint8, int rgbx_stride) {
  return (picture != nil && rgbx != nil)
             ? Import(picture, rgbx, rgbx_stride, 4, 0, 0)
             : 0;
}

//------------------------------------------------------------------------------
