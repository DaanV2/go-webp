package dsp

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Image transforms and color space conversion methods for lossless decoder.
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)
//          Urvang Joshi (urvang@google.com)

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/generics"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
	"github.com/daanv2/go-webp/pkg/vp8"
)

//------------------------------------------------------------------------------
// Image transforms.

func Average2(a0, a1 uint32) uint32 {
	return (((a0 ^ a1) & uint32(0xfefefefe)) >> 1) + (a0 & a1)
}

func Average3(a0, a1, a2 uint32) uint32 {
	return Average2(Average2(a0, a2), a1)
}

func Average4(a0, a1, a2, a3 uint32) uint32 {
	return Average2(Average2(a0, a1), Average2(a2, a3))
}

func Clip255(a uint32) uint32 {
	if a < 256 {
		return a
	}
	// return 0, when a is a negative integer.
	// return 255, when a is positive.
	return ^a >> 24
}

func AddSubtractComponentFull(a int, b int, c int) uint32 {
	return Clip255(uint32(a + b - c))
}

func ClampedAddSubtractFull(c0, c1, c2 uint32) uint32 {
	a := AddSubtractComponentFull(c0>>24, c1>>24, c2>>24)
	r := AddSubtractComponentFull((c0>>16)&0xff, (c1>>16)&0xff, (c2>>16)&0xff)
	g := AddSubtractComponentFull((c0>>8)&0xff, (c1>>8)&0xff, (c2>>8)&0xff)
	b := AddSubtractComponentFull(c0&0xff, c1&0xff, c2&0xff)
	return uint32((a << 24) | (r << 16) | (g << 8) | b)
}

func AddSubtractComponentHalf(a int, b int) int {
	return Clip255((uint32)(a + (a-b)/2))
}

func ClampedAddSubtractHalf(c0, c1, c2 uint32) uint32 {
	ave := Average2(c0, c1)
	a := AddSubtractComponentHalf(ave>>24, int(c2>>24))
	r := AddSubtractComponentHalf((ave>>16)&0xff, int(c2>>16)&0xff)
	g := AddSubtractComponentHalf((ave>>8)&0xff, int(c2>>8)&0xff)
	b := AddSubtractComponentHalf((ave>>0)&0xff, int(c2>>0)&0xff)
	return uint32((a << 24) | (r << 16) | (g << 8) | b)
}

func Sub3(a int, b int, c int) int {
	pb := b - c
	pa := a - c
	return stdlib.Abs(pb) - stdlib.Abs(pa)
}

func Select(a, b, c uint32) uint32 {
	pa_minus_pb := Sub3(int(a>>24), int(b>>24), int(c>>24)) +
		Sub3(int(a>>16)&0xff, int(b>>16)&0xff, int(c>>16)&0xff) +
		Sub3(int(a>>8)&0xff, int(b>>8)&0xff, int(c>>8)&0xff) +
		Sub3(int(a)&0xff, int(b)&0xff, int(c)&0xff)
	return tenary.If(pa_minus_pb <= 0, a, b)
}

func VP8LPredictor0_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	_ = top
	_ = left
	return constants.ARGB_BLACK
}
func VP8LPredictor1_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	_ = top
	return *left
}
func VP8LPredictor2_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	_ = left
	return top[0]
}
func VP8LPredictor3_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	_ = left
	return top[1]
}
func VP8LPredictor4_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	_ = left
	return top[-1]
}
func VP8LPredictor5_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average3(left[0], top[0], top[1])
	return pred
}
func VP8LPredictor6_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average2(left[0], top[-1])
	return pred
}
func VP8LPredictor7_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average2(left[0], top[0])
	return pred
}
func VP8LPredictor8_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average2(top[-1], top[0])
	_ = left
	return pred
}
func VP8LPredictor9_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average2(top[0], top[1])
	_ = left
	return pred
}
func VP8LPredictor10_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Average4(left[0], top[-1], top[0], top[1])
	return pred
}
func VP8LPredictor11_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := Select(top[0], *left, top[-1])
	return pred
}
func VP8LPredictor12_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := ClampedAddSubtractFull(left[0], top[0], top[-1])
	return pred
}
func VP8LPredictor13_C( /* const */ left []uint32 /*const*/, top []uint32) uint32 {
	pred := ClampedAddSubtractHalf(left[0], top[0], top[-1])
	return pred
}

