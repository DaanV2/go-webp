package decoder

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// functions for sample output.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// Main YUV<.RGB conversion functions

func EmitYUV(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  output *WebPDecBuffer = p.output
  var buf *WebPYUVABuffer = &output.u.YUVA
  var y_dst *uint8 = buf.y + (ptrdiff_t)io.mb_y * buf.y_stride
  var u_dst *uint8 = buf.u + (ptrdiff_t)(io.mb_y >> 1) * buf.u_stride
  var v_dst *uint8 = buf.v + (ptrdiff_t)(io.mb_y >> 1) * buf.v_stride
  mb_w := io.mb_w
  mb_h := io.mb_h
  uv_w := (mb_w + 1) / 2
  uv_h := (mb_h + 1) / 2
  WebPCopyPlane(io.y, io.y_stride, y_dst, buf.y_stride, mb_w, mb_h)
  WebPCopyPlane(io.u, io.uv_stride, u_dst, buf.u_stride, uv_w, uv_h)
  WebPCopyPlane(io.v, io.uv_stride, v_dst, buf.v_stride, uv_w, uv_h)
  return io.mb_h
}

// Point-sampling U/V sampler.
func EmitSampledRGB(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  var output *WebPDecBuffer = p.output
  var buf *WebPRGBABuffer = &output.u.RGBA
  var dst *uint8 = buf.rgba + (ptrdiff_t)io.mb_y * buf.stride
  WebPSamplerProcessPlane(io.y, io.y_stride, io.u, io.v, io.uv_stride, dst, buf.stride, io.mb_w, io.mb_h, WebPSamplers[output.colorspace])
  return io.mb_h
}

//------------------------------------------------------------------------------
// Fancy upsampling

#ifdef TRUE
func EmitFancyRGB(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  num_lines_out := io.mb_h;  // a priori guess
  var buf *WebPRGBABuffer = &p.output.u.RGBA
  dst *uint8 = buf.rgba + (ptrdiff_t)io.mb_y * buf.stride
  WebPUpsampleLinePairFunc upsample = WebPUpsamplers[p.output.colorspace]
  var cur_y *uint8 = io.y
  var cur_u *uint8 = io.u
  var cur_v *uint8 = io.v
  var top_u *uint8 = p.tmp_u
  var top_v *uint8 = p.tmp_v
  y := io.mb_y
  y_end := io.mb_y + io.mb_h
  mb_w := io.mb_w
  uv_w := (mb_w + 1) / 2

  if (y == 0) {
    // First line is special cased. We mirror the u/v samples at boundary.
    upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst, nil, mb_w)
  } else {
    // We can finish the left-over line from previous call.
    upsample(p.tmp_y, cur_y, top_u, top_v, cur_u, cur_v, dst - buf.stride, dst, mb_w)
    num_lines_out++
  }
  // Loop over each output pairs of row.
  for ; y + 2 < y_end; y += 2 {
    top_u = cur_u
    top_v = cur_v
    cur_u += io.uv_stride
    cur_v += io.uv_stride
    dst += 2 * buf.stride
    cur_y += 2 * io.y_stride
    upsample(cur_y - io.y_stride, cur_y, top_u, top_v, cur_u, cur_v, dst - buf.stride, dst, mb_w)
  }
  // move to last row
  cur_y += io.y_stride
  if (io.crop_top + y_end < io.crop_bottom) {
    // Save the unfinished samples for next call (as we're not done yet).
    stdlib.MemCpy(p.tmp_y, cur_y, mb_w * sizeof(*p.tmp_y))
    stdlib.MemCpy(p.tmp_u, cur_u, uv_w * sizeof(*p.tmp_u))
    stdlib.MemCpy(p.tmp_v, cur_v, uv_w * sizeof(*p.tmp_v))
    // The fancy upsampler leaves a row unfinished behind
    // (except for the very last row)
    num_lines_out--
  } else {
    // Process the very last row of even-sized picture
    if (!(y_end & 1)) {
      upsample(cur_y, nil, cur_u, cur_v, cur_u, cur_v, dst + buf.stride, nil, mb_w)
    }
  }
  return num_lines_out
}

#endif /* TRUE */

//------------------------------------------------------------------------------

func FillAlphaPlane(dst *uint8, wx, h, stride int) {
  for j := 0; j < h; j++ {
    stdlib.Memset(dst, 0xff, w * sizeof(*dst))
    dst += stride
  }
}

