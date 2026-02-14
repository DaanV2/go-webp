// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
	"github.com/daanv2/go-webp/pkg/vp8"
)

func clip_8b(v int) uint8 {
	return tenary.If(!(v & ~0xff), v, tenary.If(v < 0, 0, 255))
}

func clip_max(v, max int) {
	return tenary.If(v > max, max, v)
}

//------------------------------------------------------------------------------
// Compute susceptibility based on DCT-coeff histograms:
// the higher, the "easier" the macroblock is to compress.

var VP8DspScan = [16 + 4 + 4]int{
	// Luma
	0 + 0*constants.BPS, 4 + 0*constants.BPS, 8 + 0*constants.BPS, 12 + 0*constants.BPS, 0 + 4*constants.BPS, 4 + 4*constants.BPS, 8 + 4*constants.BPS, 12 + 4*constants.BPS, 0 + 8*constants.BPS, 4 + 8*constants.BPS, 8 + 8*constants.BPS, 12 + 8*constants.BPS, 0 + 12*constants.BPS, 4 + 12*constants.BPS, 8 + 12*constants.BPS, 12 + 12*constants.BPS,

	0 + 0*constants.BPS, 4 + 0*constants.BPS, 0 + 4*constants.BPS, 4 + 4*constants.BPS, // U
	8 + 0*constants.BPS, 12 + 0*constants.BPS, 8 + 4*constants.BPS, 12 + 4*constants.BPS, // V
}

// general-purpose util function
func VP8SetHistogramData( /* const */ distribution [MAX_COEFF_THRESH + 1]int /*const*/, histo *VP8Histogram) {
	max_value := 0
	last_non_zero := 1
	var k int
	for k = 0; k <= MAX_COEFF_THRESH; k++ {
		value := distribution[k]
		if value > 0 {
			if value > max_value {
				max_value = value
			}
			last_non_zero = k
		}
	}
	histo.max_value = max_value
	histo.last_non_zero = last_non_zero
}

func CollectHistogram_C( /* const */ ref *uint8 /*const*/, pred *uint8, start_block int, end_block int /* const */, histo *VP8Histogram) {
	var j int
	var distribution = [MAX_COEFF_THRESH + 1]int{}
	for j = start_block; j < end_block; j++ {
		var k int
		var out [16]int16

		VP8FTransform(ref+VP8DspScan[j], pred+VP8DspScan[j], out)

		// Convert coefficients to bin.
		for k = 0; k < 16; k++ {
			v := stdlib.Abs(out[k]) >> 3
			clipped_value := clip_max(v, MAX_COEFF_THRESH)
			distribution[clipped_value] = distribution[clipped_value] + 1
		}
	}
	VP8SetHistogramData(distribution, histo)
}

// var clip1 [255 + 510 + 1]uint8  // clips [-255,510] to [0,255]
//go:fix inline
func clip1(v int) uint8 {
	return clip_8b(v)
}

func STORE(x, y, v int, dst []int) {
	dst[(x)+(y)*constants.BPS] = clip_8b(ref[(x)+(y)*constants.BPS] + ((v) >> 3))
}

func ITransformOne( /* const */ ref []uint8 /*const*/, in []int16, dst []uint8) {
	var C [4 * 4]int
	var tmp []int
	var i int
	tmp = C[0:]
	for i = 0; i < 4; i++ { // vertical pass
		a := in[0] + in[8]
		b := in[0] - in[8]
		c := WEBP_TRANSFORM_AC3_MUL2(in[4]) - WEBP_TRANSFORM_AC3_MUL1(in[12])
		d := WEBP_TRANSFORM_AC3_MUL1(in[4]) + WEBP_TRANSFORM_AC3_MUL2(in[12])
		tmp[0] = a + d
		tmp[1] = b + c
		tmp[2] = b - c
		tmp[3] = a - d
		tmp = tmp[4:]
		in++
	}

	tmp = C
	for i = 0; i < 4; i++ { // horizontal pass
		dc := tmp[0] + 4
		a := dc + tmp[8]
		b := dc - tmp[8]
		c := WEBP_TRANSFORM_AC3_MUL2(tmp[4]) - WEBP_TRANSFORM_AC3_MUL1(tmp[12])
		d := WEBP_TRANSFORM_AC3_MUL1(tmp[4]) + WEBP_TRANSFORM_AC3_MUL2(tmp[12])
		STORE(0, i, a+d)
		STORE(1, i, b+c)
		STORE(2, i, b-c)
		STORE(3, i, a-d)
		tmp = tmp[1:]
	}
}

