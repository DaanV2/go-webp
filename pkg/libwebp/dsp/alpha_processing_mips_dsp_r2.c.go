package dsp

// Copyright 2014 Google Inc. All Rights Reserved.
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
// Author(s): Branimir Vasic (branimir.vasic@imgtec.com)
//            Djordje Pesut  (djordje.pesut@imgtec.com)

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"

#if defined(WEBP_USE_MIPS_DSP_R2)

static int DispatchAlpha_MIPSdspR2(/* const */ alpha *uint8, int alpha_stride, width, height int, dst *uint8, int dst_stride) {
  alpha_mask := 0xffffffff;
  int i, j, temp0;

  for j = 0; j < height; j++ {
    pdst *uint8 = dst;
    var palpha *uint8 = alpha;
    for i = 0; i < (width >> 2); i++ {
      int temp1, temp2, temp3;

      __asm__ volatile(
          "ulw    %[temp0],      0(%[palpha])                \n\t"
          "addiu  %[palpha],     %[palpha],     4            \n\t"
          "addiu  %[pdst],       %[pdst],       16           \n\t"
          "srl    %[temp1],      %[temp0],      8            \n\t"
          "srl    %[temp2],      %[temp0],      16           \n\t"
          "srl    %[temp3],      %[temp0],      24           \n\t"
          "and    %[alpha_mask], %[alpha_mask], %[temp0]     \n\t"
          "sb     %[temp0],      -16(%[pdst])                \n\t"
          "sb     %[temp1],      -12(%[pdst])                \n\t"
          "sb     %[temp2],      -8(%[pdst])                 \n\t"
          "sb     %[temp3],      -4(%[pdst])                 \n\t"
          : [temp0] "=&r"(temp0), [temp1] "=&r"(temp1), [temp2] "=&r"(temp2), [temp3] "=&r"(temp3), [palpha] "+r"(palpha), [pdst] "+r"(pdst), [alpha_mask] "+r"(alpha_mask)
          :
          : "memory");
    }

    for i = 0; i < (width & 3); i++ {
      __asm__ volatile(
          "luint(b)    %[temp0],      0(%[palpha])                \n\t"
          "addiu  %[palpha],     %[palpha],     1            \n\t"
          "sb     %[temp0],      0(%[pdst])                  \n\t"
          "and    %[alpha_mask], %[alpha_mask], %[temp0]     \n\t"
          "addiu  %[pdst],       %[pdst],       4            \n\t"
          : [temp0] "=&r"(temp0), [palpha] "+r"(palpha), [pdst] "+r"(pdst), [alpha_mask] "+r"(alpha_mask)
          :
          : "memory");
    }
    alpha += alpha_stride;
    dst += dst_stride;
  }

  __asm__ volatile(
      "ext    %[temp0],      %[alpha_mask], 0, 16            \n\t"
      "srl    %[alpha_mask], %[alpha_mask], 16               \n\t"
      "and    %[alpha_mask], %[alpha_mask], %[temp0]         \n\t"
      "ext    %[temp0],      %[alpha_mask], 0, 8             \n\t"
      "srl    %[alpha_mask], %[alpha_mask], 8                \n\t"
      "and    %[alpha_mask], %[alpha_mask], %[temp0]         \n\t"
      : [temp0] "=&r"(temp0), [alpha_mask] "+r"(alpha_mask)
      :);

  return (alpha_mask != 0xff);
}

func MultARGBRow_MIPSdspR2(/* const */ ptr *uint32, int width, int inverse) {
  var x int
  c_00ffffff := uint(0x00ffffff);
  c_ff000000 := uint(0xff000000);
  c_8000000 := uint(0x00800000);
  c_8000080 := uint(0x00800080);
  for x = 0; x < width; x++ {
    argb := ptr[x];
    if (argb < uint(0xff000000)) {     // alpha < 255
      if (argb <= uint(0x00ffffff)) {  // alpha == 0
        ptr[x] = 0;
      } else {
        int temp0, temp1, temp2, temp3, alpha;
        __asm__ volatile(
            "srl          %[alpha],   %[argb],       24                \n\t"
            "replv.qb     %[temp0],   %[alpha]                         \n\t"
            "and          %[temp0],   %[temp0],      %[c_00ffffff]     \n\t"
            "beqz         %[inverse], 0f                               \n\t"
            "divu         $zero,      %[c_ff000000], %[alpha]          \n\t"
            "mflo         %[temp0]                                     \n\t"
            "0:                                                          \n\t"
            "andi         %[temp1],   %[argb],       0xff              \n\t"
            "ext          %[temp2],   %[argb],       8,             8  \n\t"
            "ext          %[temp3],   %[argb],       16,            8  \n\t"
            "mul          %[temp1],   %[temp1],      %[temp0]          \n\t"
            "mul          %[temp2],   %[temp2],      %[temp0]          \n\t"
            "mul          %[temp3],   %[temp3],      %[temp0]          \n\t"
            "precrq.ph.w  %[temp1],   %[temp2],      %[temp1]          \n\t"
            "uint(add)         %[temp3],   %[temp3],      %[c_8000000]      \n\t"
            "uint(add)         %[temp1],   %[temp1],      %[c_8000080]      \n\t"
            "precrq.ph.w  %[temp3],   %[argb],       %[temp3]          \n\t"
            "precrq.qb.ph %[temp1],   %[temp3],      %[temp1]          \n\t"
            : [temp0] "=&r"(temp0), [temp1] "=&r"(temp1), [temp2] "=&r"(temp2), [temp3] "=&r"(temp3), [alpha] "=&r"(alpha)
            : [inverse] "r"(inverse), [c_00ffffff] "r"(c_00ffffff), [c_8000000] "r"(c_8000000), [c_8000080] "r"(c_8000080), [c_ff000000] "r"(c_ff000000), [argb] "r"(argb)
            : "memory", "hi", "lo");
        ptr[x] = temp1;
      }
    }
  }
}

