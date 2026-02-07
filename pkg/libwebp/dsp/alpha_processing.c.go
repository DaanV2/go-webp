package dsp

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Utilities for processing transparent channel.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

// Tables can be faster on some platform but incur some extra binary size (~2k).
#if !defined(USE_TABLES_FOR_ALPHA_MULT)
const USE_TABLES_FOR_ALPHA_MULT =0  // ALTERNATE_CODE
#endif

// -----------------------------------------------------------------------------

const MFIX = 24  // 24bit fixed-point arithmetic
const HALF = ((uint(1) << MFIX) >> 1)
const KINV_255 =((uint(1) << MFIX) / uint(255))

func Mult(uint8 x, uint32 mult) uint32 {
  v := (x * mult + HALF) >> MFIX;
  assert.Assert(v <= 255);  // <- 24bit precision is enough to ensure that.
  return v;
}

#if (USE_TABLES_FOR_ALPHA_MULT == 1)

var kMultTables = [2][256]uint32 = {
    // (uint(255) << MFIX) / alpha
    {0x00000000, 0xff000000, 0x7f800000, 0x55000000, 0x3fc00000, 0x33000000, 0x2a800000, 0x246db6db, 0x1fe00000, 0x1c555555, 0x19800000, 0x172e8ba2, 0x15400000, 0x139d89d8, 0x1236db6d, 0x11000000, 0x0ff00000, 0x0f000000, 0x0e2aaaaa, 0x0d6bca1a, 0x0cc00000, 0x0c249249, 0x0b9745d1, 0x0b1642c8, 0x0aa00000, 0x0a333333, 0x09cec4ec, 0x0971c71c, 0x091b6db6, 0x08cb08d3, 0x08800000, 0x0839ce73, 0x07f80000, 0x07ba2e8b, 0x07800000, 0x07492492, 0x07155555, 0x06e45306, 0x06b5e50d, 0x0689d89d, 0x06600000, 0x063831f3, 0x06124924, 0x05ee23b8, 0x05cba2e8, 0x05aaaaaa, 0x058b2164, 0x056cefa8, 0x05500000, 0x05343eb1, 0x05199999, 0x05000000, 0x04e76276, 0x04cfb2b7, 0x04b8e38e, 0x04a2e8ba, 0x048db6db, 0x0479435e, 0x04658469, 0x045270d0, 0x04400000, 0x042e29f7, 0x041ce739, 0x040c30c3, 0x03fc0000, 0x03ec4ec4, 0x03dd1745, 0x03ce540f, 0x03c00000, 0x03b21642, 0x03a49249, 0x03976fc6, 0x038aaaaa, 0x037e3f1f, 0x03722983, 0x03666666, 0x035af286, 0x034fcace, 0x0344ec4e, 0x033a5440, 0x03300000, 0x0325ed09, 0x031c18f9, 0x0312818a, 0x03092492, 0x03000000, 0x02f711dc, 0x02ee5846, 0x02e5d174, 0x02dd7baf, 0x02d55555, 0x02cd5cd5, 0x02c590b2, 0x02bdef7b, 0x02b677d4, 0x02af286b, 0x02a80000, 0x02a0fd5c, 0x029a1f58, 0x029364d9, 0x028ccccc, 0x0286562d, 0x02800000, 0x0279c952, 0x0273b13b, 0x026db6db, 0x0267d95b, 0x026217ec, 0x025c71c7, 0x0256e62a, 0x0251745d, 0x024c1bac, 0x0246db6d, 0x0241b2f9, 0x023ca1af, 0x0237a6f4, 0x0232c234, 0x022df2df, 0x02293868, 0x02249249, 0x02200000, 0x021b810e, 0x021714fb, 0x0212bb51, 0x020e739c, 0x020a3d70, 0x02061861, 0x02020408, 0x01fe0000, 0x01fa0be8, 0x01f62762, 0x01f25213, 0x01ee8ba2, 0x01ead3ba, 0x01e72a07, 0x01e38e38, 0x01e00000, 0x01dc7f10, 0x01d90b21, 0x01d5a3e9, 0x01d24924, 0x01cefa8d, 0x01cbb7e3, 0x01c880e5, 0x01c55555, 0x01c234f7, 0x01bf1f8f, 0x01bc14e5, 0x01b914c1, 0x01b61eed, 0x01b33333, 0x01b05160, 0x01ad7943, 0x01aaaaaa, 0x01a7e567, 0x01a5294a, 0x01a27627, 0x019fcbd2, 0x019d2a20, 0x019a90e7, 0x01980000, 0x01957741, 0x0192f684, 0x01907da4, 0x018e0c7c, 0x018ba2e8, 0x018940c5, 0x0186e5f0, 0x01849249, 0x018245ae, 0x01800000, 0x017dc11f, 0x017b88ee, 0x0179574e, 0x01772c23, 0x01750750, 0x0172e8ba, 0x0170d045, 0x016ebdd7, 0x016cb157, 0x016aaaaa, 0x0168a9b9, 0x0166ae6a, 0x0164b8a7, 0x0162c859, 0x0160dd67, 0x015ef7bd, 0x015d1745, 0x015b3bea, 0x01596596, 0x01579435, 0x0155c7b4, 0x01540000, 0x01523d03, 0x01507eae, 0x014ec4ec, 0x014d0fac, 0x014b5edc, 0x0149b26c, 0x01480a4a, 0x01466666, 0x0144c6af, 0x01432b16, 0x0141938b, 0x01400000, 0x013e7063, 0x013ce4a9, 0x013b5cc0, 0x0139d89d, 0x01385830, 0x0136db6d, 0x01356246, 0x0133ecad, 0x01327a97, 0x01310bf6, 0x012fa0be, 0x012e38e3, 0x012cd459, 0x012b7315, 0x012a150a, 0x0128ba2e, 0x01276276, 0x01260dd6, 0x0124bc44, 0x01236db6, 0x01222222, 0x0120d97c, 0x011f93bc, 0x011e50d7, 0x011d10c4, 0x011bd37a, 0x011a98ef, 0x0119611a, 0x01182bf2, 0x0116f96f, 0x0115c988, 0x01149c34, 0x0113716a, 0x01124924, 0x01112358, 0x01100000, 0x010edf12, 0x010dc087, 0x010ca458, 0x010b8a7d, 0x010a72f0, 0x01095da8, 0x01084a9f, 0x010739ce, 0x01062b2e, 0x01051eb8, 0x01041465, 0x01030c30, 0x01020612, 0x01010204, 0x01000000}, // alpha * KINV_255
    {0x00000000, 0x00010101, 0x00020202, 0x00030303, 0x00040404, 0x00050505, 0x00060606, 0x00070707, 0x00080808, 0x00090909, 0x000a0a0a, 0x000b0b0b, 0x000c0c0c, 0x000d0d0d, 0x000e0e0e, 0x000f0f0f, 0x00101010, 0x00111111, 0x00121212, 0x00131313, 0x00141414, 0x00151515, 0x00161616, 0x00171717, 0x00181818, 0x00191919, 0x001a1a1a, 0x001b1b1b, 0x001c1c1c, 0x001d1d1d, 0x001e1e1e, 0x001f1f1f, 0x00202020, 0x00212121, 0x00222222, 0x00232323, 0x00242424, 0x00252525, 0x00262626, 0x00272727, 0x00282828, 0x00292929, 0x002a2a2a, 0x002b2b2b, 0x002c2c2c, 0x002d2d2d, 0x002e2e2e, 0x002f2f2f, 0x00303030, 0x00313131, 0x00323232, 0x00333333, 0x00343434, 0x00353535, 0x00363636, 0x00373737, 0x00383838, 0x00393939, 0x003a3a3a, 0x003b3b3b, 0x003c3c3c, 0x003d3d3d, 0x003e3e3e, 0x003f3f3f, 0x00404040, 0x00414141, 0x00424242, 0x00434343, 0x00444444, 0x00454545, 0x00464646, 0x00474747, 0x00484848, 0x00494949, 0x004a4a4a, 0x004b4b4b, 0x004c4c4c, 0x004d4d4d, 0x004e4e4e, 0x004f4f4f, 0x00505050, 0x00515151, 0x00525252, 0x00535353, 0x00545454, 0x00555555, 0x00565656, 0x00575757, 0x00585858, 0x00595959, 0x005a5a5a, 0x005b5b5b, 0x005c5c5c, 0x005d5d5d, 0x005e5e5e, 0x005f5f5f, 0x00606060, 0x00616161, 0x00626262, 0x00636363, 0x00646464, 0x00656565, 0x00666666, 0x00676767, 0x00686868, 0x00696969, 0x006a6a6a, 0x006b6b6b, 0x006c6c6c, 0x006d6d6d, 0x006e6e6e, 0x006f6f6f, 0x00707070, 0x00717171, 0x00727272, 0x00737373, 0x00747474, 0x00757575, 0x00767676, 0x00777777, 0x00787878, 0x00797979, 0x007a7a7a, 0x007b7b7b, 0x007c7c7c, 0x007d7d7d, 0x007e7e7e, 0x007f7f7f, 0x00808080, 0x00818181, 0x00828282, 0x00838383, 0x00848484, 0x00858585, 0x00868686, 0x00878787, 0x00888888, 0x00898989, 0x008a8a8a, 0x008b8b8b, 0x008c8c8c, 0x008d8d8d, 0x008e8e8e, 0x008f8f8f, 0x00909090, 0x00919191, 0x00929292, 0x00939393, 0x00949494, 0x00959595, 0x00969696, 0x00979797, 0x00989898, 0x00999999, 0x009a9a9a, 0x009b9b9b, 0x009c9c9c, 0x009d9d9d, 0x009e9e9e, 0x009f9f9f, 0x00a0a0a0, 0x00a1a1a1, 0x00a2a2a2, 0x00a3a3a3, 0x00a4a4a4, 0x00a5a5a5, 0x00a6a6a6, 0x00a7a7a7, 0x00a8a8a8, 0x00a9a9a9, 0x00aaaaaa, 0x00ababab, 0x00acacac, 0x00adadad, 0x00aeaeae, 0x00afafaf, 0x00b0b0b0, 0x00b1b1b1, 0x00b2b2b2, 0x00b3b3b3, 0x00b4b4b4, 0x00b5b5b5, 0x00b6b6b6, 0x00b7b7b7, 0x00b8b8b8, 0x00b9b9b9, 0x00bababa, 0x00bbbbbb, 0x00bcbcbc, 0x00bdbdbd, 0x00bebebe, 0x00bfbfbf, 0x00c0c0c0, 0x00c1c1c1, 0x00c2c2c2, 0x00c3c3c3, 0x00c4c4c4, 0x00c5c5c5, 0x00c6c6c6, 0x00c7c7c7, 0x00c8c8c8, 0x00c9c9c9, 0x00cacaca, 0x00cbcbcb, 0x00cccccc, 0x00cdcdcd, 0x00cecece, 0x00cfcfcf, 0x00d0d0d0, 0x00d1d1d1, 0x00d2d2d2, 0x00d3d3d3, 0x00d4d4d4, 0x00d5d5d5, 0x00d6d6d6, 0x00d7d7d7, 0x00d8d8d8, 0x00d9d9d9, 0x00dadada, 0x00dbdbdb, 0x00dcdcdc, 0x00dddddd, 0x00dedede, 0x00dfdfdf, 0x00e0e0e0, 0x00e1e1e1, 0x00e2e2e2, 0x00e3e3e3, 0x00e4e4e4, 0x00e5e5e5, 0x00e6e6e6, 0x00e7e7e7, 0x00e8e8e8, 0x00e9e9e9, 0x00eaeaea, 0x00ebebeb, 0x00ececec, 0x00ededed, 0x00eeeeee, 0x00efefef, 0x00f0f0f0, 0x00f1f1f1, 0x00f2f2f2, 0x00f3f3f3, 0x00f4f4f4, 0x00f5f5f5, 0x00f6f6f6, 0x00f7f7f7, 0x00f8f8f8, 0x00f9f9f9, 0x00fafafa, 0x00fbfbfb, 0x00fcfcfc, 0x00fdfdfd, 0x00fefefe, 0x00ffffff}}