func ITransform_C( /* const */ ref *uint8 /*const*/, in *int16, dst []uint8, do_two int) {
	ITransformOne(ref, in, dst)
	if do_two {
		ITransformOne(ref+4, in+16, dst+4)
	}
}

func FTransform_C( /* const */ src *uint8 /*const*/, ref *uint8, out *int16) {
	var i int
	var tmp [16]int

	for i = 0; i < 4; {
		d0 := src[0] - ref[0] // 9bit dynamic range ([-255,255])
		d1 := src[1] - ref[1]
		d2 := src[2] - ref[2]
		d3 := src[3] - ref[3]
		a0 := (d0 + d3) // 10b [-510,510]
		a1 := (d1 + d2)
		a2 := (d1 - d2)
		a3 := (d0 - d3)
		tmp[0+i*4] = (a0 + a1) * 8                   // 14b [-8160,8160]
		tmp[1+i*4] = (a2*2217 + a3*5352 + 1812) >> 9 // [-7536,7542]
		tmp[2+i*4] = (a0 - a1) * 8
		tmp[3+i*4] = (a3*2217 - a2*5352 + 937) >> 9

		i++
		src += constants.BPS
		ref += constants.BPS
	}
	for i = 0; i < 4; i++ {
		a0 := (tmp[0+i] + tmp[12+i]) // 15b
		a1 := (tmp[4+i] + tmp[8+i])
		a2 := (tmp[4+i] - tmp[8+i])
		a3 := (tmp[0+i] - tmp[12+i])
		out[0+i] = (a0 + a1 + 7) >> 4 // 12b
		out[4+i] = ((a2*2217 + a3*5352 + 12000) >> 16) + (a3 != 0)
		out[8+i] = (a0 - a1 + 7) >> 4
		out[12+i] = ((a3*2217 - a2*5352 + 51000) >> 16)
	}
}

func FTransform2_C( /* const */ src *uint8 /*const*/, ref *uint8, out *int16) {
	VP8FTransform(src, ref, out)
	VP8FTransform(src+4, ref+4, out+16)
}

func FTransformWHT_C( /* const */ in *int16, out *int16) {
	// input is 12b signed
	var tmp [16]int32
	var i int

	for i = 0; i < 4; {
		a0 := (in[0*16] + in[2*16]) // 13b
		a1 := (in[1*16] + in[3*16])
		a2 := (in[1*16] - in[3*16])
		a3 := (in[0*16] - in[2*16])
		tmp[0+i*4] = a0 + a1 // 14b
		tmp[1+i*4] = a3 + a2
		tmp[2+i*4] = a3 - a2
		tmp[3+i*4] = a0 - a1

		i++
		in += 64
	}
	for i = 0; i < 4; i++ {
		a0 := (tmp[0+i] + tmp[8+i]) // 15b
		a1 := (tmp[4+i] + tmp[12+i])
		a2 := (tmp[4+i] - tmp[12+i])
		a3 := (tmp[0+i] - tmp[8+i])
		b0 := a0 + a1 // 16b
		b1 := a3 + a2
		b2 := a3 - a2
		b3 := a0 - a1
		out[0+i] = b0 >> 1 // 15b
		out[4+i] = b1 >> 1
		out[8+i] = b2 >> 1
		out[12+i] = b3 >> 1
	}
}

func Fill(dst []uint8, value int, size int) {
	var j int
	for j = 0; j < size; j++ {
		stdlib.Memset(dst+j*constants.BPS, value, size)
	}
}

