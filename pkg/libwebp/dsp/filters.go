// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"slices"

	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

var WebPUnfilters [4]func(prev []uint8, in []uint8, out []uint8, width int)
var WebPFilters [4]func(data []uint8, width, height int, stride int, filtered_data []uint8)

func init() {
	WebPUnfilters[WEBP_FILTER_NONE] = NoneUnfilter_C
	WebPUnfilters[WEBP_FILTER_HORIZONTAL] = HorizontalUnfilter_C
	WebPUnfilters[WEBP_FILTER_VERTICAL] = VerticalUnfilter_C
	WebPUnfilters[WEBP_FILTER_GRADIENT] = GradientUnfilter_C

	WebPFilters[WEBP_FILTER_NONE] = nil
	WebPFilters[WEBP_FILTER_HORIZONTAL] = HorizontalFilter_C
	WebPFilters[WEBP_FILTER_VERTICAL] = VerticalFilter_C
	WebPFilters[WEBP_FILTER_GRADIENT] = GradientFilter_C
}

func PredictLine_C( /* const */ src []uint8 /*const*/, pred []uint8, dst []uint8, length int) {
	for i = 0; i < length; i++ {
		dst[i] = uint8(src[i] - pred[i])
	}
}

func DoHorizontalFilter_C( /* const */ in []uint8, width, height int, stride int, out []uint8) {
	var preds []uint8 = in
	var row int

	// Leftmost pixel is the same as input for topmost scanline.
	out[0] = in[0]
	PredictLine_C(in[1:], preds, out[1:], width-1)
	oldpreds := preds
	preds = preds[stride:]
	in = in[stride:]
	out = out[stride:]

	// Filter line-by-line.
	for row = 1; row < height; row++ {
		// Leftmost pixel is predicted from above.
		PredictLine_C(in, oldpreds, out, 1)
		PredictLine_C(in[1:], preds, out[1:], width-1)
		oldpreds = preds
		preds = preds[stride:]
		in = in[stride:]
		out = out[stride:]
	}
}

func DoVerticalFilter_C( /* const */ in []uint8, width, height int, stride int, out []uint8) {
	var preds []uint8 = in
	var row int

	// Very first top-left pixel is copied.
	out[0] = in[0]
	// Rest of top scan-line is left-predicted.
	PredictLine_C(in[1:], preds, out[1:], width-1)
	in = in[stride:]
	out = out[stride:]

	// Filter line-by-line.
	for row = 1; row < height; row++ {
		PredictLine_C(in, preds, out, width)
		preds = preds[stride:]
		in = in[stride:]
		out = out[stride:]
	}
}

func GradientPredictor_C(a, b, c uint8) int {
	g := int(a + b - c)
	return tenary.If((g&^0xff) == 0, g, tenary.If(g < 0, 0, 255)) // clip to 8bit
}

func DoGradientFilter_C( /* const */ in []uint8, width, height, stride int, out []uint8) {
	var preds []uint8 = in
	var row int

	// left prediction for top scan-line
	out[0] = in[0]
	PredictLine_C(in[1:], preds, out[1:], width-1)
	oldpred := preds // replaces preds - stride
	preds = preds[stride:]
	in = in[stride:]
	out = out[stride:]

	// Filter line-by-line.
	for row = 1; row < height; row++ {
		var w int
		// leftmost pixel: predict from above.
		PredictLine_C(in, oldpred, out, 1)
		for w = 1; w < width; w++ {
			pred := GradientPredictor_C(preds[w-1], preds[w-stride], preds[w-stride-1])
			out[w] = uint8(int(in[w]) - pred)
		}
		oldpred = preds // replaces preds - stride
		preds = preds[stride:]
		in = in[stride:]
		out = out[stride:]
	}
}

func HorizontalFilter_C( /* const */ data []uint8, width, height int, stride int, filtered_data []uint8) {
	DoHorizontalFilter_C(data, width, height, stride, filtered_data)
}

func VerticalFilter_C( /* const */ data []uint8, width, height int, stride int, filtered_data []uint8) {
	DoVerticalFilter_C(data, width, height, stride, filtered_data)
}

func GradientFilter_C( /* const */ data []uint8, width, height int, stride int, filtered_data []uint8) {
	DoGradientFilter_C(data, width, height, stride, filtered_data)
}

//------------------------------------------------------------------------------

func NoneUnfilter_C( /* const */ prev []uint8 /*const*/, in []uint8, out []uint8, width int) {
	_ = prev
	if !slices.Equal(out, in) {
		stdlib.MemCpy2(out, in)
	}
}

func HorizontalUnfilter_C( /* const */ prev []uint8 /*const*/, in []uint8, out []uint8, width int) {
	pred := tenary.If(prev == nil, 0, prev[0])
	var i int
	for i = 0; i < width; i++ {
		out[i] = (uint8)(pred + in[i])
		pred = out[i]
	}
}

func VerticalUnfilter_C( /* const */ prev []uint8 /*const*/, in []uint8, out []uint8, width int) {
	if prev == nil {
		HorizontalUnfilter_C(nil, in, out, width)
	} else {
		for i := 0; i < width; i++ {
			out[i] = (uint8)(prev[i] + in[i])
		}
	}
}

func GradientUnfilter_C( /* const */ prev []uint8 /*const*/, in []uint8, out []uint8, width int) {
	if prev == nil {
		HorizontalUnfilter_C(nil, in, out, width)
	} else {
		top := prev[0]
		top_left := top
		left := top

		for i := 0; i < width; i++ {
			top = prev[i] // need to read this first, in case prev==out
			left = uint8(int(in[i]) + GradientPredictor_C(left, top, top_left))
			top_left = top
			out[i] = left
		}
	}
}
