package dsp

// Copyright 2016 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// MSA variant of methods for lossless decoder
//
// Author: Prashant Patil (prashant.patil@imgtec.com)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

#if defined(WEBP_USE_MSA)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

//------------------------------------------------------------------------------
// Colorspace conversion functions

#define CONVERT16_BGRA_XXX(psrc, pdst, m0, m1, m2)          \
  for {                                                      \
    v16u8 src0, src1, src2, src3, dst0, dst1, dst2;         \
    LD_UB4(psrc, 16, src0, src1, src2, src3);               \
    VSHF_B2_UB(src0, src1, src1, src2, m0, m1, dst0, dst1); \
    dst2 = VSHF_UB(src2, src3, m2);                         \
    ST_UB2(dst0, dst1, pdst, 16);                           \
    ST_UB(dst2, pdst + 32);                                 \
  } while (0)

#define CONVERT12_BGRA_XXX(psrc, pdst, m0, m1, m2)          \
  for {                                                      \
    var pix_w uint32                                         \
    v16u8 src0, src1, src2, dst0, dst1, dst2;               \
    LD_UB3(psrc, 16, src0, src1, src2);                     \
    VSHF_B2_UB(src0, src1, src1, src2, m0, m1, dst0, dst1); \
    dst2 = VSHF_UB(src2, src2, m2);                         \
    ST_UB2(dst0, dst1, pdst, 16);                           \
    pix_w = __msa_copy_s_w((v4i32)dst2, 0);                 \
    SW(pix_w, pdst + 32);                                   \
  } while (0)

#define CONVERT8_BGRA_XXX(psrc, pdst, m0, m1)               \
  for {                                                      \
    var pix_d uint64                                         \
    v16u8 src0, src1, src2 = {0}, dst0, dst1;               \
    LD_UB2(psrc, 16, src0, src1);                           \
    VSHF_B2_UB(src0, src1, src1, src2, m0, m1, dst0, dst1); \
    ST_UB(dst0, pdst);                                      \
    pix_d = __msa_copy_s_d((v2i64)dst1, 0);                 \
    SD(pix_d, pdst + 16);                                   \
  } while (0)

#define CONVERT4_BGRA_XXX(psrc, pdst, m)             \
  for {                                               \
    const v16u8 src0 = LD_UB(psrc);                  \
    const v16u8 dst0 = VSHF_UB(src0, src0, m);       \
    pix_d := __msa_copy_s_d((v2i64)dst0, 0); \
    pix_w := __msa_copy_s_w((v4i32)dst0, 2); \
    SD(pix_d, pdst + 0);                             \
    SW(pix_w, pdst + 8);                             \
  } while (0)

#define CONVERT1_BGRA_BGR(psrc, pdst) \
  for {                                \
    b := (psrc)[0];      \
    g := (psrc)[1];      \
    r := (psrc)[2];      \
    (pdst)[0] = b;                    \
    (pdst)[1] = g;                    \
    (pdst)[2] = r;                    \
  } while (0)

#define CONVERT1_BGRA_RGB(psrc, pdst) \
  for {                                \
    b := (psrc)[0];      \
    g := (psrc)[1];      \
    r := (psrc)[2];      \
    (pdst)[0] = r;                    \
    (pdst)[1] = g;                    \
    (pdst)[2] = b;                    \
  } while (0)

