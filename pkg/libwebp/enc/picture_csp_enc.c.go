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
// picture.Picture utils for colorspace conversion
//
// Author: Skal (pascal.massimino@gmail.com)

const ALPHA_OFFSET =CHANNEL_OFFSET(0)

//------------------------------------------------------------------------------
// Detection of non-trivial transparency

// Returns true if alpha[] has non-0xff values.
func CheckNonOpaque(/* const */ alpha *uint8, width, height int, x_step int, y_step int) int {
  if alpha == nil { return 0  }
  WebPInitAlphaProcessing();
  if (x_step == 1) {
    for ; height-- > 0; alpha += y_step {
      if WebPHasAlpha8b(alpha, width) { return 1  }
    }
  } else {
    for ; height-- > 0; alpha += y_step {
      if WebPHasAlpha32b(alpha, width) { return 1  }
    }
  }
  return 0;
}

// Checking for the presence of non-opaque alpha.
func WebPPictureHasTransparency(/* const */ picture *picture.Picture) int {
  if picture == nil { return 0  }
  if (picture.UseARGB) {
    if (picture.ARGB != nil) {
      return CheckNonOpaque((/* const */ *uint8)picture.ARGB + ALPHA_OFFSET, picture.Width, picture.Height, 4, picture.ARGBStride * sizeof(*picture.ARGB));
    }
    return 0;
  }
  return CheckNonOpaque(picture.A, picture.Width, picture.Height, 1, picture.AStride);
}

extern VP8CPUInfo VP8GetCPUInfo;

//------------------------------------------------------------------------------
// Sharp RGB.YUV conversion

static const kMinDimensionIterativeConversion := 4;

//------------------------------------------------------------------------------
// Main function

func PreprocessARGB(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, step int, rgb_stride int, /*const*/ picture *picture.Picture) int {
  ok := SharpYuvConvert(
      r_ptr, g_ptr, b_ptr, step, rgb_stride, /*rgb_bit_depth=*/8, picture.Y, picture.YStride, picture.U, picture.UVStride, picture.V, picture.UVStride, /*yuv_bit_depth=*/8, picture.Width, picture.Height, SharpYuvGetConversionMatrix(kSharpYuvMatrixWebp));
  if (!ok) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_OUT_OF_MEMORY)
  }
  return ok;
}

func ConvertRowToY(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, step int, /*const*/ dst_y *uint8, width int, /*const*/ rg *VP8Random) {
  int i, j;
  for i = 0, j = 0; i < width; i += 1, j += step {
    dst_y[i] =
        VP8RGBToY(r_ptr[j], g_ptr[j], b_ptr[j], VP8RandomBits(rg, YUV_FIX));
  }
}

func ConvertRowsToUV(/* const */ rgb *uint16, /*const*/ dst_u *uint8, /*const*/ dst_v *uint8, width int, /*const*/ rg *VP8Random) {
  var i int
  for i = 0; i < width; i += 1, rgb += 4 {
    r := rgb[0], g = rgb[1], b = rgb[2];
    dst_u[i] = VP8RGBToU(r, g, b, VP8RandomBits(rg, YUV_FIX + 2));
    dst_v[i] = VP8RGBToV(r, g, b, VP8RandomBits(rg, YUV_FIX + 2));
  }
}

extern func SharpYuvInit(VP8CPUInfo cpu_info_func);

