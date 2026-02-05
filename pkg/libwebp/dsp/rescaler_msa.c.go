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
// MSA version of rescaling functions
//
// Author: Prashant Patil (prashant.patil@imgtec.com)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

#if defined(WEBP_USE_MSA) && !defined(WEBP_REDUCE_SIZE)

import "github.com/daanv2/go-webp/pkg/assert"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"

const ROUNDER = (WEBP_RESCALER_ONE >> 1)
#define MULT_FIX(x, y) (((uint64)(x) * (y) + ROUNDER) >> WEBP_RESCALER_RFIX)
#define MULT_FIX_FLOOR(x, y) (((uint64)(x) * (y)) >> WEBP_RESCALER_RFIX)

#define CALC_MULT_FIX_16(in0, in1, in2, in3, scale, shift, dst) \
  for {                                                          \
    v4u32 tmp0, tmp1, tmp2, tmp3;                               \
    v16u8 t0, t1, t2, t3, t4, t5;                               \
    v2u64 out0, out1, out2, out3;                               \
    ILVRL_W2_UW(zero, in0, tmp0, tmp1);                         \
    ILVRL_W2_UW(zero, in1, tmp2, tmp3);                         \
    DOTP_UW2_UD(tmp0, tmp1, scale, scale, out0, out1);          \
    DOTP_UW2_UD(tmp2, tmp3, scale, scale, out2, out3);          \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                  \
    PCKEV_B2_UB(out1, out0, out3, out2, t0, t1);                \
    ILVRL_W2_UW(zero, in2, tmp0, tmp1);                         \
    ILVRL_W2_UW(zero, in3, tmp2, tmp3);                         \
    DOTP_UW2_UD(tmp0, tmp1, scale, scale, out0, out1);          \
    DOTP_UW2_UD(tmp2, tmp3, scale, scale, out2, out3);          \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                  \
    PCKEV_B2_UB(out1, out0, out3, out2, t2, t3);                \
    PCKEV_B2_UB(t1, t0, t3, t2, t4, t5);                        \
    dst = (v16u8)__msa_pckev_b((v16i8)t5, (v16i8)t4);           \
  } while (0)

#define CALC_MULT_FIX_4(in0, scale, shift, dst)        \
  for {                                                 \
    v4u32 tmp0, tmp1;                                  \
    v16i8 t0, t1;                                      \
    v2u64 out0, out1;                                  \
    ILVRL_W2_UW(zero, in0, tmp0, tmp1);                \
    DOTP_UW2_UD(tmp0, tmp1, scale, scale, out0, out1); \
    SRAR_D2_UD(out0, out1, shift);                     \
    t0 = __msa_pckev_b((v16i8)out1, (v16i8)out0);      \
    t1 = __msa_pckev_b(t0, t0);                        \
    t0 = __msa_pckev_b(t1, t1);                        \
    dst = __msa_copy_s_w((v4i32)t0, 0);                \
  } while (0)

#define CALC_MULT_FIX1_16(in0, in1, in2, in3, fyscale, shift, dst0, dst1, \
                          dst2, dst3)                                     \
  for {                                                                    \
    v4u32 tmp0, tmp1, tmp2, tmp3;                                         \
    v2u64 out0, out1, out2, out3;                                         \
    ILVRL_W2_UW(zero, in0, tmp0, tmp1);                                   \
    ILVRL_W2_UW(zero, in1, tmp2, tmp3);                                   \
    DOTP_UW2_UD(tmp0, tmp1, fyscale, fyscale, out0, out1);                \
    DOTP_UW2_UD(tmp2, tmp3, fyscale, fyscale, out2, out3);                \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                            \
    PCKEV_W2_UW(out1, out0, out3, out2, dst0, dst1);                      \
    ILVRL_W2_UW(zero, in2, tmp0, tmp1);                                   \
    ILVRL_W2_UW(zero, in3, tmp2, tmp3);                                   \
    DOTP_UW2_UD(tmp0, tmp1, fyscale, fyscale, out0, out1);                \
    DOTP_UW2_UD(tmp2, tmp3, fyscale, fyscale, out2, out3);                \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                            \
    PCKEV_W2_UW(out1, out0, out3, out2, dst2, dst3);                      \
  } while (0)

