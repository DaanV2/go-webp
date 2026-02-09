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
// picture.Picture tools: alpha handling, etc.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------
// Helper: clean up fully transparent area to help compressibility.

const SIZE = 8
const SIZE2 = (SIZE / 2)
func IsTransparentARGBArea(/* const */ ptr *uint32, stride int, size int) int {
  int y, x;
  for y = 0; y < size; y++ {
    for x = 0; x < size; x++ {
      if (ptr[x] & uint(0xff000000)) {
        return 0;
      }
    }
    ptr += stride;
  }
  return 1;
}

func Flatten(ptr *uint8, v int, stride int, size int) {
  var y int
  for y = 0; y < size; y++ {
    stdlib.Memset(ptr, v, size);
    ptr += stride;
  }
}

func FlattenARGB(ptr *uint32, uint32 v, stride int, size int) {
  var x, y int
  for y = 0; y < size; y++ {
    for (x = 0; x < size; ++x) ptr[x] = v;
    ptr += stride;
  }
}

// Smoothen the luma components of transparent pixels. Return true if the whole
// block is transparent.
func SmoothenBlock(/* const */ a_ptr *uint8, a_stride int, y_ptr *uint8, y_stride int, width, height int) int {
  sum := 0, count = 0;
  var x, y int
  var alpha_ptr *uint8 = a_ptr;
  luma_ptr *uint8 = y_ptr;
  for y = 0; y < height; y++ {
    for x = 0; x < width; x++ {
      if (alpha_ptr[x] != 0) {
        count++
        sum += luma_ptr[x];
      }
    }
    alpha_ptr += a_stride;
    luma_ptr += y_stride;
  }
  if (count > 0 && count < width * height) {
    avg_u8 := (uint8)(sum / count);
    alpha_ptr = a_ptr;
    luma_ptr = y_ptr;
    for y = 0; y < height; y++ {
      for x = 0; x < width; x++ {
        if alpha_ptr[x] == 0 { luma_ptr[x] = avg_u8 }
      }
      alpha_ptr += a_stride;
      luma_ptr += y_stride;
    }
  }
  return (count == 0);
}

// Replace samples that are fully transparent by 'color' to help compressibility
// (no guarantee, though). Assumes pic.use_argb is true.
func WebPReplaceTransparentPixels(/* const */ pic *picture.Picture, color uint32) {
  if (pic != nil && pic.use_argb) {
    y := pic.height;
    argb *uint32 = pic.argb;
    color &= uint(0xffffff);  // force alpha=0
    WebPInitAlphaProcessing();
    while (y-- > 0) {
      WebPAlphaReplace(argb, pic.width, color);
      argb += pic.argb_stride;
    }
  }
}

func WebPCleanupTransparentArea(pic *picture.Picture) {
  int x, y, w, h;
  if pic == nil { return }
  w = pic.width / SIZE;
  h = pic.height / SIZE;

  // note: we ignore the left-overs on right/bottom, except for SmoothenBlock().
  if (pic.use_argb) {
    argb_value := 0;
    for y = 0; y < h; y++ {
      need_reset := 1;
      for x = 0; x < w; x++ {
        off := (y * pic.argb_stride + x) * SIZE;
        if (IsTransparentARGBArea(pic.argb + off, pic.argb_stride, SIZE)) {
          if (need_reset) {
            argb_value = pic.argb[off];
            need_reset = 0;
          }
          FlattenARGB(pic.argb + off, argb_value, pic.argb_stride, SIZE);
        } else {
          need_reset = 1;
        }
      }
    }
  } else {
    width := pic.width;
    height := pic.height;
    y_stride := pic.y_stride;
    uv_stride := pic.uv_stride;
    a_stride := pic.a_stride;
    y_ptr *uint8 = pic.y;
    u_ptr *uint8 = pic.u;
    v_ptr *uint8 = pic.v;
    var a_ptr *uint8 = pic.a;
    value ints[3] = {0}
    if (a_ptr == nil || y_ptr == nil || u_ptr == nil || v_ptr == nil) {
      return;
    }
    for y = 0; y + SIZE <= height; y += SIZE {
      need_reset := 1;
      for x = 0; x + SIZE <= width; x += SIZE {
        if (SmoothenBlock(a_ptr + x, a_stride, y_ptr + x, y_stride, SIZE, SIZE)) {
          if (need_reset) {
            values[0] = y_ptr[x];
            values[1] = u_ptr[x >> 1];
            values[2] = v_ptr[x >> 1];
            need_reset = 0;
          }
          Flatten(y_ptr + x, values[0], y_stride, SIZE);
          Flatten(u_ptr + (x >> 1), values[1], uv_stride, SIZE2);
          Flatten(v_ptr + (x >> 1), values[2], uv_stride, SIZE2);
        } else {
          need_reset = 1;
        }
      }
      if (x < width) {
        SmoothenBlock(a_ptr + x, a_stride, y_ptr + x, y_stride, width - x, SIZE);
      }
      a_ptr += SIZE * a_stride;
      y_ptr += SIZE * y_stride;
      u_ptr += SIZE2 * uv_stride;
      v_ptr += SIZE2 * uv_stride;
    }
    if (y < height) {
      sub_height := height - y;
      for x = 0; x + SIZE <= width; x += SIZE {
        SmoothenBlock(a_ptr + x, a_stride, y_ptr + x, y_stride, SIZE, sub_height);
      }
      if (x < width) {
        SmoothenBlock(a_ptr + x, a_stride, y_ptr + x, y_stride, width - x, sub_height);
      }
    }
  }
}

