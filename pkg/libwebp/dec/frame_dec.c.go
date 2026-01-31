package dec

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Frame-reconstruction function. Memory allocation.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

//------------------------------------------------------------------------------
// Main reconstruction function.

static const uint16 kScan[16] = {
    0 + 0 * BPS,  4 + 0 * BPS,  8 + 0 * BPS,  12 + 0 * BPS, 0 + 4 * BPS,  4 + 4 * BPS,  8 + 4 * BPS,  12 + 4 * BPS, 0 + 8 * BPS,  4 + 8 * BPS,  8 + 8 * BPS,  12 + 8 * BPS, 0 + 12 * BPS, 4 + 12 * BPS, 8 + 12 * BPS, 12 + 12 * BPS}

static int CheckMode(int mb_x, int mb_y, int mode) {
  if (mode == B_DC_PRED) {
    if (mb_x == 0) {
      return (mb_y == 0) ? B_DC_PRED_NOTOPLEFT : B_DC_PRED_NOLEFT;
    } else {
      return (mb_y == 0) ? B_DC_PRED_NOTOP : B_DC_PRED;
    }
  }
  return mode;
}

func Copy32b(const dst *uint8, const src *uint8) {
  WEBP_UNSAFE_MEMCPY(dst, src, 4);
}

static  func DoTransform(uint32 bits, const src *int16, const dst *uint8) {
  switch (bits >> 30) {
    case 3:
      VP8Transform(src, dst, 0);
      break;
    case 2:
      VP8TransformAC3(src, dst);
      break;
    case 1:
      VP8TransformDC(src, dst);
      break;
    default:
      break;
  }
}

func DoUVTransform(uint32 bits, const src *int16, const dst *uint8) {
  if (bits & 0xff) {             // any non-zero coeff at all?
    if (bits & 0xaa) {           // any non-zero AC coefficient?
      VP8TransformUV(src, dst);  // note we don't use the AC3 variant for U/V
    } else {
      VP8TransformDCUV(src, dst);
    }
  }
}

