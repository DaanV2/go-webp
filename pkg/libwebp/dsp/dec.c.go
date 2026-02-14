package dsp

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Speed-critical decoding functions, default plain-C implementations.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------

static  uint8 clip_8b(int v) {
  return (!(v & ~0xff)) ? v : tenary.If(v < 0, 0, 255)
}

//------------------------------------------------------------------------------
// Transforms (Paragraph 14.4)

#define STORE(x, y, v) \
  dst[(x) + (y) * BPS] = clip_8b(dst[(x) + (y) * BPS] + ((v) >> 3))

#define STORE2(y, dc, d, c) \
  for {                      \
    DC := (dc);    \
    STORE(0, y, DC + (d));  \
    STORE(1, y, DC + (c));  \
    STORE(2, y, DC - (c));  \
    STORE(3, y, DC - (d));  \
  } while (0)

#if !WEBP_NEON_OMIT_C_CODE
func TransformOne_C(/* const */ in *int16, dst *uint8) {
  int C[4 * 4], *tmp
  var i int
  tmp = C
  for i = 0; i < 4; i++ {       // vertical pass
    a := in[0] + in[8];  // [-4096, 4094]
    b := in[0] - in[8];  // [-4095, 4095]
    c := WEBP_TRANSFORM_AC3_MUL2(in[4]) -
                  WEBP_TRANSFORM_AC3_MUL1(in[12]);  // [-3783, 3783]
    d := WEBP_TRANSFORM_AC3_MUL1(in[4]) +
                  WEBP_TRANSFORM_AC3_MUL2(in[12]);  // [-3785, 3781]
    tmp[0] = a + d;                                 // [-7881, 7875]
    tmp[1] = b + c;                                 // [-7878, 7878]
    tmp[2] = b - c;                                 // [-7878, 7878]
    tmp[3] = a - d;                                 // [-7877, 7879]
    tmp += 4
    in++
  }
  // Each pass is expanding the dynamic range by ~3.85 (upper bound).
  // The exact value is (2. + (20091 + 35468) / 65536).
  // After the second pass, maximum interval is [-3794, 3794], assuming
  // an input in [-2048, 2047] interval. We then need to add a dst value
  // in the [0, 255] range.
  // In the worst case scenario, the input to clip_8b() can be as large as
  // [-60713, 60968].
  tmp = C
  for i = 0; i < 4; i++ {  // horizontal pass
    dc := tmp[0] + 4
    a := dc + tmp[8]
    b := dc - tmp[8]
    c := WEBP_TRANSFORM_AC3_MUL2(tmp[4]) - WEBP_TRANSFORM_AC3_MUL1(tmp[12])
    d := WEBP_TRANSFORM_AC3_MUL1(tmp[4]) + WEBP_TRANSFORM_AC3_MUL2(tmp[12])
    STORE(0, 0, a + d)
    STORE(1, 0, b + c)
    STORE(2, 0, b - c)
    STORE(3, 0, a - d)
    tmp++
    dst += BPS
  }
}

// Simplified transform when only in[0], in[1] and in[4] are non-zero
func TransformAC3_C(/* const */ in *int16, dst *uint8) {
  a := in[0] + 4
  c4 := WEBP_TRANSFORM_AC3_MUL2(in[4])
  d4 := WEBP_TRANSFORM_AC3_MUL1(in[4])
  c1 := WEBP_TRANSFORM_AC3_MUL2(in[1])
  d1 := WEBP_TRANSFORM_AC3_MUL1(in[1])
  STORE2(0, a + d4, d1, c1)
  STORE2(1, a + c4, d1, c1)
  STORE2(2, a - c4, d1, c1)
  STORE2(3, a - d4, d1, c1)
}
#undef STORE2