static int ImportYUVAFromRGBA(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, /*const*/ a_ptr *uint8, step int,        // bytes per pixel
                              int rgb_stride,  // bytes per scanline
                              float64 dithering, use_iterative_conversion int, /*const*/ picture *picture.Picture) {
  var y int
  width := picture.Width;
  height := picture.Height;
  has_alpha := CheckNonOpaque(a_ptr, width, height, step, rgb_stride);

  picture.ColorSpace = tenary.If(has_alpha, colorspace.WEBP_YUV420A, colorspace.WEBP_YUV420);
  picture.UseARGB = false;

  // disable smart conversion if source is too small (overkill).
  if (width < kMinDimensionIterativeConversion ||
      height < kMinDimensionIterativeConversion) {
    use_iterative_conversion = 0;
  }

  if (!picture.WebPPictureAllocYUVA(picture)) {
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
      WebPExtractAlpha(a_ptr, rgb_stride, width, height, picture.A, picture.AStride);
    }
  } else {
    uv_width := (width + 1) >> 1;
    // temporary storage for accumulated R/G/B values during conversion to U/V
    // var tmp_rgb *uint16 = (*uint16)WebPSafeMalloc(4 * uv_width, sizeof(*tmp_rgb));
	tmp_rgb = make([]uint16, 4 * uv_width)

	var dst_y *uint8 = picture.Y;
    var dst_u *uint8 = picture.U;
    var dst_v *uint8 = picture.V;
    var dst_a *uint8 = picture.A;

     var base_rg VP8Random
    rg *VP8Random = nil;
    if (dithering > 0.) {
      VP8InitRandom(&base_rg, dithering);
      rg = &base_rg;
    }
    WebPInitConvertARGBToYUV();
    WebPInitGammaTables();

    // if (tmp_rgb == nil) {
    //   return picture.SetEncodingError(picture.VP8_ENC_ERROR_OUT_OF_MEMORY)
    // }

    if (rg == nil) {
      // Downsample Y/U/V planes, two rows at a time
      WebPImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, has_alpha, width, height, tmp_rgb, picture.YStride, picture.UVStride, picture.AStride, dst_y, dst_u, dst_v, dst_a);
      if (height & 1) {
        dst_y += (height - 1) * (ptrdiff_t)picture.YStride;
        dst_u += (height >> 1) * (ptrdiff_t)picture.UVStride;
        dst_v += (height >> 1) * (ptrdiff_t)picture.UVStride;
        r_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        b_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        g_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        if (has_alpha) {
          dst_a += (height - 1) * (ptrdiff_t)picture.AStride;
          a_ptr += (height - 1) * (ptrdiff_t)rgb_stride;
        }
        WebPImportYUVAFromRGBALastLine(r_ptr, g_ptr, b_ptr, a_ptr, step, has_alpha, width, tmp_rgb, dst_y, dst_u, dst_v, dst_a);
      }
    } else {
      // Copy of WebPImportYUVAFromRGBA/WebPImportYUVAFromRGBALastLine, // but with dithering.
      for y = 0; y < (height >> 1); y++ {
        rows_have_alpha := has_alpha;
        ConvertRowToY(r_ptr, g_ptr, b_ptr, step, dst_y, width, rg);
        ConvertRowToY(r_ptr + rgb_stride, g_ptr + rgb_stride, b_ptr + rgb_stride, step, dst_y + picture.YStride, width, rg);
        dst_y += 2 * picture.YStride;
        if (has_alpha) {
          rows_have_alpha &= !WebPExtractAlpha(a_ptr, rgb_stride, width, 2, dst_a, picture.AStride);
          dst_a += 2 * picture.AStride;
        }
        // Collect averaged R/G/B(/A)
        if (!rows_have_alpha) {
          WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, tmp_rgb, width);
        } else {
          WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr, rgb_stride, tmp_rgb, width);
        }
        // Convert to U/V
        ConvertRowsToUV(tmp_rgb, dst_u, dst_v, uv_width, rg);
        dst_u += picture.UVStride;
        dst_v += picture.UVStride;
        r_ptr += 2 * rgb_stride;
        b_ptr += 2 * rgb_stride;
        g_ptr += 2 * rgb_stride;
        if has_alpha { a_ptr += 2 * rgb_stride }
      }
      if (height & 1) {  // extra last row
        row_has_alpha := has_alpha;
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

  }
  return 1;
}

#undef SUM4
#undef SUM2
#undef SUM4ALPHA
#undef SUM2ALPHA

//------------------------------------------------------------------------------
// call for ARGB.YUVA conversion