func PredictorAdd0_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	_ = upper
	for x = 0; x < num_pixels; x++ {
		out[x] = VP8LAddPixels(in[x], constants.ARGB_BLACK)
	}
}
func PredictorAdd1_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var i int
	left := out[-1]
	_ = upper
	for i = 0; i < num_pixels; i++ {
		v = VP8LAddPixels(in[i], left)
		out[i] = v
		left = v
	}
}

func PredictorAdd2_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor2_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd3_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor3_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd4_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor4_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd5_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor5_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd6_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor6_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd7_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor7_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd8_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor8_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd9_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor9_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd10_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor10_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd11_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor11_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd12_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor12_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

func PredictorAdd13_C( /* const */ in []uint32 /*const*/, upper []uint32, num_pixels int, out []uint32) {
	var x int
	assert.Assert(upper != nil)

	for x = 0; x < num_pixels; x++ {
		pred := VP8LPredictor13_C(&out[x-1], upper+x)
		out[x] = VP8LAddPixels(in[x], pred)
	}
}

// Inverse prediction.
func PredictorInverseTransform_C( /* const */ transform *vp8.VP8LTransform, y_start int, y_end int /*const*/, in []uint32, out []uint32) {
	width := transform.xsize
	if y_start == 0 { // First Row follows the L (mode=1) mode.
		PredictorAdd0_C(in, nil, 1, out)
		PredictorAdd1_C(in+1, nil, width-1, out+1)
		in += width
		out += width
		y_start++
	}

	{
		y := y_start
		tile_width := 1 << transform.bits
		mask := tile_width - 1
		tiles_per_row := VP8LSubSampleSize(width, transform.bits)
		var pred_mode_base []uint32 = transform.data + (y>>transform.bits)*tiles_per_row

		for y < y_end {
			var pred_mode_src []uint32 = pred_mode_base
			x := 1
			// First pixel follows the T (mode=2) mode.
			PredictorAdd2_C(in, out-width, 1, out)
			// .. the rest:
			for x < width {
				v := pred_mode_src[0]
				pred_mode_src = pred_mode_src[1:]

				var pred_func VP8LPredictorAddSubFunc = vp8.VP8LPredictorsAdd[(v>>8)&0xf]
				x_end := (x & ~mask) + tile_width
				if x_end > width {
					x_end = width
				}
				pred_func(in+x, out+x-width, x_end-x, out+x)
				x = x_end
			}
			in += width
			out += width
			y++
			if (y & mask) == 0 { // Use the same mask, since tiles are squares.
				pred_mode_base += tiles_per_row
			}
		}
	}
}

// Add green to blue and red channels (i.e. perform the inverse transform of
// 'subtract green').
func VP8LAddGreenToBlueAndRed_C( /* const */ src []uint32, num_pixels int, dst []uint32) {
	var i int
	for i = 0; i < num_pixels; i++ {
		argb := src[i]
		green := ((argb >> 8) & 0xff)
		red_blue := (argb & uint(0x00ff00ff))
		red_blue += (green << 16) | green
		red_blue &= uint(0x00ff00ff)
		dst[i] = (argb & uint(0xff00ff00)) | red_blue
	}
}

func ColorTransformDelta(color_pred int8, color int8) int {
	return (color_pred * color) >> 5
}

func ColorCodeToMultipliers(uint32 color_code /*const*/, m *VP8LMultipliers) {
	m.green_to_red = (color_code >> 0) & 0xff
	m.green_to_blue = (color_code >> 8) & 0xff
	m.red_to_blue = (color_code >> 16) & 0xff
}