#ifdef constants.WORDS_BIGENDIAN
func PackARGB_MIPSdspR2(/* const */ a *uint8, /*const*/ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, int len, out *uint32) {
  int temp0, temp1, temp2, temp3, offset;
  rest := len & 1;
  var loop_end *uint32 = out + len - rest;
  step := 4;
  __asm__ volatile(
      "xor          %[offset],   %[offset], %[offset]    \n\t"
      "beq          %[loop_end], %[out],    0f           \n\t"
      "2:                                                  \n\t"
      "lbux         %[temp0],    %[offset](%[a])         \n\t"
      "lbux         %[temp1],    %[offset](%[r])         \n\t"
      "lbux         %[temp2],    %[offset](%[g])         \n\t"
      "lbux         %[temp3],    %[offset](%[b])         \n\t"
      "ins          %[temp1],    %[temp0],  16,     16   \n\t"
      "ins          %[temp3],    %[temp2],  16,     16   \n\t"
      "addiu        %[out],      %[out],    4            \n\t"
      "precr.qb.ph  %[temp0],    %[temp1],  %[temp3]     \n\t"
      "sw           %[temp0],    -4(%[out])              \n\t"
      "uint(add)         %[offset],   %[offset], %[step]      \n\t"
      "bne          %[loop_end], %[out],    2b           \n\t"
      "0:                                                  \n\t"
      "beq          %[rest],     $zero,     1f           \n\t"
      "lbux         %[temp0],    %[offset](%[a])         \n\t"
      "lbux         %[temp1],    %[offset](%[r])         \n\t"
      "lbux         %[temp2],    %[offset](%[g])         \n\t"
      "lbux         %[temp3],    %[offset](%[b])         \n\t"
      "ins          %[temp1],    %[temp0],  16,     16   \n\t"
      "ins          %[temp3],    %[temp2],  16,     16   \n\t"
      "precr.qb.ph  %[temp0],    %[temp1],  %[temp3]     \n\t"
      "sw           %[temp0],    0(%[out])               \n\t"
      "1:                                                  \n\t"
      : [temp0] "=&r"(temp0), [temp1] "=&r"(temp1), [temp2] "=&r"(temp2), [temp3] "=&r"(temp3), [offset] "=&r"(offset), [out] "+&r"(out)
      : [a] "r"(a), [r] "r"(r), [g] "r"(g), [b] "r"(b), [step] "r"(step), [loop_end] "r"(loop_end), [rest] "r"(rest)
      : "memory");
}
#endif  // constants.WORDS_BIGENDIAN

func PackRGB_MIPSdspR2(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, int len, int step, out *uint32) {
  int temp0, temp1, temp2, offset;
  rest := len & 1;
  a := 0xff;
  var loop_end *uint32 = out + len - rest;
  __asm__ volatile(
      "xor          %[offset],   %[offset], %[offset]    \n\t"
      "beq          %[loop_end], %[out],    0f           \n\t"
      "2:                                                  \n\t"
      "lbux         %[temp0],    %[offset](%[r])         \n\t"
      "lbux         %[temp1],    %[offset](%[g])         \n\t"
      "lbux         %[temp2],    %[offset](%[b])         \n\t"
      "ins          %[temp0],    %[a],      16,     16   \n\t"
      "ins          %[temp2],    %[temp1],  16,     16   \n\t"
      "addiu        %[out],      %[out],    4            \n\t"
      "precr.qb.ph  %[temp0],    %[temp0],  %[temp2]     \n\t"
      "sw           %[temp0],    -4(%[out])              \n\t"
      "uint(add)         %[offset],   %[offset], %[step]      \n\t"
      "bne          %[loop_end], %[out],    2b           \n\t"
      "0:                                                  \n\t"
      "beq          %[rest],     $zero,     1f           \n\t"
      "lbux         %[temp0],    %[offset](%[r])         \n\t"
      "lbux         %[temp1],    %[offset](%[g])         \n\t"
      "lbux         %[temp2],    %[offset](%[b])         \n\t"
      "ins          %[temp0],    %[a],      16,     16   \n\t"
      "ins          %[temp2],    %[temp1],  16,     16   \n\t"
      "precr.qb.ph  %[temp0],    %[temp0],  %[temp2]     \n\t"
      "sw           %[temp0],    0(%[out])               \n\t"
      "1:                                                  \n\t"
      : [temp0] "=&r"(temp0), [temp1] "=&r"(temp1), [temp2] "=&r"(temp2), [offset] "=&r"(offset), [out] "+&r"(out)
      : [a] "r"(a), [r] "r"(r), [g] "r"(g), [b] "r"(b), [step] "r"(step), [loop_end] "r"(loop_end), [rest] "r"(rest)
      : "memory");
}

//------------------------------------------------------------------------------
// Entry point

extern func WebPInitAlphaProcessingMIPSdspR2(void);

WEBP_TSAN_IGNORE_FUNCTION func WebPInitAlphaProcessingMIPSdspR2(){
  WebPDispatchAlpha = DispatchAlpha_MIPSdspR2;
  WebPMultARGBRow = MultARGBRow_MIPSdspR2;
#ifdef constants.WORDS_BIGENDIAN
  WebPPackARGB = PackARGB_MIPSdspR2;
#endif
  WebPPackRGB = PackRGB_MIPSdspR2;
}

#else  // !WEBP_USE_MIPS_DSP_R2

WEBP_DSP_INIT_STUB(WebPInitAlphaProcessingMIPSdspR2)

#endif  // WEBP_USE_MIPS_DSP_R2