#define CALC_MULT_FIX1_4(in0, scale, shift, dst)          \
  for {                                                    \
    v4u32 tmp0, tmp1;                                     \
    v2u64 out0, out1;                                     \
    ILVRL_W2_UW(zero, in0, tmp0, tmp1);                   \
    DOTP_UW2_UD(tmp0, tmp1, scale, scale, out0, out1);    \
    SRAR_D2_UD(out0, out1, shift);                        \
    dst = (v4u32)__msa_pckev_w((v4i32)out1, (v4i32)out0); \
  } while (0)

#define CALC_MULT_FIX2_16(in0, in1, in2, in3, mult, scale, shift, dst0, dst1) \
  for {                                                                        \
    v4u32 tmp0, tmp1, tmp2, tmp3;                                             \
    v2u64 out0, out1, out2, out3;                                             \
    ILVRL_W2_UW(in0, in2, tmp0, tmp1);                                        \
    ILVRL_W2_UW(in1, in3, tmp2, tmp3);                                        \
    DOTP_UW2_UD(tmp0, tmp1, mult, mult, out0, out1);                          \
    DOTP_UW2_UD(tmp2, tmp3, mult, mult, out2, out3);                          \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                                \
    DOTP_UW2_UD(out0, out1, scale, scale, out0, out1);                        \
    DOTP_UW2_UD(out2, out3, scale, scale, out2, out3);                        \
    SRAR_D4_UD(out0, out1, out2, out3, shift);                                \
    PCKEV_B2_UB(out1, out0, out3, out2, dst0, dst1);                          \
  } while (0)

#define CALC_MULT_FIX2_4(in0, in1, mult, scale, shift, dst) \
  for {                                                      \
    v4u32 tmp0, tmp1;                                       \
    v2u64 out0, out1;                                       \
    v16i8 t0, t1;                                           \
    ILVRL_W2_UW(in0, in1, tmp0, tmp1);                      \
    DOTP_UW2_UD(tmp0, tmp1, mult, mult, out0, out1);        \
    SRAR_D2_UD(out0, out1, shift);                          \
    DOTP_UW2_UD(out0, out1, scale, scale, out0, out1);      \
    SRAR_D2_UD(out0, out1, shift);                          \
    t0 = __msa_pckev_b((v16i8)out1, (v16i8)out0);           \
    t1 = __msa_pckev_b(t0, t0);                             \
    t0 = __msa_pckev_b(t1, t1);                             \
    dst = __msa_copy_s_w((v4i32)t0, 0);                     \
  } while (0)

static  func ExportRowExpand_0(
    const WEBP_RESTRICT frow *uint32, WEBP_RESTRICT dst *uint8, int length, WEBP_RESTRICT const wrk *WebPRescaler) {
  const v4u32 scale = (v4u32)__msa_fill_w(wrk.fy_scale);
  const v4u32 shift = (v4u32)__msa_fill_w(WEBP_RESCALER_RFIX);
  const v4i32 zero = {0}

  while (length >= 16) {
    v4u32 src0, src1, src2, src3;
    v16u8 out;
    LD_UW4(frow, 4, src0, src1, src2, src3);
    CALC_MULT_FIX_16(src0, src1, src2, src3, scale, shift, out);
    ST_UB(out, dst);
    length -= 16;
    frow += 16;
    dst += 16;
  }
  if (length > 0) {
    var x_out int
    if (length >= 12) {
      uint32 val0_m, val1_m, val2_m;
      v4u32 src0, src1, src2;
      LD_UW3(frow, 4, src0, src1, src2);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      CALC_MULT_FIX_4(src1, scale, shift, val1_m);
      CALC_MULT_FIX_4(src2, scale, shift, val2_m);
      SW3(val0_m, val1_m, val2_m, dst, 4);
      length -= 12;
      frow += 12;
      dst += 12;
    } else if (length >= 8) {
      uint32 val0_m, val1_m;
      v4u32 src0, src1;
      LD_UW2(frow, 4, src0, src1);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      CALC_MULT_FIX_4(src1, scale, shift, val1_m);
      SW2(val0_m, val1_m, dst, 4);
      length -= 8;
      frow += 8;
      dst += 8;
    } else if (length >= 4) {
      uint32 val0_m;
      const v4u32 src0 = LD_UW(frow);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      SW(val0_m, dst);
      length -= 4;
      frow += 4;
      dst += 4;
    }
    for x_out = 0; x_out < length; x_out++ {
      J := frow[x_out];
      v := (int)MULT_FIX(J, wrk.fy_scale);
      dst[x_out] = (v > 255) ? uint(255) : (uint8)v;
    }
  }
}

