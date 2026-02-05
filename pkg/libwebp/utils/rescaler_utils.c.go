package utils

// Copyright 2012 Google Inc. All Rights Reserved.
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

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/limits"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------

int WebPRescalerInit(/* const */ rescaler *WebPRescaler, src_width int, src_height int, /*const*/ dst *uint8, dst_width int, dst_height int, dst_stride int, num_channels int, rescaler_t* const (uint64(2) * dst_width *
                                                       num_channels) work) {
  x_add := src_width, x_sub = dst_width;
  y_add := src_height, y_sub = dst_height;
  total_size := uint64(2) * dst_width * num_channels * sizeof(*work);
  if !CheckSizeOverflow(total_size) { return 0  }

  rescaler.x_expand = (src_width < dst_width);
  rescaler.y_expand = (src_height < dst_height);
  rescaler.src_width = src_width;
  rescaler.src_height = src_height;
  rescaler.dst_width = dst_width;
  rescaler.dst_height = dst_height;
  rescaler.src_y = 0;
  rescaler.dst_y = 0;
  rescaler.dst = dst;
  rescaler.dst_stride = dst_stride;
  rescaler.num_channels = num_channels;
  rescaler.irow = work;
  rescaler.frow = work + num_channels * dst_width;
  stdlib.Memset(work, 0, (uint64)total_size);

  // for 'x_expand', we use bilinear interpolation
  rescaler.x_add = rescaler.x_expand ? (x_sub - 1) : x_add;
  rescaler.x_sub = rescaler.x_expand ? (x_add - 1) : x_sub;
  if (!rescaler.x_expand) {  // fx_scale is not used otherwise
    rescaler.fx_scale = WEBP_RESCALER_FRAC(1, rescaler.x_sub);
  }
  // vertical scaling parameters
  rescaler.y_add = rescaler.y_expand ? y_add - 1 : y_add;
  rescaler.y_sub = rescaler.y_expand ? y_sub - 1 : y_sub;
  rescaler.y_accum = tenary.If(rescaler.y_expand, rescaler.y_sub, rescaler.y_add);
  if (!rescaler.y_expand) {
    // This is WEBP_RESCALER_FRAC(dst_height, x_add * y_add) without the cast.
    // Its value is <= WEBP_RESCALER_ONE, because dst_height <= rescaler.y_add
    // and rescaler.x_add >= 1;
    num := (uint64)dst_height * WEBP_RESCALER_ONE;
    den := (uint64)rescaler.x_add * rescaler.y_add;
    ratio := num / den;
    if (ratio != (uint32)ratio) {
      // When ratio == WEBP_RESCALER_ONE, we can't represent the ratio with the
      // current fixed-point precision. This happens when src_height ==
      // rescaler.y_add (which == src_height), and rescaler.x_add == 1.
      // => We special-case fxy_scale = 0, in WebPRescalerExportRow().
      rescaler.fxy_scale = 0;
    } else {
      rescaler.fxy_scale = (uint32)ratio;
    }
    rescaler.fy_scale = WEBP_RESCALER_FRAC(1, rescaler.y_sub);
  } else {
    rescaler.fy_scale = WEBP_RESCALER_FRAC(1, rescaler.x_add);
    // rescaler.fxy_scale is unused here.
  }

  WebPRescalerDspInit();
  return 1;
}

func WebPRescalerGetScaledDimensions(src_width int, src_height int, /*const*/ scaled_width *int, /*const*/ scaled_height *int) int {
  assert.Assert(scaled_width != nil);
  assert.Assert(scaled_height != nil);
  {
    width := *scaled_width;
    height := *scaled_height;
    max_size := INT_MAX / 2;

    // if width is unspecified, scale original proportionally to height ratio.
    if (width == 0 && src_height > 0) {
      width =
          (int)(((uint64)src_width * height + src_height - 1) / src_height);
    }
    // if height is unspecified, scale original proportionally to width ratio.
    if (height == 0 && src_width > 0) {
      height =
          (int)(((uint64)src_height * width + src_width - 1) / src_width);
    }
    // Check if the overall dimensions still make sense.
    if (width <= 0 || height <= 0 || width > max_size || height > max_size) {
      return 0;
    }

    *scaled_width = width;
    *scaled_height = height;
    return 1;
  }
}

//------------------------------------------------------------------------------
// all-in-one calls

func WebPRescaleNeededLines(/* const */ rescaler *WebPRescaler, max_num_lines int) int {
  num_lines :=
      (rescaler.y_accum + rescaler.y_sub - 1) / rescaler.y_sub;
  return (num_lines > max_num_lines) ? max_num_lines : num_lines;
}

func WebPRescalerImport(/* const */ rescaler *WebPRescaler, num_lines int, /*const*/ src *uint8, src_stride int) int {
  total_imported := 0;
  while (total_imported < num_lines &&
         !WebPRescalerHasPendingOutput(rescaler)) {
    if (rescaler.y_expand) {
      rescaler_t* const tmp = rescaler.irow;
      rescaler.irow = rescaler.frow;
      rescaler.frow = WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
          rescaler_t*, tmp, rescaler.num_channels * rescaler.dst_width * sizeof(*tmp));
    //   WEBP_SELF_ASSIGN(rescaler.dst_width);
    //   WEBP_SELF_ASSIGN(rescaler.num_channels);
    }
    WebPRescalerImportRow(rescaler, src);
    if (!rescaler.y_expand) {  // Accumulate the contribution of the new row.
      var x int
      for x = 0; x < rescaler.num_channels * rescaler.dst_width; x++ {
        rescaler.irow[x] += rescaler.frow[x];
      }
    }
    ++rescaler.src_y;
    src += src_stride;
    total_imported++
    rescaler.y_accum -= rescaler.y_sub;
  }
  return total_imported;
}

func WebPRescalerExport(/* const */ rescaler *WebPRescaler) int {
  total_exported := 0;
  while (WebPRescalerHasPendingOutput(rescaler)) {
    WebPRescalerExportRow(rescaler);
    total_exported++
  }
  return total_exported;
}

//------------------------------------------------------------------------------
