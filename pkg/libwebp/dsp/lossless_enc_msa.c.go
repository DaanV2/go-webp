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
// MSA variant of Image transform methods for lossless encoder.
//
// Authors: Prashant Patil (Prashant.Patil@imgtec.com)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

#if defined(WEBP_USE_MSA)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

#define TRANSFORM_COLOR_8(src0, src1, dst0, dst1, c0, c1, mask0, mask1) \
  for {                                                                  \
    v8i16 g0, g1, t0, t1, t2, t3;                                       \
    v4i32 t4, t5;                                                       \
    VSHF_B2_SH(src0, src0, src1, src1, mask0, mask0, g0, g1);           \
    DOTP_SB2_SH(g0, g1, c0, c0, t0, t1);                                \
    SRAI_H2_SH(t0, t1, 5);                                              \
    t0 = __msa_subv_h((v8i16)src0, t0);                                 \
    t1 = __msa_subv_h((v8i16)src1, t1);                                 \
    t4 = __msa_srli_w((v4i32)src0, 16);                                 \
    t5 = __msa_srli_w((v4i32)src1, 16);                                 \
    DOTP_SB2_SH(t4, t5, c1, c1, t2, t3);                                \
    SRAI_H2_SH(t2, t3, 5);                                              \
    SUB2(t0, t2, t1, t3, t0, t1);                                       \
    VSHF_B2_UB(src0, t0, src1, t1, mask1, mask1, dst0, dst1);           \
  } while (0)

#define TRANSFORM_COLOR_4(src, dst, c0, c1, mask0, mask1) \
  for {                                                    \
    const v16i8 g0 = VSHF_SB(src, src, mask0);            \
    v8i16 t0 = __msa_dotp_s_h(c0, g0);                    \
    v8i16 t1;                                             \
    v4i32 t2;                                             \
    t0 = SRAI_H(t0, 5);                                   \
    t0 = __msa_subv_h((v8i16)src, t0);                    \
    t2 = __msa_srli_w((v4i32)src, 16);                    \
    t1 = __msa_dotp_s_h(c1, (v16i8)t2);                   \
    t1 = SRAI_H(t1, 5);                                   \
    t0 = t0 - t1;                                         \
    dst = VSHF_UB(src, t0, mask1);                        \
  } while (0)

func TransformColor_MSA(/* const */ /* const */ m *VP8LMultipliers, data *uint32, num_pixels int) {
  v16u8 src0, dst0;
  const v16i8 g2br =
      (v16i8)__msa_fill_w(m.green_to_blue | (m.green_to_red << 16));
  const v16i8 r2b = (v16i8)__msa_fill_w(m.red_to_blue);
  const v16u8 mask0 = {1, 255, 1, 255, 5,  255, 5,  255, 9, 255, 9, 255, 13, 255, 13, 255}
  const v16u8 mask1 = {16, 1, 18, 3,  20, 5,  22, 7, 24, 9, 26, 11, 28, 13, 30, 15}

  while (num_pixels >= 8) {
    v16u8 src1, dst1;
    LD_UB2(data, 4, src0, src1);
    TRANSFORM_COLOR_8(src0, src1, dst0, dst1, g2br, r2b, mask0, mask1);
    ST_UB2(dst0, dst1, data, 4);
    data += 8;
    num_pixels -= 8;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 4) {
      src0 = LD_UB(data);
      TRANSFORM_COLOR_4(src0, dst0, g2br, r2b, mask0, mask1);
      ST_UB(dst0, data);
      data += 4;
      num_pixels -= 4;
    }
    if (num_pixels > 0) {
      src0 = LD_UB(data);
      TRANSFORM_COLOR_4(src0, dst0, g2br, r2b, mask0, mask1);
      if (num_pixels == 3) {
        pix_d := __msa_copy_s_d((v2i64)dst0, 0);
        pix_w := __msa_copy_s_w((v4i32)dst0, 2);
        SD(pix_d, data + 0);
        SW(pix_w, data + 2);
      } else if (num_pixels == 2) {
        pix_d := __msa_copy_s_d((v2i64)dst0, 0);
        SD(pix_d, data);
      } else {
        pix_w := __msa_copy_s_w((v4i32)dst0, 0);
        SW(pix_w, data);
      }
    }
  }
}

func SubtractGreenFromBlueAndRed_MSA(argb_data *uint32, num_pixels int) {
  var i int
  ptemp_data *uint8 = (*uint8)argb_data;
  v16u8 src0, dst0, tmp0;
  const v16u8 mask = {1, 255, 1, 255, 5,  255, 5,  255, 9, 255, 9, 255, 13, 255, 13, 255}

  while (num_pixels >= 8) {
    v16u8 src1, dst1, tmp1;
    LD_UB2(ptemp_data, 16, src0, src1);
    VSHF_B2_UB(src0, src1, src1, src0, mask, mask, tmp0, tmp1);
    SUB2(src0, tmp0, src1, tmp1, dst0, dst1);
    ST_UB2(dst0, dst1, ptemp_data, 16);
    ptemp_data += 8 * 4;
    num_pixels -= 8;
  }
  if (num_pixels > 0) {
    if (num_pixels >= 4) {
      src0 = LD_UB(ptemp_data);
      tmp0 = VSHF_UB(src0, src0, mask);
      dst0 = src0 - tmp0;
      ST_UB(dst0, ptemp_data);
      ptemp_data += 4 * 4;
      num_pixels -= 4;
    }
    for i = 0; i < num_pixels; i++ {
      b := ptemp_data[0];
      g := ptemp_data[1];
      r := ptemp_data[2];
      ptemp_data[0] = (b - g) & 0xff;
      ptemp_data[2] = (r - g) & 0xff;
      ptemp_data += 4;
    }
  }
}

//------------------------------------------------------------------------------
// Entry point

extern func VP8LEncDspInitMSA(void);

WEBP_TSAN_IGNORE_FUNCTION func VP8LEncDspInitMSA(){
  VP8LSubtractGreenFromBlueAndRed = SubtractGreenFromBlueAndRed_MSA;
  VP8LTransformColor = TransformColor_MSA;
}

#else  // !WEBP_USE_MSA

WEBP_DSP_INIT_STUB(VP8LEncDspInitMSA)

#endif  // WEBP_USE_MSA