func VerticalPred(dst []uint8 /*const*/, top []uint8, size int) {
	var j int
	if top != nil {
		for j = 0; j < size; j++ {
			stdlib.MemCpy(dst+j*constants.BPS, top, size)
		}
	} else {
		Fill(dst, 127, size)
	}
}

func HorizontalPred(dst []uint8 /*const*/, left []uint8, size int) {
	if left != nil {
		var j int
		for j = 0; j < size; j++ {
			stdlib.Memset(dst+j*constants.BPS, left[j], size)
		}
	} else {
		Fill(dst, 129, size)
	}
}

func TrueMotion(dst []uint8 /*const*/, left []uint8 /*const*/, top []uint8, size int) {
	var y int
	if left != nil {
		if top != nil {
			for y = 0; y < size; y++ {
				var x int
				for x = 0; x < size; x++ {
					dst[x] = clip1(top[x])
				}
				dst += constants.BPS
			}
		} else {
			HorizontalPred(dst, left, size)
		}
	} else {
		// true motion without left samples (hence: with default 129 value)
		// is equivalent to VE prediction where you just copy the top samples.
		// Note that if top samples are not available, the default value is
		// then 129, and not 127 as in the VerticalPred case.
		if top != nil {
			VerticalPred(dst, top, size)
		} else {
			Fill(dst, 129, size)
		}
	}
}

func DCMode(dst []uint8 /*const*/, left []uint8 /*const*/, top []uint8, size int, round int, shift int) {
	DC := 0
	var j int
	if top != nil {
		for j := 0; j < size; j++ {
			DC += top[j]
		}
		if left != nil { // top and left present
			for j = 0; j < size; j++ {
				DC += left[j]
			}
		} else { // top, but no left
			DC += DC
		}
		DC = (DC + round) >> shift
	} else if left != nil { // left but no top
		for j = 0; j < size; j++ {
			DC += left[j]
		}
		DC += DC
		DC = (DC + round) >> shift
	} else { // no top, no left, nothing.
		DC = 0x80
	}
	Fill(dst, DC, size)
}

func IntraChromaPreds_C(dst []uint8 /*const*/, left []uint8 /*const*/, top []uint8) {
	// U block
	DCMode(C8DC8+dst, left, top, 8, 8, 4)
	VerticalPred(C8VE8+dst, top, 8)
	HorizontalPred(C8HE8+dst, left, 8)
	TrueMotion(C8TM8+dst, left, top, 8)
	// V block
	dst += 8
	if top != nil {
		top += 8
	}
	if left != nil {
		left += 16
	}
	DCMode(C8DC8+dst, left, top, 8, 8, 4)
	VerticalPred(C8VE8+dst, top, 8)
	HorizontalPred(C8HE8+dst, left, 8)
	TrueMotion(C8TM8+dst, left, top, 8)
}

func Intra16Preds_C(dst []uint8 /*const*/, left []uint8 /*const*/, top []uint8) {
	DCMode(I16DC16+dst, left, top, 16, 16, 5)
	VerticalPred(I16VE16+dst, top, 16)
	HorizontalPred(I16HE16+dst, left, 16)
	TrueMotion(I16TM16+dst, left, top, 16)
}

func DSTIndex(x, y int) int    { return (x) + (y)*constants.BPS }
func AVG3(a, b, c uint8) uint8 { return uint8((((a) + 2*(b) + (c) + 2) >> 2)) }
func AVG2(a, b uint8) uint8    { return (((a) + (b) + 1) >> 1) }

// vertical
func VE4(dst []uint8 /*const*/, top []uint8) {
	var vals = [4]int{
		AVG3(top[-1], top[0], top[1]),
		AVG3(top[0], top[1], top[2]),
		AVG3(top[1], top[2], top[3]),
		AVG3(top[2], top[3], top[4]),
	}
	var i int
	for i = 0; i < 4; i++ {
		stdlib.MemCpy(dst+i*constants.BPS, vals, 4)
	}
}