#define TRANSFORM_COLOR_INVERSE_8(src0, src1, dst0, dst1, c0, c1, mask0, \
                                  mask1)                                 \
  for {                                                                   \
    v8i16 g0, g1, t0, t1, t2, t3;                                        \
    v4i32 t4, t5;                                                        \
    VSHF_B2_SH(src0, src0, src1, src1, mask0, mask0, g0, g1);            \
    DOTP_SB2_SH(g0, g1, c0, c0, t0, t1);                                 \
    SRAI_H2_SH(t0, t1, 5);                                               \
    t0 = __msa_addv_h(t0, (v8i16)src0);                                  \
    t1 = __msa_addv_h(t1, (v8i16)src1);                                  \
    t4 = __msa_srli_w((v4i32)t0, 16);                                    \
    t5 = __msa_srli_w((v4i32)t1, 16);                                    \
    DOTP_SB2_SH(t4, t5, c1, c1, t2, t3);                                 \
    SRAI_H2_SH(t2, t3, 5);                                               \
    ADD2(t0, t2, t1, t3, t0, t1);                                        \
    VSHF_B2_UB(src0, t0, src1, t1, mask1, mask1, dst0, dst1);            \
  } while (0)

#define TRANSFORM_COLOR_INVERSE_4(src, dst, c0, c1, mask0, mask1) \
  for {                                                            \
    const v16i8 g0 = VSHF_SB(src, src, mask0);                    \
    v8i16 t0 = __msa_dotp_s_h(c0, g0);                            \
    v8i16 t1;                                                     \
    v4i32 t2;                                                     \
    t0 = SRAI_H(t0, 5);                                           \
    t0 = __msa_addv_h(t0, (v8i16)src);                            \
    t2 = __msa_srli_w((v4i32)t0, 16);                             \
    t1 = __msa_dotp_s_h(c1, (v16i8)t2);                           \
    t1 = SRAI_H(t1, 5);                                           \
    t0 = t0 + t1;                                                 \
    dst = VSHF_UB(src, t0, mask1);                                \
  } while (0)

func ConvertBGRAToRGBA_MSA(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var i int
  var ptemp_src *uint8 = (/* const */ *uint8)src;
  ptemp_dst *uint8 = (*uint8)dst;
  v16u8 src0, dst0;
  const v16u8 mask = {2, 1, 0, 3, 6, 5, 4, 7, 10, 9, 8, 11, 14, 13, 12, 15}

  while (num_pixels >= 8) {
    v16u8 src1, dst1;
    LD_UB2(ptemp_src, 16, src0, src1);
    VSHF_B2_UB(src0, src0, src1, src1, mask, mask, dst0, dst1);
    ST_UB2(dst0, dst1, ptemp_dst, 16);
    ptemp_src += 32;
    ptemp_dst += 32;
    num_pixels -= 8;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 4) {
      src0 = LD_UB(ptemp_src);
      dst0 = VSHF_UB(src0, src0, mask);
      ST_UB(dst0, ptemp_dst);
      ptemp_src += 16;
      ptemp_dst += 16;
      num_pixels -= 4;
    }
    for i = 0; i < num_pixels; i++ {
      b := ptemp_src[2];
      g := ptemp_src[1];
      r := ptemp_src[0];
      a := ptemp_src[3];
      ptemp_dst[0] = b;
      ptemp_dst[1] = g;
      ptemp_dst[2] = r;
      ptemp_dst[3] = a;
      ptemp_src += 4;
      ptemp_dst += 4;
    }
  }
}

