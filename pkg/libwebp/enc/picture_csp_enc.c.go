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
const MinDimensionIterativeConversion = 4

//------------------------------------------------------------------------------
// Main function

func PreprocessARGB(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, step int, rgb_stride int, /*const*/ picture *picture.Picture) int {
  ok := SharpYuvConvert(
      r_ptr, g_ptr, b_ptr, step, rgb_stride, /*rgb_bit_depth=*/8, picture.Y, picture.YStride, picture.U, picture.UVStride, picture.V, picture.UVStride, /*yuv_bit_depth=*/8, picture.Width, picture.Height, SharpYuvGetConversionMatrix(kSharpYuvMatrixWebp))
  if (!ok) {
    return picture.SetEncodingError(picture.ENC_ERROR_OUT_OF_MEMORY)
  }
  return ok
}

func ConvertRowToY(/* const */ r_ptr *uint8, /*const*/ g_ptr *uint8, /*const*/ b_ptr *uint8, step int, /*const*/ dst_y *uint8, width int, /*const*/ rg *VP8Random) {
  var i, j int = 0, 0
  for ; i < width; {
    dst_y[i] = VP8RGBToY(r_ptr[j], g_ptr[j], b_ptr[j], VP8RandomBits(rg, YUV_FIX))

		i += 1
		j += step
  }
}

func ConvertRowsToUV(/* const */ rgb *uint16, /*const*/ dst_u *uint8, /*const*/ dst_v *uint8, width int, /*const*/ rg *VP8Random) {
  var i int
  for i = 0; i < width; i += 1, rgb += 4 {
    r := rgb[0], g = rgb[1], b = rgb[2]
    dst_u[i] = VP8RGBToU(r, g, b, VP8RandomBits(rg, YUV_FIX + 2))
    dst_v[i] = VP8RGBToV(r, g, b, VP8RandomBits(rg, YUV_FIX + 2))
  }
}

extern func SharpYuvInit(VP8CPUInfo cpu_info_func)



//------------------------------------------------------------------------------
// call for ARGB.YUVA conversion



//------------------------------------------------------------------------------
// call for YUVA . ARGB conversion

func WebPPictureYUVAToARGB(picture *picture.Picture) int {
  if picture == nil { return 0  }
  if (picture.Y == nil || picture.U == nil || picture.V == nil) {
    return picture.SetEncodingError(picture.ENC_ERROR_nil_PARAMETER)
  }
  if ((picture.ColorSpace & colorspace.WEBP_CSP_ALPHA_BIT) && picture.A == nil) {
    return picture.SetEncodingError(picture.ENC_ERROR_nil_PARAMETER)
  }
  if ((picture.ColorSpace & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420) {
    return picture.SetEncodingError(picture.ENC_ERROR_INVALID_CONFIGURATION)
  }
  // Allocate a new argb buffer (discarding the previous one).
  if !picture.WebPPictureAllocARGB(picture) { return 0  }
  picture.UseARGB = true

  // Convert
  {
    var y int
    width := picture.Width
    height := picture.Height
    argb_stride := 4 * picture.ARGBStride
    dst []uint8 = (*uint8)picture.ARGB
    *cur_u = picture.U, *cur_v = picture.V, *cur_y := picture.Y
    WebPUpsampleLinePairFunc upsample = WebPGetLinePairConverter(ALPHA_OFFSET > 0)

    // First row, with replicated top samples.
    upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width)
    cur_y += picture.YStride
    dst += argb_stride
    // Center rows.
    for y = 1; y + 1 < height; y += 2 {
      var top_u *uint8 = cur_u
      var top_v *uint8 = cur_v
      cur_u += picture.UVStride
      cur_v += picture.UVStride
      upsample(cur_y, cur_y + picture.YStride, top_u, top_v, cur_u, cur_v, dst, dst + argb_stride, width)
      cur_y += 2 * picture.YStride
      dst += 2 * argb_stride
    }
    // Last row (if needed), with replicated bottom samples.
    if (height > 1 && !(height & 1)) {
      upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, width)
    }
    // Insert alpha values if needed, in replacement for the default 0xff ones.
    if (picture.ColorSpace & colorspace.WEBP_CSP_ALPHA_BIT) {
      for y = 0; y < height; y++ {
        var argb_dst *uint32 = picture.ARGB + y * picture.ARGBStride
        var src *uint8 = picture.A + y * picture.AStride
        var x int
        for x = 0; x < width; x++ {
          argb_dst[x] = (argb_dst[x] & uint(0x00ffffff)) | ((uint32)src[x] << 24)
        }
      }
    }
  }
  return 1
}