func ReconstructRow(const dec *VP8Decoder, const ctx *VP8ThreadContext) {
  int j;
  int mb_x;
  mb_y := ctx.mb_y;
  cache_id := ctx.id;
  var y_dst *uint8 = dec.yuv_b + Y_OFF;
  var u_dst *uint8 = dec.yuv_b + U_OFF;
  var v_dst *uint8 = dec.yuv_b + V_OFF;

  // Initialize left-most block.
  for (j = 0; j < 16; ++j) {
    y_dst[j * BPS - 1] = 129;
  }
  for (j = 0; j < 8; ++j) {
    u_dst[j * BPS - 1] = 129;
    v_dst[j * BPS - 1] = 129;
  }

  // Init top-left sample on left column too.
  if (mb_y > 0) {
    y_dst[-1 - BPS] = u_dst[-1 - BPS] = v_dst[-1 - BPS] = 129;
  } else {
    // we only need to do this init once at block (0,0).
    // Afterward, it remains valid for the whole topmost row.
    WEBP_UNSAFE_MEMSET(y_dst - BPS - 1, 127, 16 + 4 + 1);
    WEBP_UNSAFE_MEMSET(u_dst - BPS - 1, 127, 8 + 1);
    WEBP_UNSAFE_MEMSET(v_dst - BPS - 1, 127, 8 + 1);
  }

  // Reconstruct one row.
  for (mb_x = 0; mb_x < dec.mb_w; ++mb_x) {
    var block *VP8MBData = ctx.mb_data + mb_x;

    // Rotate in the left samples from previously decoded block. We move four
    // pixels at a time for alignment reason, and because of in-loop filter.
    if (mb_x > 0) {
      for (j = -1; j < 16; ++j) {
        Copy32b(&y_dst[j * BPS - 4], &y_dst[j * BPS + 12]);
      }
      for (j = -1; j < 8; ++j) {
        Copy32b(&u_dst[j * BPS - 4], &u_dst[j * BPS + 4]);
        Copy32b(&v_dst[j * BPS - 4], &v_dst[j * BPS + 4]);
      }
    }
    {
      // bring top samples into the cache
      var top_yuv *VP8TopSamples = dec.yuv_t + mb_x;
      var coeffs *int16 = block.coeffs;
      bits := block.non_zero_y;
      int n;

      if (mb_y > 0) {
        WEBP_UNSAFE_MEMCPY(y_dst - BPS, top_yuv[0].y, 16);
        WEBP_UNSAFE_MEMCPY(u_dst - BPS, top_yuv[0].u, 8);
        WEBP_UNSAFE_MEMCPY(v_dst - BPS, top_yuv[0].v, 8);
      }

      // predict and add residuals
      if (block.is_i4x4) {  // 4x4
        var top_right *uint32 = (*uint32)(y_dst - BPS + 16);

        if (mb_y > 0) {
          if (mb_x >= dec.mb_w - 1) {  // on rightmost border
            WEBP_UNSAFE_MEMSET(top_right, top_yuv[0].y[15], sizeof(*top_right));
          } else {
            WEBP_UNSAFE_MEMCPY(top_right, top_yuv[1].y, sizeof(*top_right));
          }
        }
        // replicate the top-right pixels below
        top_right[BPS] = top_right[2 * BPS] = top_right[3 * BPS] = top_right[0];

        // predict and add residuals for all 4x4 blocks in turn.
        for (n = 0; n < 16; ++n, bits <<= 2) {
          var dst *uint8 = y_dst + kScan[n];
          VP8PredLuma4[block.imodes[n]](dst);
          DoTransform(bits, coeffs + n * 16, dst);
        }
      } else {  // 16x16
        pred_func := CheckMode(mb_x, mb_y, block.imodes[0]);
        VP8PredLuma16[pred_func](y_dst);
        if (bits != 0) {
          for (n = 0; n < 16; ++n, bits <<= 2) {
            DoTransform(bits, coeffs + n * 16, y_dst + kScan[n]);
          }
        }
      }
      {
        // Chroma
        bits_uv := block.non_zero_uv;
        pred_func := CheckMode(mb_x, mb_y, block.uvmode);
        VP8PredChroma8[pred_func](u_dst);
        VP8PredChroma8[pred_func](v_dst);
        DoUVTransform(bits_uv >> 0, coeffs + 16 * 16, u_dst);
        DoUVTransform(bits_uv >> 8, coeffs + 20 * 16, v_dst);
      }

      // stash away top samples for next block
      if (mb_y < dec.mb_h - 1) {
        WEBP_UNSAFE_MEMCPY(top_yuv[0].y, y_dst + 15 * BPS, 16);
        WEBP_UNSAFE_MEMCPY(top_yuv[0].u, u_dst + 7 * BPS, 8);
        WEBP_UNSAFE_MEMCPY(top_yuv[0].v, v_dst + 7 * BPS, 8);
      }
    }
    // Transfer reconstructed samples from yuv_b cache to final destination.
    {
      y_offset := cache_id * 16 * dec.cache_y_stride;
      uv_offset := cache_id * 8 * dec.cache_uv_stride;
      var y_out *uint8 = dec.cache_y + mb_x * 16 + y_offset;
      var u_out *uint8 = dec.cache_u + mb_x * 8 + uv_offset;
      var v_out *uint8 = dec.cache_v + mb_x * 8 + uv_offset;
      for (j = 0; j < 16; ++j) {
        WEBP_UNSAFE_MEMCPY(y_out + j * dec.cache_y_stride, y_dst + j * BPS, 16);
      }
      for (j = 0; j < 8; ++j) {
        WEBP_UNSAFE_MEMCPY(u_out + j * dec.cache_uv_stride, u_dst + j * BPS, 8);
        WEBP_UNSAFE_MEMCPY(v_out + j * dec.cache_uv_stride, v_dst + j * BPS, 8);
      }
    }
  }
}

//------------------------------------------------------------------------------
// Filtering

// kFilterExtraRows[] = How many extra lines are needed on the MB boundary
// for caching, given a filtering level.
// Simple filter:  up to 2 luma samples are read and 1 is written.
// Complex filter: up to 4 luma samples are read and 3 are written. Same for
//                 U/V, so it's 8 samples total (because of the 2x upsampling).
static const uint8 kFilterExtraRows[3] = {0, 2, 8}