func VP8LTransformColorInverse_C( /* const */ m *VP8LMultipliers /*const*/, src []uint32, num_pixels int, dst *uint32) {
	var i int
	for i = 0; i < num_pixels; i++ {
		argb := src[i]
		green := (int8)(argb >> 8)
		red := argb >> 16
		new_red := red & 0xff
		new_blue := argb & 0xff
		new_red += ColorTransformDelta(int8(m.green_to_red), green)
		new_red &= 0xff
		new_blue += ColorTransformDelta(int8(m.green_to_blue), green)
		new_blue += ColorTransformDelta(int8(m.red_to_blue), int8(new_red))
		new_blue &= 0xff
		dst[i] = (argb & uint(0xff00ff00)) | (new_red << 16) | (new_blue)
	}
}

// Color space inverse transform.
func ColorSpaceInverseTransform_C( /* const */ transform *VP8LTransform, y_start int, y_end int /*const*/, src []uint32, dst *uint32) {
	width := transform.xsize
	tile_width := 1 << transform.bits
	mask := tile_width - 1
	safe_width := width & ~mask
	remaining_width := width - safe_width
	tiles_per_row := VP8LSubSampleSize(width, transform.bits)
	y := y_start
	var pred_row *uint32 = transform.data + (y>>transform.bits)*tiles_per_row

	for y < y_end {
		var pred *uint32 = pred_row
		var m = VP8LMultipliers{0, 0, 0}
		var src_safe_end *uint32 = src + safe_width
		var src_end *uint32 = src + width
		for src < src_safe_end {
			v := pred[0]
			pred = pred[1:]
			ColorCodeToMultipliers(v, &m)
			VP8LTransformColorInverse(&m, src, tile_width, dst)
			src += tile_width
			dst += tile_width
		}
		if src < src_end { // Left-overs using C-version.
			v := pred[0]
			pred = pred[1:]
			ColorCodeToMultipliers(v, &m)
			VP8LTransformColorInverse(&m, src, remaining_width, dst)
			src += remaining_width
			dst += remaining_width
		}
		y++
		if (y & mask) == 0 {
			pred_row += tiles_per_row
		}
	}
}

func MapARGB_C( /* const */ src []uint32 /*const*/, color_map []uint32, dst []uint32, y_start int, y_end int, width int) {
	var y int
	for y = y_start; y < y_end; y++ {
		var x int
		for x = 0; x < width; x++ {
			v := src[0]
			src = src[1:]
			dst[i] = VP8GetARGBValue(color_map[VP8GetARGBIndex()])
			i++
		}
	}
}

func ColorIndexInverseTransform_C( /* const */ transform *VP8LTransform, y_start int, y_end int /*const*/, src []uint32, dst *uint32) {
	var y int
	bits_per_pixel := 8 >> transform.bits
	width := transform.xsize
	var color_map *uint32 = transform.data
	if bits_per_pixel < 8 {
		pixels_per_byte := 1 << transform.bits
		count_mask := pixels_per_byte - 1
		bit_mask := (1 << bits_per_pixel) - 1
		for y = y_start; y < y_end; y++ {
			packed_pixels := 0
			var x int
			for x = 0; x < width; x++ {
				/* We need to load fresh 'packed_pixels' once every                */
				/* 'pixels_per_byte' increments of x. Fortunately, pixels_per_byte */
				/* is a power of 2, so can just use a mask for that, instead of    */
				/* decrementing a counter.                                         */

				v := src[0]
				src = src[1:]
				if (x & count_mask) == 0 {
					packed_pixels = VP8GetARGBIndex(v)
				}
				dst[i] = VP8GetARGBValue(color_map[packed_pixels&bit_mask])
				i++
				packed_pixels >>= bits_per_pixel
			}
		}
	} else {
		VP8LMapColor32b(src, color_map, dst, y_start, y_end, width)
	}
}

func MapAlpha_C( /* const */ src *uint8 /*const*/, color_map *uint32, dst []uint8, y_start int, y_end int, width int) {
	var y int
	for y = y_start; y < y_end; y++ {
		var x int
		for x = 0; x < width; x++ {

			v := src[0]
			src = src[1:]
			dst[i] = VP8GetAlphaValue(color_map[VP8GetAlphaIndex(v)])
			i++
		}
	}
}