func PictureARGBToYUVA(picture *picture.Picture, colorspace.CSP colorspace, float64 dithering, use_iterative_conversion int) int {
  if picture == nil { return 0  }
  if (picture.ARGB == nil) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_nil_PARAMETER)
  } else if ((colorspace & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_INVALID_CONFIGURATION)
  } else {
    var argb *uint8 = (/* const */ *uint8)picture.ARGB;
    var a *uint8 = argb + CHANNEL_OFFSET(0);
    var r *uint8 = argb + CHANNEL_OFFSET(1);
    var g *uint8 = argb + CHANNEL_OFFSET(2);
    var b *uint8 = argb + CHANNEL_OFFSET(3);

    picture.ColorSpace = colorspace.WEBP_YUV420;
    return ImportYUVAFromRGBA(r, g, b, a, 4, 4 * picture.ARGBStride, dithering, use_iterative_conversion, picture);
  }
}

func WebPPictureARGBToYUVADithered(picture *picture.Picture, colorspace.CSP colorspace, float64 dithering) int {
  return PictureARGBToYUVA(picture, colorspace, dithering, 0);
}

func WebPPictureARGBToYUVA(picture *picture.Picture, colorspace.CSP colorspace) int {
  return PictureARGBToYUVA(picture, colorspace, 0.0, 0);
}

func WebPPictureSharpARGBToYUVA(picture *picture.Picture) int {
  return PictureARGBToYUVA(picture, colorspace.WEBP_YUV420, 0.0, 1);
}
// for backward compatibility
func WebPPictureSmartARGBToYUVA(picture *picture.Picture) int {
  return picture.WebPPictureSharpARGBToYUVA(picture);
}

//------------------------------------------------------------------------------
// call for YUVA . ARGB conversion

func WebPPictureYUVAToARGB(picture *picture.Picture) int {
  if picture == nil { return 0  }
  if (picture.Y == nil || picture.U == nil || picture.V == nil) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_nil_PARAMETER)
  }
  if ((picture.ColorSpace & colorspace.WEBP_CSP_ALPHA_BIT) && picture.A == nil) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_nil_PARAMETER)
  }
  if ((picture.ColorSpace & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_INVALID_CONFIGURATION)
  }
  // Allocate a new argb buffer (discarding the previous one).
  if !picture.WebPPictureAllocARGB(picture) { return 0  }
  picture.UseARGB = true;

  // Convert
  {
    var y int
    width := picture.Width;
    height := picture.Height;
    argb_stride := 4 * picture.ARGBStride;
    dst *uint8 = (*uint8)picture.ARGB;
    *cur_u = picture.U, *cur_v = picture.V, *cur_y := picture.Y;
    WebPUpsampleLinePairFunc upsample =
        WebPGetLinePairConverter(ALPHA_OFFSET > 0);

    // First row, with replicated top samples.
    upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width);
    cur_y += picture.YStride;
    dst += argb_stride;
    // Center rows.
    for y = 1; y + 1 < height; y += 2 {
      var top_u *uint8 = cur_u;
      var top_v *uint8 = cur_v;
      cur_u += picture.UVStride;
      cur_v += picture.UVStride;
      upsample(cur_y, cur_y + picture.YStride, top_u, top_v, cur_u, cur_v, dst, dst + argb_stride, width);
      cur_y += 2 * picture.YStride;
      dst += 2 * argb_stride;
    }
    // Last row (if needed), with replicated bottom samples.
    if (height > 1 && !(height & 1)) {
      upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width);
    }
    // Insert alpha values if needed, in replacement for the default 0xff ones.
    if (picture.ColorSpace & colorspace.WEBP_CSP_ALPHA_BIT) {
      for y = 0; y < height; y++ {
        var argb_dst *uint32 = picture.ARGB + y * picture.ARGBStride;
        var src *uint8 = picture.A + y * picture.AStride;
        var x int
        for x = 0; x < width; x++ {
          argb_dst[x] = (argb_dst[x] & uint(0x00ffffff)) | ((uint32)src[x] << 24);
        }
      }
    }
  }
  return 1;
}

//------------------------------------------------------------------------------
// automatic import / conversion