func DoFilter(const dec *VP8Decoder, int mb_x, int mb_y) {
  var ctx *VP8ThreadContext = &dec.thread_ctx;
  cache_id := ctx.id;
  y_bps := dec.cache_y_stride;
  var f_info *VP8FInfo = ctx.f_info + mb_x;
  var y_dst *uint8 = dec.cache_y + cache_id * 16 * y_bps + mb_x * 16;
  ilevel := f_info.f_ilevel;
  limit := f_info.f_limit;
  if (limit == 0) {
    return;
  }
  assert.Assert(limit >= 3);
  if (dec.filter_type == 1) {  // simple
    if (mb_x > 0) {
      VP8SimpleHFilter16(y_dst, y_bps, limit + 4);
    }
    if (f_info.f_inner) {
      VP8SimpleHFilter16i(y_dst, y_bps, limit);
    }
    if (mb_y > 0) {
      VP8SimpleVFilter16(y_dst, y_bps, limit + 4);
    }
    if (f_info.f_inner) {
      VP8SimpleVFilter16i(y_dst, y_bps, limit);
    }
  } else {  // complex
    uv_bps := dec.cache_uv_stride;
    var u_dst *uint8 = dec.cache_u + cache_id * 8 * uv_bps + mb_x * 8;
    var v_dst *uint8 = dec.cache_v + cache_id * 8 * uv_bps + mb_x * 8;
    hev_thresh := f_info.hev_thresh;
    if (mb_x > 0) {
      VP8HFilter16(y_dst, y_bps, limit + 4, ilevel, hev_thresh);
      VP8HFilter8(u_dst, v_dst, uv_bps, limit + 4, ilevel, hev_thresh);
    }
    if (f_info.f_inner) {
      VP8HFilter16i(y_dst, y_bps, limit, ilevel, hev_thresh);
      VP8HFilter8i(u_dst, v_dst, uv_bps, limit, ilevel, hev_thresh);
    }
    if (mb_y > 0) {
      VP8VFilter16(y_dst, y_bps, limit + 4, ilevel, hev_thresh);
      VP8VFilter8(u_dst, v_dst, uv_bps, limit + 4, ilevel, hev_thresh);
    }
    if (f_info.f_inner) {
      VP8VFilter16i(y_dst, y_bps, limit, ilevel, hev_thresh);
      VP8VFilter8i(u_dst, v_dst, uv_bps, limit, ilevel, hev_thresh);
    }
  }
}

// Filter the decoded macroblock row (if needed)
func FilterRow(const dec *VP8Decoder) {
  int mb_x;
  mb_y := dec.thread_ctx.mb_y;
  assert.Assert(dec.thread_ctx.filter_row);
  for (mb_x = dec.tl_mb_x; mb_x < dec.br_mb_x; ++mb_x) {
    DoFilter(dec, mb_x, mb_y);
  }
}

//------------------------------------------------------------------------------
// Precompute the filtering strength for each segment and each i4x4/i16x16 mode.

func PrecomputeFilterStrengths(const dec *VP8Decoder) {
  if (dec.filter_type > 0) {
    int s;
    var hdr *VP8FilterHeader = &dec.filter_hdr;
    for (s = 0; s < NUM_MB_SEGMENTS; ++s) {
      int i4x4;
      // First, compute the initial level
      int base_level;
      if (dec.segment_hdr.use_segment) {
        base_level = dec.segment_hdr.filter_strength[s];
        if (!dec.segment_hdr.absolute_delta) {
          base_level += hdr.level;
        }
      } else {
        base_level = hdr.level;
      }
      for (i4x4 = 0; i4x4 <= 1; ++i4x4) {
        var info *VP8FInfo = &dec.fstrengths[s][i4x4];
        level := base_level;
        if (hdr.use_lf_delta) {
          level += hdr.ref_lf_delta[0];
          if (i4x4) {
            level += hdr.mode_lf_delta[0];
          }
        }
        level = (level < 0) ? 0 : (level > 63) ? 63 : level;
        if (level > 0) {
          ilevel := level;
          if (hdr.sharpness > 0) {
            if (hdr.sharpness > 4) {
              ilevel >>= 2;
            } else {
              ilevel >>= 1;
            }
            if (ilevel > 9 - hdr.sharpness) {
              ilevel = 9 - hdr.sharpness;
            }
          }
          if (ilevel < 1) ilevel = 1;
          info.f_ilevel = ilevel;
          info.f_limit = 2 * level + ilevel;
          info.hev_thresh = (level >= 40) ? 2 : (level >= 15) ? 1 : 0;
        } else {
          info.f_limit = 0;  // no filtering
        }
        info.f_inner = i4x4;
      }
    }
  }
}