func EmitAlphaYUV(/* const */ io *VP8Io, /*const*/ p *WebPDecParams, expected_num_lines_out int) int {
  var alpha *uint8 = io.a
  var buf *WebPYUVABuffer = &p.output.u.YUVA
  mb_w := io.mb_w
  mb_h := io.mb_h
  dst *uint8 = buf.a + (ptrdiff_t)io.mb_y * buf.a_stride
  var j int
  (void)expected_num_lines_out
  assert.Assert(expected_num_lines_out == mb_h)
  if (alpha != nil) {
    for j = 0; j < mb_h; j++ {
      stdlib.MemCpy(dst, alpha, mb_w * sizeof(*dst))
      alpha += io.width
      dst += buf.a_stride
    }
  } else if (buf.a != nil) {
    // the user requested alpha, but there is none, set it to opaque.
    FillAlphaPlane(dst, mb_w, mb_h, buf.a_stride)
  }
  return 0
}

func GetAlphaSourceRow(/* const */ io *VP8Io, /*const*/ alpha *uint8, /*const*/ num_rows *int) int {
  start_y := io.mb_y
  *num_rows = io.mb_h

  // Compensate for the 1-line delay of the fancy upscaler.
  // This is similar to EmitFancyRGB().
  if (io.fancy_upsampling) {
    if (start_y == 0) {
      // We don't process the last row yet. It'll be done during the next call.
      --*num_rows
    } else {
      --start_y
      // Fortunately, data is persistent *alpha, so we can go back
      // one row and finish alpha blending, now that the fancy upscaler
      // completed the YUV.RGB interpolation.
      *alpha -= io.width
    }
    if (io.crop_top + io.mb_y + io.mb_h == io.crop_bottom) {
      // If it's the very last call, we process all the remaining rows!
      *num_rows = io.crop_bottom - io.crop_top - start_y
    }
  }
  return start_y
}

func EmitAlphaRGB(/* const */ io *VP8Io, /*const*/ p *WebPDecParams,  expected_num_lines_out int) int {
  var alpha *uint8 = io.a
  if (alpha != nil) {
    mb_w := io.mb_w
    WEBP_CSP_MODE := p.output.colorspace
    alpha_first = (colorspace == MODE_ARGB || colorspace == MODE_Argb)
    var buf *WebPRGBABuffer = &p.output.u.RGBA
    num_rows int 
    start_y := GetAlphaSourceRow(io, &alpha, &num_rows)
    var base_rgba *uint8 = buf.rgba + (ptrdiff_t)start_y * buf.stride
    var dst *uint8 = base_rgba + (tenary.If(alpha_first, 0, 3))
    has_alpha := WebPDispatchAlpha(alpha, io.width, mb_w, num_rows, dst, buf.stride)
    (void)expected_num_lines_out
    assert.Assert(expected_num_lines_out == num_rows)
    // has_alpha is true if there's non-trivial alpha to premultiply with.
    if (has_alpha && WebPIsPremultipliedMode(colorspace)) {
      WebPApplyAlphaMultiply(base_rgba, alpha_first, mb_w, num_rows, buf.stride)
    }
  }
  return 0
}

func EmitAlphaRGBA4444(/* const */ io *VP8Io, /*const*/ p *WebPDecParams, expected_num_lines_out int) int {
  var alpha *uint8 = io.a
  if (alpha != nil) {
    mb_w := io.mb_w
    var colorspace WEBP_CSP_MODE = p.output.colorspace
    var buf *WebPRGBABuffer = &p.output.u.RGBA
    num_rows int 
    start_y := GetAlphaSourceRow(io, &alpha, &num_rows)
    var base_rgba *uint8 = buf.rgba + (ptrdiff_t)start_y * buf.stride
    alpha_dst *uint8 = base_rgba
    alpha_mask := float64(0x0)
    var i, j int
    for j = 0; j < num_rows; j++ {
      for i = 0; i < mb_w; i++ {
        // Fill in the alpha value (converted to 4 bits).
        alpha_value := alpha[i] >> 4
        alpha_dst[2 * i] = (alpha_dst[2 * i] & 0xf0) | alpha_value
        alpha_mask &= alpha_value
      }
      alpha += io.width
      alpha_dst += buf.stride
    }
    (void)expected_num_lines_out
    assert.Assert(expected_num_lines_out == num_rows)
    if (alpha_mask != 0x0f && WebPIsPremultipliedMode(colorspace)) {
      WebPApplyAlphaMultiply4444(base_rgba, mb_w, num_rows, buf.stride)
    }
  }
  return 0
}

