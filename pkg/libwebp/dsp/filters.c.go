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
// Spatial prediction using various filters
//
// Author: Urvang (urvang@google.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"



#if !WEBP_NEON_OMIT_C_CODE
func PredictLine_C(/* const */ src *uint8, /*const*/ pred *uint8, dst *uint8, length int) {
  var i int
  for (i = 0; i < length; ++i) dst[i] = (uint8)(src[i] - pred[i])
}

//------------------------------------------------------------------------------
// Horizontal filter.

func DoHorizontalFilter_C(/* const */ in *uint8, width, height int, stride int, out *uint8) {
  var preds *uint8 = in
  row int

  // Leftmost pixel is the same as input for topmost scanline.
  out[0] = in[0]
  PredictLine_C(in + 1, preds, out + 1, width - 1)
  preds += stride
  in += stride
  out += stride

  // Filter line-by-line.
  for row = 1; row < height; row++ {
    // Leftmost pixel is predicted from above.
    PredictLine_C(in, preds - stride, out, 1)
    PredictLine_C(in + 1, preds, out + 1, width - 1)
    preds += stride
    in += stride
    out += stride
  }
}

//------------------------------------------------------------------------------
// Vertical filter.

func DoVerticalFilter_C(/* const */ in *uint8, width, height int, stride int, out *uint8) {
  var preds *uint8 = in
  row int

  // Very first top-left pixel is copied.
  out[0] = in[0]
  // Rest of top scan-line is left-predicted.
  PredictLine_C(in + 1, preds, out + 1, width - 1)
  in += stride
  out += stride

  // Filter line-by-line.
  for row = 1; row < height; row++ {
    PredictLine_C(in, preds, out, width)
    preds += stride
    in += stride
    out += stride
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------
// Gradient filter.

static  int GradientPredictor_C(uint8 a, uint8 b, uint8 c) {
  g := a + b - c
  return ((g & ~0xff) == 0) ? g : tenary.If(g < 0, 0, 255)  // clip to 8bit
}

#if !WEBP_NEON_OMIT_C_CODE
func DoGradientFilter_C(/* const */ in *uint8, width, height int, stride int, out *uint8) {
  var preds *uint8 = in
  row int

  // left prediction for top scan-line
  out[0] = in[0]
  PredictLine_C(in + 1, preds, out + 1, width - 1)
  preds += stride
  in += stride
  out += stride

  // Filter line-by-line.
  for row = 1; row < height; row++ {
    var w int
    // leftmost pixel: predict from above.
    PredictLine_C(in, preds - stride, out, 1)
    for w = 1; w < width; w++ {
      pred := GradientPredictor_C(preds[w - 1], preds[w - stride], preds[w - stride - 1])
      out[w] = (uint8)(in[w] - pred)
    }
    preds += stride
    in += stride
    out += stride
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE


//------------------------------------------------------------------------------

#if !WEBP_NEON_OMIT_C_CODE
func HorizontalFilter_C(/* const */ data *uint8, width, height int, stride int, filtered_data *uint8) {
  DoHorizontalFilter_C(data, width, height, stride, filtered_data)
}

func VerticalFilter_C(/* const */ data *uint8, width, height int, stride int, filtered_data *uint8) {
  DoVerticalFilter_C(data, width, height, stride, filtered_data)
}

func GradientFilter_C(/* const */ data *uint8, width, height int, stride int, filtered_data *uint8) {
  DoGradientFilter_C(data, width, height, stride, filtered_data)
}
#endif  // !WEBP_NEON_OMIT_C_CODE

//------------------------------------------------------------------------------

func NoneUnfilter_C(/* const */ prev *uint8, /*const*/ in *uint8, out *uint8, width int) {
  _ = prev
  if out != in { memcpy(out, in, width * sizeof(*out)) }
}

func HorizontalUnfilter_C(/* const */ prev *uint8, /*const*/ in *uint8, out *uint8, width int) {
  pred := tenary.If(prev == nil, 0, prev[0])
  var i int
  for i = 0; i < width; i++ {
    out[i] = (uint8)(pred + in[i])
    pred = out[i]
  }
}

#if !WEBP_NEON_OMIT_C_CODE
func VerticalUnfilter_C(/* const */ prev *uint8, /*const*/ in *uint8, out *uint8, width int) {
  if (prev == nil) {
    HorizontalUnfilter_C(nil, in, out, width)
  } else {
    var i int
    for (i = 0; i < width; ++i) out[i] = (uint8)(prev[i] + in[i])
  }
}
#endif  // !WEBP_NEON_OMIT_C_CODE

func GradientUnfilter_C(/* const */ prev *uint8, /*const*/ in *uint8, out *uint8, width int) {
  if (prev == nil) {
    HorizontalUnfilter_C(nil, in, out, width)
  } else {
    top := prev[0], top_left = top, left = top
    var i int
    for i = 0; i < width; i++ {
      top = prev[i];  // need to read this first, in case prev==out
      left = (uint8)(in[i] + GradientPredictor_C(left, top, top_left))
      top_left = top
      out[i] = left
    }
  }
}

//------------------------------------------------------------------------------
// Init function

WebPFilterFunc WebPFilters[WEBP_FILTER_LAST]
WebPUnfilterFunc WebPUnfilters[WEBP_FILTER_LAST]


extern func VP8FiltersInitMIPSdspR2(void)
extern func VP8FiltersInitMSA(void)
extern func VP8FiltersInitNEON(void)
extern func VP8FiltersInitSSE2(void)

WEBP_DSP_INIT_FUNC(VP8FiltersInit) {
  WebPUnfilters[WEBP_FILTER_NONE] = NoneUnfilter_C
#if !WEBP_NEON_OMIT_C_CODE
  WebPUnfilters[WEBP_FILTER_HORIZONTAL] = HorizontalUnfilter_C
  WebPUnfilters[WEBP_FILTER_VERTICAL] = VerticalUnfilter_C
#endif
  WebPUnfilters[WEBP_FILTER_GRADIENT] = GradientUnfilter_C

  WebPFilters[WEBP_FILTER_NONE] = nil
#if !WEBP_NEON_OMIT_C_CODE
  WebPFilters[WEBP_FILTER_HORIZONTAL] = HorizontalFilter_C
  WebPFilters[WEBP_FILTER_VERTICAL] = VerticalFilter_C
  WebPFilters[WEBP_FILTER_GRADIENT] = GradientFilter_C
#endif

  if (VP8GetCPUInfo != nil) {
#if false
    if (VP8GetCPUInfo(kSSE2)) {
      VP8FiltersInitSSE2()
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      VP8FiltersInitMIPSdspR2()
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      VP8FiltersInitMSA()
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    VP8FiltersInitNEON()
  }
#endif

  assert.Assert(WebPUnfilters[WEBP_FILTER_NONE] != nil)
  assert.Assert(WebPUnfilters[WEBP_FILTER_HORIZONTAL] != nil)
  assert.Assert(WebPUnfilters[WEBP_FILTER_VERTICAL] != nil)
  assert.Assert(WebPUnfilters[WEBP_FILTER_GRADIENT] != nil)
  assert.Assert(WebPFilters[WEBP_FILTER_HORIZONTAL] != nil)
  assert.Assert(WebPFilters[WEBP_FILTER_VERTICAL] != nil)
  assert.Assert(WebPFilters[WEBP_FILTER_GRADIENT] != nil)
}