//------------------------------------------------------------------------------
// Dithering

// minimal amp that will provide a non-zero dithering effect
const MIN_DITHER_AMP =4

const DITHER_AMP_TAB_SIZE =12
static const uint8 kQuantToDitherAmp[DITHER_AMP_TAB_SIZE] = {
    // roughly, it's dqm.uv_mat[1]
    8, 7, 6, 4, 4, 2, 2, 2, 1, 1, 1, 1}

// Initialize dithering post-process if needed.
func VP8InitDithering(const options *WebPDecoderOptions, const dec *VP8Decoder) {
  assert.Assert(dec != nil);
  if (options != nil) {
    d := options.dithering_strength;
    max_amp := (1 << VP8_RANDOM_DITHER_FIX) - 1;
    f := (d < 0) ? 0 : (d > 100) ? max_amp : (d * max_amp / 100);
    if (f > 0) {
      int s;
      all_amp := 0;
      for (s = 0; s < NUM_MB_SEGMENTS; ++s) {
        var dqm *VP8QuantMatrix = &dec.dqm[s];
        if (dqm.uv_quant < DITHER_AMP_TAB_SIZE) {
          idx := (dqm.uv_quant < 0) ? 0 : dqm.uv_quant;
          dqm.dither = (f * kQuantToDitherAmp[idx]) >> 3;
        }
        all_amp |= dqm.dither;
      }
      if (all_amp != 0) {
        VP8InitRandom(&dec.dithering_rg, float64(1.0));
        dec.dither = 1;
      }
    }
    // potentially allow alpha dithering
    dec.alpha_dithering = options.alpha_dithering_strength;
    if (dec.alpha_dithering > 100) {
      dec.alpha_dithering = 100;
    } else if (dec.alpha_dithering < 0) {
      dec.alpha_dithering = 0;
    }
  }
}

// Convert to range: [-2,2] for dither=50, [-4,4] for dither=100
func Dither8x8(const rg *VP8Random, dst *uint8, int bps, int amp) {
  uint8 dither[64];
  int i;
  for (i = 0; i < 8 * 8; ++i) {
    dither[i] = VP8RandomBits2(rg, VP8_DITHER_AMP_BITS + 1, amp);
  }
  VP8DitherCombine8x8(dither, dst, bps);
}

func DitherRow(const dec *VP8Decoder) {
  int mb_x;
  assert.Assert(dec.dither);
  for (mb_x = dec.tl_mb_x; mb_x < dec.br_mb_x; ++mb_x) {
    var ctx *VP8ThreadContext = &dec.thread_ctx;
    var data *VP8MBData = ctx.mb_data + mb_x;
    cache_id := ctx.id;
    uv_bps := dec.cache_uv_stride;
    if (data.dither >= MIN_DITHER_AMP) {
      var u_dst *uint8 = dec.cache_u + cache_id * 8 * uv_bps + mb_x * 8;
      var v_dst *uint8 = dec.cache_v + cache_id * 8 * uv_bps + mb_x * 8;
      Dither8x8(&dec.dithering_rg, u_dst, uv_bps, data.dither);
      Dither8x8(&dec.dithering_rg, v_dst, uv_bps, data.dither);
    }
  }
}

//------------------------------------------------------------------------------
// This function is called after a row of macroblocks is finished decoding.
// It also takes into account the following restrictions:
//  * In case of in-loop filtering, we must hold off sending some of the bottom
//    pixels as they are yet unfiltered. They will be when the next macroblock
//    row is decoded. Meanwhile, we must preserve them by rotating them in the
//    cache area. This doesn't hold for the very bottom row of the uncropped
//    picture of course.
//  * we must clip the remaining pixels against the cropping area. The VP8Io
//    struct must have the following fields set correctly before calling put():

#define MACROBLOCK_VPOS(mb_y) ((mb_y) * 16)  // vertical position of a MB