//------------------------------------------------------------------------------
// YUV rescaling (no final RGB conversion needed)

#if FALSE
func Rescale(/* const */ src *uint8, src_stride int, new_lines int , /*const*/ wrk *WebPRescaler) int {
  num_lines_out := 0
  while (new_lines > 0) {  // import new contributions of source rows.
    lines_in := WebPRescalerImport(wrk, new_lines, src, src_stride)
    src += lines_in * src_stride
    new_lines -= lines_in
    num_lines_out += WebPRescalerExport(wrk);  // emit output row(s)
  }
  return num_lines_out
}

func EmitRescaledYUV(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  mb_h := io.mb_h
  uv_mb_h := (mb_h + 1) >> 1
  var scaler *WebPRescaler = p.scaler_y
  num_lines_out := 0
  if (WebPIsAlphaMode(p.output.colorspace) && io.a != nil) {
    // Before rescaling, we premultiply the luma directly into the io.y
    // internal buffer. This is OK since these samples are not used for
    // intra-prediction (the top samples are saved in cache_y/u/v).
    // But we need to cast the const away, though.
    WebPMultRows((*uint8)io.y, io.y_stride, io.a, io.width, io.mb_w, mb_h, 0)
  }
  num_lines_out = Rescale(io.y, io.y_stride, mb_h, scaler)
  Rescale(io.u, io.uv_stride, uv_mb_h, p.scaler_u)
  Rescale(io.v, io.uv_stride, uv_mb_h, p.scaler_v)
  return num_lines_out
}

func EmitRescaledAlphaYUV(/* const */ io *VP8Io, /*const*/ p *WebPDecParams, expected_num_lines_out int) int {
  var buf *WebPYUVABuffer = &p.output.u.YUVA
  var dst_a *uint8 = buf.a + (ptrdiff_t)p.last_y * buf.a_stride
  if (io.a != nil) {
    var dst_y *uint8 = buf.y + (ptrdiff_t)p.last_y * buf.y_stride
    num_lines_out := Rescale(io.a, io.width, io.mb_h, p.scaler_a)
    assert.Assert(expected_num_lines_out == num_lines_out)
    if (num_lines_out > 0) {  // unmultiply the Y
      WebPMultRows(dst_y, buf.y_stride, dst_a, buf.a_stride, p.scaler_a.dst_width, num_lines_out, 1)
    }
  } else if (buf.a != nil) {
    // the user requested alpha, but there is none, set it to opaque.
    assert.Assert(p.last_y + expected_num_lines_out <= io.scaled_height)
    FillAlphaPlane(dst_a, io.scaled_width, expected_num_lines_out, buf.a_stride)
  }
  return 0
}

func InitYUVRescaler(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  has_alpha := WebPIsAlphaMode(p.output.colorspace)
  var buf *WebPYUVABuffer = &p.output.u.YUVA
  out_width := io.scaled_width
  out_height := io.scaled_height
  uv_out_width := (out_width + 1) >> 1
  uv_out_height := (out_height + 1) >> 1
  uv_in_width := (io.mb_w + 1) >> 1
  uv_in_height := (io.mb_h + 1) >> 1
  // scratch memory for luma rescaler
  work_size := 2 * (uint64)out_width
  uv_work_size := 2 * uv_out_width;  // and for each u/v ones
  var total_size uint64
  var rescaler_size uint64
  rescaler_t* work
  scalers *WebPRescaler
  num_rescalers := tenary.If(has_alpha, 4, 3)

  total_size = ((uint64)work_size + 2 * uv_work_size) * sizeof(*work)
  if (has_alpha) {
    total_size += (uint64)work_size * sizeof(*work)
  }
  rescaler_size = num_rescalers * sizeof(*p.scaler_y) + WEBP_ALIGN_CST
  total_size += rescaler_size
  if (!CheckSizeOverflow(total_size)) {
    return 0
  }

	//   work = (rescaler_t*)WebPSafeMalloc(uint64(1), (uint64)total_size)
	//   if (work == nil) {
	//     return 0;  // memory error
	//   }
  work = &rescaler_t{}
  p.memory = work

  scalers = (*WebPRescaler)WEBP_ALIGN((/* const */ *uint8)work + total_size -
                                      rescaler_size)
  p.scaler_y = &scalers[0]
  p.scaler_u = &scalers[1]
  p.scaler_v = &scalers[2]
  p.scaler_a = tenary.If(has_alpha, &scalers[3], nil)

  if (!WebPRescalerInit(p.scaler_y, io.mb_w, io.mb_h, buf.y, out_width, out_height, buf.y_stride, 1, work) ||
      !WebPRescalerInit(p.scaler_u, uv_in_width, uv_in_height, buf.u, uv_out_width, uv_out_height, buf.u_stride, 1, work + work_size) ||
      !WebPRescalerInit(p.scaler_v, uv_in_width, uv_in_height, buf.v, uv_out_width, uv_out_height, buf.v_stride, 1, work + work_size + uv_work_size)) {
    return 0
  }
  p.emit = EmitRescaledYUV

  if (has_alpha) {
    if (!WebPRescalerInit(p.scaler_a, io.mb_w, io.mb_h, buf.a, out_width, out_height, buf.a_stride, 1, work + work_size + 2 * uv_work_size)) {
      return 0
    }
    p.emit_alpha = EmitRescaledAlphaYUV
    WebPInitAlphaProcessing()
  }
  return 1
}