// horizontal
func HE4(dst []uint8 /*const*/, top []uint8) {
	X := top[-1]
	I := top[-2]
	J := top[-3]
	K := top[-4]
	L := top[-5]
	WebPUint32ToMem(dst+0*constants.BPS, uint(0x01010101)*AVG3(X, I, J))
	WebPUint32ToMem(dst+1*constants.BPS, uint(0x01010101)*AVG3(I, J, K))
	WebPUint32ToMem(dst+2*constants.BPS, uint(0x01010101)*AVG3(J, K, L))
	WebPUint32ToMem(dst+3*constants.BPS, uint(0x01010101)*AVG3(K, L, L))
}

func DC4(dst []uint8 /*const*/, top []uint8) {
	dc := 4
	var i int
	for i = 0; i < 4; i++ {
		dc += top[i] + top[-5+i]
	}
	Fill(dst, dc>>3, 4)
}

func RD4(dst []uint8 /*const*/, top []uint8) {
	X := top[-1]
	I := top[-2]
	J := top[-3]
	K := top[-4]
	L := top[-5]
	A := top[0]
	B := top[1]
	C := top[2]
	D := top[3]
	dst[DSTIndex(0, 3)] = AVG3(J, K, L)

	v := AVG3(I, J, K)
	dst[DSTIndex(0, 2)] = v
	dst[DSTIndex(1, 3)] = v

	v = AVG3(X, I, J)
	dst[DSTIndex(0, 1)] = v
	dst[DSTIndex(1, 2)] = v
	dst[DSTIndex(2, 3)] = v

	v = AVG3(A, X, I)
	dst[DSTIndex(0, 0)] = v
	dst[DSTIndex(1, 1)] = v
	dst[DSTIndex(2, 2)] = v
	dst[DSTIndex(3, 3)] = v

	v = AVG3(B, A, X)
	dst[DSTIndex(1, 0)] = v
	dst[DSTIndex(2, 1)] = v
	dst[DSTIndex(3, 2)] = v

	v = AVG3(C, B, A)
	dst[DSTIndex(2, 0)] = v
	dst[DSTIndex(3, 1)] = v

	dst[DSTIndex(3, 0)] = AVG3(D, C, B)
}

func LD4(dst []uint8 /*const*/, top []uint8) {
	A := top[0]
	B := top[1]
	C := top[2]
	D := top[3]
	E := top[4]
	F := top[5]
	G := top[6]
	H := top[7]
	dst[DSTIndex(0, 0)] = AVG3(A, B, C)

	v := AVG3(B, C, D)
	dst[DSTIndex(1, 0)] = v
	dst[DSTIndex(0, 1)] = v

	v = AVG3(C, D, E)
	dst[DSTIndex(2, 0)] = v
	dst[DSTIndex(1, 1)] = v
	dst[DSTIndex(0, 2)] = v

	v = AVG3(D, E, F)
	dst[DSTIndex(3, 0)] = v
	dst[DSTIndex(2, 1)] = v
	dst[DSTIndex(1, 2)] = v
	dst[DSTIndex(0, 3)] = v

	v = AVG3(E, F, G)
	dst[DSTIndex(3, 1)] = v
	dst[DSTIndex(2, 2)] = v
	dst[DSTIndex(1, 3)] = v

	v = AVG3(F, G, H)
	dst[DSTIndex(3, 2)] = v
	dst[DSTIndex(2, 3)] = v

	dst[DSTIndex(3, 3)] = AVG3(G, H, H)
}