func VP8LColorIndexInverseTransformAlpha( /* const */ transform *VP8LTransform, y_start int, y_end int /*const*/, src *uint8, dst []uint8) {
	var y int
	bits_per_pixel := 8 >> transform.bits
	width := transform.xsize
	var color_map *uint32 = transform.data
	if bits_per_pixel < 8 {
		pixels_per_byte := 1 << transform.bits
		count_mask := pixels_per_byte - 1
		bit_mask := (1 << bits_per_pixel) - 1
		for y = y_start; y < y_end; y++ {
			packed_pixels := 0
			var x int
			for x = 0; x < width; x++ {
				// We need to load fresh 'packed_pixels' once every
				// 'pixels_per_byte' increments of x. Fortunately, pixels_per_byte
				// is a power of 2, so can just use a mask for that, instead of
				// decrementing a counter.
				v := src[0]
				src = src[1:]
				if (x & count_mask) == 0 {
					packed_pixels = VP8GetAlphaIndex(v)
				}
				dst[i] = VP8GetAlphaValue(color_map[packed_pixels&bit_mask])
				i++
				packed_pixels >>= bits_per_pixel
			}
		}
	} else {
		VP8LMapColor8b(src, color_map, dst, y_start, y_end, width)
	}
}

// Performs inverse transform of data given transform information, start and end
// rows. Transform will be applied to rows [row_start, row_end[.
// The and pointers refer *in to *out source and destination data respectively
// corresponding to the intermediate row (row_start).
func VP8LInverseTransform( /* const */ transform *vp8.VP8LTransform, row_start int, row_end int /*const*/, in []uint32 /*const*/, out []uint32) {
	width := transform.xsize
	assert.Assert(row_start < row_end)
	assert.Assert(row_end <= transform.ysize)
	switch transform.vtype {
	case vp8.SUBTRACT_GREEN_TRANSFORM:
		VP8LAddGreenToBlueAndRed(in, (row_end-row_start)*width, out)
		break
	case vp8.PREDICTOR_TRANSFORM:
		PredictorInverseTransform_C(transform, row_start, row_end, in, out)
		if row_end != transform.ysize {
			// The last predicted row in this iteration will be the top-pred row
			// for the first row in next iteration.
			stdlib.MemCpy(out-width, out+(row_end-row_start-1)*width, width*sizeof(*out))
		}
		break
	case vp8.CROSS_COLOR_TRANSFORM:
		ColorSpaceInverseTransform_C(transform, row_start, row_end, in, out)
		break
	case vp8.COLOR_INDEXING_TRANSFORM:
		if in == out && transform.bits > 0 {
			// Move packed pixels to the end of unpacked region, so that unpacking
			// can occur seamlessly.
			// Also, note that this is the only transform that applies on
			// the effective width of VP8LSubSampleSize(xsize, bits). All other
			// transforms work on effective width of 'xsize'.
			out_stride := (row_end - row_start) * width
			in_stride := (row_end - row_start) * VP8LSubSampleSize(transform.xsize, transform.bits)
			var src []uint32 = out + out_stride - in_stride
			stdlib.MemMove(src, out, in_stride*sizeof(*src))
			ColorIndexInverseTransform_C(transform, row_start, row_end, src, out)
		} else {
			ColorIndexInverseTransform_C(transform, row_start, row_end, in, out)
		}
		break
	}
}

func VP8LConvertBGRAToRGB_C( /* const */ src []uint32, num_pixels int, dst []uint8) {
	var i int
	for _, argb := range src[:num_pixels] {
		dst[i] = uint8(argb>>16) & 0xff
		i++
		dst[i] = uint8(argb>>8) & 0xff
		i++
		dst[i] = uint8(argb>>0) & 0xff
		i++
	}
}

func VP8LConvertBGRAToRGBA_C( /* const */ src []uint32, num_pixels int, dst []uint8) {
	var i int
	for _, argb := range src[:num_pixels] {
		dst[i] = uint8(argb>>16) & 0xff
		i++
		dst[i] = uint8(argb>>8) & 0xff
		i++
		dst[i] = uint8(argb>>0) & 0xff
		i++
		dst[i] = uint8(argb>>24) & 0xff
		i++
	}
}