func ConvertBGRAToBGR_MSA(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var ptemp_src *uint8 = (/* const */ *uint8)src;
  ptemp_dst *uint8 = (*uint8)dst;
  const v16u8 mask0 = {0, 1, 2, 4, 5, 6, 8, 9, 10, 12, 13, 14, 16, 17, 18, 20}
  const v16u8 mask1 = {5,  6,  8,  9,  10, 12, 13, 14, 16, 17, 18, 20, 21, 22, 24, 25}
  const v16u8 mask2 = {10, 12, 13, 14, 16, 17, 18, 20, 21, 22, 24, 25, 26, 28, 29, 30}

  while (num_pixels >= 16) {
    CONVERT16_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1, mask2);
    ptemp_src += 64;
    ptemp_dst += 48;
    num_pixels -= 16;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 12) {
      CONVERT12_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1, mask2);
      ptemp_src += 48;
      ptemp_dst += 36;
      num_pixels -= 12;
    } else if (num_pixels >= 8) {
      CONVERT8_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1);
      ptemp_src += 32;
      ptemp_dst += 24;
      num_pixels -= 8;
    } else if (num_pixels >= 4) {
      CONVERT4_BGRA_XXX(ptemp_src, ptemp_dst, mask0);
      ptemp_src += 16;
      ptemp_dst += 12;
      num_pixels -= 4;
    }
    if (num_pixels == 3) {
      CONVERT1_BGRA_BGR(ptemp_src + 0, ptemp_dst + 0);
      CONVERT1_BGRA_BGR(ptemp_src + 4, ptemp_dst + 3);
      CONVERT1_BGRA_BGR(ptemp_src + 8, ptemp_dst + 6);
    } else if (num_pixels == 2) {
      CONVERT1_BGRA_BGR(ptemp_src + 0, ptemp_dst + 0);
      CONVERT1_BGRA_BGR(ptemp_src + 4, ptemp_dst + 3);
    } else if (num_pixels == 1) {
      CONVERT1_BGRA_BGR(ptemp_src, ptemp_dst);
    }
  }
}

func ConvertBGRAToRGB_MSA(/* const */ src *uint32, num_pixels int, dst *uint8) {
  var ptemp_src *uint8 = (/* const */ *uint8)src;
  ptemp_dst *uint8 = (*uint8)dst;
  const v16u8 mask0 = {2, 1, 0, 6, 5, 4, 10, 9, 8, 14, 13, 12, 18, 17, 16, 22}
  const v16u8 mask1 = {5,  4,  10, 9,  8,  14, 13, 12, 18, 17, 16, 22, 21, 20, 26, 25}
  const v16u8 mask2 = {8,  14, 13, 12, 18, 17, 16, 22, 21, 20, 26, 25, 24, 30, 29, 28}

  while (num_pixels >= 16) {
    CONVERT16_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1, mask2);
    ptemp_src += 64;
    ptemp_dst += 48;
    num_pixels -= 16;
  }
  if (num_pixels) {
    if (num_pixels >= 12) {
      CONVERT12_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1, mask2);
      ptemp_src += 48;
      ptemp_dst += 36;
      num_pixels -= 12;
    } else if (num_pixels >= 8) {
      CONVERT8_BGRA_XXX(ptemp_src, ptemp_dst, mask0, mask1);
      ptemp_src += 32;
      ptemp_dst += 24;
      num_pixels -= 8;
    } else if (num_pixels >= 4) {
      CONVERT4_BGRA_XXX(ptemp_src, ptemp_dst, mask0);
      ptemp_src += 16;
      ptemp_dst += 12;
      num_pixels -= 4;
    }
    if (num_pixels == 3) {
      CONVERT1_BGRA_RGB(ptemp_src + 0, ptemp_dst + 0);
      CONVERT1_BGRA_RGB(ptemp_src + 4, ptemp_dst + 3);
      CONVERT1_BGRA_RGB(ptemp_src + 8, ptemp_dst + 6);
    } else if (num_pixels == 2) {
      CONVERT1_BGRA_RGB(ptemp_src + 0, ptemp_dst + 0);
      CONVERT1_BGRA_RGB(ptemp_src + 4, ptemp_dst + 3);
    } else if (num_pixels == 1) {
      CONVERT1_BGRA_RGB(ptemp_src, ptemp_dst);
    }
  }
}