func VR4(dst []uint8 /*const*/, top []uint8) {
	X := top[-1]
	I := top[-2]
	J := top[-3]
	K := top[-4]
	A := top[0]
	B := top[1]
	C := top[2]
	D := top[3]

	v = AVG2(X, A)
	dst[DSTIndex(1, 2)] = v
	dst[DSTIndex(0, 0)] = v

	v = AVG2(A, B)
	dst[DSTIndex(2, 2)] = v
	dst[DSTIndex(1, 0)] = v

	v = AVG2(B, C)
	dst[DSTIndex(3, 2)] = v
	dst[DSTIndex(2, 0)] = v
	dst[DSTIndex(3, 0)] = AVG2(C, D)

	dst[DSTIndex(0, 3)] = AVG3(K, J, I)
	dst[DSTIndex(0, 2)] = AVG3(J, I, X)

	v = AVG3(I, X, A)
	dst[DSTIndex(0, 1)] = v
	dst[DSTIndex(1, 3)] = v

	v = AVG3(X, A, B)
	dst[DSTIndex(1, 1)] = v
	dst[DSTIndex(2, 3)] = v

	v = AVG3(A, B, C)
	dst[DSTIndex(2, 1)] = v
	dst[DSTIndex(3, 3)] = v

	dst[DSTIndex(3, 1)] = AVG3(B, C, D)
}

func VL4(dst []uint8 /*const*/, top []uint8) {
	A := top[0]
	B := top[1]
	C := top[2]
	D := top[3]
	E := top[4]
	F := top[5]
	G := top[6]
	H := top[7]
	dst[DSTIndex(0, 0)] = AVG2(A, B)

	v = AVG2(B, C)
	dst[DSTIndex(1, 0)] = v
	dst[DSTIndex(0, 2)] = v

	v = AVG2(C, D)
	dst[DSTIndex(2, 0)] = v
	dst[DSTIndex(1, 2)] = v

	v = AVG2(D, E)
	dst[DSTIndex(3, 0)] = v
	dst[DSTIndex(2, 2)] = v

	dst[DSTIndex(0, 1)] = AVG3(A, B, C)

	v = AVG3(B, C, D)
	dst[DSTIndex(1, 1)] = v
	dst[DSTIndex(0, 3)] = v

	v = AVG3(C, D, E)
	dst[DSTIndex(2, 1)] = v
	dst[DSTIndex(1, 3)] = v

	v = AVG3(D, E, F)
	dst[DSTIndex(3, 1)] = v
	dst[DSTIndex(2, 3)] = v
	dst[DSTIndex(3, 2)] = AVG3(E, F, G)
	dst[DSTIndex(3, 3)] = AVG3(F, G, H)
}

func HU4(dst []uint8 /*const*/, top []uint8) {
	I := top[-2]
	J := top[-3]
	K := top[-4]
	L := top[-5]
	dst[DSTIndex(0, 0)] = AVG2(I, J)

	v := AVG2(J, K)
	dst[DSTIndex(2, 0)] = v
	dst[DSTIndex(0, 1)] = v

	v = AVG2(K, L)
	dst[DSTIndex(2, 1)] = v
	dst[DSTIndex(0, 2)] = v
	dst[DSTIndex(1, 0)] = AVG3(I, J, K)

	v = AVG3(J, K, L)
	dst[DSTIndex(3, 0)] = v
	dst[DSTIndex(1, 1)] = v

	v = AVG3(K, L, L)
	dst[DSTIndex(3, 1)] = v
	dst[DSTIndex(1, 2)] = v

	dst[DSTIndex(3, 2)] = L
	dst[DSTIndex(2, 2)] = L
	dst[DSTIndex(0, 3)] = L
	dst[DSTIndex(1, 3)] = L
	dst[DSTIndex(2, 3)] = L
	dst[DSTIndex(3, 3)] = L
}

func HD4(dst []uint8 /*const*/, top []uint8) {
	X := top[-1]
	I := top[-2]
	J := top[-3]
	K := top[-4]
	L := top[-5]
	A := top[0]
	B := top[1]
	C := top[2]

	v = AVG2(I, X)
	dst[DSTIndex(0, 0)] = v
	dst[DSTIndex(2, 1)] = v

	v = AVG2(J, I)
	dst[DSTIndex(0, 1)] = v
	dst[DSTIndex(2, 2)] = v

	v = AVG2(K, J)
	dst[DSTIndex(0, 2)] = v
	dst[DSTIndex(2, 3)] = v

	dst[DSTIndex(0, 3)] = AVG2(L, K)

	dst[DSTIndex(3, 0)] = AVG3(A, B, C)
	dst[DSTIndex(2, 0)] = AVG3(X, A, B)

	v = AVG3(I, X, A)
	dst[DSTIndex(1, 0)] = v
	dst[DSTIndex(3, 1)] = v

	v = AVG3(J, I, X)
	dst[DSTIndex(1, 1)] = v
	dst[DSTIndex(3, 2)] = v

	v = AVG3(K, J, I)
	dst[DSTIndex(1, 2)] = v
	dst[DSTIndex(3, 3)] = v
	dst[DSTIndex(1, 3)] = AVG3(L, K, J)
}