static  func ExportRowExpand_1(
    const WEBP_RESTRICT frow *uint32, WEBP_RESTRICT irow *uint32, WEBP_RESTRICT dst *uint8, int length, WEBP_RESTRICT const wrk *WebPRescaler) {
  B := WEBP_RESCALER_FRAC(-wrk.y_accum, wrk.y_sub);
  A := (uint32)(WEBP_RESCALER_ONE - B);
  const v4i32 B1 = __msa_fill_w(B);
  const v4i32 A1 = __msa_fill_w(A);
  const v4i32 AB = __msa_ilvr_w(A1, B1);
  const v4u32 scale = (v4u32)__msa_fill_w(wrk.fy_scale);
  const v4u32 shift = (v4u32)__msa_fill_w(WEBP_RESCALER_RFIX);

  while (length >= 16) {
    v4u32 frow0, frow1, frow2, frow3, irow0, irow1, irow2, irow3;
    v16u8 t0, t1, t2, t3, t4, t5;
    LD_UW4(frow, 4, frow0, frow1, frow2, frow3);
    LD_UW4(irow, 4, irow0, irow1, irow2, irow3);
    CALC_MULT_FIX2_16(frow0, frow1, irow0, irow1, AB, scale, shift, t0, t1);
    CALC_MULT_FIX2_16(frow2, frow3, irow2, irow3, AB, scale, shift, t2, t3);
    PCKEV_B2_UB(t1, t0, t3, t2, t4, t5);
    t0 = (v16u8)__msa_pckev_b((v16i8)t5, (v16i8)t4);
    ST_UB(t0, dst);
    frow += 16;
    irow += 16;
    dst += 16;
    length -= 16;
  }
  if (length > 0) {
    var x_out int
    if (length >= 12) {
      uint32 val0_m, val1_m, val2_m;
      v4u32 frow0, frow1, frow2, irow0, irow1, irow2;
      LD_UW3(frow, 4, frow0, frow1, frow2);
      LD_UW3(irow, 4, irow0, irow1, irow2);
      CALC_MULT_FIX2_4(frow0, irow0, AB, scale, shift, val0_m);
      CALC_MULT_FIX2_4(frow1, irow1, AB, scale, shift, val1_m);
      CALC_MULT_FIX2_4(frow2, irow2, AB, scale, shift, val2_m);
      SW3(val0_m, val1_m, val2_m, dst, 4);
      frow += 12;
      irow += 12;
      dst += 12;
      length -= 12;
    } else if (length >= 8) {
      uint32 val0_m, val1_m;
      v4u32 frow0, frow1, irow0, irow1;
      LD_UW2(frow, 4, frow0, frow1);
      LD_UW2(irow, 4, irow0, irow1);
      CALC_MULT_FIX2_4(frow0, irow0, AB, scale, shift, val0_m);
      CALC_MULT_FIX2_4(frow1, irow1, AB, scale, shift, val1_m);
      SW2(val0_m, val1_m, dst, 4);
      frow += 4;
      irow += 4;
      dst += 4;
      length -= 4;
    } else if (length >= 4) {
      uint32 val0_m;
      const v4u32 frow0 = LD_UW(frow + 0);
      const v4u32 irow0 = LD_UW(irow + 0);
      CALC_MULT_FIX2_4(frow0, irow0, AB, scale, shift, val0_m);
      SW(val0_m, dst);
      frow += 4;
      irow += 4;
      dst += 4;
      length -= 4;
    }
    for x_out = 0; x_out < length; x_out++ {
      I := (uint64)A * frow[x_out] + (uint64)B * irow[x_out];
      J := (uint32)((I + ROUNDER) >> WEBP_RESCALER_RFIX);
      v := (int)MULT_FIX(J, wrk.fy_scale);
      dst[x_out] = (v > 255) ? uint(255) : (uint8)v;
    }
  }
}