func VP8LConvertBGRAToRGBA4444_C( /* const */ src []uint32, num_pixels int, dst []uint8) {
	var i int
	for _, argb := range src[:num_pixels] {
		rg := ((argb >> 16) & 0xf0) | ((argb >> 12) & 0xf)
		ba := ((argb >> 0) & 0xf0) | ((argb >> 28) & 0xf)
		dst[i] = uint8(ba)
		i++
		dst[i] = uint8(rg)
		i++
	}
}

func VP8LConvertBGRAToRGB565_C( /* const */ src []uint32, num_pixels int, dst []uint8) {
	var i int
	for _, argb := range src[:num_pixels] {
		rg := ((argb >> 16) & 0xf8) | ((argb >> 13) & 0x7)
		gb := ((argb >> 5) & 0xe0) | ((argb >> 3) & 0x1f)
		dst[i] = uint8(gb)
		i++
		dst[i] = uint8(rg)
		i++
	}
}

func VP8LConvertBGRAToBGR_C( /* const */ src []uint32, num_pixels int, dst []uint8) {
	var i int
	for _, argb := range src[:num_pixels] {
		dst[i] = uint8(argb>>0) & 0xff
		i++
		dst[i] = uint8(argb>>8) & 0xff
		i++
		dst[i] = uint8(argb>>16) & 0xff
		i++
	}
}

func CopyOrSwap( /* const */ src []uint32, num_pixels int, dst []uint8, swap_on_big_endian bool) {
	if false == swap_on_big_endian {
		for _, argb := range src[:num_pixels] {
			WebPUint32ToMem(dst, BSwap32(argb))
			dst = dst[generics.SizeOf(argb):]
		}
	} else {
		stdlib.MemCpy(dst, src, num_pixels*sizeof(*src))
	}
}

// Converts from BGRA to other color spaces.
func VP8LConvertFromBGRA( /* const */ in_data []uint32, num_pixels int, out_colorspace webp.WEBP_CSP_MODE /*const*/, rgba []uint8) {
	switch out_colorspace {
	case webp.MODE_RGB:
		VP8LConvertBGRAToRGB(in_data, num_pixels, rgba)
		break
	case webp.MODE_RGBA:
		VP8LConvertBGRAToRGBA(in_data, num_pixels, rgba)
		break
	case webp.MODE_rgbA:
		VP8LConvertBGRAToRGBA(in_data, num_pixels, rgba)
		WebPApplyAlphaMultiply(rgba, 0, num_pixels, 1, 0)
		break
	case webp.MODE_BGR:
		VP8LConvertBGRAToBGR(in_data, num_pixels, rgba)
		break
	case webp.MODE_BGRA:
		CopyOrSwap(in_data, num_pixels, rgba, 1)
		break
	case webp.MODE_bgrA:
		CopyOrSwap(in_data, num_pixels, rgba, 1)
		WebPApplyAlphaMultiply(rgba, 0, num_pixels, 1, 0)
		break
	case webp.MODE_ARGB:
		CopyOrSwap(in_data, num_pixels, rgba, 0)
		break
	case webp.MODE_Argb:
		CopyOrSwap(in_data, num_pixels, rgba, 0)
		WebPApplyAlphaMultiply(rgba, 1, num_pixels, 1, 0)
		break
	case webp.MODE_RGBA_4444:
		VP8LConvertBGRAToRGBA4444(in_data, num_pixels, rgba)
		break
	case webp.MODE_rgbA_4444:
		VP8LConvertBGRAToRGBA4444(in_data, num_pixels, rgba)
		WebPApplyAlphaMultiply4444(rgba, num_pixels, 1, 0)
		break
	case webp.MODE_RGB_565:
		VP8LConvertBGRAToRGB565(in_data, num_pixels, rgba)
		break
	default:
		panic("Unsupported output colorspace")
		// assert.Assert(0) // Code flow should not reach here.
	}
}