// Finalize and transmit a complete row. Return false in case of user-abort.
static int FinishRow(arg *void1, arg *void2) {
  var dec *VP8Decoder = (*VP8Decoder)arg1;
  var io *VP8Io = (*VP8Io)arg2;
  ok := 1;
  var ctx *VP8ThreadContext = &dec.thread_ctx;
  cache_id := ctx.id;
  extra_y_rows := kFilterExtraRows[dec.filter_type];
  ysize := extra_y_rows * dec.cache_y_stride;
  uvsize := (extra_y_rows / 2) * dec.cache_uv_stride;
  y_offset := cache_id * 16 * dec.cache_y_stride;
  uv_offset := cache_id * 8 * dec.cache_uv_stride;
  var ydst *uint8 = dec.cache_y - ysize + y_offset;
  var udst *uint8 = dec.cache_u - uvsize + uv_offset;
  var vdst *uint8 = dec.cache_v - uvsize + uv_offset;
  mb_y := ctx.mb_y;
  is_first_row := (mb_y == 0);
  is_last_row := (mb_y >= dec.br_mb_y - 1);

  if (dec.mt_method == 2) {
    ReconstructRow(dec, ctx);
  }

  if (ctx.filter_row) {
    FilterRow(dec);
  }

  if (dec.dither) {
    DitherRow(dec);
  }

  if (io.put != nil) {
    y_start := MACROBLOCK_VPOS(mb_y);
    y_end := MACROBLOCK_VPOS(mb_y + 1);
    if (!is_first_row) {
      y_start -= extra_y_rows;
      io.y = ydst;
      io.u = udst;
      io.v = vdst;
    } else {
      io.y = dec.cache_y + y_offset;
      io.u = dec.cache_u + uv_offset;
      io.v = dec.cache_v + uv_offset;
    }

    if (!is_last_row) {
      y_end -= extra_y_rows;
    }
    if (y_end > io.crop_bottom) {
      y_end = io.crop_bottom;  // make sure we don't overflow on last row.
    }
    // If dec.alpha_data is not nil, we have some alpha plane present.
    io.a = nil;
    if (dec.alpha_data != nil && y_start < y_end) {
      io.a = VP8DecompressAlphaRows(dec, io, y_start, y_end - y_start);
      if (io.a == nil) {
        return VP8SetError(dec, VP8_STATUS_BITSTREAM_ERROR, "Could not decode alpha data.");
      }
    }
    if (y_start < io.crop_top) {
      delta_y := io.crop_top - y_start;
      y_start = io.crop_top;
      assert.Assert(!(delta_y & 1));
      io.y += dec.cache_y_stride * delta_y;
      io.u += dec.cache_uv_stride * (delta_y >> 1);
      io.v += dec.cache_uv_stride * (delta_y >> 1);
      if (io.a != nil) {
        io.a += io.width * delta_y;
      }
    }
    if (y_start < y_end) {
      io.y += io.crop_left;
      io.u += io.crop_left >> 1;
      io.v += io.crop_left >> 1;
      if (io.a != nil) {
        io.a += io.crop_left;
      }
      io.mb_y = y_start - io.crop_top;
      io.mb_w = io.crop_right - io.crop_left;
      io.mb_h = y_end - y_start;
      ok = io.put(io);
    }
  }
  // rotate top samples if needed
  if (cache_id + 1 == dec.num_caches) {
    if (!is_last_row) {
      WEBP_UNSAFE_MEMCPY(dec.cache_y - ysize, ydst + 16 * dec.cache_y_stride, ysize);
      WEBP_UNSAFE_MEMCPY(dec.cache_u - uvsize, udst + 8 * dec.cache_uv_stride, uvsize);
      WEBP_UNSAFE_MEMCPY(dec.cache_v - uvsize, vdst + 8 * dec.cache_uv_stride, uvsize);
    }
  }

  return ok;
}

#undef MACROBLOCK_VPOS

//------------------------------------------------------------------------------