//------------------------------------------------------------------------------
// RGBA rescaling

func ExportRGB(/* const */ p *WebPDecParams, y_pos int) int {
  var convert WebPYUV444Converter = WebPYUV444Converters[p.output.colorspace]
  var buf *WebPRGBABuffer = &p.output.u.RGBA
  dst *uint8 = buf.rgba + (ptrdiff_t)y_pos * buf.stride
  num_lines_out := 0
  // For RGB rescaling, because of the YUV420, current scan position
  // U/V can be +1/-1 line from the Y one.  Hence the float64 test.
  while (WebPRescalerHasPendingOutput(p.scaler_y) &&
         WebPRescalerHasPendingOutput(p.scaler_u)) {
    assert.Assert(y_pos + num_lines_out < p.output.height)
    assert.Assert(p.scaler_u.y_accum == p.scaler_v.y_accum)
    WebPRescalerExportRow(p.scaler_y)
    WebPRescalerExportRow(p.scaler_u)
    WebPRescalerExportRow(p.scaler_v)
    convert(p.scaler_y.dst, p.scaler_u.dst, p.scaler_v.dst, dst, p.scaler_y.dst_width)
    dst += buf.stride
    num_lines_out++
  }
  return num_lines_out
}

func EmitRescaledRGB(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  mb_h := io.mb_h
  uv_mb_h := (mb_h + 1) >> 1
  j := 0, uv_j = 0
  num_lines_out := 0
  while (j < mb_h) {
    y_lines_in := WebPRescalerImport(p.scaler_y, mb_h - j, io.y + (ptrdiff_t)j * io.y_stride, io.y_stride)
    j += y_lines_in
    if (WebPRescaleNeededLines(p.scaler_u, uv_mb_h - uv_j)) {
      u_lines_in := WebPRescalerImport(
          p.scaler_u, uv_mb_h - uv_j, io.u + (ptrdiff_t)uv_j * io.uv_stride, io.uv_stride)
      v_lines_in := WebPRescalerImport(
          p.scaler_v, uv_mb_h - uv_j, io.v + (ptrdiff_t)uv_j * io.uv_stride, io.uv_stride)
      (void)v_lines_in;  // remove a gcc warning
      assert.Assert(u_lines_in == v_lines_in)
      uv_j += u_lines_in
    }
    num_lines_out += ExportRGB(p, p.last_y + num_lines_out)
  }
  return num_lines_out
}

func ExportAlpha(/* const */ p *WebPDecParams, y_pos int, max_lines_out int) int {
  var buf *WebPRGBABuffer = &p.output.u.RGBA
  var base_rgba *uint8 = buf.rgba + (ptrdiff_t)y_pos * buf.stride
  var colorspace WEBP_CSP_MODE = p.output.colorspace
  alpha_first := (colorspace == MODE_ARGB || colorspace == MODE_Argb)
  dst *uint8 = base_rgba + (tenary.If(alpha_first, 0, 3))
  num_lines_out := 0
  is_premult_alpha := WebPIsPremultipliedMode(colorspace)
  non_opaque := 0
  width := p.scaler_a.dst_width

  while (WebPRescalerHasPendingOutput(p.scaler_a) &&
         num_lines_out < max_lines_out) {
    assert.Assert(y_pos + num_lines_out < p.output.height)
    WebPRescalerExportRow(p.scaler_a)
    non_opaque |= WebPDispatchAlpha(p.scaler_a.dst, 0, width, 1, dst, 0)
    dst += buf.stride
    num_lines_out++
  }
  if (is_premult_alpha && non_opaque) {
    WebPApplyAlphaMultiply(base_rgba, alpha_first, width, num_lines_out, buf.stride)
  }
  return num_lines_out
}