func TransformTwo_C(/* const */ in *int16, dst *uint8, do_two int) {
  TransformOne_C(in, dst)
  if (do_two) {
    TransformOne_C(in + 16, dst + 4)
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func TransformUV_C(/* const */ in *int16, dst *uint8) {
  VP8Transform(in + 0 * 16, dst, 1)
  VP8Transform(in + 2 * 16, dst + 4 * BPS, 1)
}

#if !WEBP_NEON_OMIT_C_CODE
func TransformDC_C(/* const */ in *int16, dst *uint8) {
  DC := in[0] + 4
  var i, j int
  for j = 0; j < 4; j++ {
    for i = 0; i < 4; i++ {
      STORE(i, j, DC)
    }
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func TransformDCUV_C(/* const */ in *int16, dst *uint8) {
  if in[0 * 16] { VP8TransformDC(in + 0 * 16, dst) }
  if in[1 * 16] { VP8TransformDC(in + 1 * 16, dst + 4) }
  if in[2 * 16] { VP8TransformDC(in + 2 * 16, dst + 4 * BPS) }
  if in[3 * 16] { VP8TransformDC(in + 3 * 16, dst + 4 * BPS + 4) }
}

#undef STORE

//------------------------------------------------------------------------------
// Paragraph 14.3

#if !WEBP_NEON_OMIT_C_CODE
func TransformWHT_C(/* const */ in *int16, out *int16) {
  int tmp[16]
  var i int
  for i = 0; i < 4; i++ {
    a0 := in[0 + i] + in[12 + i]
    a1 := in[4 + i] + in[8 + i]
    a2 := in[4 + i] - in[8 + i]
    a3 := in[0 + i] - in[12 + i]
    tmp[0 + i] = a0 + a1
    tmp[8 + i] = a0 - a1
    tmp[4 + i] = a3 + a2
    tmp[12 + i] = a3 - a2
  }
  for i = 0; i < 4; i++ {
    dc := tmp[0 + i * 4] + 3;  // w/ rounder
    a0 := dc + tmp[3 + i * 4]
    a1 := tmp[1 + i * 4] + tmp[2 + i * 4]
    a2 := tmp[1 + i * 4] - tmp[2 + i * 4]
    a3 := dc - tmp[3 + i * 4]
    out[0] = (a0 + a1) >> 3
    out[16] = (a3 + a2) >> 3
    out[32] = (a0 - a1) >> 3
    out[48] = (a3 - a2) >> 3
    out += 64
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

VP8WHT VP8TransformWHT

//------------------------------------------------------------------------------
// Intra predictions

#define DST(x, y) dst[(x) + (y) * BPS]

#if !WEBP_NEON_OMIT_C_CODE
func TrueMotion(dst *uint8, size int) {
  var top *uint8 = dst - BPS
  var clip *uint80 = VP8kclip1 - top[-1]
  var y int
  for y = 0; y < size; y++ {
    var clip *uint8 = clip0 + dst[-1]
    var x int
    for x = 0; x < size; x++ {
      dst[x] = clip[top[x]]
    }
    dst += BPS
  }
}
func TM4_C(dst *uint8) { TrueMotion(dst, 4); }
func TM8uv_C(dst *uint8) { TrueMotion(dst, 8); }
func TM16_C(dst *uint8) { TrueMotion(dst, 16); }

//------------------------------------------------------------------------------
// 16x16

func VE16_C(dst *uint8) {  // vertical
  var j int
  for j = 0; j < 16; j++ {
    memcpy(dst + j * BPS, dst - BPS, 16)
  }
}

func HE16_C(dst *uint8) {  // horizontal
  var j int
  for j = 16; j > 0; --j {
    stdlib.Memset(dst, dst[-1], 16)
    dst += BPS
  }
}

func Put16(int v, dst *uint8) {
  var j int
  for j = 0; j < 16; j++ {
    stdlib.Memset(dst + j * BPS, v, 16)
  }
}

func DC16_C(dst *uint8) {  // DC
  DC := 16
  var j int
  for j = 0; j < 16; j++ {
    DC += dst[-1 + j * BPS] + dst[j - BPS]
  }
  Put16(DC >> 5, dst)
}

func DC16NoTop_C(dst *uint8) {  // DC with top samples not available
  DC := 8
  var j int
  for j = 0; j < 16; j++ {
    DC += dst[-1 + j * BPS]
  }
  Put16(DC >> 4, dst)
}

func DC16NoLeft_C(dst *uint8) {  // DC with left samples not available
  DC := 8
  var i int
  for i = 0; i < 16; i++ {
    DC += dst[i - BPS]
  }
  Put16(DC >> 4, dst)
}

func DC16NoTopLeft_C(dst *uint8) {  // DC with no top and left samples
  Put16(0x80, dst)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

VP8PredFunc VP8PredLuma16[NUM_B_DC_MODES]

//------------------------------------------------------------------------------
// 4x4

#define AVG3(a, b, c) ((uint8)(((a) + 2 * (b) + (c) + 2) >> 2))
#define AVG2(a, b) (((a) + (b) + 1) >> 1)

#if !WEBP_NEON_OMIT_C_CODE
func VE4_C(dst *uint8) {  // vertical
  var top *uint8 = dst - BPS
  vals[4] := {
      AVG3(top[-1], top[0], top[1]), AVG3(top[0], top[1], top[2]), AVG3(top[1], top[2], top[3]), AVG3(top[2], top[3], top[4]), }
  var i int
  for i = 0; i < 4; i++ {
    memcpy(dst + i * BPS, vals, sizeof(vals))
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func HE4_C(dst *uint8) {  // horizontal
  A := dst[-1 - BPS]
  B := dst[-1]
  C := dst[-1 + BPS]
  D := dst[-1 + 2 * BPS]
  E := dst[-1 + 3 * BPS]
  WebPUint32ToMem(dst + 0 * BPS, uint(0x01010101) * AVG3(A, B, C))
  WebPUint32ToMem(dst + 1 * BPS, uint(0x01010101) * AVG3(B, C, D))
  WebPUint32ToMem(dst + 2 * BPS, uint(0x01010101) * AVG3(C, D, E))
  WebPUint32ToMem(dst + 3 * BPS, uint(0x01010101) * AVG3(D, E, E))
}

#if !WEBP_NEON_OMIT_C_CODE
func DC4_C(dst *uint8) {  // DC
  dc := 4
  var i int
  for (i = 0; i < 4; ++i) dc += dst[i - BPS] + dst[-1 + i * BPS]
  dc >>= 3
  for (i = 0; i < 4; ++i) stdlib.Memset(dst + i * BPS, dc, 4)
}

func RD4_C(dst *uint8) {  // Down-right
  I := dst[-1 + 0 * BPS]
  J := dst[-1 + 1 * BPS]
  K := dst[-1 + 2 * BPS]
  L := dst[-1 + 3 * BPS]
  X := dst[-1 - BPS]
  A := dst[0 - BPS]
  B := dst[1 - BPS]
  C := dst[2 - BPS]
  D := dst[3 - BPS]
  DST(0, 3) = AVG3(J, K, L)
  DST(1, 3) = DST(0, 2) = AVG3(I, J, K)
  DST(2, 3) = DST(1, 2) = DST(0, 1) = AVG3(X, I, J)
  DST(3, 3) = DST(2, 2) = DST(1, 1) = DST(0, 0) = AVG3(A, X, I)
  DST(3, 2) = DST(2, 1) = DST(1, 0) = AVG3(B, A, X)
  DST(3, 1) = DST(2, 0) = AVG3(C, B, A)
  DST(3, 0) = AVG3(D, C, B)
}

func LD4_C(dst *uint8) {  // Down-Left
  A := dst[0 - BPS]
  B := dst[1 - BPS]
  C := dst[2 - BPS]
  D := dst[3 - BPS]
  E := dst[4 - BPS]
  F := dst[5 - BPS]
  G := dst[6 - BPS]
  H := dst[7 - BPS]
  DST(0, 0) = AVG3(A, B, C)
  DST(1, 0) = DST(0, 1) = AVG3(B, C, D)
  DST(2, 0) = DST(1, 1) = DST(0, 2) = AVG3(C, D, E)
  DST(3, 0) = DST(2, 1) = DST(1, 2) = DST(0, 3) = AVG3(D, E, F)
  DST(3, 1) = DST(2, 2) = DST(1, 3) = AVG3(E, F, G)
  DST(3, 2) = DST(2, 3) = AVG3(F, G, H)
  DST(3, 3) = AVG3(G, H, H)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func VR4_C(dst *uint8) {  // Vertical-Right
  I := dst[-1 + 0 * BPS]
  J := dst[-1 + 1 * BPS]
  K := dst[-1 + 2 * BPS]
  X := dst[-1 - BPS]
  A := dst[0 - BPS]
  B := dst[1 - BPS]
  C := dst[2 - BPS]
  D := dst[3 - BPS]
  DST(0, 0) = DST(1, 2) = AVG2(X, A)
  DST(1, 0) = DST(2, 2) = AVG2(A, B)
  DST(2, 0) = DST(3, 2) = AVG2(B, C)
  DST(3, 0) = AVG2(C, D)

  DST(0, 3) = AVG3(K, J, I)
  DST(0, 2) = AVG3(J, I, X)
  DST(0, 1) = DST(1, 3) = AVG3(I, X, A)
  DST(1, 1) = DST(2, 3) = AVG3(X, A, B)
  DST(2, 1) = DST(3, 3) = AVG3(A, B, C)
  DST(3, 1) = AVG3(B, C, D)
}

func VL4_C(dst *uint8) {  // Vertical-Left
  A := dst[0 - BPS]
  B := dst[1 - BPS]
  C := dst[2 - BPS]
  D := dst[3 - BPS]
  E := dst[4 - BPS]
  F := dst[5 - BPS]
  G := dst[6 - BPS]
  H := dst[7 - BPS]
  DST(0, 0) = AVG2(A, B)
  DST(1, 0) = DST(0, 2) = AVG2(B, C)
  DST(2, 0) = DST(1, 2) = AVG2(C, D)
  DST(3, 0) = DST(2, 2) = AVG2(D, E)

  DST(0, 1) = AVG3(A, B, C)
  DST(1, 1) = DST(0, 3) = AVG3(B, C, D)
  DST(2, 1) = DST(1, 3) = AVG3(C, D, E)
  DST(3, 1) = DST(2, 3) = AVG3(D, E, F)
  DST(3, 2) = AVG3(E, F, G)
  DST(3, 3) = AVG3(F, G, H)
}

func HU4_C(dst *uint8) {  // Horizontal-Up
  I := dst[-1 + 0 * BPS]
  J := dst[-1 + 1 * BPS]
  K := dst[-1 + 2 * BPS]
  L := dst[-1 + 3 * BPS]
  DST(0, 0) = AVG2(I, J)
  DST(2, 0) = DST(0, 1) = AVG2(J, K)
  DST(2, 1) = DST(0, 2) = AVG2(K, L)
  DST(1, 0) = AVG3(I, J, K)
  DST(3, 0) = DST(1, 1) = AVG3(J, K, L)
  DST(3, 1) = DST(1, 2) = AVG3(K, L, L)
  DST(3, 2) = DST(2, 2) = DST(0, 3) = DST(1, 3) = DST(2, 3) = DST(3, 3) = L
}

func HD4_C(dst *uint8) {  // Horizontal-Down
  I := dst[-1 + 0 * BPS]
  J := dst[-1 + 1 * BPS]
  K := dst[-1 + 2 * BPS]
  L := dst[-1 + 3 * BPS]
  X := dst[-1 - BPS]
  A := dst[0 - BPS]
  B := dst[1 - BPS]
  C := dst[2 - BPS]

  DST(0, 0) = DST(2, 1) = AVG2(I, X)
  DST(0, 1) = DST(2, 2) = AVG2(J, I)
  DST(0, 2) = DST(2, 3) = AVG2(K, J)
  DST(0, 3) = AVG2(L, K)

  DST(3, 0) = AVG3(A, B, C)
  DST(2, 0) = AVG3(X, A, B)
  DST(1, 0) = DST(3, 1) = AVG3(I, X, A)
  DST(1, 1) = DST(3, 2) = AVG3(J, I, X)
  DST(1, 2) = DST(3, 3) = AVG3(K, J, I)
  DST(1, 3) = AVG3(L, K, J)
}

#undef DST
#undef AVG3
#undef AVG2

VP8PredFunc VP8PredLuma4[NUM_BMODES]

//------------------------------------------------------------------------------
// Chroma

#if !WEBP_NEON_OMIT_C_CODE
func VE8uv_C(dst *uint8) {  // vertical
  var j int
  for j = 0; j < 8; j++ {
    memcpy(dst + j * BPS, dst - BPS, 8)
  }
}

func HE8uv_C(dst *uint8) {  // horizontal
  var j int
  for j = 0; j < 8; j++ {
    stdlib.Memset(dst, dst[-1], 8)
    dst += BPS
  }
}

// helper for chroma-DC predictions
func Put8x8uv(uint8 value, dst *uint8) {
  var j int
  for j = 0; j < 8; j++ {
    stdlib.Memset(dst + j * BPS, value, 8)
  }
}

func DC8uv_C(dst *uint8) {  // DC
  int dc0 = 8
  var i int
  for i = 0; i < 8; i++ {
    dc0 += dst[i - BPS] + dst[-1 + i * BPS]
  }
  Put8x8uv(dc0 >> 4, dst)
}

func DC8uvNoLeft_C(dst *uint8) {  // DC with no left samples
  int dc0 = 4
  var i int
  for i = 0; i < 8; i++ {
    dc0 += dst[i - BPS]
  }
  Put8x8uv(dc0 >> 3, dst)
}

func DC8uvNoTop_C(dst *uint8) {  // DC with no top samples
  int dc0 = 4
  var i int
  for i = 0; i < 8; i++ {
    dc0 += dst[-1 + i * BPS]
  }
  Put8x8uv(dc0 >> 3, dst)
}

func DC8uvNoTopLeft_C(dst *uint8) {  // DC with nothing
  Put8x8uv(0x80, dst)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

VP8PredFunc VP8PredChroma8[NUM_B_DC_MODES]

//------------------------------------------------------------------------------
// Edge filtering functions

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
// 4 pixels in, 2 pixels out
func DoFilter2_C(p *uint8, step int) {
  p1 = p[-2 * step], p0 = p[-step], q0 = p[0], q1 := p[step]
  a := 3 * (q0 - p0) + VP8ksclip1[p1 - q1];  // in [-893,892]
  a1 := VP8ksclip2[(a + 4) >> 3];            // in [-16,15]
  a2 := VP8ksclip2[(a + 3) >> 3]
  p[-step] = VP8kclip1[p0 + a2]
  p[0] = VP8kclip1[q0 - a1]
}

// 4 pixels in, 4 pixels out
func DoFilter4_C(p *uint8, step int) {
  p1 = p[-2 * step], p0 = p[-step], q0 = p[0], q1 := p[step]
  a := 3 * (q0 - p0)
  a1 := VP8ksclip2[(a + 4) >> 3]
  a2 := VP8ksclip2[(a + 3) >> 3]
  a3 := (a1 + 1) >> 1
  p[-2 * step] = VP8kclip1[p1 + a3]
  p[-step] = VP8kclip1[p0 + a2]
  p[0] = VP8kclip1[q0 - a1]
  p[step] = VP8kclip1[q1 - a3]
}

// 6 pixels in, 6 pixels out
func DoFilter6_C(p *uint8, step int) {
  p2 = p[-3 * step], p1 = p[-2 * step], p0 := p[-step]
  q0 = p[0], q1 = p[step], q2 := p[2 * step]
  a := VP8ksclip1[3 * (q0 - p0) + VP8ksclip1[p1 - q1]]
  // a is in [-128,127], a1 in [-27,27], a2 in [-18,18] and a3 in [-9,9]
  a1 := (27 * a + 63) >> 7;  // eq. to ((3 * a + 7) * 9) >> 7
  a2 := (18 * a + 63) >> 7;  // eq. to ((2 * a + 7) * 9) >> 7
  a3 := (9 * a + 63) >> 7;   // eq. to ((1 * a + 7) * 9) >> 7
  p[-3 * step] = VP8kclip1[p2 + a3]
  p[-2 * step] = VP8kclip1[p1 + a2]
  p[-step] = VP8kclip1[p0 + a1]
  p[0] = VP8kclip1[q0 - a1]
  p[step] = VP8kclip1[q1 - a2]
  p[2 * step] = VP8kclip1[q2 - a3]
}

func Hev(/* const */ p *uint8, step int, thresh int) int {
  p1 = p[-2 * step], p0 = p[-step], q0 = p[0], q1 := p[step]
  return (VP8kabs0[p1 - p0] > thresh) || (VP8kabs0[q1 - q0] > thresh)
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

#if !WEBP_NEON_OMIT_C_CODE
static  int NeedsFilter_C(/* const */ p *uint8, step int, t int) {
  p1 = p[-2 * step], p0 = p[-step], q0 = p[0], q1 := p[step]
  return ((4 * VP8kabs0[p0 - q0] + VP8kabs0[p1 - q1]) <= t)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
static  int NeedsFilter2_C(/* const */ p *uint8, step int, t int, it int) {
  p3 = p[-4 * step], p2 = p[-3 * step], p1 := p[-2 * step]
  p0 = p[-step], q0 := p[0]
  q1 = p[step], q2 = p[2 * step], q3 := p[3 * step]
  if (4 * VP8kabs0[p0 - q0] + VP8kabs0[p1 - q1]) > t { return 0  }
  return VP8kabs0[p3 - p2] <= it && VP8kabs0[p2 - p1] <= it &&
         VP8kabs0[p1 - p0] <= it && VP8kabs0[q3 - q2] <= it &&
         VP8kabs0[q2 - q1] <= it && VP8kabs0[q1 - q0] <= it
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

//------------------------------------------------------------------------------
// Simple In-loop filtering (Paragraph 15.2)

#if !WEBP_NEON_OMIT_C_CODE
func SimpleVFilter16_C(p *uint8, stride int, thresh int) {
  var i int
  thresh2 := 2 * thresh + 1
  for i = 0; i < 16; i++ {
    if (NeedsFilter_C(p + i, stride, thresh2)) {
      DoFilter2_C(p + i, stride)
    }
  }
}

func SimpleHFilter16_C(p *uint8, stride int, thresh int) {
  var i int
  thresh2 := 2 * thresh + 1
  for i = 0; i < 16; i++ {
    if (NeedsFilter_C(p + i * stride, 1, thresh2)) {
      DoFilter2_C(p + i * stride, 1)
    }
  }
}

func SimpleVFilter16i_C(p *uint8, stride int, thresh int) {
  var k int
  for k = 3; k > 0; --k {
    p += 4 * stride
    SimpleVFilter16_C(p, stride, thresh)
  }
}

func SimpleHFilter16i_C(p *uint8, stride int, thresh int) {
  var k int
  for k = 3; k > 0; --k {
    p += 4
    SimpleHFilter16_C(p, stride, thresh)
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------
// Complex In-loop filtering (Paragraph 15.3)

func FilterLoop26_C(p *uint8, hstride int, vstride int, size int, thresh int, ithresh int, hev_thresh int) {
  thresh2 := 2 * thresh + 1
  while (size-- > 0) {
    if (NeedsFilter2_C(p, hstride, thresh2, ithresh)) {
      if (Hev(p, hstride, hev_thresh)) {
        DoFilter2_C(p, hstride)
      } else {
        DoFilter6_C(p, hstride)
      }
    }
    p += vstride
  }
}

func FilterLoop24_C(p *uint8, hstride int, vstride int, size int, thresh int, ithresh int, hev_thresh int) {
  thresh2 := 2 * thresh + 1
  while (size-- > 0) {
    if (NeedsFilter2_C(p, hstride, thresh2, ithresh)) {
      if (Hev(p, hstride, hev_thresh)) {
        DoFilter2_C(p, hstride)
      } else {
        DoFilter4_C(p, hstride)
      }
    }
    p += vstride
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

#if !WEBP_NEON_OMIT_C_CODE
// on macroblock edges
func VFilter16_C(p *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop26_C(p, stride, 1, 16, thresh, ithresh, hev_thresh)
}

func HFilter16_C(p *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop26_C(p, 1, stride, 16, thresh, ithresh, hev_thresh)
}

// on three inner edges
func VFilter16i_C(p *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  var k int
  for k = 3; k > 0; --k {
    p += 4 * stride
    FilterLoop24_C(p, stride, 1, 16, thresh, ithresh, hev_thresh)
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
func HFilter16i_C(p *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  var k int
  for k = 3; k > 0; --k {
    p += 4
    FilterLoop24_C(p, 1, stride, 16, thresh, ithresh, hev_thresh)
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

#if !WEBP_NEON_OMIT_C_CODE
// 8-pixels wide variant, for chroma filtering
func VFilter8_C(u *uint8, v *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop26_C(u, stride, 1, 8, thresh, ithresh, hev_thresh)
  FilterLoop26_C(v, stride, 1, 8, thresh, ithresh, hev_thresh)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
func HFilter8_C(u *uint8, v *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop26_C(u, 1, stride, 8, thresh, ithresh, hev_thresh)
  FilterLoop26_C(v, 1, stride, 8, thresh, ithresh, hev_thresh)
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

#if !WEBP_NEON_OMIT_C_CODE
func VFilter8i_C(u *uint8, v *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop24_C(u + 4 * stride, stride, 1, 8, thresh, ithresh, hev_thresh)
  FilterLoop24_C(v + 4 * stride, stride, 1, 8, thresh, ithresh, hev_thresh)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
func HFilter8i_C(u *uint8, v *uint8, stride int, thresh int, ithresh int, hev_thresh int) {
  FilterLoop24_C(u + 4, 1, stride, 8, thresh, ithresh, hev_thresh)
  FilterLoop24_C(v + 4, 1, stride, 8, thresh, ithresh, hev_thresh)
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

//------------------------------------------------------------------------------

func DitherCombine8x8_C(/* const */ dither *uint8, dst *uint8, dst_stride int) {
  var i, j int
  for j = 0; j < 8; j++ {
    for i = 0; i < 8; i++ {
      delta0 := dither[i] - VP8_DITHER_AMP_CENTER
      delta1 := (delta0 + VP8_DITHER_DESCALE_ROUNDER) >> VP8_DITHER_DESCALE
      dst[i] = clip_8b((int)dst[i] + delta1)
    }
    dst += dst_stride
    dither += 8
  }
}

//------------------------------------------------------------------------------

VP8DecIdct2 VP8Transform
VP8DecIdct VP8TransformAC3
VP8DecIdct VP8TransformUV
VP8DecIdct VP8TransformDC
VP8DecIdct VP8TransformDCUV

VP8LumaFilterFunc VP8VFilter16
VP8LumaFilterFunc VP8HFilter16
VP8ChromaFilterFunc VP8VFilter8
VP8ChromaFilterFunc VP8HFilter8
VP8LumaFilterFunc VP8VFilter16i
VP8LumaFilterFunc VP8HFilter16i
VP8ChromaFilterFunc VP8VFilter8i
VP8ChromaFilterFunc VP8HFilter8i
VP8SimpleFilterFunc VP8SimpleVFilter16
VP8SimpleFilterFunc VP8SimpleHFilter16
VP8SimpleFilterFunc VP8SimpleVFilter16i
VP8SimpleFilterFunc VP8SimpleHFilter16i

func (*VP8DitherCombine8x8)(/* const */ dither *uint8, dst *uint8, dst_stride int)


extern func VP8DspInitSSE2(void)
extern func VP8DspInitSSE41(void)
extern func VP8DspInitNEON(void)
extern func VP8DspInitMIPS32(void)
extern func VP8DspInitMIPSdspR2(void)
extern func VP8DspInitMSA(void)

WEBP_DSP_INIT_FUNC(VP8DspInit) {
  VP8InitClipTables()

#if !WEBP_NEON_OMIT_C_CODE
  VP8TransformWHT = TransformWHT_C
  VP8Transform = TransformTwo_C
  VP8TransformDC = TransformDC_C
  VP8TransformAC3 = TransformAC3_C
#endif
  VP8TransformUV = TransformUV_C
  VP8TransformDCUV = TransformDCUV_C

#if !WEBP_NEON_OMIT_C_CODE
  VP8VFilter16 = VFilter16_C
  VP8VFilter16i = VFilter16i_C
  VP8HFilter16 = HFilter16_C
  VP8VFilter8 = VFilter8_C
  VP8VFilter8i = VFilter8i_C
  VP8SimpleVFilter16 = SimpleVFilter16_C
  VP8SimpleHFilter16 = SimpleHFilter16_C
  VP8SimpleVFilter16i = SimpleVFilter16i_C
  VP8SimpleHFilter16i = SimpleHFilter16i_C
#endif

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
  VP8HFilter16i = HFilter16i_C
  VP8HFilter8 = HFilter8_C
  VP8HFilter8i = HFilter8i_C
#endif

#if !WEBP_NEON_OMIT_C_CODE
  VP8PredLuma4[0] = DC4_C
  VP8PredLuma4[1] = TM4_C
  VP8PredLuma4[2] = VE4_C
  VP8PredLuma4[4] = RD4_C
  VP8PredLuma4[6] = LD4_C
#endif

  VP8PredLuma4[3] = HE4_C
  VP8PredLuma4[5] = VR4_C
  VP8PredLuma4[7] = VL4_C
  VP8PredLuma4[8] = HD4_C
  VP8PredLuma4[9] = HU4_C

#if !WEBP_NEON_OMIT_C_CODE
  VP8PredLuma16[0] = DC16_C
  VP8PredLuma16[1] = TM16_C
  VP8PredLuma16[2] = VE16_C
  VP8PredLuma16[3] = HE16_C
  VP8PredLuma16[4] = DC16NoTop_C
  VP8PredLuma16[5] = DC16NoLeft_C
  VP8PredLuma16[6] = DC16NoTopLeft_C

  VP8PredChroma8[0] = DC8uv_C
  VP8PredChroma8[1] = TM8uv_C
  VP8PredChroma8[2] = VE8uv_C
  VP8PredChroma8[3] = HE8uv_C
  VP8PredChroma8[4] = DC8uvNoTop_C
  VP8PredChroma8[5] = DC8uvNoLeft_C
  VP8PredChroma8[6] = DC8uvNoTopLeft_C
#endif

  VP8DitherCombine8x8 = DitherCombine8x8_C

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if false
    if (VP8GetCPUInfo(kSSE2)) {
      VP8DspInitSSE2()
#if false
      if (VP8GetCPUInfo(kSSE4_1)) {
        VP8DspInitSSE41()
      }
#endif
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPS32)) {
      VP8DspInitMIPS32()
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      VP8DspInitMIPSdspR2()
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      VP8DspInitMSA()
    }
#endif
  }

#if FALSE
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    VP8DspInitNEON()
  }
#endif

  assert.Assert(VP8TransformWHT != nil)
  assert.Assert(VP8Transform != nil)
  assert.Assert(VP8TransformDC != nil)
  assert.Assert(VP8TransformAC3 != nil)
  assert.Assert(VP8TransformUV != nil)
  assert.Assert(VP8TransformDCUV != nil)
  assert.Assert(VP8VFilter16 != nil)
  assert.Assert(VP8HFilter16 != nil)
  assert.Assert(VP8VFilter8 != nil)
  assert.Assert(VP8HFilter8 != nil)
  assert.Assert(VP8VFilter16i != nil)
  assert.Assert(VP8HFilter16i != nil)
  assert.Assert(VP8VFilter8i != nil)
  assert.Assert(VP8HFilter8i != nil)
  assert.Assert(VP8SimpleVFilter16 != nil)
  assert.Assert(VP8SimpleHFilter16 != nil)
  assert.Assert(VP8SimpleVFilter16i != nil)
  assert.Assert(VP8SimpleHFilter16i != nil)
  assert.Assert(VP8PredLuma4[0] != nil)
  assert.Assert(VP8PredLuma4[1] != nil)
  assert.Assert(VP8PredLuma4[2] != nil)
  assert.Assert(VP8PredLuma4[3] != nil)
  assert.Assert(VP8PredLuma4[4] != nil)
  assert.Assert(VP8PredLuma4[5] != nil)
  assert.Assert(VP8PredLuma4[6] != nil)
  assert.Assert(VP8PredLuma4[7] != nil)
  assert.Assert(VP8PredLuma4[8] != nil)
  assert.Assert(VP8PredLuma4[9] != nil)
  assert.Assert(VP8PredLuma16[0] != nil)
  assert.Assert(VP8PredLuma16[1] != nil)
  assert.Assert(VP8PredLuma16[2] != nil)
  assert.Assert(VP8PredLuma16[3] != nil)
  assert.Assert(VP8PredLuma16[4] != nil)
  assert.Assert(VP8PredLuma16[5] != nil)
  assert.Assert(VP8PredLuma16[6] != nil)
  assert.Assert(VP8PredChroma8[0] != nil)
  assert.Assert(VP8PredChroma8[1] != nil)
  assert.Assert(VP8PredChroma8[2] != nil)
  assert.Assert(VP8PredChroma8[3] != nil)
  assert.Assert(VP8PredChroma8[4] != nil)
  assert.Assert(VP8PredChroma8[5] != nil)
  assert.Assert(VP8PredChroma8[6] != nil)
  assert.Assert(VP8DitherCombine8x8 != nil)
}