//------------------------------------------------------------------------------
// automatic import / conversion

func Import(/* const */ picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int, step int, swap_rb int, import_alpha int) int {
  var y int
  // swap_rb . b,g,r,a , !swap_rb . r,g,b,a
  var r_ptr *uint8 = rgb + (tenary.If(swap_rb, 2, 0))
  var g_ptr *uint8 = rgb + 1
  var b_ptr *uint8 = rgb + (tenary.If(swap_rb, 0, 2))
  width := picture.Width
  height := picture.Height

  if abs(rgb_stride) < (tenary.If(import_alpha, 4, 3)) * width { return 0  }

  if (!picture.UseARGB) {
    var a_ptr *uint8 = import_alpha ? rgb + 3 : nil
    return ImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, 0.0 /* no dithering */, 0, picture)
  }
  if !picture.WebPPictureAlloc(picture) { return 0  }

  VP8LDspInit()
  WebPInitAlphaProcessing()

  if (import_alpha) {
    // dst[] byte order is {a,r,g,b} for big-endian, {b,g,r,a} for little endian
    dst *uint32 = picture.ARGB
    do_copy := (ALPHA_OFFSET == 3) && swap_rb
    assert.Assert(step == 4)
    if (do_copy) {
      for y = 0; y < height; y++ {
        stdlib.MemCpy(dst, rgb, width * 4)
        rgb += rgb_stride
        dst += picture.ARGBStride
      }
    } else {
      for y = 0; y < height; y++ {
#ifdef constants.FALSE
        // BGRA or RGBA input order.
        var a_ptr *uint8 = rgb + 3
        WebPPackARGB(a_ptr, r_ptr, g_ptr, b_ptr, width, dst)
        r_ptr += rgb_stride
        g_ptr += rgb_stride
        b_ptr += rgb_stride
#else
        // RGBA input order. Need to swap R and B.
        VP8LConvertBGRAToRGBA((/* const */ *uint32)rgb, width, (*uint8)dst)
#endif
        rgb += rgb_stride
        dst += picture.ARGBStride
      }
    }
  } else {
    dst *uint32 = picture.ARGB
    assert.Assert(step >= 3)
    for y = 0; y < height; y++ {
      WebPPackRGB(r_ptr, g_ptr, b_ptr, width, step, dst)
      r_ptr += rgb_stride
      g_ptr += rgb_stride
      b_ptr += rgb_stride
      dst += picture.ARGBStride
    }
  }
  return 1
}

// Public API

#if !defined(WEBP_REDUCE_CSP)

func WebPPictureImportBGR(picture *picture.Picture, /*const*/ bgr *uint8, bgr_stride int) int {
  return (picture != nil && bgr != nil)
             ? Import(picture, bgr, bgr_stride, 3, 1, 0)
             : 0
}

func WebPPictureImportBGRA(picture *picture.Picture, /*const*/ bgra *uint8, bgra_stride int) int {
  return (picture != nil && bgra != nil)
             ? Import(picture, bgra, bgra_stride, 4, 1, 1)
             : 0
}

func WebPPictureImportBGRX(picture *picture.Picture, /*const*/ bgrx *uint8, bgrx_stride int) int {
  return (picture != nil && bgrx != nil)
             ? Import(picture, bgrx, bgrx_stride, 4, 1, 0)
             : 0
}

#endif  // WEBP_REDUCE_CSP

func WebPPictureImportRGB(picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int) int {
  return (picture != nil && rgb != nil)
             ? Import(picture, rgb, rgb_stride, 3, 0, 0)
             : 0
}

func WebPPictureImportRGBA(picture *picture.Picture, /*const*/ rgba []uint8, rgba_stride int) int {
  return (picture != nil && rgba != nil)
             ? Import(picture, rgba, rgba_stride, 4, 0, 1)
             : 0
}

func WebPPictureImportRGBX(picture *picture.Picture, /*const*/ rgbx *uint8, rgbx_stride int) int {
  return (picture != nil && rgbx != nil)
             ? Import(picture, rgbx, rgbx_stride, 4, 0, 0)
             : 0
}

//------------------------------------------------------------------------------
