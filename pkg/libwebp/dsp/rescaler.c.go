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
// Rescaling functions
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------
// Implementations of critical functions ImportRow / ExportRow

const ROUNDER = (WEBP_RESCALER_ONE >> 1)

func MULT_FIX(x, y uint64) {
	return uint64(((x) * (y) + ROUNDER) >> WEBP_RESCALER_RFIX)
}
func MULT_FIX_FLOOR(x, y) {
	return uint64(((x) * (y)) >> WEBP_RESCALER_RFIX)
}

//------------------------------------------------------------------------------
// Row import

func WebPRescalerImportRowExpand_C(/* const */ wrk *WebPRescaler, /*const*/ src *uint8) {
  x_stride := wrk.num_channels;
  x_out_max := wrk.dst_width * wrk.num_channels;
  var channel int
  assert.Assert(!WebPRescalerInputDone(wrk));
  assert.Assert(wrk.x_expand);
  for channel = 0; channel < x_stride; channel++ {
    x_in := channel;
    x_out := channel;
    // simple bilinear interpolation
    accum := wrk.x_add;
    var left rescaler_t = rescaler_t(src[x_in])
    var right rescaler_t = tenary.If(wrk.src_width > 1, rescaler_t(src[x_in + x_stride], left))
    x_in += x_stride;
    for {
      wrk.frow[x_out] = right * wrk.x_add + (left - right) * accum;
      x_out += x_stride;
      if x_out >= x_out_max { break }
      accum -= wrk.x_sub;
      if (accum < 0) {
        left = right;
        x_in += x_stride;
        assert.Assert(x_in < wrk.src_width * x_stride);
        right = rescaler_t(src[x_in])
        accum += wrk.x_add;
      }
    }
    assert.Assert(wrk.x_sub == 0 /* <- special case for src_width=1 */ || accum == 0);
  }
}

func WebPRescalerImportRowShrink_C(/* const */ wrk *WebPRescaler, /*const*/ src *uint8) {
  x_stride := wrk.num_channels;
  x_out_max := wrk.dst_width * wrk.num_channels;
  var channel int
  assert.Assert(!WebPRescalerInputDone(wrk));
  assert.As
  sert(!wrk.x_expand);
  for channel = 0; channel < x_stride; channel++ {
    x_in := channel;
    x_out := channel;
    sum := 0;
    accum := 0;
    for (x_out < x_out_max) {
      base := 0;
      accum += wrk.x_add;
      for (accum > 0) {
        accum -= wrk.x_sub;
        assert.Assert(x_in < wrk.src_width * x_stride);
        base = src[x_in];
        sum += base;
        x_in += x_stride;
      }
      {  // Emit next horizontal pixel.
         = base * (-accum);
        wrk.frow[x_out] = sum * wrk.x_sub - frac;
        // fresh fractional start for next pixel
        sum = (int)MULT_FIX(frac, wrk.fx_scale);
      }
      x_out += x_stride;
    }
    assert.Assert(accum == 0);
  }
}

//------------------------------------------------------------------------------
// Row export

func WebPRescalerExportRowExpand_C(/* const */ wrk *WebPRescaler) {
  var x_out int
  var dst *uint8 = wrk.dst;
  rescaler_t* const irow = wrk.irow;
  x_out_max := wrk.dst_width * wrk.num_channels;
  const rescaler_t* const frow = wrk.frow;
  assert.Assert(!WebPRescalerOutputDone(wrk));
  assert.Assert(wrk.y_accum <= 0);
  assert.Assert(wrk.y_expand);
  assert.Assert(wrk.y_sub != 0);
  if (wrk.y_accum == 0) {
    for x_out = 0; x_out < x_out_max; x_out++ {
      J := frow[x_out];
      v := (int)MULT_FIX(J, wrk.fy_scale);
      dst[x_out] = tenary.If(v > 255, uint(255), (uint8)v);
    }
  } else {
    B := WEBP_RESCALER_FRAC(-wrk.y_accum, wrk.y_sub);
    A := (uint32)(WEBP_RESCALER_ONE - B);
    for x_out = 0; x_out < x_out_max; x_out++ {
      I := (uint64)A * frow[x_out] + (uint64)B * irow[x_out];
      J := (uint32)((I + ROUNDER) >> WEBP_RESCALER_RFIX);
      v := (int)MULT_FIX(J, wrk.fy_scale);
      dst[x_out] = tenary.If(v > 255, uint(255), (uint8)v);
    }
  }
}