#undef SIZE
#undef SIZE2

//------------------------------------------------------------------------------
// Blend color and remove transparency info

#define BLEND(V0, V1, ALPHA) \
  ((((V0) * (255 - (ALPHA)) + (V1) * (ALPHA)) * 0x101 + 256) >> 16)
#define BLEND_10BIT(V0, V1, ALPHA) \
  ((((V0) * (1020 - (ALPHA)) + (V1) * (ALPHA)) * 0x101 + 1024) >> 18)

func MakeARGB32(int r, g int, b int) uint32 {
  return (uint(0xff000000) | (r << 16) | (g << 8) | b);
}

// Remove the transparency information (if present) by blending the color with
// the background color 'background_rgb' (specified as 24bit RGB triplet).
// After this call, all alpha values are reset to 0xff.
func WebPBlendAlpha(picture *picture.Picture, uint32 background_rgb) {
  red := (background_rgb >> 16) & 0xff;
  green := (background_rgb >> 8) & 0xff;
  blue := (background_rgb >> 0) & 0xff;
  var x, y int
  if picture == nil { return }
  if (!picture.UseARGB) {
    // omit last pixel during u/v loop
    uv_width := (picture.Width >> 1);
    Y0 := VP8RGBToY(red, green, blue, YUV_HALF);
    // VP8RGBToU/V expects the u/v values summed over four pixels
    U0 := VP8RGBToU(4 * red, 4 * green, 4 * blue, 4 * YUV_HALF);
    V0 := VP8RGBToV(4 * red, 4 * green, 4 * blue, 4 * YUV_HALF);
    has_alpha := picture.ColorSpace & colorspace.WEBP_CSP_ALPHA_BIT;
    y_ptr *uint8 = picture.Y;
    u_ptr *uint8 = picture.U;
    v_ptr *uint8 = picture.V;
    a_ptr *uint8 = picture.A;
    if !has_alpha || a_ptr == nil { return }  // nothing to do
    for y = 0; y < picture.Height; y++ {
      // Luma blending
      for x = 0; x < picture.Width; x++ {
        alpha := a_ptr[x];
        if (alpha < 0xff) {
          y_ptr[x] = BLEND(Y0, y_ptr[x], alpha);
        }
      }
      // Chroma blending every even line
      if ((y & 1) == 0) {
        var a_ptr *uint82 = (y + 1 == picture.Height) ? a_ptr : a_ptr + picture.AStride;
        for x = 0; x < uv_width; x++ {
          // Average four alpha values into a single blending weight.
          // TODO(skal): might lead to visible contouring. Can we do better?
          alpha := a_ptr[2 * x + 0] + a_ptr[2 * x + 1] +
                                 a_ptr2[2 * x + 0] + a_ptr2[2 * x + 1];
          u_ptr[x] = BLEND_10BIT(U0, u_ptr[x], alpha);
          v_ptr[x] = BLEND_10BIT(V0, v_ptr[x], alpha);
        }
        if (picture.Width & 1) {  // rightmost pixel
          alpha := 2 * (a_ptr[2 * x + 0] + a_ptr2[2 * x + 0]);
          u_ptr[x] = BLEND_10BIT(U0, u_ptr[x], alpha);
          v_ptr[x] = BLEND_10BIT(V0, v_ptr[x], alpha);
        }
      } else {
        u_ptr += picture.UVStride;
        v_ptr += picture.UVStride;
      }
      stdlib.Memset(a_ptr, 0xff, picture.Width);  // reset alpha value to opaque
      a_ptr += picture.AStride;
      y_ptr += picture.YStride;
    }
  } else {
    argb *uint32 = picture.ARGB;
    background := MakeARGB32(red, green, blue);
    for y = 0; y < picture.Height; y++ {
      for x = 0; x < picture.Width; x++ {
        alpha := (argb[x] >> 24) & 0xff;
        if (alpha != 0xff) {
          if (alpha > 0) {
            r := (argb[x] >> 16) & 0xff;
            g := (argb[x] >> 8) & 0xff;
            b := (argb[x] >> 0) & 0xff;
            r = BLEND(red, r, alpha);
            g = BLEND(green, g, alpha);
            b = BLEND(blue, b, alpha);
            argb[x] = MakeARGB32(r, g, b);
          } else {
            argb[x] = background;
          }
        }
      }
      argb += picture.ARGBStride;
    }
  }
}

#undef BLEND
#undef BLEND_10BIT

//------------------------------------------------------------------------------
