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
// picture.Picture tools: copy, crop, rescaling and view.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if !defined(WEBP_REDUCE_SIZE)
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
#endif  // !defined(WEBP_REDUCE_SIZE)

#define HALVE(x) (((x) + 1) >> 1)

// Grab the 'specs' (writer, *opaque, width, height...) from 'src' and copy them
// into 'dst'. Mark 'dst' as not owning any memory.
func PictureGrabSpecs(/* const */ src *picture.Picture, /*const*/ dst *picture.Picture) {
  assert.Assert(src != nil && dst != nil);
  *dst = *src;
  picture.WebPPictureResetBuffers(dst);
}

//------------------------------------------------------------------------------

// Adjust top-left corner to chroma sample position.
func SnapTopLeftPosition(/* const */ pic *picture.Picture, /*const*/ left *int, /*const*/ top *int) {
  if (!pic.use_argb) {
    *left &= ~1;
    *top &= ~1;
  }
}

// Adjust top-left corner and verify that the sub-rectangle is valid.
func AdjustAndCheckRectangle(/* const */ pic *picture.Picture, /*const*/ left *int, /*const*/ top *int, width, height int) int {
  SnapTopLeftPosition(pic, left, top);
  if (*left) < 0 || (*top) < 0 { return 0  }
  if width <= 0 || height <= 0 { return 0  }
  if (*left) + width > pic.width { return 0  }
  if (*top) + height > pic.height { return 0  }
  return 1;
}

func WebPPictureIsView(/* const */ picture *picture.Picture) int {
  if picture == nil { return 0  }
  if (picture.UseARGB) {
    return (picture.memory_argb_ == nil);
  }
  return (picture.memory_ == nil);
}

func WebPPictureView(/* const */ src *picture.Picture, left int, top int, width, height int, dst *picture.Picture) int {
  if src == nil || dst == nil { return 0  }

  // verify rectangle position.
  if !AdjustAndCheckRectangle(src, &left, &top, width, height) { return 0  }

  if (src != dst) {  // beware of aliasing! We don't want to leak 'memory_'.
    PictureGrabSpecs(src, dst);
  }
  dst.width = width;
  dst.height = height;
  if (!src.use_argb) {
    dst.y = src.y + top * src.y_stride + left;
    dst.u = src.u + (top >> 1) * src.uv_stride + (left >> 1);
    dst.v = src.v + (top >> 1) * src.uv_stride + (left >> 1);
    dst.y_stride = src.y_stride;
    dst.uv_stride = src.uv_stride;
    if (src.a != nil) {
      dst.a = src.a + top * src.a_stride + left;
      dst.a_stride = src.a_stride;
    }
  } else {
    dst.argb = src.argb + top * src.argb_stride + left;
    dst.argb_stride = src.argb_stride;
  }
  return 1;
}

#if !defined(WEBP_REDUCE_SIZE)
//------------------------------------------------------------------------------
// Picture cropping

func WebPPictureCrop(pic *picture.Picture, left int, top int, width, height int) int {
   var tmp picture.Picture

  if pic == nil { return 0  }
  if !AdjustAndCheckRectangle(pic, &left, &top, width, height) { return 0  }

  PictureGrabSpecs(pic, &tmp);
  tmp.width = width;
  tmp.height = height;
  if (!picture.WebPPictureAlloc(&tmp)) {
    return pic.SetEncodingError(picture.tmp.ErrorCode)
  }

  if (!pic.use_argb) {
    y_offset := top * pic.y_stride + left;
    uv_offset := (top / 2) * pic.uv_stride + left / 2;
    WebPCopyPlane(pic.y + y_offset, pic.y_stride, tmp.y, tmp.y_stride, width, height);
    WebPCopyPlane(pic.u + uv_offset, pic.uv_stride, tmp.u, tmp.uv_stride, HALVE(width), HALVE(height));
    WebPCopyPlane(pic.v + uv_offset, pic.uv_stride, tmp.v, tmp.uv_stride, HALVE(width), HALVE(height));

    if (tmp.a != nil) {
      a_offset := top * pic.a_stride + left;
      WebPCopyPlane(pic.a + a_offset, pic.a_stride, tmp.a, tmp.a_stride, width, height);
    }
  } else {
    const src *uint8 =
        (/* const */ *uint8)(pic.argb + top * pic.argb_stride + left);
    WebPCopyPlane(src, pic.argb_stride * 4, (*uint8)tmp.argb, tmp.argb_stride * 4, width * 4, height);
  }
  picture.WebPPictureFree(pic);
  *pic = tmp;
  return 1;
}

//------------------------------------------------------------------------------
// Simple picture rescaler