func WebPRescalerExportRowShrink_C(/* const */ wrk *WebPRescaler) {
  var x_out int
  var dst *uint8 = wrk.dst;
  rescaler_t* const irow = wrk.irow;
  x_out_max := wrk.dst_width * wrk.num_channels;
  const rescaler_t* const frow = wrk.frow;
  yscale := wrk.fy_scale * (-wrk.y_accum);
  assert.Assert(!WebPRescalerOutputDone(wrk));
  assert.Assert(wrk.y_accum <= 0);
  assert.Assert(!wrk.y_expand);
  if (yscale) {
    for x_out = 0; x_out < x_out_max; x_out++ {
      frac := (uint32)MULT_FIX_FLOOR(frow[x_out], yscale);
      v := (int)MULT_FIX(irow[x_out] - frac, wrk.fxy_scale);
      dst[x_out] = tenary.If(v > 255, uint(255), (uint8)v);
      irow[x_out] = frac;  // new fractional start
    }
  } else {
    for x_out = 0; x_out < x_out_max; x_out++ {
      v := (int)MULT_FIX(irow[x_out], wrk.fxy_scale);
      dst[x_out] = tenary.If(v > 255, uint(255), (uint8)v);
      irow[x_out] = 0;
    }
  }
}

#undef MULT_FIX_FLOOR
#undef MULT_FIX
#undef ROUNDER

//------------------------------------------------------------------------------
// Main entry calls

func WebPRescalerImportRow(/* const */ wrk *WebPRescaler, /*const*/ src *uint8) {
  assert.Assert(!WebPRescalerInputDone(wrk));
  if (!wrk.x_expand) {
    WebPRescalerImportRowShrink(wrk, src);
  } else {
    WebPRescalerImportRowExpand(wrk, src);
  }
}

func WebPRescalerExportRow(/* const */ wrk *WebPRescaler) {
  if (wrk.y_accum <= 0) {
    assert.Assert(!WebPRescalerOutputDone(wrk));
    if (wrk.y_expand) {
      WebPRescalerExportRowExpand(wrk);
    } else if (wrk.fxy_scale) {
      WebPRescalerExportRowShrink(wrk);
    } else {  // special case
      var i int
      assert.Assert(wrk.src_height == wrk.dst_height && wrk.x_add == 1);
      assert.Assert(wrk.src_width == 1 && wrk.dst_width <= 2);
      for i = 0; i < wrk.num_channels * wrk.dst_width; i++ {
        wrk.dst[i] = wrk.irow[i];
        wrk.irow[i] = 0;
      }
    }
    wrk.y_accum += wrk.y_add;
    wrk.dst += wrk.dst_stride;
    ++wrk.dst_y;
  }
}

//------------------------------------------------------------------------------

WebPRescalerImportRowFunc WebPRescalerImportRowExpand;
WebPRescalerImportRowFunc WebPRescalerImportRowShrink;

WebPRescalerExportRowFunc WebPRescalerExportRowExpand;
WebPRescalerExportRowFunc WebPRescalerExportRowShrink;

extern VP8CPUInfo VP8GetCPUInfo;
extern func WebPRescalerDspInitSSE2(void);
extern func WebPRescalerDspInitMIPS32(void);
extern func WebPRescalerDspInitMIPSdspR2(void);
extern func WebPRescalerDspInitMSA(void);
extern func WebPRescalerDspInitNEON(void);

WEBP_DSP_INIT_FUNC(WebPRescalerDspInit) {
#if !defined(WEBP_REDUCE_SIZE)
#if !WEBP_NEON_OMIT_C_CODE
  WebPRescalerExportRowExpand = WebPRescalerExportRowExpand_C;
  WebPRescalerExportRowShrink = WebPRescalerExportRowShrink_C;
#endif

  WebPRescalerImportRowExpand = WebPRescalerImportRowExpand_C;
  WebPRescalerImportRowShrink = WebPRescalerImportRowShrink_C;

  if (VP8GetCPUInfo != nil) {
#if false
    if (VP8GetCPUInfo(kSSE2)) {
      WebPRescalerDspInitSSE2();
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPS32)) {
      WebPRescalerDspInitMIPS32();
    }
#endif
#if false
    if (VP8GetCPUInfo(kMIPSdspR2)) {
      WebPRescalerDspInitMIPSdspR2();
    }
#endif
#if defined(WEBP_USE_MSA)
    if (VP8GetCPUInfo(kMSA)) {
      WebPRescalerDspInitMSA();
    }
#endif
  }

#if defined(WEBP_HAVE_NEON)
  if (WEBP_NEON_OMIT_C_CODE ||
      (VP8GetCPUInfo != nil && VP8GetCPUInfo(kNEON))) {
    WebPRescalerDspInitNEON();
  }
#endif

  assert.Assert(WebPRescalerExportRowExpand != nil);
  assert.Assert(WebPRescalerExportRowShrink != nil);
  assert.Assert(WebPRescalerImportRowExpand != nil);
  assert.Assert(WebPRescalerImportRowShrink != nil);
#endif  // WEBP_REDUCE_SIZE
}