func GetScale(uint32 a, inverse int) uint32 {
  return kMultTables[!inverse][a];
}

#else

func GetScale(uint32 a, inverse int) uint32 {
  return inverse ? (uint(255) << MFIX) / a : a * KINV_255;
}

#endif  // USE_TABLES_FOR_ALPHA_MULT

func WebPMultARGBRow_C(/* const */ ptr *uint32, width int, inverse int) {
  var x int
  for x = 0; x < width; x++ {
    argb := ptr[x];
    if (argb < uint(0xff000000)) {     // alpha < 255
      if (argb <= uint(0x00ffffff)) {  // alpha == 0
        ptr[x] = 0;
      } else {
        alpha := (argb >> 24) & 0xff;
        scale := GetScale(alpha, inverse);
        out := argb & uint(0xff000000);
        out |= Mult(argb >> 0, scale) << 0;
        out |= Mult(argb >> 8, scale) << 8;
        out |= Mult(argb >> 16, scale) << 16;
        ptr[x] = out;
      }
    }
  }
}

func WebPMultRow_C(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int) {
  var x int
  for x = 0; x < width; x++ {
    a := alpha[x];
    if (a != 255) {
      if (a == 0) {
        ptr[x] = 0;
      } else {
        scale := GetScale(a, inverse);
        ptr[x] = Mult(ptr[x], scale);
      }
    }
  }
}