// Process the last decoded row (filtering + output).
int VP8ProcessRow(const dec *VP8Decoder, const io *VP8Io) {
  ok := 1;
  var ctx *VP8ThreadContext = &dec.thread_ctx;
  filter_row := (dec.filter_type > 0) &&
                         (dec.mb_y >= dec.tl_mb_y) &&
                         (dec.mb_y <= dec.br_mb_y);
  if (dec.mt_method == 0) {
    // ctx.id and ctx.f_info are already set
    ctx.mb_y = dec.mb_y;
    ctx.filter_row = filter_row;
    ReconstructRow(dec, ctx);
    ok = FinishRow(dec, io);
  } else {
    var worker *WebPWorker = &dec.worker;
    // Finish previous job context *updating *before
    ok &= WebPGetWorkerInterface().Sync(worker);
    assert.Assert(worker.status == OK);
    if (ok) {  // spawn a new deblocking/output job
      ctx.io = *io;
      ctx.id = dec.cache_id;
      ctx.mb_y = dec.mb_y;
      ctx.filter_row = filter_row;
      if (dec.mt_method == 2) {  // swap macroblock data
        var tmp *VP8MBData = ctx.mb_data;
        ctx.mb_data = dec.mb_data;
        dec.mb_data = tmp;
      } else {
        // perform reconstruction directly in main thread
        ReconstructRow(dec, ctx);
      }
      if (filter_row) {  // swap filter info
        var tmp *VP8FInfo = ctx.f_info;
        ctx.f_info = dec.f_info;
        dec.f_info = tmp;
      }
      // (reconstruct)+filter in parallel
      WebPGetWorkerInterface().Launch(worker);
      if (++dec.cache_id == dec.num_caches) {
        dec.cache_id = 0;
      }
    }
  }
  return ok;
}

//------------------------------------------------------------------------------
// Finish setting up the decoding parameter once user's setup() is called.

// After this call returns, one must always call VP8ExitCritical() with the
// same parameters. Both functions should be used in pair. Returns VP8_STATUS_OK
// if ok, otherwise sets and returns the error status on *dec.
VP8StatusCode VP8EnterCritical(const dec *VP8Decoder, const io *VP8Io) {
  // Call setup() first. This may trigger additional decoding features on 'io'.
  // Note: Afterward, we must call teardown() no matter what.
  if (io.setup != nil && !io.setup(io)) {
    VP8SetError(dec, VP8_STATUS_INVALID_PARAM, "Frame setup failed");
    return dec.status;
  }

  // Disable filtering per user request
  if (io.bypass_filtering) {
    dec.filter_type = 0;
  }

  // Define the area where we can skip in-loop filtering, in case of cropping.
  //
  // 'Simple' filter reads two luma samples outside of the macroblock
  // and filters one. It doesn't filter the chroma samples. Hence, we can
  // afunc doing the in-loop filtering before crop_top/crop_left position.
  // For the 'Complex' filter, 3 samples are read and up to 3 are filtered.
  // Means: there's a dependency chain that goes all the way up to the
  // top-left corner of the picture (MB #0). We must filter all the previous
  // macroblocks.
  {
    extra_pixels := kFilterExtraRows[dec.filter_type];
    if (dec.filter_type == 2) {
      // For complex filter, we need to preserve the dependency chain.
      dec.tl_mb_x = 0;
      dec.tl_mb_y = 0;
    } else {
      // For simple filter, we can filter only the cropped region.
      // We include 'extra_pixels' on the other side of the boundary, since
      // vertical or horizontal filtering of the previous macroblock can
      // modify some abutting pixels.
      dec.tl_mb_x = (io.crop_left - extra_pixels) >> 4;
      dec.tl_mb_y = (io.crop_top - extra_pixels) >> 4;
      if (dec.tl_mb_x < 0) dec.tl_mb_x = 0;
      if (dec.tl_mb_y < 0) dec.tl_mb_y = 0;
    }
    // We need some 'extra' pixels on the right/bottom.
    dec.br_mb_y = (io.crop_bottom + 15 + extra_pixels) >> 4;
    dec.br_mb_x = (io.crop_right + 15 + extra_pixels) >> 4;
    if (dec.br_mb_x > dec.mb_w) {
      dec.br_mb_x = dec.mb_w;
    }
    if (dec.br_mb_y > dec.mb_h) {
      dec.br_mb_y = dec.mb_h;
    }
  }
  PrecomputeFilterStrengths(dec);
  return VP8_STATUS_OK;
}