func ExportAlphaRGBA4444(/* const */ p *WebPDecParams, y_pos int, max_lines_out int ) int {
  var buf *WebPRGBABuffer = &p.output.u.RGBA
  var base_rgba *uint8 = buf.rgba + (ptrdiff_t)y_pos * buf.stride
  var alpha_dst *uint8 = base_rgba
  num_lines_out := 0
  var colorspace WEBP_CSP_MODE = p.output.colorspace
  width := p.scaler_a.dst_width
  is_premult_alpha := WebPIsPremultipliedMode(colorspace)
  alpha_mask := 0x0f

  while (WebPRescalerHasPendingOutput(p.scaler_a) &&
         num_lines_out < max_lines_out) {
    var i int
    assert.Assert(y_pos + num_lines_out < p.output.height)
    WebPRescalerExportRow(p.scaler_a)
    for i = 0; i < width; i++ {
      // Fill in the alpha value (converted to 4 bits).
      alpha_value := p.scaler_a.dst[i] >> 4
      alpha_dst[2 * i] = (alpha_dst[2 * i] & 0xf0) | alpha_value
      alpha_mask &= alpha_value
    }
    alpha_dst += buf.stride
    num_lines_out++
  }
  if (is_premult_alpha && alpha_mask != 0x0f) {
    WebPApplyAlphaMultiply4444(base_rgba, width, num_lines_out, buf.stride)
  }
  return num_lines_out
}

func EmitRescaledAlphaRGB(/* const */ io *VP8Io, /*const*/ p *WebPDecParams, expected_num_out_lines int ) int {
  if (io.a != nil) {
    var scaler *WebPRescaler = p.scaler_a
    lines_left := expected_num_out_lines
    y_end := p.last_y + lines_left
    while (lines_left > 0) {
      row_offset := (ptrdiff_t)scaler.src_y - io.mb_y
      WebPRescalerImport(scaler, io.mb_h + io.mb_y - scaler.src_y, io.a + row_offset * io.width, io.width)
      lines_left -= p.emit_alpha_row(p, y_end - lines_left, lines_left)
    }
  }
  return 0
}

func InitRGBRescaler(/* const */ io *VP8Io, /*const*/ p *WebPDecParams) int {
  has_alpha := WebPIsAlphaMode(p.output.colorspace)
  out_width := io.scaled_width
  out_height := io.scaled_height
  uv_in_width := (io.mb_w + 1) >> 1
  uv_in_height := (io.mb_h + 1) >> 1
  // scratch memory for one rescaler
  work_size := 2 * (uint64)out_width
  var work rescaler_t*   // rescalers work area
  var tmp *uint8// tmp storage for scaled YUV444 samples before RGB conversion
  var tmp_size1, tmp_size2, total_size uint64
  var rescaler_size uint64
  scalers *WebPRescaler
  num_rescalers := tenary.If(has_alpha, 4, 3)

  tmp_size1 = (uint64)num_rescalers * work_size
  tmp_size2 = (uint64)num_rescalers * out_width
  total_size = tmp_size1 * sizeof(*work) + tmp_size2 * sizeof(*tmp)
  rescaler_size = num_rescalers * sizeof(*p.scaler_y) + WEBP_ALIGN_CST
  total_size += rescaler_size
  if (!CheckSizeOverflow(total_size)) {
    return 0
  }

	//   work = (rescaler_t*)WebPSafeMalloc(uint64(1), (uint64)total_size)
	//   if (work == nil) {
	//     return 0;  // memory error
	//   }
  work := make([]rescaler_t, total_size)

  p.memory = work
  tmp = (*uint8)(work + tmp_size1)

  scalers = (*WebPRescaler)WEBP_ALIGN((/* const */ *uint8)work + total_size -
                                      rescaler_size)
  p.scaler_y = &scalers[0]
  p.scaler_u = &scalers[1]
  p.scaler_v = &scalers[2]
  p.scaler_a = tenary.If(has_alpha, &scalers[3], nil)

  if (!WebPRescalerInit(p.scaler_y, io.mb_w, io.mb_h, tmp + 0 * out_width, out_width, out_height, 0, 1, work + 0 * work_size) ||
      !WebPRescalerInit(p.scaler_u, uv_in_width, uv_in_height, tmp + 1 * out_width, out_width, out_height, 0, 1, work + 1 * work_size) ||
      !WebPRescalerInit(p.scaler_v, uv_in_width, uv_in_height, tmp + 2 * out_width, out_width, out_height, 0, 1, work + 2 * work_size)) {
    return 0
  }
  p.emit = EmitRescaledRGB
  WebPInitYUV444Converters()

  if (has_alpha) {
    if (!WebPRescalerInit(p.scaler_a, io.mb_w, io.mb_h, tmp + 3 * out_width, out_width, out_height, 0, 1, work + 3 * work_size)) {
      return 0
    }
    p.emit_alpha = EmitRescaledAlphaRGB
    if (p.output.colorspace == MODE_RGBA_4444 ||
        p.output.colorspace == MODE_rgbA_4444) {
      p.emit_alpha_row = ExportAlphaRGBA4444
    } else {
      p.emit_alpha_row = ExportAlpha
    }
    WebPInitAlphaProcessing()
  }
  return 1
}