func RescalePlane(/* const */ src *uint8, src_width int , src_height int, src_stride int, dst *uint8, dst_width int, dst_height int, dst_stride int, work *rescaler_t, num_channels int) int {
   var rescaler WebPRescaler
  y := 0;
  if (!WebPRescalerInit(&rescaler, src_width, src_height, dst, dst_width, dst_height, dst_stride, num_channels, work)) {
    return 0;
  }
  while (y < src_height) {
    y += WebPRescalerImport(&rescaler, src_height - y, src + y * src_stride, src_stride);
    WebPRescalerExport(&rescaler);
  }
  return 1;
}

func AlphaMultiplyARGB(/* const */ pic *picture.Picture, inverse int) {
  assert.Assert(pic.argb != nil);
  WebPMultARGBRows((*uint8)pic.argb, pic.argb_stride * sizeof(*pic.argb), pic.width, pic.height, inverse);
}

func AlphaMultiplyY(/* const */ pic *picture.Picture, inverse int) {
  if (pic.a != nil) {
    WebPMultRows(pic.y, pic.y_stride, pic.a, pic.a_stride, pic.width, pic.height, inverse);
  }
}

func WebPPictureRescale(picture *picture.Picture, width, height int) int {
   var tmp picture.Picture
  int prev_width, prev_height;
  rescaler_t* work;
  status := ENC_OK;

  if picture == nil { return 0  }
  prev_width = picture.Width;
  prev_height = picture.Height;
  if (!WebPRescalerGetScaledDimensions(prev_width, prev_height, &width, &height)) {
    return picture.SetEncodingError(picture.ENC_ERROR_BAD_DIMENSION)
  }

  PictureGrabSpecs(picture, &tmp);
  tmp.width = width;
  tmp.height = height;
  if (!picture.WebPPictureAlloc(&tmp)) {
    return picture.SetEncodingError(picture.tmp.ErrorCode)
  }

  if (!picture.UseARGB) {
    // work = (rescaler_t*)WebPSafeMalloc(uint64(2) * width, sizeof(*work));
    // if (work == nil) {
    //   status = ENC_ERROR_OUT_OF_MEMORY;
    //   goto Cleanup;
    // }
	work = make([]rescaler_t, 2 * width)
    // If present, we need to rescale alpha first (for AlphaMultiplyY).
    if (picture.A != nil) {
      WebPInitAlphaProcessing();
      if (!RescalePlane(picture.A, prev_width, prev_height, picture.AStride, tmp.a, width, height, tmp.a_stride, work, 1)) {
        status = ENC_ERROR_BAD_DIMENSION;
        goto Cleanup;
      }
    }

    // We take transparency into account on the luma plane only. That's not
    // totally exact blending, but still is a good approximation.
    AlphaMultiplyY(picture, 0);
    if (!RescalePlane(picture.Y, prev_width, prev_height, picture.YStride, tmp.y, width, height, tmp.y_stride, work, 1) ||
        !RescalePlane(picture.U, HALVE(prev_width), HALVE(prev_height), picture.UVStride, tmp.u, HALVE(width), HALVE(height), tmp.uv_stride, work, 1) ||
        !RescalePlane(picture.V, HALVE(prev_width), HALVE(prev_height), picture.UVStride, tmp.v, HALVE(width), HALVE(height), tmp.uv_stride, work, 1)) {
      status = ENC_ERROR_BAD_DIMENSION;
      goto Cleanup;
    }
    AlphaMultiplyY(&tmp, 1);
  } else {
    // work = (rescaler_t*)WebPSafeMalloc(uint64(2) * width * 4, sizeof(*work));
    // if (work == nil) {
    //   status = ENC_ERROR_BAD_DIMENSION;
    //   goto Cleanup;
    // }
	work := make([]rescaler_t, 2 * width * 4)

    // In order to correctly interpolate colors, we need to apply the alpha
    // weighting first (black-matting), scale the RGB values, and remove
    // the premultiplication afterward (while preserving the alpha channel).
    WebPInitAlphaProcessing();
    AlphaMultiplyARGB(picture, 0);
    if (!RescalePlane((/* const */ *uint8)picture.ARGB, prev_width, prev_height, picture.ARGBStride * 4, (*uint8)tmp.argb, width, height, tmp.argb_stride * 4, work, 4)) {
      status = ENC_ERROR_BAD_DIMENSION;
      goto Cleanup;
    }
    AlphaMultiplyARGB(&tmp, 1);
  }

Cleanup:
  if (status != ENC_OK) {
    picture.WebPPictureFree(&tmp);
    return picture.SetEncodingError(picture.status)
  }

  picture.WebPPictureFree(picture);
  *picture = tmp;
  return 1;
}