func TM4(dst []uint8 /*const*/, top []uint8) {
	var x, y int
	for y = 0; y < 4; y++ {
		for x = 0; x < 4; x++ {
			dst[x] = clip1(int(top[x]))
		}
		dst = dst[constants.BPS:]
	}
}

// Left samples are top[-5 .. -2], top_left is top[-1], top are
// located at top[0..3], and top right is top[4..7]
func Intra4Preds_C(dst []uint8 /*const*/, top []uint8) {
	DC4(dst[vp8.I4DC4:], top)
	TM4(dst[vp8.I4TM4:], top)
	VE4(dst[vp8.I4VE4:], top)
	HE4(dst[vp8.I4HE4:], top)
	RD4(dst[vp8.I4RD4:], top)
	VR4(dst[vp8.I4VR4:], top)
	LD4(dst[vp8.I4LD4:], top)
	VL4(dst[vp8.I4VL4:], top)
	HD4(dst[vp8.I4HD4:], top)
	HU4(dst[vp8.I4HU4:], top)
}

func GetSSE( /* const */ a []uint8 /*const*/, b []uint8, w int, h int) int {
	count := 0
	var y, x int
	for y = 0; y < h; y++ {
		for x = 0; x < w; x++ {
			diff := int(a[x] - b[x])
			count += diff * diff
		}
		a = a[constants.BPS:]
		b = b[constants.BPS:]
	}
	return count
}

func SSE16x16_C(a, b []uint8) int {
	return GetSSE(a, b, 16, 16)
}
func SSE16x8_C(a, b []uint8) int {
	return GetSSE(a, b, 16, 8)
}
func SSE8x8_C(a, b []uint8) int {
	return GetSSE(a, b, 8, 8)
}
func SSE4x4_C(a, b []uint8) int {
	return GetSSE(a, b, 4, 4)
}

func Mean16x4_C( /* const */ ref []uint8, dc [4]uint32) {
	var k, x, y int
	for k = 0; k < 4; k++ {
		avg := 0
		for y = 0; y < 4; y++ {
			for x = 0; x < 4; x++ {
				avg += int(ref[x+y*constants.BPS])
			}
		}
		dc[k] = uint32(avg)
		ref = ref[4:] // go to next 4x4 block.
	}
}

// Hadamard transform
// Returns the weighted sum of the absolute value of transformed coefficients.
// w[] contains a row-major 4 by 4 symmetric matrix.
func TTransform( /* const */ in []uint8 /*const*/, w []uint16) int {
	sum := 0
	var tmp [16]int
	var i int
	// horizontal pass
	for i = 0; i < 4; {
		a0 := in[0] + in[2]
		a1 := in[1] + in[3]
		a2 := in[1] - in[3]
		a3 := in[0] - in[2]
		tmp[0+i*4] = int(a0 + a1)
		tmp[1+i*4] = int(a3 + a2)
		tmp[2+i*4] = int(a3 - a2)
		tmp[3+i*4] = int(a0 - a1)

		i++
		in = in[constants.BPS:]
	}
	// vertical pass
	for i = 0; i < 4; {
		a0 := tmp[0+i] + tmp[8+i]
		a1 := tmp[4+i] + tmp[12+i]
		a2 := tmp[4+i] - tmp[12+i]
		a3 := tmp[0+i] - tmp[8+i]
		b0 := a0 + a1
		b1 := a3 + a2
		b2 := a3 - a2
		b3 := a0 - a1

		sum += int(w[0]) * stdlib.Abs(b0)
		sum += int(w[4]) * stdlib.Abs(b1)
		sum += int(w[8]) * stdlib.Abs(b2)
		sum += int(w[12]) * stdlib.Abs(b3)

		i++
		w = w[1:]
	}
	return sum
}