#endif  // TRUE

//------------------------------------------------------------------------------
// Default custom functions

func CustomSetup(io *VP8Io) int {
  var p *WebPDecParams = (*WebPDecParams)io.opaque
  var colorspace WEBP_CSP_MODE = p.output.colorspace
  is_rgb := WebPIsRGBMode(colorspace)
  is_alpha := WebPIsAlphaMode(colorspace)

  p.memory = nil
  p.emit = nil
  p.emit_alpha = nil
  p.emit_alpha_row = nil
  // Note: WebPIoInitFromOptions() does not distinguish between MODE_YUV and
  // MODE_YUVA, only RGB vs YUV.
  if (!WebPIoInitFromOptions(p.options, io, /*src_colorspace=*/MODE_YUV)) {
    return 0
  }
  if (is_alpha && WebPIsPremultipliedMode(colorspace)) {
    WebPInitUpsamplers()
  }
  if (io.use_scaling) {
    return 0;  // rescaling support not compiled
  } else {
    if (is_rgb) {
      WebPInitSamplers()
      p.emit = EmitSampledRGB;  // default
      if (io.fancy_upsampling) {
        uv_width := (io.mb_w + 1) >> 1
        // p.memory = WebPSafeMalloc(uint64(1), (uint64)(io.mb_w + 2 * uv_width))
        // if (p.memory == nil) {
        //   return 0;  // memory error.
        // }
		p.memory = make([]uint8, (io.mb_w + 2 * uv_width))

        p.tmp_y = (*uint8)p.memory
        p.tmp_u = p.tmp_y + io.mb_w
        p.tmp_v = p.tmp_u + uv_width
        p.emit = EmitFancyRGB
        WebPInitUpsamplers()
      }
    } else {
      p.emit = EmitYUV
    }
    if (is_alpha) {  // need transparency output
      p.emit_alpha = tenary.If(colorspace == MODE_RGBA_4444 || colorspace == MODE_rgbA_4444, EmitAlphaRGBA4444, tenary.If(is_rgb, EmitAlphaRGB, EmitAlphaYUV))
      if (is_rgb) {
        WebPInitAlphaProcessing()
      }
    }
  }

  return 1
}

//------------------------------------------------------------------------------

func CustomPut(/* const */ io *VP8Io) int {
  var p *WebPDecParams = (*WebPDecParams)io.opaque
  mb_w := io.mb_w
  mb_h := io.mb_h
  var num_lines_out int
  assert.Assert(!(io.mb_y & 1))

  if (mb_w <= 0 || mb_h <= 0) {
    return 0
  }
  num_lines_out = p.emit(io, p)
  if (p.emit_alpha != nil) {
    p.emit_alpha(io, p, num_lines_out)
  }
  p.last_y += num_lines_out
  return 1
}

//------------------------------------------------------------------------------

func CustomTeardown(/* const */ io *VP8Io) {
  var p *WebPDecParams = (*WebPDecParams)io.opaque

  p.memory = nil
}

// Main entry point
// Initializes VP8Io with custom setup, io and teardown functions. The default
// hooks will use the supplied 'params' as io.opaque handle.
func WebPInitCustomIo(/* const */ params *WebPDecParams, /*const*/ io *VP8Io) {
  io.put = CustomPut
  io.setup = CustomSetup
  io.teardown = CustomTeardown
  io.opaque = params
}