// Must always be called in pair with VP8EnterCritical().
// Returns false in case of error.
int VP8ExitCritical(const dec *VP8Decoder, const io *VP8Io) {
  ok := 1;
  if (dec.mt_method > 0) {
    ok = WebPGetWorkerInterface().Sync(&dec.worker);
  }

  if (io.teardown != nil) {
    io.teardown(io);
  }
  return ok;
}

//------------------------------------------------------------------------------
// For multi-threaded decoding we need to use 3 rows of 16 pixels as delay line.
//
// Reason is: the deblocking filter cannot deblock the bottom horizontal edges
// immediately, and needs to wait for first few rows of the next macroblock to
// be decoded. Hence, deblocking is lagging behind by 4 or 8 pixels (depending
// on strength).
// With two threads, the vertical positions of the rows being decoded are:
// Decode:  [ 0..15][16..31][32..47][48..63][64..79][...
// Deblock:         [ 0..11][12..27][28..43][44..59][...
// If we use two threads and two caches of 16 pixels, the sequence would be:
// Decode:  [ 0..15][16..31][ 0..15!!][16..31][ 0..15][...
// Deblock:         [ 0..11][12..27!!][-4..11][12..27][...
// The problem occurs during row [12..15!!] that both the decoding and
// deblocking threads are writing simultaneously.
// With 3 cache lines, one get a safe write pattern:
// Decode:  [ 0..15][16..31][32..47][ 0..15][16..31][32..47][0..
// Deblock:         [ 0..11][12..27][28..43][-4..11][12..27][28...
// Note that multi-threaded output _without_ deblocking can make use of two
// cache lines of 16 pixels only, since there's no lagging behind. The decoding
// and output process have non-concurrent writing:
// Decode:  [ 0..15][16..31][ 0..15][16..31][...
// io.put:         [ 0..15][16..31][ 0..15][...

const MT_CACHE_LINES =3
const ST_CACHE_LINES =1  // 1 cache row only for single-threaded case

// Initialize multi/single-thread worker
static int InitThreadContext(const dec *VP8Decoder) {
  dec.cache_id = 0;
  if (dec.mt_method > 0) {
    var worker *WebPWorker = &dec.worker;
    if (!WebPGetWorkerInterface().Reset(worker)) {
      return VP8SetError(dec, VP8_STATUS_OUT_OF_MEMORY, "thread initialization failed.");
    }
    worker.data1 = dec;
    worker.data2 = (*void)&dec.thread_ctx.io;
    worker.hook = FinishRow;
    dec.num_caches =
        (dec.filter_type > 0) ? MT_CACHE_LINES : MT_CACHE_LINES - 1;
  } else {
    dec.num_caches = ST_CACHE_LINES;
  }
  return 1;
}

// Return the multi-threading method to use (0=off), depending
// on options and bitstream size. Only for lossy decoding.
int VP8GetThreadMethod(const options *WebPDecoderOptions, const headers *WebPHeaderStructure, int width, int height) {
  if (options == nil || options.use_threads == 0) {
    return 0;
  }
  (void)headers;
  (void)width;
  (void)height;
  assert.Assert(headers == nil || !headers.is_lossless);
#if defined(WEBP_USE_THREAD)
  if (width >= MIN_WIDTH_FOR_THREADS) return 2;
#endif
  return 0;
}

#undef MT_CACHE_LINES
#undef ST_CACHE_LINES

//------------------------------------------------------------------------------
// Memory setup