func RescalerExportRowExpand_MIPSdspR2(/* const */ wrk *WebPRescaler) {
  dst *uint8 = wrk.dst;
  rescaler_t* irow = wrk.irow;
  x_out_max := wrk.dst_width * wrk.num_channels;
  const rescaler_t* frow = wrk.frow;
  assert.Assert(!WebPRescalerOutputDone(wrk));
  assert.Assert(wrk.y_accum <= 0);
  assert.Assert(wrk.y_expand);
  assert.Assert(wrk.y_sub != 0);
  if (wrk.y_accum == 0) {
    ExportRowExpand_0(frow, dst, x_out_max, wrk);
  } else {
    ExportRowExpand_1(frow, irow, dst, x_out_max, wrk);
  }
}

#if 0   // disabled for now. TODO(skal): make match the C-code
static  func ExportRowShrink_0(
    const WEBP_RESTRICT frow *uint32, WEBP_RESTRICT irow *uint32, WEBP_RESTRICT dst *uint8, int length, /*const*/ uint32 yscale, WEBP_RESTRICT const wrk *WebPRescaler) {
  const v4u32 y_scale = (v4u32)__msa_fill_w(yscale);
  const v4u32 fxyscale = (v4u32)__msa_fill_w(wrk.fxy_scale);
  const v4u32 shiftval = (v4u32)__msa_fill_w(WEBP_RESCALER_RFIX);
  const v4i32 zero = { 0 }

  while (length >= 16) {
    v4u32 src0, src1, src2, src3, frac0, frac1, frac2, frac3;
    v16u8 out;
    LD_UW4(frow, 4, src0, src1, src2, src3);
    CALC_MULT_FIX1_16(src0, src1, src2, src3, y_scale, shiftval, frac0, frac1, frac2, frac3);
    LD_UW4(irow, 4, src0, src1, src2, src3);
    SUB4(src0, frac0, src1, frac1, src2, frac2, src3, frac3, src0, src1, src2, src3);
    CALC_MULT_FIX_16(src0, src1, src2, src3, fxyscale, shiftval, out);
    ST_UB(out, dst);
    ST_UW4(frac0, frac1, frac2, frac3, irow, 4);
    frow   += 16;
    irow   += 16;
    dst    += 16;
    length -= 16;
  }
  if (length > 0) {
    var x_out int
    if (length >= 12) {
      uint32 val0_m, val1_m, val2_m;
      v4u32 src0, src1, src2, frac0, frac1, frac2;
      LD_UW3(frow, 4, src0, src1, src2);
      CALC_MULT_FIX1_4(src0, y_scale, shiftval, frac0);
      CALC_MULT_FIX1_4(src1, y_scale, shiftval, frac1);
      CALC_MULT_FIX1_4(src2, y_scale, shiftval, frac2);
      LD_UW3(irow, 4, src0, src1, src2);
      SUB3(src0, frac0, src1, frac1, src2, frac2, src0, src1, src2);
      CALC_MULT_FIX_4(src0, fxyscale, shiftval, val0_m);
      CALC_MULT_FIX_4(src1, fxyscale, shiftval, val1_m);
      CALC_MULT_FIX_4(src2, fxyscale, shiftval, val2_m);
      SW3(val0_m, val1_m, val2_m, dst, 4);
      ST_UW3(frac0, frac1, frac2, irow, 4);
      frow   += 12;
      irow   += 12;
      dst    += 12;
      length -= 12;
    } else if (length >= 8) {
      uint32 val0_m, val1_m;
      v4u32 src0, src1, frac0, frac1;
      LD_UW2(frow, 4, src0, src1);
      CALC_MULT_FIX1_4(src0, y_scale, shiftval, frac0);
      CALC_MULT_FIX1_4(src1, y_scale, shiftval, frac1);
      LD_UW2(irow, 4, src0, src1);
      SUB2(src0, frac0, src1, frac1, src0, src1);
      CALC_MULT_FIX_4(src0, fxyscale, shiftval, val0_m);
      CALC_MULT_FIX_4(src1, fxyscale, shiftval, val1_m);
      SW2(val0_m, val1_m, dst, 4);
      ST_UW2(frac0, frac1, irow, 4);
      frow   += 8;
      irow   += 8;
      dst    += 8;
      length -= 8;
    } else if (length >= 4) {
      uint32 val0_m;
      v4u32 frac0;
      v4u32 src0 = LD_UW(frow);
      CALC_MULT_FIX1_4(src0, y_scale, shiftval, frac0);
      src0 = LD_UW(irow);
      src0 = src0 - frac0;
      CALC_MULT_FIX_4(src0, fxyscale, shiftval, val0_m);
      SW(val0_m, dst);
      ST_UW(frac0, irow);
      frow   += 4;
      irow   += 4;
      dst    += 4;
      length -= 4;
    }
    for x_out = 0; x_out < length; x_out++ {
      frac := (uint32)MULT_FIX_FLOOR(frow[x_out], yscale);
      v := (int)MULT_FIX(irow[x_out] - frac, wrk.fxy_scale);
      dst[x_out] = (v > 255) ? uint(255) : (uint8)v;
      irow[x_out] = frac;
    }
  }
}