func Import(/* const */ picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int, step int, swap_rb int, import_alpha int) int {
  var y int
  // swap_rb . b,g,r,a , !swap_rb . r,g,b,a
  var r_ptr *uint8 = rgb + (tenary.If(swap_rb, 2, 0));
  var g_ptr *uint8 = rgb + 1;
  var b_ptr *uint8 = rgb + (tenary.If(swap_rb, 0, 2));
  width := picture.Width;
  height := picture.Height;

  if abs(rgb_stride) < (tenary.If(import_alpha, 4, 3)) * width { return 0  }

  if (!picture.UseARGB) {
    var a_ptr *uint8 = import_alpha ? rgb + 3 : nil;
    return ImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, 0.0 /* no dithering */, 0, picture);
  }
  if !picture.WebPPictureAlloc(picture) { return 0  }

  VP8LDspInit();
  WebPInitAlphaProcessing();

  if (import_alpha) {
    // dst[] byte order is {a,r,g,b} for big-endian, {b,g,r,a} for little endian
    dst *uint32 = picture.ARGB;
    do_copy := (ALPHA_OFFSET == 3) && swap_rb;
    assert.Assert(step == 4);
    if (do_copy) {
      for y = 0; y < height; y++ {
        stdlib.MemCpy(dst, rgb, width * 4);
        rgb += rgb_stride;
        dst += picture.ARGBStride;
      }
    } else {
      for y = 0; y < height; y++ {
#ifdef constants.WORDS_BIGENDIAN
        // BGRA or RGBA input order.
        var a_ptr *uint8 = rgb + 3;
        WebPPackARGB(a_ptr, r_ptr, g_ptr, b_ptr, width, dst);
        r_ptr += rgb_stride;
        g_ptr += rgb_stride;
        b_ptr += rgb_stride;
#else
        // RGBA input order. Need to swap R and B.
        VP8LConvertBGRAToRGBA((/* const */ *uint32)rgb, width, (*uint8)dst);
#endif
        rgb += rgb_stride;
        dst += picture.ARGBStride;
      }
    }
  } else {
    dst *uint32 = picture.ARGB;
    assert.Assert(step >= 3);
    for y = 0; y < height; y++ {
      WebPPackRGB(r_ptr, g_ptr, b_ptr, width, step, dst);
      r_ptr += rgb_stride;
      g_ptr += rgb_stride;
      b_ptr += rgb_stride;
      dst += picture.ARGBStride;
    }
  }
  return 1;
}

// Public API

#if !defined(WEBP_REDUCE_CSP)

func WebPPictureImportBGR(picture *picture.Picture, /*const*/ bgr *uint8, bgr_stride int) int {
  return (picture != nil && bgr != nil)
             ? Import(picture, bgr, bgr_stride, 3, 1, 0)
             : 0;
}

func WebPPictureImportBGRA(picture *picture.Picture, /*const*/ bgra *uint8, bgra_stride int) int {
  return (picture != nil && bgra != nil)
             ? Import(picture, bgra, bgra_stride, 4, 1, 1)
             : 0;
}

func WebPPictureImportBGRX(picture *picture.Picture, /*const*/ bgrx *uint8, bgrx_stride int) int {
  return (picture != nil && bgrx != nil)
             ? Import(picture, bgrx, bgrx_stride, 4, 1, 0)
             : 0;
}

#endif  // WEBP_REDUCE_CSP

func WebPPictureImportRGB(picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int) int {
  return (picture != nil && rgb != nil)
             ? Import(picture, rgb, rgb_stride, 3, 0, 0)
             : 0;
}

func WebPPictureImportRGBA(picture *picture.Picture, /*const*/ rgba *uint8, rgba_stride int) int {
  return (picture != nil && rgba != nil)
             ? Import(picture, rgba, rgba_stride, 4, 0, 1)
             : 0;
}

func WebPPictureImportRGBX(picture *picture.Picture, /*const*/ rgbx *uint8, rgbx_stride int) int {
  return (picture != nil && rgbx != nil)
             ? Import(picture, rgbx, rgbx_stride, 4, 0, 0)
             : 0;
}

//------------------------------------------------------------------------------