func Disto4x4_C(a []uint8, b []uint8, w []uint16) int {
	sum1 := TTransform(a, w)
	sum2 := TTransform(b, w)
	return stdlib.Abs(sum2-sum1) >> 5
}

func Disto16x16_C(a []uint8, b []uint8, w []uint16) int {
	D := 0
	var x, y int
	for y = 0; y < 16*constants.BPS; y += 4 * constants.BPS {
		for x = 0; x < 16; x += 4 {
			D += Disto4x4_C(a[x+y:], b[x+y:], w)
		}
	}
	return D
}

var kZigzag = [16]uint8{0, 1, 4, 8, 5, 2, 3, 6, 9, 12, 13, 10, 7, 11, 14, 15}

// Simple quantization
func QuantizeBlock_C(in [16]int16, out [16]int16, mtx *vp8.VP8Matrix) bool {
	last := -1
	var n int
	for n = 0; n < 16; n++ {
		j := kZigzag[n]
		sign := (in[j] < 0)
		coeff := tenary.If(sign, -in[j], in[j]) + mtx.sharpen[j]
		if coeff > mtx.zthresh[j] {
			Q := mtx.q[j]
			iQ := mtx.iq[j]
			B := mtx.bias[j]
			level := QUANTDIV(coeff, iQ, B)
			if level > vp8.MAX_LEVEL {
				level = vp8.MAX_LEVEL
			}
			if sign {
				level = -level
			}
			in[j] = level * int(Q)
			out[n] = level
			if level {
				last = n
			}
		} else {
			out[n] = 0
			in[j] = 0
		}
	}
	return (last >= 0)
}

func Quantize2Blocks_C(in [32]int16, out [32]int16, mtx *vp8.VP8Matrix) int {
	var nz int
	nz = VP8EncQuantizeBlock(in[0*16:], out[0*16:], mtx) << 0
	nz |= VP8EncQuantizeBlock(in[1*16:], out[1*16:], mtx) << 1
	return nz
}

func Copy( /* const */ src []uint8, dst []uint8, w int, h int) {
	var y int
	for y = 0; y < h; y++ {
		stdlib.MemCpy(dst, src, w)
		src = src[constants.BPS:]
		dst = dst[constants.BPS:]
	}
}

func Copy4x4_C( /* const */ src []uint8, dst []uint8) {
	Copy(src, dst, 4, 4)
}

func Copy16x8_C( /* const */ src []uint8, dst []uint8) {
	Copy(src, dst, 16, 8)
}

var (
	VP8ITransform          = ITransform_C
	VP8FTransform          = FTransform_C
	VP8FTransformWHT       = FTransformWHT_C
	VP8TDisto4x4           = Disto4x4_C
	VP8TDisto16x16         = Disto16x16_C
	VP8CollectHistogram    = CollectHistogram_C
	VP8SSE16x16            = SSE16x16_C
	VP8SSE16x8             = SSE16x8_C
	VP8SSE8x8              = SSE8x8_C
	VP8SSE4x4              = SSE4x4_C
	VP8EncQuantizeBlock    = QuantizeBlock_C
	VP8EncQuantize2Blocks  = Quantize2Blocks_C
	VP8EncQuantizeBlockWHT = QuantizeBlock_C
	VP8EncPredLuma4        = Intra4Preds_C
	VP8EncPredLuma16       = Intra16Preds_C
	VP8FTransform2         = FTransform2_C
	VP8EncPredChroma8      = IntraChromaPreds_C
	VP8Mean16x4            = Mean16x4_C
	VP8Copy4x4             = Copy4x4_C
	VP8Copy16x8            = Copy16x8_C
)