static  func ExportRowShrink_1(
    WEBP_RESTRICT irow *uint32, WEBP_RESTRICT dst *uint8, int length, WEBP_RESTRICT const wrk *WebPRescaler) {
  const v4u32 scale = (v4u32)__msa_fill_w(wrk.fxy_scale);
  const v4u32 shift = (v4u32)__msa_fill_w(WEBP_RESCALER_RFIX);
  const v4i32 zero = { 0 }

  while (length >= 16) {
    v4u32 src0, src1, src2, src3;
    v16u8 dst0;
    LD_UW4(irow, 4, src0, src1, src2, src3);
    CALC_MULT_FIX_16(src0, src1, src2, src3, scale, shift, dst0);
    ST_UB(dst0, dst);
    ST_SW4(zero, zero, zero, zero, irow, 4);
    length -= 16;
    irow   += 16;
    dst    += 16;
  }
  if (length > 0) {
    var x_out int
    if (length >= 12) {
      uint32 val0_m, val1_m, val2_m;
      v4u32 src0, src1, src2;
      LD_UW3(irow, 4, src0, src1, src2);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      CALC_MULT_FIX_4(src1, scale, shift, val1_m);
      CALC_MULT_FIX_4(src2, scale, shift, val2_m);
      SW3(val0_m, val1_m, val2_m, dst, 4);
      ST_SW3(zero, zero, zero, irow, 4);
      length -= 12;
      irow   += 12;
      dst    += 12;
    } else if (length >= 8) {
      uint32 val0_m, val1_m;
      v4u32 src0, src1;
      LD_UW2(irow, 4, src0, src1);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      CALC_MULT_FIX_4(src1, scale, shift, val1_m);
      SW2(val0_m, val1_m, dst, 4);
      ST_SW2(zero, zero, irow, 4);
      length -= 8;
      irow   += 8;
      dst    += 8;
    } else if (length >= 4) {
      uint32 val0_m;
      const v4u32 src0 = LD_UW(irow + 0);
      CALC_MULT_FIX_4(src0, scale, shift, val0_m);
      SW(val0_m, dst);
      ST_SW(zero, irow);
      length -= 4;
      irow   += 4;
      dst    += 4;
    }
    for x_out = 0; x_out < length; x_out++ {
      v := (int)MULT_FIX(irow[x_out], wrk.fxy_scale);
      dst[x_out] = (v > 255) ? uint(255) : (uint8)v;
      irow[x_out] = 0;
    }
  }
}

func RescalerExportRowShrink_MIPSdspR2(/* const */ wrk *WebPRescaler) {
  dst *uint8 = wrk.dst;
  rescaler_t* irow = wrk.irow;
  x_out_max := wrk.dst_width * wrk.num_channels;
  const rescaler_t* frow = wrk.frow;
  yscale := wrk.fy_scale * (-wrk.y_accum);
  assert.Assert(!WebPRescalerOutputDone(wrk));
  assert.Assert(wrk.y_accum <= 0);
  assert.Assert(!wrk.y_expand);
  if (yscale) {
    ExportRowShrink_0(frow, irow, dst, x_out_max, yscale, wrk);
  } else {
    ExportRowShrink_1(irow, dst, x_out_max, wrk);
  }
}
#endif  // 0

//------------------------------------------------------------------------------
// Entry point

extern func WebPRescalerDspInitMSA(void);

WEBP_TSAN_IGNORE_FUNCTION func WebPRescalerDspInitMSA(){
  WebPRescalerExportRowExpand = RescalerExportRowExpand_MIPSdspR2;
  //  WebPRescalerExportRowShrink = RescalerExportRowShrink_MIPSdspR2;
}

#else  // !WEBP_USE_MSA

WEBP_DSP_INIT_STUB(WebPRescalerDspInitMSA)

#endif  // WEBP_USE_MSA