#undef KINV_255
#undef HALF
#undef MFIX

func (*WebPMultARGBRow)(/* const */ ptr *uint32, width int, inverse int);
func (*WebPMultRow)(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int);

//------------------------------------------------------------------------------
// Generic per-plane calls

func WebPMultARGBRows(ptr *uint8, stride int, width int, num_rows int , inverse int) {
  var n int
  for n = 0; n < num_rows; n++ {
    WebPMultARGBRow((*uint32)ptr, width, inverse);
    ptr += stride;
  }
}

func WebPMultRows(ptr *uint8, stride int, /*const*/ alpha *uint8, alpha_stride int, width int, num_rows int , inverse int) {
  var n int
  for n = 0; n < num_rows; n++ {
    WebPMultRow(ptr, alpha, width, inverse);
    ptr += stride;
    alpha += alpha_stride;
  }
}

//------------------------------------------------------------------------------
// Premultiplied modes

// non dithered-modes

// (x * a * 32897) >> 23 is bit-wise equivalent to (int)(x * a / 255.)
// for all 8bit x or a. For bit-wise equivalence to (int)(x * a / 255. + .5),
// one can use instead: (x * a * 65793 + (1 << 23)) >> 24
#if 1  // (int)(x * a / 255.)
#define MULTIPLIER(a) ((a) * uint(32897))
#define PREMULTIPLY(x, m) (((x) * (m)) >> 23)
#else  // (int)(x * a / 255. + .5)
#define MULTIPLIER(a) ((a) * uint(65793))
#define PREMULTIPLY(x, m) (((x) * (m) + (uint(1) << 23)) >> 24)
#endif

