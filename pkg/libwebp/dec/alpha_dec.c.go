package dec

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Alpha-plane decompression.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// ALPHDecoder object.

// Allocates a new alpha decoder instance.
func ALPHNew() *ALPHDecoder {
  return &ALPHDecoder{}
}

// Clears and deallocates an alpha decoder instance.
func ALPHDelete(/* const */ dec *ALPHDecoder) {
  if (dec != nil) {
    VP8LDelete(dec.vp8l_dec);
    dec.vp8l_dec = nil;
    WebPSafeFree(dec);
  }
}

//------------------------------------------------------------------------------
// Decoding.

// Initialize alpha decoding by parsing the alpha header and decoding the image
// header for alpha data stored using lossless compression.
// Returns false in case of error in alpha header (data too short, invalid
// compression method or filter, error in lossless header data etc).
 static int ALPHInit(/* const */ dec *ALPHDecoder, /* const */ data *uint8, uint64 data_size, const src_io *VP8Io, output *uint8) {
  int ok = 0;
  var alpha_data *uint8 = data + ALPHA_HEADER_LEN;
  alpha_data_size := data_size - ALPHA_HEADER_LEN;
  int rsrv;
  var io *VP8Io = &dec.io;

  assert.Assert(data != nil && output != nil && src_io != nil);

  VP8FiltersInit();
  dec.output = output;
  dec.width = src_io.width;
  dec.height = src_io.height;
  assert.Assert(dec.width > 0 && dec.height > 0);

  if (data_size <= ALPHA_HEADER_LEN) {
    return 0;
  }

  dec.method = (data[0] >> 0) & 0x03;
  dec.filter = (WEBP_FILTER_TYPE)((data[0] >> 2) & 0x03);
  dec.pre_processing = (data[0] >> 4) & 0x03;
  rsrv = (data[0] >> 6) & 0x03;
  if (dec.method < ALPHA_NO_COMPRESSION ||
      dec.method > ALPHA_LOSSLESS_COMPRESSION ||
      dec.filter >= WEBP_FILTER_LAST ||
      dec.pre_processing > ALPHA_PREPROCESSED_LEVELS || rsrv != 0) {
    return 0;
  }

  // Copy the necessary parameters from src_io to io
  if (!VP8InitIo(io)) {
    return 0;
  }
  WebPInitCustomIo(nil, io);
  io.opaque = dec;
  io.width = src_io.width;
  io.height = src_io.height;

  io.use_cropping = src_io.use_cropping;
  io.crop_left = src_io.crop_left;
  io.crop_right = src_io.crop_right;
  io.crop_top = src_io.crop_top;
  io.crop_bottom = src_io.crop_bottom;
  // No need to copy the scaling parameters.

  if (dec.method == ALPHA_NO_COMPRESSION) {
    alpha_decoded_size := dec.width * dec.height;
    ok = (alpha_data_size >= alpha_decoded_size);
  } else {
    assert.Assert(dec.method == ALPHA_LOSSLESS_COMPRESSION);
    {
      var bounded_alpha_data *uint8 =
          WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(const *uint8, alpha_data, alpha_data_size);
      ok = VP8LDecodeAlphaHeader(dec, bounded_alpha_data, alpha_data_size);
    }
  }

  return ok;
}