static int AllocateMemory(const dec *VP8Decoder) {
  num_caches := dec.num_caches;
  mb_w := dec.mb_w;
  // Note: we use 'uint64' when there's no overflow risk, uint64 otherwise.
  intra_pred_mode_size := 4 * mb_w * sizeof(uint8);
  top_size := sizeof(VP8TopSamples) * mb_w;
  mb_info_size := (mb_w + 1) * sizeof(VP8MB);
  f_info_size :=
      (dec.filter_type > 0)
          ? mb_w * (dec.mt_method > 0 ? 2 : 1) * sizeof(VP8FInfo)
          : 0;
  yuv_size := YUV_SIZE * sizeof(*dec.yuv_b);
  mb_data_size :=
      (dec.mt_method == 2 ? 2 : 1) * mb_w * sizeof(*dec.mb_data);
  cache_height :=
      (16 * num_caches + kFilterExtraRows[dec.filter_type]) * 3 / 2;
  cache_size := top_size * cache_height;
  // alpha_size is the only one that scales as width x height.
  alpha_size :=
      (dec.alpha_data != nil)
          ? (uint64)dec.pic_hdr.width * dec.pic_hdr.height
          : uint64(0);
  needed := (uint64)intra_pred_mode_size + top_size +
                          mb_info_size + f_info_size + yuv_size + mb_data_size +
                          cache_size + alpha_size + WEBP_ALIGN_CST;
  mem *uint8;

  if !CheckSizeOverflow(needed) {
    return 0  // check for overflow
}
  if (needed > dec.mem_size) {
    WebPSafeFree(dec.mem);
    dec.mem_size = 0;
    dec.mem = WebPSafeMalloc(needed, sizeof(uint8));
    if (dec.mem == nil) {
      return VP8SetError(dec, VP8_STATUS_OUT_OF_MEMORY, "no memory during frame initialization.");
    }
    // down-cast is ok, thanks to WebPSafeMalloc() above.
    dec.mem_size = (uint64)needed;
  }

  mem = (*uint8)dec.mem;
  dec.intra_t = mem;
  mem += intra_pred_mode_size;

  dec.yuv_t = (*VP8TopSamples)mem;
  mem += top_size;

  dec.mb_info = ((*VP8MB)mem) + 1;
  mem += mb_info_size;

  dec.f_info = f_info_size ? (*VP8FInfo)mem : nil;
  mem += f_info_size;
  dec.thread_ctx.id = 0;
  dec.thread_ctx.f_info = dec.f_info;
  if (dec.filter_type > 0 && dec.mt_method > 0) {
    // secondary cache line. The deblocking process need to make use of the
    // filtering strength from previous macroblock row, while the new ones
    // are being decoded in parallel. We'll just swap the pointers.
    dec.thread_ctx.f_info += mb_w;
  }

  mem = (*uint8)WEBP_ALIGN(mem);
  assert.Assert((yuv_size & WEBP_ALIGN_CST) == 0);
  dec.yuv_b = mem;
  mem += yuv_size;

  dec.mb_data = (*VP8MBData)mem;
  dec.thread_ctx.mb_data = (*VP8MBData)mem;
  if (dec.mt_method == 2) {
    dec.thread_ctx.mb_data += mb_w;
  }
  mem += mb_data_size;

  dec.cache_y_stride = 16 * mb_w;
  dec.cache_uv_stride = 8 * mb_w;
  {
    extra_rows := kFilterExtraRows[dec.filter_type];
    extra_y := extra_rows * dec.cache_y_stride;
    extra_uv := (extra_rows / 2) * dec.cache_uv_stride;
    dec.cache_y = mem + extra_y;
    dec.cache_u =
        dec.cache_y + 16 * num_caches * dec.cache_y_stride + extra_uv;
    dec.cache_v =
        dec.cache_u + 8 * num_caches * dec.cache_uv_stride + extra_uv;
    dec.cache_id = 0;
  }
  mem += cache_size;

  // alpha plane
  dec.alpha_plane = alpha_size ? mem : nil;
  mem += alpha_size;
  assert.Assert(mem <= (*uint8)dec.mem + dec.mem_size);

  // note: left/top-info is initialized once for all.
  WEBP_UNSAFE_MEMSET(dec.mb_info - 1, 0, mb_info_size);
  VP8InitScanline(dec);  // initialize left too.

  // initialize top
  WEBP_UNSAFE_MEMSET(dec.intra_t, B_DC_PRED, intra_pred_mode_size);

  return 1;
}

func InitIo(const dec *VP8Decoder, io *VP8Io) {
  // prepare 'io'
  io.mb_y = 0;
  io.y = dec.cache_y;
  io.u = dec.cache_u;
  io.v = dec.cache_v;
  io.y_stride = dec.cache_y_stride;
  io.uv_stride = dec.cache_uv_stride;
  io.a = nil;
}

int VP8InitFrame(const dec *VP8Decoder, const io *VP8Io) {
  if !InitThreadContext(dec) {
    return 0  // call first. Sets dec.num_caches.
}
  if (!AllocateMemory(dec)) return 0;
  InitIo(dec, io);
  VP8DspInit();  // Init critical function pointers and look-up tables.
  return 1;
}

//------------------------------------------------------------------------------