#if !WEBP_NEON_OMIT_C_CODE
func ApplyAlphaMultiply_C(rgba *uint8, alpha_first int, w int, h int, stride int) {
  while (h-- > 0) {
    var rgb *uint8 = rgba + (tenary.If(alpha_first, 1, 0));
    var alpha *uint8 = rgba + (tenary.If(alpha_first, 0, 3));
    var i int
    for i = 0; i < w; i++ {
      a := alpha[4 * i];
      if (a != 0xff) {
        mult := MULTIPLIER(a);
        rgb[4 * i + 0] = PREMULTIPLY(rgb[4 * i + 0], mult);
        rgb[4 * i + 1] = PREMULTIPLY(rgb[4 * i + 1], mult);
        rgb[4 * i + 2] = PREMULTIPLY(rgb[4 * i + 2], mult);
      }
    }
    rgba += stride;
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE
#undef MULTIPLIER
#undef PREMULTIPLY

// rgbA4444

#define MULTIPLIER(a) ((a) * 0x1111)  // 0x1111 ~= (1 << 16) / 15

static  uint8 dither_hi(uint8 x) {
  return (x & 0xf0) | (x >> 4);
}

static  uint8 dither_lo(uint8 x) {
  return (x & 0x0f) | (x << 4);
}

static  uint8 multiply(uint8 x, uint32 m) {
  return (x * m) >> 16;
}

static  func ApplyAlphaMultiply4444_C(rgba *uint84444, w int, h int, stride int, rg_byte_pos int /* 0 or 1 */) {
  while (h-- > 0) {
    var i int
    for i = 0; i < w; i++ {
      rg := rgba4444[2 * i + rg_byte_pos];
      ba := rgba4444[2 * i + (rg_byte_pos ^ 1)];
      a := ba & 0x0f;
      mult := MULTIPLIER(a);
      r := multiply(dither_hi(rg), mult);
      g := multiply(dither_lo(rg), mult);
      b := multiply(dither_hi(ba), mult);
      rgba4444[2 * i + rg_byte_pos] = (r & 0xf0) | ((g >> 4) & 0x0f);
      rgba4444[2 * i + (rg_byte_pos ^ 1)] = (b & 0xf0) | a;
    }
    rgba4444 += stride;
  }
}
#undef MULTIPLIER

func ApplyAlphaMultiply_16b_C(rgba *uint84444, w int, h int, stride int) {
#if (WEBP_SWAP_16BIT_CSP == 1)
  ApplyAlphaMultiply4444_C(rgba4444, w, h, stride, 1);
#else
  ApplyAlphaMultiply4444_C(rgba4444, w, h, stride, 0);
#endif
}

#if !WEBP_NEON_OMIT_C_CODE
func DispatchAlpha_C(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint8, dst_stride int) int {
  alpha_mask := 0xff;
  int i, j;

  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      alpha_value := alpha[i];
      dst[4 * i] = alpha_value;
      alpha_mask &= alpha_value;
    }
    alpha += alpha_stride;
    dst += dst_stride;
  }

  return (alpha_mask != 0xff);
}