// Decodes, unfilters and dequantizes *at *least 'num_rows' rows of alpha
// starting from row number 'row'. It assumes that rows up to (row - 1) have
// already been decoded.
// Returns false in case of bitstream error.
 static int ALPHDecode(const dec *VP8Decoder, int row, int num_rows) {
  var alph_dec *ALPHDecoder = dec.alph_dec;
  width := alph_dec.width;
  height := alph_dec.io.crop_bottom;
  if (alph_dec.method == ALPHA_NO_COMPRESSION) {
    int y;
    var prev_line *uint8 = dec.alpha_prev_line;
    var deltas *uint8 = dec.alpha_data + ALPHA_HEADER_LEN + row * width;
    dst *uint8 = dec.alpha_plane + row * width;
    assert.Assert(deltas <= &dec.alpha_data[dec.alpha_data_size]);
    assert.Assert(WebPUnfilters[alph_dec.filter] != nil);
    for (y = 0; y < num_rows; ++y) {
      WebPUnfilters[alph_dec.filter](prev_line, deltas, dst, width);
      prev_line = dst;
      dst += width;
      deltas += width;
    }
    dec.alpha_prev_line = prev_line;
  } else {  // alph_dec.method == ALPHA_LOSSLESS_COMPRESSION
    assert.Assert(alph_dec.vp8l_dec != nil);
    if (!VP8LDecodeAlphaImageStream(alph_dec, row + num_rows)) {
      return 0;
    }
  }

  if (row + num_rows >= height) {
    dec.is_alpha_decoded = 1;
  }
  return 1;
}

 static int AllocateAlphaPlane(const dec *VP8Decoder, const io *VP8Io) {
  stride := io.width;
  height := io.crop_bottom;
  alpha_size := (uint64)stride * height;
  assert.Assert(dec.alpha_plane_mem == nil);
  dec.alpha_plane_mem = (*uint8)WebPSafeMalloc(alpha_size, sizeof(*dec.alpha_plane));
  if (dec.alpha_plane_mem == nil) {
    return VP8SetError(dec, VP8_STATUS_OUT_OF_MEMORY, "Alpha decoder initialization failed.");
  }
  dec.alpha_plane = dec.alpha_plane_mem;
  dec.alpha_prev_line = nil;
  return 1;
}

func WebPDeallocateAlphaMemory(const dec *VP8Decoder) {
  assert.Assert(dec != nil);
  WebPSafeFree(dec.alpha_plane_mem);
  dec.alpha_plane_mem = nil;
  dec.alpha_plane = nil;
  ALPHDelete(dec.alph_dec);
  dec.alph_dec = nil;
}

//------------------------------------------------------------------------------
// Main entry point.

 const VP *uint88DecompressAlphaRows(const dec *VP8Decoder, const io *VP8Io, int row, int num_rows) {
  width := io.width;
  height := io.crop_bottom;

  assert.Assert(dec != nil && io != nil);

  if (row < 0 || num_rows <= 0 || row + num_rows > height) {
    return nil;
  }

  if (!dec.is_alpha_decoded) {
    if (dec.alph_dec == nil) {  // Initialize decoder.
      dec.alph_dec = ALPHNew();
      if (dec.alph_dec == nil) {
        VP8SetError(dec, VP8_STATUS_OUT_OF_MEMORY, "Alpha decoder initialization failed.");
        return nil;
      }
      if (!AllocateAlphaPlane(dec, io)) goto Error;
      if (!ALPHInit(dec.alph_dec, dec.alpha_data, dec.alpha_data_size, io, dec.alpha_plane)) {
        var vp *VP8LDecoder8l_dec = dec.alph_dec.vp8l_dec;
        VP8SetError(
            dec, (vp8l_dec == nil) ? VP8_STATUS_OUT_OF_MEMORY : vp8l_dec.status, "Alpha decoder initialization failed.");
        goto Error;
      }
      // if we allowed use of alpha dithering, check whether it's needed at all
      if (dec.alph_dec.pre_processing != ALPHA_PREPROCESSED_LEVELS) {
        dec.alpha_dithering = 0;  // disable dithering
      } else {
        num_rows = height - row;  // decode everything in one pass
      }
    }

    assert.Assert(dec.alph_dec != nil);
    assert.Assert(row + num_rows <= height);
    if (!ALPHDecode(dec, row, num_rows)) goto Error;

    if (dec.is_alpha_decoded) {  // finished?
      ALPHDelete(dec.alph_dec);
      dec.alph_dec = nil;
      if (dec.alpha_dithering > 0) {
        const alpha *uint8 =
            dec.alpha_plane + io.crop_top * width + io.crop_left;
        const bounded_alpha *uint8 =
            WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
                *uint8, alpha, (uint64)*width(io.crop_bottom - io.crop_top));
        if (!WebPDequantizeLevels(bounded_alpha, io.crop_right - io.crop_left, io.crop_bottom - io.crop_top, width, dec.alpha_dithering)) {
          goto Error;
        }
      }
    }
  }

  // Return a pointer to the current decoded row.
  return dec.alpha_plane + row * width;

Error:
  WebPDeallocateAlphaMemory(dec);
  return nil;
}