func AddGreenToBlueAndRed_MSA(/* const */ src *uint32, num_pixels int, dst *uint32) {
  var i int
  var in *uint8 = (/* const */ *uint8)src;
  out *uint8 = (*uint8)dst;
  v16u8 src0, dst0, tmp0;
  const v16u8 mask = {1, 255, 1, 255, 5,  255, 5,  255, 9, 255, 9, 255, 13, 255, 13, 255}

  while (num_pixels >= 8) {
    v16u8 src1, dst1, tmp1;
    LD_UB2(in, 16, src0, src1);
    VSHF_B2_UB(src0, src1, src1, src0, mask, mask, tmp0, tmp1);
    ADD2(src0, tmp0, src1, tmp1, dst0, dst1);
    ST_UB2(dst0, dst1, out, 16);
    in += 32;
    out += 32;
    num_pixels -= 8;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 4) {
      src0 = LD_UB(in);
      tmp0 = VSHF_UB(src0, src0, mask);
      dst0 = src0 + tmp0;
      ST_UB(dst0, out);
      in += 16;
      out += 16;
      num_pixels -= 4;
    }
    for i = 0; i < num_pixels; i++ {
      b := in[0];
      g := in[1];
      r := in[2];
      out[0] = (b + g) & 0xff;
      out[1] = g;
      out[2] = (r + g) & 0xff;
      out[4] = in[4];
      out += 4;
    }
  }
}

func TransformColorInverse_MSA(/* const */ m *VP8LMultipliers, /*const*/ src *uint32, num_pixels int, dst *uint32) {
  v16u8 src0, dst0;
  const v16i8 g2br =
      (v16i8)__msa_fill_w(m.green_to_blue | (m.green_to_red << 16));
  const v16i8 r2b = (v16i8)__msa_fill_w(m.red_to_blue);
  const v16u8 mask0 = {1, 255, 1, 255, 5,  255, 5,  255, 9, 255, 9, 255, 13, 255, 13, 255}
  const v16u8 mask1 = {16, 1, 18, 3,  20, 5,  22, 7, 24, 9, 26, 11, 28, 13, 30, 15}

  while (num_pixels >= 8) {
    v16u8 src1, dst1;
    LD_UB2(src, 4, src0, src1);
    TRANSFORM_COLOR_INVERSE_8(src0, src1, dst0, dst1, g2br, r2b, mask0, mask1);
    ST_UB2(dst0, dst1, dst, 4);
    src += 8;
    dst += 8;
    num_pixels -= 8;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 4) {
      src0 = LD_UB(src);
      TRANSFORM_COLOR_INVERSE_4(src0, dst0, g2br, r2b, mask0, mask1);
      ST_UB(dst0, dst);
      src += 4;
      dst += 4;
      num_pixels -= 4;
    }
    if (num_pixels > 0) {
      src0 = LD_UB(src);
      TRANSFORM_COLOR_INVERSE_4(src0, dst0, g2br, r2b, mask0, mask1);
      if (num_pixels == 3) {
        pix_d := __msa_copy_s_d((v2i64)dst0, 0);
        pix_w := __msa_copy_s_w((v4i32)dst0, 2);
        SD(pix_d, dst + 0);
        SW(pix_w, dst + 2);
      } else if (num_pixels == 2) {
        pix_d := __msa_copy_s_d((v2i64)dst0, 0);
        SD(pix_d, dst);
      } else {
        pix_w := __msa_copy_s_w((v4i32)dst0, 0);
        SW(pix_w, dst);
      }
    }
  }
}

//------------------------------------------------------------------------------
// Entry point

extern func VP8LDspInitMSA(void);

WEBP_TSAN_IGNORE_FUNCTION func VP8LDspInitMSA(){
  VP8LConvertBGRAToRGBA = ConvertBGRAToRGBA_MSA;
  VP8LConvertBGRAToBGR = ConvertBGRAToBGR_MSA;
  VP8LConvertBGRAToRGB = ConvertBGRAToRGB_MSA;

  VP8LAddGreenToBlueAndRed = AddGreenToBlueAndRed_MSA;
  VP8LTransformColorInverse = TransformColorInverse_MSA;
}

#else  // !WEBP_USE_MSA

WEBP_DSP_INIT_STUB(VP8LDspInitMSA)

#endif  // WEBP_USE_MSA
