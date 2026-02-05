package dsp

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Speed-critical encoding functions.
//
// Author: Skal (pascal.massimino@gmail.com)

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/libwebp/dsp"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
	"github.com/daanv2/go-webp/pkg/libwebp/utils"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/string"
) // for abs()

func clip_8b(v int) uint8 {
  return tenary.If(!(v & ~0xff),  v, tenary.If(v < 0, 0, 255))
}

func clip_max(v, max int ) { 
	return tenary.If(v > max, max, v)
}

//------------------------------------------------------------------------------
// Compute susceptibility based on DCT-coeff histograms:
// the higher, the "easier" the macroblock is to compress.

var VP8DspScan = [16 + 4 + 4]int{
    // Luma
    0 + 0 * constants.BPS,  4 + 0 * constants.BPS,  8 + 0 * constants.BPS,  12 + 0 * constants.BPS, 0 + 4 * constants.BPS,  4 + 4 * constants.BPS,  8 + 4 * constants.BPS,  12 + 4 * constants.BPS, 0 + 8 * constants.BPS,  4 + 8 * constants.BPS,  8 + 8 * constants.BPS,  12 + 8 * constants.BPS, 0 + 12 * constants.BPS, 4 + 12 * constants.BPS, 8 + 12 * constants.BPS, 12 + 12 * constants.BPS,

    0 + 0 * constants.BPS,  4 + 0 * constants.BPS,  0 + 4 * constants.BPS,  4 + 4 * constants.BPS,  // U
    8 + 0 * constants.BPS,  12 + 0 * constants.BPS, 8 + 4 * constants.BPS,  12 + 4 * constants.BPS,  // V
}

// general-purpose util function
func VP8SetHistogramData(/* const */ distribution [MAX_COEFF_THRESH + 1]int, /*const*/ histo *VP8Histogram) {
  max_value := 0
  last_non_zero := 1;
  var k int
  for k = 0; k <= MAX_COEFF_THRESH; k++ {
    value := distribution[k];
    if (value > 0) {
      if value > max_value { max_value = value }
      last_non_zero = k;
    }
  }
  histo.max_value = max_value;
  histo.last_non_zero = last_non_zero;
}