func DispatchAlphaToGreen_C(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint32, dst_stride int) {
  int i, j;
  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      dst[i] = alpha[i] << 8;  // leave A/R/B channels zero'd.
    }
    alpha += alpha_stride;
    dst += dst_stride;
  }
}

func ExtractAlpha_C(/* const */ argb *uint8, argb_stride int, width, height int, alpha *uint8, alpha_stride int) int {
  alpha_mask := 0xff;
  int i, j;

  for j = 0; j < height; j++ {
    for i = 0; i < width; i++ {
      alpha_value := argb[4 * i];
      alpha[i] = alpha_value;
      alpha_mask &= alpha_value;
    }
    argb += argb_stride;
    alpha += alpha_stride;
  }
  return (alpha_mask == 0xff);
}

func ExtractGreen_C(/* const */ argb *uint32, alpha *uint8, size int) {
  var i int
  for (i = 0; i < size; ++i) alpha[i] = argb[i] >> 8;
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------



//------------------------------------------------------------------------------
// Simple channel manipulations.

func MakeARGB32(int a, r int, g int, b int) uint32 {
  return (((uint32)a << 24) | (r << 16) | (g << 8) | b);
}

#ifdef constants.WORDS_BIGENDIAN
func PackARGB_C(/* const */ a *uint8, /*const*/ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, out *uint32) {
  var i int
  for i = 0; i < len; i++ {
    out[i] = MakeARGB32(a[4 * i], r[4 * i], g[4 * i], b[4 * i]);
  }
}
#endif

func PackRGB_C(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, step int, out *uint32) {
  int i, offset = 0;
  for i = 0; i < len; i++ {
    out[i] = MakeARGB32(0xff, r[offset], g[offset], b[offset]);
    offset += step;
  }
}

func (*WebPApplyAlphaMultiply)(*uint8, int, int, int, int);
func (*WebPApplyAlphaMultiply4444)(*uint8, int, int, int);
int (*WebPDispatchAlpha)(/* const */ *uint8, int, int, int, *uint8, int);
func (*WebPDispatchAlphaToGreen)(/* const */ *uint8, int, int, int, *uint32, int);
int (*WebPExtractAlpha)(/* const */ *uint8, int, int, int, *uint8, int);
func (*WebPExtractGreen)(/* const */ argb *uint32, alpha *uint8, size int);
#ifdef constants.WORDS_BIGENDIAN
func (*WebPPackARGB)(/* const */ a *uint8, /*const*/ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, int, *uint32);
#endif
func (*WebPPackRGB)(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, step int, out *uint32);

int (*WebPHasAlpha8b)(/* const */ src *uint8, length int);
int (*WebPHasAlpha32b)(/* const */ src *uint8, length int);
func (*WebPAlphaReplace)(src *uint32, length int, color uint32);

//------------------------------------------------------------------------------
// Init function

extern VP8CPUInfo VP8GetCPUInfo;
extern func WebPInitAlphaProcessingMIPSdspR2(void);
extern func WebPInitAlphaProcessingSSE2(void);
extern func WebPInitAlphaProcessingSSE41(void);
extern func WebPInitAlphaProcessingNEON(void);

WEBP_DSP_INIT_FUNC(WebPInitAlphaProcessing) {
  WebPMultARGBRow = WebPMultARGBRow_C;
  WebPMultRow = WebPMultRow_C;
  WebPApplyAlphaMultiply4444 = ApplyAlphaMultiply_16b_C;

#ifdef constants.WORDS_BIGENDIAN
  WebPPackARGB = PackARGB_C;
#endif
  WebPPackRGB = PackRGB_C;
#if !WEBP_NEON_OMIT_C_CODE
  WebPApplyAlphaMultiply = ApplyAlphaMultiply_C;
  WebPDispatchAlpha = DispatchAlpha_C;
  WebPDispatchAlphaToGreen = DispatchAlphaToGreen_C;
  WebPExtractAlpha = ExtractAlpha_C;
  WebPExtractGreen = ExtractGreen_C;
#endif

  WebPHasAlpha8b = HasAlpha8b_C;
  WebPHasAlpha32b = HasAlpha32b_C;
  WebPAlphaReplace = AlphaReplace_C;

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if defined(WEBP_HAVE_SSE2)
    if (VP8GetCPUInfo(kSSE2)) {
      WebPInitAlphaProcessingSSE2();
#if defined(WEBP_HAVE_SSE41)
      if (VP8GetCPUInfo(kSSE4_1)) {
        WebPInitAlphaProcessingSSE41();
      }
#endif
    }
#endif
#if defined(WEBP_USE_MIPS_DSP_R2)
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      WebPInitAlphaProcessingMIPSdspR2();
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    WebPInitAlphaProcessingNEON();
  }
#endif

  assert.Assert(WebPMultARGBRow != nil);
  assert.Assert(WebPMultRow != nil);
  assert.Assert(WebPApplyAlphaMultiply != nil);
  assert.Assert(WebPApplyAlphaMultiply4444 != nil);
  assert.Assert(WebPDispatchAlpha != nil);
  assert.Assert(WebPDispatchAlphaToGreen != nil);
  assert.Assert(WebPExtractAlpha != nil);
  assert.Assert(WebPExtractGreen != nil);
#ifdef constants.WORDS_BIGENDIAN
  assert.Assert(WebPPackARGB != nil);
#endif
  assert.Assert(WebPPackRGB != nil);
  assert.Assert(WebPHasAlpha8b != nil);
  assert.Assert(WebPHasAlpha32b != nil);
  assert.Assert(WebPAlphaReplace != nil);
}