#if !WEBP_NEON_OMIT_C_CODE
func CollectHistogram_C(/* const */ ref *uint8, /*const*/ pred *uint8, start_block int, end_block int, /* const */ histo *VP8Histogram) {
  var j int
  int distribution[MAX_COEFF_THRESH + 1] = {0}
  for j = start_block; j < end_block; j++ {
    var k int
    int16 out[16];

    VP8FTransform(ref + VP8DspScan[j], pred + VP8DspScan[j], out);

    // Convert coefficients to bin.
    for k = 0; k < 16; k++ {
      v := abs(out[k]) >> 3;
      clipped_value := clip_max(v, MAX_COEFF_THRESH);
      ++distribution[clipped_value];
    }
  }
  VP8SetHistogramData(distribution, histo);
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------
// run-time tables (~4k)

static uint8 clip1[255 + 510 + 1];  // clips [-255,510] to [0,255]

// We declare this variable 'volatile' to prevent instruction reordering
// and make sure it's set to true _last_ (so as to be thread-safe)
static volatile tables_ok := 0;

static WEBP_TSAN_IGNORE_FUNCTION func InitTables(){
  if (!tables_ok) {
    var i int
    for i = -255; i <= 255 + 255; i++ {
      clip1[255 + i] = clip_8b(i);
    }
    tables_ok = 1;
  }
}

//------------------------------------------------------------------------------
// Transforms (Paragraph 14.4)

#if !WEBP_NEON_OMIT_C_CODE

#define STORE(x, y, v) \
  dst[(x) + (y) * constants.BPS] = clip_8b(ref[(x) + (y) * constants.BPS] + ((v) >> 3))

func ITransformOne(/* const */ ref *uint8, /*const*/ in *int16, dst *uint8) {
  int C[4 * 4], *tmp;
  var i int
  tmp = C;
  for i = 0; i < 4; i++ {  // vertical pass
    a := in[0] + in[8];
    b := in[0] - in[8];
    c :=
        WEBP_TRANSFORM_AC3_MUL2(in[4]) - WEBP_TRANSFORM_AC3_MUL1(in[12]);
    d :=
        WEBP_TRANSFORM_AC3_MUL1(in[4]) + WEBP_TRANSFORM_AC3_MUL2(in[12]);
    tmp[0] = a + d;
    tmp[1] = b + c;
    tmp[2] = b - c;
    tmp[3] = a - d;
    tmp += 4;
    in++;
  }

  tmp = C;
  for i = 0; i < 4; i++ {  // horizontal pass
    dc := tmp[0] + 4;
    a := dc + tmp[8];
    b := dc - tmp[8];
    c :=
        WEBP_TRANSFORM_AC3_MUL2(tmp[4]) - WEBP_TRANSFORM_AC3_MUL1(tmp[12]);
    d :=
        WEBP_TRANSFORM_AC3_MUL1(tmp[4]) + WEBP_TRANSFORM_AC3_MUL2(tmp[12]);
    STORE(0, i, a + d);
    STORE(1, i, b + c);
    STORE(2, i, b - c);
    STORE(3, i, a - d);
    tmp++;
  }
}

func ITransform_C(/* const */ ref *uint8, /*const*/ in *int16, dst *uint8, do_two int) {
  ITransformOne(ref, in, dst);
  if (do_two) {
    ITransformOne(ref + 4, in + 16, dst + 4);
  }
}

func FTransform_C(/* const */ src *uint8, /*const*/ ref *uint8, out *int16) {
  var i int
  int tmp[16];
  for i = 0; i < 4; ++i, src += constants.BPS, ref += constants.BPS {
    d0 := src[0] - ref[0];  // 9bit dynamic range ([-255,255])
    d1 := src[1] - ref[1];
    d2 := src[2] - ref[2];
    d3 := src[3] - ref[3];
    a0 := (d0 + d3);  // 10b [-510,510]
    a1 := (d1 + d2);
    a2 := (d1 - d2);
    a3 := (d0 - d3);
    tmp[0 + i * 4] = (a0 + a1) * 8;                        // 14b [-8160,8160]
    tmp[1 + i * 4] = (a2 * 2217 + a3 * 5352 + 1812) >> 9;  // [-7536,7542]
    tmp[2 + i * 4] = (a0 - a1) * 8;
    tmp[3 + i * 4] = (a3 * 2217 - a2 * 5352 + 937) >> 9;
  }
  for i = 0; i < 4; i++ {
    a0 := (tmp[0 + i] + tmp[12 + i]);  // 15b
    a1 := (tmp[4 + i] + tmp[8 + i]);
    a2 := (tmp[4 + i] - tmp[8 + i]);
    a3 := (tmp[0 + i] - tmp[12 + i]);
    out[0 + i] = (a0 + a1 + 7) >> 4;  // 12b
    out[4 + i] = ((a2 * 2217 + a3 * 5352 + 12000) >> 16) + (a3 != 0);
    out[8 + i] = (a0 - a1 + 7) >> 4;
    out[12 + i] = ((a3 * 2217 - a2 * 5352 + 51000) >> 16);
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func FTransform2_C(/* const */ src *uint8, /*const*/ ref *uint8, out *int16) {
  VP8FTransform(src, ref, out);
  VP8FTransform(src + 4, ref + 4, out + 16);
}

#if !WEBP_NEON_OMIT_C_CODE
func FTransformWHT_C(/* const */ in *int16, out *int16) {
  // input is 12b signed
  int32 tmp[16];
  var i int
  for i = 0; i < 4; ++i, in += 64 {
    a0 := (in[0 * 16] + in[2 * 16]);  // 13b
    a1 := (in[1 * 16] + in[3 * 16]);
    a2 := (in[1 * 16] - in[3 * 16]);
    a3 := (in[0 * 16] - in[2 * 16]);
    tmp[0 + i * 4] = a0 + a1;  // 14b
    tmp[1 + i * 4] = a3 + a2;
    tmp[2 + i * 4] = a3 - a2;
    tmp[3 + i * 4] = a0 - a1;
  }
  for i = 0; i < 4; i++ {
    a0 := (tmp[0 + i] + tmp[8 + i]);  // 15b
    a1 := (tmp[4 + i] + tmp[12 + i]);
    a2 := (tmp[4 + i] - tmp[12 + i]);
    a3 := (tmp[0 + i] - tmp[8 + i]);
    b0 := a0 + a1;  // 16b
    b1 := a3 + a2;
    b2 := a3 - a2;
    b3 := a0 - a1;
    out[0 + i] = b0 >> 1;  // 15b
    out[4 + i] = b1 >> 1;
    out[8 + i] = b2 >> 1;
    out[12 + i] = b3 >> 1;
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

#undef STORE

//------------------------------------------------------------------------------
// Intra predictions

func Fill(dst *uint8, value int, size int) {
  var j int
  for j = 0; j < size; j++ {
    stdlib.Memset(dst + j * constants.BPS, value, size);
  }
}

func VerticalPred(dst *uint8, /*const*/ top *uint8, size int) {
  var j int
  if (top != nil) {
    for (j = 0; j < size; ++j) memcpy(dst + j * constants.BPS, top, size);
  } else {
    Fill(dst, 127, size);
  }
}

func HorizontalPred(dst *uint8, /*const*/ left *uint8, size int) {
  if (left != nil) {
    var j int
    for j = 0; j < size; j++ {
      stdlib.Memset(dst + j * constants.BPS, left[j], size);
    }
  } else {
    Fill(dst, 129, size);
  }
}

func TrueMotion(dst *uint8, /*const*/ left *uint8, /*const*/ top *uint8, size int) {
  var y int
  if (left != nil) {
    if (top != nil) {
      var clip *uint8 = clip1 + 255 - left[-1];
      for y = 0; y < size; y++ {
        var clip_table *uint8 = clip + left[y];
        var x int
        for x = 0; x < size; x++ {
          dst[x] = clip_table[top[x]];
        }
        dst += constants.BPS;
      }
    } else {
      HorizontalPred(dst, left, size);
    }
  } else {
    // true motion without left samples (hence: with default 129 value)
    // is equivalent to VE prediction where you just copy the top samples.
    // Note that if top samples are not available, the default value is
    // then 129, and not 127 as in the VerticalPred case.
    if (top != nil) {
      VerticalPred(dst, top, size);
    } else {
      Fill(dst, 129, size);
    }
  }
}

func DCMode(dst *uint8, /*const*/ left *uint8, /*const*/ top *uint8, size int, round int, shift int) {
  DC := 0;
  var j int
  if (top != nil) {
    for (j = 0; j < size; ++j) DC += top[j];
    if (left != nil) {  // top and left present
      for (j = 0; j < size; ++j) DC += left[j];
    } else {  // top, but no left
      DC += DC;
    }
    DC = (DC + round) >> shift;
  } else if (left != nil) {  // left but no top
    for (j = 0; j < size; ++j) DC += left[j];
    DC += DC;
    DC = (DC + round) >> shift;
  } else {  // no top, no left, nothing.
    DC = 0x80;
  }
  Fill(dst, DC, size);
}

//------------------------------------------------------------------------------
// Chroma 8x8 prediction (paragraph 12.2)

func IntraChromaPreds_C(dst *uint8, /*const*/ left *uint8, /*const*/ top *uint8) {
  // U block
  DCMode(C8DC8 + dst, left, top, 8, 8, 4);
  VerticalPred(C8VE8 + dst, top, 8);
  HorizontalPred(C8HE8 + dst, left, 8);
  TrueMotion(C8TM8 + dst, left, top, 8);
  // V block
  dst += 8;
  if top != nil { top += 8 }
  if left != nil { left += 16 }
  DCMode(C8DC8 + dst, left, top, 8, 8, 4);
  VerticalPred(C8VE8 + dst, top, 8);
  HorizontalPred(C8HE8 + dst, left, 8);
  TrueMotion(C8TM8 + dst, left, top, 8);
}

//------------------------------------------------------------------------------
// luma 16x16 prediction (paragraph 12.3)

#if !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64
func Intra16Preds_C(dst *uint8, /*const*/ left *uint8, /*const*/ top *uint8) {
  DCMode(I16DC16 + dst, left, top, 16, 16, 5);
  VerticalPred(I16VE16 + dst, top, 16);
  HorizontalPred(I16HE16 + dst, left, 16);
  TrueMotion(I16TM16 + dst, left, top, 16);
}
#endif  // !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64

//------------------------------------------------------------------------------
// luma 4x4 prediction

#if !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64 || constants.BPS != 32

#define DST(x, y) dst[(x) + (y) * constants.BPS]
#define AVG3(a, b, c) ((uint8)(((a) + 2 * (b) + (c) + 2) >> 2))
#define AVG2(a, b) (((a) + (b) + 1) >> 1)

// vertical
func VE4(dst *uint8, /*const*/ top *uint8) {
  vals[4] := {
      AVG3(top[-1], top[0], top[1]), AVG3(top[0], top[1], top[2]), AVG3(top[1], top[2], top[3]), AVG3(top[2], top[3], top[4]), }
  var i int
  for i = 0; i < 4; i++ {
    memcpy(dst + i * constants.BPS, vals, 4);
  }
}

// horizontal
func HE4(dst *uint8, /*const*/ top *uint8) {
  X := top[-1];
  I := top[-2];
  J := top[-3];
  K := top[-4];
  L := top[-5];
  WebPUint32ToMem(dst + 0 * constants.BPS, uint(0x01010101) * AVG3(X, I, J));
  WebPUint32ToMem(dst + 1 * constants.BPS, uint(0x01010101) * AVG3(I, J, K));
  WebPUint32ToMem(dst + 2 * constants.BPS, uint(0x01010101) * AVG3(J, K, L));
  WebPUint32ToMem(dst + 3 * constants.BPS, uint(0x01010101) * AVG3(K, L, L));
}

func DC4(dst *uint8, /*const*/ top *uint8) {
  dc := 4;
  var i int
  for (i = 0; i < 4; ++i) dc += top[i] + top[-5 + i];
  Fill(dst, dc >> 3, 4);
}

func RD4(dst *uint8, /*const*/ top *uint8) {
  X := top[-1];
  I := top[-2];
  J := top[-3];
  K := top[-4];
  L := top[-5];
  A := top[0];
  B := top[1];
  C := top[2];
  D := top[3];
  DST(0, 3) = AVG3(J, K, L);
  DST(0, 2) = DST(1, 3) = AVG3(I, J, K);
  DST(0, 1) = DST(1, 2) = DST(2, 3) = AVG3(X, I, J);
  DST(0, 0) = DST(1, 1) = DST(2, 2) = DST(3, 3) = AVG3(A, X, I);
  DST(1, 0) = DST(2, 1) = DST(3, 2) = AVG3(B, A, X);
  DST(2, 0) = DST(3, 1) = AVG3(C, B, A);
  DST(3, 0) = AVG3(D, C, B);
}

func LD4(dst *uint8, /*const*/ top *uint8) {
  A := top[0];
  B := top[1];
  C := top[2];
  D := top[3];
  E := top[4];
  F := top[5];
  G := top[6];
  H := top[7];
  DST(0, 0) = AVG3(A, B, C);
  DST(1, 0) = DST(0, 1) = AVG3(B, C, D);
  DST(2, 0) = DST(1, 1) = DST(0, 2) = AVG3(C, D, E);
  DST(3, 0) = DST(2, 1) = DST(1, 2) = DST(0, 3) = AVG3(D, E, F);
  DST(3, 1) = DST(2, 2) = DST(1, 3) = AVG3(E, F, G);
  DST(3, 2) = DST(2, 3) = AVG3(F, G, H);
  DST(3, 3) = AVG3(G, H, H);
}

func VR4(dst *uint8, /*const*/ top *uint8) {
  X := top[-1];
  I := top[-2];
  J := top[-3];
  K := top[-4];
  A := top[0];
  B := top[1];
  C := top[2];
  D := top[3];
  DST(0, 0) = DST(1, 2) = AVG2(X, A);
  DST(1, 0) = DST(2, 2) = AVG2(A, B);
  DST(2, 0) = DST(3, 2) = AVG2(B, C);
  DST(3, 0) = AVG2(C, D);

  DST(0, 3) = AVG3(K, J, I);
  DST(0, 2) = AVG3(J, I, X);
  DST(0, 1) = DST(1, 3) = AVG3(I, X, A);
  DST(1, 1) = DST(2, 3) = AVG3(X, A, B);
  DST(2, 1) = DST(3, 3) = AVG3(A, B, C);
  DST(3, 1) = AVG3(B, C, D);
}

func VL4(dst *uint8, /*const*/ top *uint8) {
  A := top[0];
  B := top[1];
  C := top[2];
  D := top[3];
  E := top[4];
  F := top[5];
  G := top[6];
  H := top[7];
  DST(0, 0) = AVG2(A, B);
  DST(1, 0) = DST(0, 2) = AVG2(B, C);
  DST(2, 0) = DST(1, 2) = AVG2(C, D);
  DST(3, 0) = DST(2, 2) = AVG2(D, E);

  DST(0, 1) = AVG3(A, B, C);
  DST(1, 1) = DST(0, 3) = AVG3(B, C, D);
  DST(2, 1) = DST(1, 3) = AVG3(C, D, E);
  DST(3, 1) = DST(2, 3) = AVG3(D, E, F);
  DST(3, 2) = AVG3(E, F, G);
  DST(3, 3) = AVG3(F, G, H);
}

func HU4(dst *uint8, /*const*/ top *uint8) {
  I := top[-2];
  J := top[-3];
  K := top[-4];
  L := top[-5];
  DST(0, 0) = AVG2(I, J);
  DST(2, 0) = DST(0, 1) = AVG2(J, K);
  DST(2, 1) = DST(0, 2) = AVG2(K, L);
  DST(1, 0) = AVG3(I, J, K);
  DST(3, 0) = DST(1, 1) = AVG3(J, K, L);
  DST(3, 1) = DST(1, 2) = AVG3(K, L, L);
  DST(3, 2) = DST(2, 2) = DST(0, 3) = DST(1, 3) = DST(2, 3) = DST(3, 3) = L;
}

func HD4(dst *uint8, /*const*/ top *uint8) {
  X := top[-1];
  I := top[-2];
  J := top[-3];
  K := top[-4];
  L := top[-5];
  A := top[0];
  B := top[1];
  C := top[2];

  DST(0, 0) = DST(2, 1) = AVG2(I, X);
  DST(0, 1) = DST(2, 2) = AVG2(J, I);
  DST(0, 2) = DST(2, 3) = AVG2(K, J);
  DST(0, 3) = AVG2(L, K);

  DST(3, 0) = AVG3(A, B, C);
  DST(2, 0) = AVG3(X, A, B);
  DST(1, 0) = DST(3, 1) = AVG3(I, X, A);
  DST(1, 1) = DST(3, 2) = AVG3(J, I, X);
  DST(1, 2) = DST(3, 3) = AVG3(K, J, I);
  DST(1, 3) = AVG3(L, K, J);
}

func TM4(dst *uint8, /*const*/ top *uint8) {
  var x, y int
  var clip *uint8 = clip1 + 255 - top[-1];
  for y = 0; y < 4; y++ {
    var clip_table *uint8 = clip + top[-2 - y];
    for x = 0; x < 4; x++ {
      dst[x] = clip_table[top[x]];
    }
    dst += constants.BPS;
  }
}

#undef DST
#undef AVG3
#undef AVG2

// Left samples are top[-5 .. -2], top_left is top[-1], top are
// located at top[0..3], and top right is top[4..7]
func Intra4Preds_C(dst *uint8, /*const*/ top *uint8) {
  DC4(I4DC4 + dst, top);
  TM4(I4TM4 + dst, top);
  VE4(I4VE4 + dst, top);
  HE4(I4HE4 + dst, top);
  RD4(I4RD4 + dst, top);
  VR4(I4VR4 + dst, top);
  LD4(I4LD4 + dst, top);
  VL4(I4VL4 + dst, top);
  HD4(I4HD4 + dst, top);
  HU4(I4HU4 + dst, top);
}

#endif  // !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64 || constants.BPS != 32

//------------------------------------------------------------------------------
// Metric

#if !WEBP_NEON_OMIT_C_CODE
func GetSSE(/* const */ a *uint8, /*const*/ b *uint8, w int, h int) int {
  count := 0;
  int y, x;
  for y = 0; y < h; y++ {
    for x = 0; x < w; x++ {
      diff := (int)a[x] - b[x];
      count += diff * diff;
    }
    a += constants.BPS;
    b += constants.BPS;
  }
  return count;
}

func SSE16x16_C(/* const */ a *uint8, /*const*/ b *uint8) int {
  return GetSSE(a, b, 16, 16);
}
func SSE16x8_C(/* const */ a *uint8, /*const*/ b *uint8) int {
  return GetSSE(a, b, 16, 8);
}
func SSE8x8_C(/* const */ a *uint8, /*const*/ b *uint8) int {
  return GetSSE(a, b, 8, 8);
}
func SSE4x4_C(/* const */ a *uint8, /*const*/ b *uint8) int {
  return GetSSE(a, b, 4, 4);
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func Mean16x4_C(/* const */ ref *uint8, uint32 dc[4]) {
  int k, x, y;
  for k = 0; k < 4; k++ {
    avg := 0;
    for y = 0; y < 4; y++ {
      for x = 0; x < 4; x++ {
        avg += ref[x + y * constants.BPS];
      }
    }
    dc[k] = avg;
    ref += 4;  // go to next 4x4 block.
  }
}

//------------------------------------------------------------------------------
// Texture distortion
//
// We try to match the spectral content (weighted) between source and
// reconstructed samples.

#if !WEBP_NEON_OMIT_C_CODE
// Hadamard transform
// Returns the weighted sum of the absolute value of transformed coefficients.
// w[] contains a row-major 4 by 4 symmetric matrix.
func TTransform(/* const */ in *uint8, /*const*/ w *uint16) int {
  sum := 0;
  int tmp[16];
  var i int
  // horizontal pass
  for i = 0; i < 4; ++i, in += constants.BPS {
    a0 := in[0] + in[2];
    a1 := in[1] + in[3];
    a2 := in[1] - in[3];
    a3 := in[0] - in[2];
    tmp[0 + i * 4] = a0 + a1;
    tmp[1 + i * 4] = a3 + a2;
    tmp[2 + i * 4] = a3 - a2;
    tmp[3 + i * 4] = a0 - a1;
  }
  // vertical pass
  for i = 0; i < 4; ++i, w++ {
    a0 := tmp[0 + i] + tmp[8 + i];
    a1 := tmp[4 + i] + tmp[12 + i];
    a2 := tmp[4 + i] - tmp[12 + i];
    a3 := tmp[0 + i] - tmp[8 + i];
    b0 := a0 + a1;
    b1 := a3 + a2;
    b2 := a3 - a2;
    b3 := a0 - a1;

    sum += w[0] * abs(b0);
    sum += w[4] * abs(b1);
    sum += w[8] * abs(b2);
    sum += w[12] * abs(b3);
  }
  return sum;
}

func Disto4x4_C(/* const */ /* const */ a *uint8, /*const*/ /* const */ b *uint8, /*const*/ /* const */ w *uint16) int {
  sum1 := TTransform(a, w);
  sum2 := TTransform(b, w);
  return abs(sum2 - sum1) >> 5;
}

func Disto16x16_C(/* const */ /* const */ a *uint8, /*const*/ /* const */ b *uint8, /*const*/ /* const */ w *uint16) int {
  D := 0;
  var x, y int
  for y = 0; y < 16 * constants.BPS; y += 4 * constants.BPS {
    for x = 0; x < 16; x += 4 {
      D += Disto4x4_C(a + x + y, b + x + y, w);
    }
  }
  return D;
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------
// Quantization
//

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
var kZigzag = [16]uint8 = {0, 1,  4,  8,  5, 2,  3,  6, 9, 12, 13, 10, 7, 11, 14, 15}

// Simple quantization
func QuantizeBlock_C(int16 in[16], int16 out[16], /*const*/ /* const */ mtx *VP8Matrix) int {
  last := -1;
  var n int
  for n = 0; n < 16; n++ {
    j := kZigzag[n];
    sign := (in[j] < 0);
    coeff := (sign ? -in[j] : in[j]) + mtx.sharpen[j];
    if (coeff > mtx.zthresh[j]) {
      Q := mtx.q[j];
      iQ := mtx.iq[j];
      B := mtx.bias[j];
      level := QUANTDIV(coeff, iQ, B);
      if level > MAX_LEVEL { level = MAX_LEVEL }
      if sign { level = -level }
      in[j] = level * (int)Q;
      out[n] = level;
      if level { last = n }
    } else {
      out[n] = 0;
      in[j] = 0;
    }
  }
  return (last >= 0);
}

func Quantize2Blocks_C(int16 in[32], int16 out[32], /*const*/ /* const */ mtx *VP8Matrix) int {
  var nz int
  nz = VP8EncQuantizeBlock(in + 0 * 16, out + 0 * 16, mtx) << 0;
  nz |= VP8EncQuantizeBlock(in + 1 * 16, out + 1 * 16, mtx) << 1;
  return nz;
}
#endif  // !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC

//------------------------------------------------------------------------------
// Block copy

func Copy(/* const */ src *uint8, dst *uint8, w int, h int) {
  var y int
  for y = 0; y < h; y++ {
    memcpy(dst, src, w);
    src += constants.BPS;
    dst += constants.BPS;
  }
}

func Copy4x4_C(/* const */ src *uint8, dst *uint8) {
  Copy(src, dst, 4, 4);
}

func Copy16x8_C(/* const */ src *uint8, dst *uint8) {
  Copy(src, dst, 16, 8);
}

//------------------------------------------------------------------------------
// Initialization

// Speed-critical function pointers. We have to initialize them to the default
// implementations within VP8EncDspInit().
VP8CHisto VP8CollectHistogram;
VP8Idct VP8ITransform;
VP8Fdct VP8FTransform;
VP8Fdct VP8FTransform2;
VP8WHT VP8FTransformWHT;
VP8Intra4Preds VP8EncPredLuma4;
VP8IntraPreds VP8EncPredLuma16;
VP8IntraPreds VP8EncPredChroma8;
VP8Metric VP8SSE16x16;
VP8Metric VP8SSE8x8;
VP8Metric VP8SSE16x8;
VP8Metric VP8SSE4x4;
VP8WMetric VP8TDisto4x4;
VP8WMetric VP8TDisto16x16;
VP8MeanMetric VP8Mean16x4;
VP8QuantizeBlock VP8EncQuantizeBlock;
VP8Quantize2Blocks VP8EncQuantize2Blocks;
VP8QuantizeBlockWHT VP8EncQuantizeBlockWHT;
VP8BlockCopy VP8Copy4x4;
VP8BlockCopy VP8Copy16x8;

extern VP8CPUInfo VP8GetCPUInfo;
extern func VP8EncDspInitSSE2(void);
extern func VP8EncDspInitSSE41(void);
extern func VP8EncDspInitNEON(void);
extern func VP8EncDspInitMIPS32(void);
extern func VP8EncDspInitMIPSdspR2(void);
extern func VP8EncDspInitMSA(void);

WEBP_DSP_INIT_FUNC(VP8EncDspInit) {
  VP8DspInit();  // common inverse transforms
  InitTables();

  // default C implementations
#if !WEBP_NEON_OMIT_C_CODE
  VP8ITransform = ITransform_C;
  VP8FTransform = FTransform_C;
  VP8FTransformWHT = FTransformWHT_C;
  VP8TDisto4x4 = Disto4x4_C;
  VP8TDisto16x16 = Disto16x16_C;
  VP8CollectHistogram = CollectHistogram_C;
  VP8SSE16x16 = SSE16x16_C;
  VP8SSE16x8 = SSE16x8_C;
  VP8SSE8x8 = SSE8x8_C;
  VP8SSE4x4 = SSE4x4_C;
#endif

#if !WEBP_NEON_OMIT_C_CODE || WEBP_NEON_WORK_AROUND_GCC
  VP8EncQuantizeBlock = QuantizeBlock_C;
  VP8EncQuantize2Blocks = Quantize2Blocks_C;
  VP8EncQuantizeBlockWHT = QuantizeBlock_C;
#endif

#if !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64 || constants.BPS != 32
  VP8EncPredLuma4 = Intra4Preds_C;
#endif
#if !WEBP_NEON_OMIT_C_CODE || !WEBP_AARCH64
  VP8EncPredLuma16 = Intra16Preds_C;
#endif

  VP8FTransform2 = FTransform2_C;
  VP8EncPredChroma8 = IntraChromaPreds_C;
  VP8Mean16x4 = Mean16x4_C;
  VP8Copy4x4 = Copy4x4_C;
  VP8Copy16x8 = Copy16x8_C;

  // If defined, use CPUInfo() to overwrite some pointers with faster versions.
  if (VP8GetCPUInfo != nil) {
#if defined(WEBP_HAVE_SSE2)
    if (VP8GetCPUInfo(kSSE2)) {
      VP8EncDspInitSSE2();
#if defined(WEBP_HAVE_SSE41)
      if (VP8GetCPUInfo(kSSE4_1)) {
        VP8EncDspInitSSE41();
      }
#endif
    }
#endif
#if defined(WEBP_USE_MIPS32)
    if (VP8GetCPUInfo(kMIPS32)) {
      VP8EncDspInitMIPS32();
    }
#endif
#if defined(WEBP_USE_MIPS_DSP_R2)
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      VP8EncDspInitMIPSdspR2();
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      VP8EncDspInitMSA();
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    VP8EncDspInitNEON();
  }
#endif

  assert.Assert(VP8ITransform != nil);
  assert.Assert(VP8FTransform != nil);
  assert.Assert(VP8FTransformWHT != nil);
  assert.Assert(VP8TDisto4x4 != nil);
  assert.Assert(VP8TDisto16x16 != nil);
  assert.Assert(VP8CollectHistogram != nil);
  assert.Assert(VP8SSE16x16 != nil);
  assert.Assert(VP8SSE16x8 != nil);
  assert.Assert(VP8SSE8x8 != nil);
  assert.Assert(VP8SSE4x4 != nil);
  assert.Assert(VP8EncQuantizeBlock != nil);
  assert.Assert(VP8EncQuantize2Blocks != nil);
  assert.Assert(VP8FTransform2 != nil);
  assert.Assert(VP8EncPredLuma4 != nil);
  assert.Assert(VP8EncPredLuma16 != nil);
  assert.Assert(VP8EncPredChroma8 != nil);
  assert.Assert(VP8Mean16x4 != nil);
  assert.Assert(VP8EncQuantizeBlockWHT != nil);
  assert.Assert(VP8Copy4x4 != nil);
  assert.Assert(VP8Copy16x8 != nil);
}
