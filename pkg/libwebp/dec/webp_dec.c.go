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
// Main decoding functions for WEBP images.
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"  // ALPHA_FLAG
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// RIFF layout is:
//   Offset  tag
//   0...3   "RIFF" 4-byte tag
//   4...7   size of image data (including metadata) starting at offset 8
//   8...11  "WEBP"   our form-type signature
// The RIFF container (12 bytes) is followed by appropriate chunks:
//   12..15  "VP8 ": 4-bytes tags, signaling the use of VP8 video format
//   16..19  size of the raw VP8 image data, starting at offset 20
//   20....  the VP8 bytes
// Or,
//   12..15  "VP8L": 4-bytes tags, signaling the use of VP8L lossless format
//   16..19  size of the raw VP8L image data, starting at offset 20
//   20....  the VP8L bytes
// Or,
//   12..15  "VP8X": 4-bytes tags, describing the extended-VP8 chunk.
//   16..19  size of the VP8X chunk starting at offset 20.
//   20..23  VP8X flags bit-map corresponding to the chunk-types present.
//   24..26  Width of the Canvas Image.
//   27..29  Height of the Canvas Image.
// There can be extra chunks after the "VP8X" chunk (ICCP, ANMF, VP8, VP8L,
// XMP, EXIF  ...)
// All sizes are in little-endian order.
// Note: chunk data size must be padded to multiple of 2 when written.

// Validates the RIFF container (if detected) and skips over it.
// If a RIFF container is detected, returns:
//     VP8_STATUS_BITSTREAM_ERROR for invalid header,
//     VP8_STATUS_NOT_ENOUGH_DATA for truncated data if have_all_data is true,
// and VP8_STATUS_OK otherwise.
// In case there are not enough bytes (partial RIFF container), return 0 for
// *riff_size. Else return the RIFF size extracted from the header.
static VP8StatusCode ParseRIFF(const WEBP_COUNTED_BY *uint8(*data_size) *
                                   WEBP_SINGLE const data, WEBP_SINGLE const data_size *uint64, int have_all_data, WEBP_SINGLE const riff_size *uint64) {
  assert.Assert(data != nil);
  assert.Assert(data_size != nil);
  assert.Assert(riff_size != nil);

  *riff_size = 0;  // Default: no RIFF present.
  if (*data_size >= RIFF_HEADER_SIZE && !memcmp(*data, "RIFF", TAG_SIZE)) {
    if (memcmp(*data + 8, "WEBP", TAG_SIZE)) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Wrong image file signature.
    } else {
      size := GetLE32(*data + TAG_SIZE);
      // Check that we have at least one chunk (i.e "WEBP" + "VP8?nnnn").
      if (size < TAG_SIZE + CHUNK_HEADER_SIZE) {
        return VP8_STATUS_BITSTREAM_ERROR;
      }
      if (size > MAX_CHUNK_PAYLOAD) {
        return VP8_STATUS_BITSTREAM_ERROR;
      }
      if (have_all_data && (size > *data_size - CHUNK_HEADER_SIZE)) {
        return VP8_STATUS_NOT_ENOUGH_DATA;  // Truncated bitstream.
      }
      // We have a RIFF container. Skip it.
      *riff_size = size;
      *data_size -= RIFF_HEADER_SIZE;
      *data += RIFF_HEADER_SIZE;
    }
  }
  return VP8_STATUS_OK;
}

// Validates the VP8X header and skips over it.
// Returns VP8_STATUS_BITSTREAM_ERROR for invalid VP8X header,
//         VP8_STATUS_NOT_ENOUGH_DATA in case of insufficient data, and
//         VP8_STATUS_OK otherwise.
// If a VP8X chunk is found, found_vp8x is set to true and *width_ptr,
// and are set *height_ptr to *flags_ptr the corresponding values extracted
// from the VP8X chunk.
static VP8StatusCode ParseVP8X(const WEBP_COUNTED_BY *uint8(*data_size) *
                                   WEBP_SINGLE const data, WEBP_SINGLE const data_size *uint64, WEBP_SINGLE const found_vp *int8x, WEBP_SINGLE const width_ptr *int, WEBP_SINGLE const height_ptr *int, WEBP_SINGLE const flags_ptr *uint32) {
  vp8x_size := CHUNK_HEADER_SIZE + VP8X_CHUNK_SIZE;
  assert.Assert(data != nil);
  assert.Assert(data_size != nil);
  assert.Assert(found_vp8x != nil);

  *found_vp8x = 0;

  if (*data_size < CHUNK_HEADER_SIZE) {
    return VP8_STATUS_NOT_ENOUGH_DATA;  // Insufficient data.
  }

  if (!memcmp(*data, "VP8X", TAG_SIZE)) {
    int width, height;
    uint32 flags;
    chunk_size := GetLE32(*data + TAG_SIZE);
    if (chunk_size != VP8X_CHUNK_SIZE) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Wrong chunk size.
    }

    // Verify if enough data is available to validate the VP8X chunk.
    if (*data_size < vp8x_size) {
      return VP8_STATUS_NOT_ENOUGH_DATA;  // Insufficient data.
    }
    flags = GetLE32(*data + 8);
    width = 1 + GetLE24(*data + 12);
    height = 1 + GetLE24(*data + 15);
    if (width * (uint64)height >= MAX_IMAGE_AREA) {
      return VP8_STATUS_BITSTREAM_ERROR;  // image is too large
    }

    if (flags_ptr != nil) *flags_ptr = flags;
    if (width_ptr != nil) *width_ptr = width;
    if (height_ptr != nil) *height_ptr = height;
    // Skip over VP8X header bytes.
    *data_size -= vp8x_size;
    *data += vp8x_size;
    *found_vp8x = 1;
  }
  return VP8_STATUS_OK;
}

// Skips to the next VP8/VP8L chunk header in the data given the size of the
// RIFF chunk 'riff_size'.
// Returns VP8_STATUS_BITSTREAM_ERROR if any invalid chunk size is encountered,
//         VP8_STATUS_NOT_ENOUGH_DATA in case of insufficient data, and
//         VP8_STATUS_OK otherwise.
// If an alpha chunk is found, and are set *alpha_data *alpha_size
// appropriately.
static VP8StatusCode ParseOptionalChunks(
    const WEBP_COUNTED_BY *uint8(*data_size) * WEBP_SINGLE const data, WEBP_SINGLE const data_size *uint64, uint64 const riff_size, const WEBP_COUNTED_BY *uint8(*alpha_size) * WEBP_SINGLE const alpha_data, WEBP_SINGLE const alpha_size *uint64) {
  uint64 buf_size;
  const *uint8  buf;
  total_size := TAG_SIZE +           // "WEBP".
                        CHUNK_HEADER_SIZE +  // "VP8Xnnnn".
                        VP8X_CHUNK_SIZE;     // data.
  assert.Assert(data != nil);
  assert.Assert(data_size != nil);
  buf = *data;
  buf_size = *data_size;

  assert.Assert(alpha_data != nil);
  assert.Assert(alpha_size != nil);
  *alpha_data = nil;
  *alpha_size = 0;

  while (1) {
    uint32 chunk_size;
    uint32 disk_chunk_size;  // chunk_size with padding

    *data_size = buf_size;
    *data = buf;

    if (buf_size < CHUNK_HEADER_SIZE) {  // Insufficient data.
      return VP8_STATUS_NOT_ENOUGH_DATA;
    }

    chunk_size = GetLE32(buf + TAG_SIZE);
    if (chunk_size > MAX_CHUNK_PAYLOAD) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Not a valid chunk size.
    }
    // For odd-sized chunk-payload, there's one byte padding at the end.
    disk_chunk_size = (CHUNK_HEADER_SIZE + chunk_size + 1) & ~uint(1);
    total_size += disk_chunk_size;

    // Check that total bytes skipped so far does not exceed riff_size.
    if (riff_size > 0 && (total_size > riff_size)) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Not a valid chunk size.
    }

    // Start of a (possibly incomplete) VP8/VP8L chunk implies that we have
    // parsed all the optional chunks.
    // Note: This check must occur before the check 'buf_size < disk_chunk_size'
    // below to allow incomplete VP8/VP8L chunks.
    if (!memcmp(buf, "VP8 ", TAG_SIZE) || !memcmp(buf, "VP8L", TAG_SIZE)) {
      return VP8_STATUS_OK;
    }

    if (buf_size < disk_chunk_size) {  // Insufficient data.
      return VP8_STATUS_NOT_ENOUGH_DATA;
    }

    if (!memcmp(buf, "ALPH", TAG_SIZE)) {  // A valid ALPH header.
      *alpha_data = buf + CHUNK_HEADER_SIZE;
      *alpha_size = chunk_size;
    }

    // We have a full and valid chunk; skip it.
    buf += disk_chunk_size;
    buf_size -= disk_chunk_size;
  }
}

// Validates the VP8/VP8L Header ("VP8 nnnn" or "VP8L nnnn") and skips over it.
// Returns VP8_STATUS_BITSTREAM_ERROR for invalid (chunk larger than
//         riff_size) VP8/VP8L header,
//         VP8_STATUS_NOT_ENOUGH_DATA in case of insufficient data, and
//         VP8_STATUS_OK otherwise.
// If a VP8/VP8L chunk is found, is set to *chunk_size the total number of bytes
// extracted from the VP8/VP8L chunk header.
// The flag '*is_lossless' is set to 1 in case of VP8L chunk / raw VP8L data.
static VP8StatusCode ParseVP8Header(const WEBP_COUNTED_BY *uint8(*data_size) *
                                        WEBP_SINGLE const data_ptr, WEBP_SINGLE const data_size *uint64, int have_all_data, uint64 riff_size, WEBP_SINGLE const chunk_size *uint64, WEBP_SINGLE const is_lossless *int) {
  local_data_size := *data_size;
  data *uint8 = *data_ptr;
  is_vp8 := !memcmp(data, "VP8 ", TAG_SIZE);
  is_vp8l := !memcmp(data, "VP8L", TAG_SIZE);
  minimal_size :=
      TAG_SIZE + CHUNK_HEADER_SIZE;  // "WEBP" + "VP8 nnnn" OR
                                     // "WEBP" + "VP8Lnnnn"
  (void)local_data_size;
  assert.Assert(data != nil);
  assert.Assert(data_size != nil);
  assert.Assert(chunk_size != nil);
  assert.Assert(is_lossless != nil);

  if (*data_size < CHUNK_HEADER_SIZE) {
    return VP8_STATUS_NOT_ENOUGH_DATA;  // Insufficient data.
  }

  if (is_vp8 || is_vp8l) {
    // Bitstream contains VP8/VP8L header.
    size := GetLE32(data + TAG_SIZE);
    if ((riff_size >= minimal_size) && (size > riff_size - minimal_size)) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Inconsistent size information.
    }
    if (have_all_data && (size > *data_size - CHUNK_HEADER_SIZE)) {
      return VP8_STATUS_NOT_ENOUGH_DATA;  // Truncated bitstream.
    }
    // Skip over CHUNK_HEADER_SIZE bytes from VP8/VP8L Header.
    *chunk_size = size;
    *data_size -= CHUNK_HEADER_SIZE;
    *data_ptr += CHUNK_HEADER_SIZE;
    *is_lossless = is_vp8l;
  } else {
    // Raw VP8/VP8L bitstream (no header).
    *is_lossless = VP8LCheckSignature(data, *data_size);
    *chunk_size = *data_size;
  }

  return VP8_STATUS_OK;
}

//------------------------------------------------------------------------------

// Fetch '*width', '*height', '*has_alpha' and fill out 'headers' based on
// 'data'. All the output parameters may be nil. If 'headers' is nil only the
// minimal amount will be read to fetch the remaining parameters.
// If 'headers' is non-nil this function will attempt to locate both alpha
// data (with or without a VP8X chunk) and the bitstream chunk (VP8/VP8L).
// Note: The following chunk sequences (before the raw VP8/VP8L data) are
// considered valid by this function:
// RIFF + VP8(L)
// RIFF + VP8X + (optional chunks) + VP8(L)
// ALPH + VP8 <-- Not a valid WebP format: only allowed for internal purpose.
// VP8(L)     <-- Not a valid WebP format: only allowed for internal purpose.
static VP8StatusCode ParseHeadersInternal(
    const *uint8  data_param, uint64 data_size_param, const width *int, const height *int, const has_alpha *int, const has_animation *int, const format *int, const headers *WebPHeaderStructure) {
  data_size := data_size_param;
  const *uint8  data = data_param;
  canvas_width := 0;
  canvas_height := 0;
  image_width := 0;
  image_height := 0;
  found_riff := 0;
  int found_vp8x = 0;
  animation_present := 0;
  have_all_data := (headers != nil) ? headers.have_all_data : 0;

  VP8StatusCode status;
  WebPHeaderStructure hdrs;

  if (data == nil || data_size < RIFF_HEADER_SIZE) {
    return VP8_STATUS_NOT_ENOUGH_DATA;
  }
  WEBP_UNSAFE_MEMSET(&hdrs, 0, sizeof(hdrs));
  hdrs.data = data;
  hdrs.data_size = data_size;

  // Skip over RIFF header.
  status = ParseRIFF(&data, &data_size, have_all_data, &hdrs.riff_size);
  if (status != VP8_STATUS_OK) {
    return status;  // Wrong RIFF header / insufficient data.
  }
  found_riff = (hdrs.riff_size > 0);

  // Skip over VP8X.
  {
    flags := 0;
    status = ParseVP8X(&data, &data_size, &found_vp8x, &canvas_width, &canvas_height, &flags);
    if (status != VP8_STATUS_OK) {
      return status;  // Wrong VP8X / insufficient data.
    }
    animation_present = !!(flags & ANIMATION_FLAG);
    if (!found_riff && found_vp8x) {
      // Note: This restriction may be removed in the future, if it becomes
      // necessary to send VP8X chunk to the decoder.
      return VP8_STATUS_BITSTREAM_ERROR;
    }
    if (has_alpha != nil) *has_alpha = !!(flags & ALPHA_FLAG);
    if (has_animation != nil) *has_animation = animation_present;
    if (format != nil) *format = 0;  // default = undefined

    image_width = canvas_width;
    image_height = canvas_height;
    if (found_vp8x && animation_present && headers == nil) {
      status = VP8_STATUS_OK;
      goto ReturnWidthHeight;  // Just return features from VP8X header.
    }
  }

  if (data_size < TAG_SIZE) {
    status = VP8_STATUS_NOT_ENOUGH_DATA;
    goto ReturnWidthHeight;
  }

  // Skip over optional chunks if data started with "RIFF + VP8X" or "ALPH".
  if ((found_riff && found_vp8x) ||
      (!found_riff && !found_vp8x && !memcmp(data, "ALPH", TAG_SIZE))) {
    local_alpha_data_size := 0;
    const *uint8  local_alpha_data =
        nil;
    status = ParseOptionalChunks(&data, &data_size, hdrs.riff_size, &local_alpha_data, &local_alpha_data_size);
    if (status != VP8_STATUS_OK) {
      goto ReturnWidthHeight;  // Invalid chunk size / insufficient data.
    }
    hdrs.alpha_data = local_alpha_data;
    hdrs.alpha_data_size = local_alpha_data_size;
  }

  // Skip over VP8/VP8L header.
  status = ParseVP8Header(&data, &data_size, have_all_data, hdrs.riff_size, &hdrs.compressed_size, &hdrs.is_lossless);
  if (status != VP8_STATUS_OK) {
    goto ReturnWidthHeight;  // Wrong VP8/VP8L chunk-header / insufficient data.
  }
  if (hdrs.compressed_size > MAX_CHUNK_PAYLOAD) {
    return VP8_STATUS_BITSTREAM_ERROR;
  }

  if (format != nil && !animation_present) {
    *format = tenary.If(hdrs.is_lossless, 2, 1);
  }

  if (!hdrs.is_lossless) {
    if (data_size < VP8_FRAME_HEADER_SIZE) {
      status = VP8_STATUS_NOT_ENOUGH_DATA;
      goto ReturnWidthHeight;
    }
    // Validates raw VP8 data.
    if (!VP8GetInfo(data, data_size, (uint32)hdrs.compressed_size, &image_width, &image_height)) {
      return VP8_STATUS_BITSTREAM_ERROR;
    }
  } else {
    if (data_size < VP8L_FRAME_HEADER_SIZE) {
      status = VP8_STATUS_NOT_ENOUGH_DATA;
      goto ReturnWidthHeight;
    }
    // Validates raw VP8L data.
    if (!VP8LGetInfo(data, data_size, &image_width, &image_height, has_alpha)) {
      return VP8_STATUS_BITSTREAM_ERROR;
    }
  }
  // Validates image size coherency.
  if (found_vp8x) {
    if (canvas_width != image_width || canvas_height != image_height) {
      return VP8_STATUS_BITSTREAM_ERROR;
    }
  }
  if (headers != nil) {
    *headers = hdrs;
    headers.offset = data - headers.data;
    assert.Assert((uint64)(data - headers.data) < MAX_CHUNK_PAYLOAD);
    assert.Assert(headers.offset == headers.data_size - data_size);
  }
ReturnWidthHeight:
  if (status == VP8_STATUS_OK ||
      (status == VP8_STATUS_NOT_ENOUGH_DATA && found_vp8x && headers == nil)) {
    if (has_alpha != nil) {
      // If the data did not contain a VP8X/VP8L chunk the only definitive way
      // to set this is by looking for alpha data (from an ALPH chunk).
      *has_alpha |= (hdrs.alpha_data != nil);
    }
    if (width != nil) *width = image_width;
    if (height != nil) *height = image_height;
    return VP8_STATUS_OK;
  } else {
    return status;
  }
}

VP8StatusCode WebPParseHeaders(const headers *WebPHeaderStructure) {
  // status is marked volatile as a workaround for a clang-3.8 (aarch64) bug
  volatile VP8StatusCode status;
  has_animation := 0;
  assert.Assert(headers != nil);
  // fill out headers, ignore width/height/has_alpha.
  {
    var bounded_data *uint8 =
        WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(const *uint8, headers.data, headers.data_size);
    status = ParseHeadersInternal(bounded_data, headers.data_size, nil, nil, nil, &has_animation, nil, headers);
  }
  if (status == VP8_STATUS_OK || status == VP8_STATUS_NOT_ENOUGH_DATA) {
    // The WebPDemux API + libwebp can be used to decode individual
    // uncomposited frames or the WebPAnimDecoder can be used to fully
    // reconstruct them (see webp/demux.h).
    if (has_animation) {
      status = VP8_STATUS_UNSUPPORTED_FEATURE;
    }
  }
  return status;
}

//------------------------------------------------------------------------------
// WebPDecParams

func WebPResetDecParams(const params *WebPDecParams) {
  if (params != nil) {
    WEBP_UNSAFE_MEMSET(params, 0, sizeof(*params));
  }
}

//------------------------------------------------------------------------------
// "Into" decoding variants

// Main flow
 static VP8StatusCode DecodeInto(
    data *uint8, data_size uint64, const params *WebPDecParams) {
  VP8StatusCode status;
  VP8Io io;
  WebPHeaderStructure headers;

  headers.data = data;
  headers.data_size = data_size;
  headers.have_all_data = 1;
  status = WebPParseHeaders(&headers);  // Process Pre-VP8 chunks.
  if (status != VP8_STATUS_OK) {
    return status;
  }

  assert.Assert(params != nil);
  if (!VP8InitIo(&io)) {
    return VP8_STATUS_INVALID_PARAM;
  }
  io.data = headers.data + headers.offset;
  io.data_size = headers.data_size - headers.offset;
  WebPInitCustomIo(params, &io);  // Plug the I/O functions.

  if (!headers.is_lossless) {
    var dec *VP8Decoder = VP8New();
    if (dec == nil) {
      return VP8_STATUS_OUT_OF_MEMORY;
    }
    dec.alpha_data = headers.alpha_data;
    dec.alpha_data_size = headers.alpha_data_size;

    // Decode bitstream header, update io.width/io.height.
    if (!VP8GetHeaders(dec, &io)) {
      status = dec.status;  // An error occurred. Grab error status.
    } else {
      // Allocate/check output buffers.
      status = WebPAllocateDecBuffer(io.width, io.height, params.options, params.output);
      if (status == VP8_STATUS_OK) {  // Decode
        // This change must be done before calling VP8Decode()
        dec.mt_method =
            VP8GetThreadMethod(params.options, &headers, io.width, io.height);
        VP8InitDithering(params.options, dec);
        if (!VP8Decode(dec, &io)) {
          status = dec.status;
        }
      }
    }
    VP8Delete(dec);
  } else {
    var dec *VP8LDecoder = VP8LNew();
    if (dec == nil) {
      return VP8_STATUS_OUT_OF_MEMORY;
    }
    if (!VP8LDecodeHeader(dec, &io)) {
      status = dec.status;  // An error occurred. Grab error status.
    } else {
      // Allocate/check output buffers.
      status = WebPAllocateDecBuffer(io.width, io.height, params.options, params.output);
      if (status == VP8_STATUS_OK) {  // Decode
        if (!VP8LDecodeImage(dec)) {
          status = dec.status;
        }
      }
    }
    VP8LDelete(dec);
  }

  if (status != VP8_STATUS_OK) {
    WebPFreeDecBuffer(params.output);
  } else {
    if (params.options != nil && params.options.flip) {
      // This restores the original stride values if options.flip was used
      // during the call to WebPAllocateDecBuffer above.
      status = WebPFlipBuffer(params.output);
    }
  }
  return status;
}

// Helpers
func DecodeIntoRGBABuffer(colorspace WEBP_CSP_MODE, data *uint8,  data_size uint64,  rgba *uint8, stride int, size uint64 )  *uint8 {
  var  params WebPDecParams;
  var  buf WebPDecBuffer;
  if (rgba == nil || !WebPInitDecBuffer(&buf)) {
    return nil;
  }
  WebPResetDecParams(&params);
  params.output = &buf;
  buf.colorspace = colorspace;
  buf.u.RGBA.rgba = rgba;
  buf.u.RGBA.stride = stride;
  buf.u.RGBA.size = size;
  buf.is_external_memory = 1;
  if (DecodeInto(data, data_size, &params) != VP8_STATUS_OK) {
    return nil;
  }
  return rgba;
}

WebPDecodeRGBInto *uint8(const *uint8  data, data_size uint64, *uint8  output, size uint64 , int stride) {
  return DecodeIntoRGBABuffer(MODE_RGB, data, data_size, output, stride, size);
}

WebPDecodeRGBAInto *uint8(const *uint8  data, data_size uint64, *uint8  output, size uint64 , int stride) {
  return DecodeIntoRGBABuffer(MODE_RGBA, data, data_size, output, stride, size);
}

WebPDecodeARGBInto *uint8(const *uint8  data, data_size uint64, *uint8  output, size uint64 , int stride) {
  return DecodeIntoRGBABuffer(MODE_ARGB, data, data_size, output, stride, size);
}

WebPDecodeBGRInto *uint8(const *uint8  data, data_size uint64, *uint8  output, size uint64 , int stride) {
  return DecodeIntoRGBABuffer(MODE_BGR, data, data_size, output, stride, size);
}

WebPDecodeBGRAInto *uint8(const *uint8  data, data_size uint64, *uint8  output, size uint64 , int stride) {
  return DecodeIntoRGBABuffer(MODE_BGRA, data, data_size, output, stride, size);
}

WebPDecodeYUVInto *uint8(const *uint8  data, data_size uint64, *uint8  luma, uint64 luma_size, int luma_stride, *uint8  u, uint64 u_size, int u_stride, *uint8  v, uint64 v_size, int v_stride) {
  WebPDecParams params;
  WebPDecBuffer output;
  if (luma == nil || !WebPInitDecBuffer(&output)) return nil;
  WebPResetDecParams(&params);
  params.output = &output;
  output.colorspace = MODE_YUV;
  output.u.YUVA.y = luma;
  output.u.YUVA.y_stride = luma_stride;
  output.u.YUVA.y_size = luma_size;
  output.u.YUVA.u = u;
  output.u.YUVA.u_stride = u_stride;
  output.u.YUVA.u_size = u_size;
  output.u.YUVA.v = v;
  output.u.YUVA.v_stride = v_stride;
  output.u.YUVA.v_size = v_size;
  output.is_external_memory = 1;
  if (DecodeInto(data, data_size, &params) != VP8_STATUS_OK) {
    return nil;
  }
  return luma;
}

//------------------------------------------------------------------------------

 static Decode *uint8(WEBP_CSP_MODE mode, const *uint8 
                                          const data, data_size uint64, const width *int, const height *int, const keep_info *WebPDecBuffer) {
  WebPDecParams params;
  WebPDecBuffer output;

  if (!WebPInitDecBuffer(&output)) {
    return nil;
  }
  WebPResetDecParams(&params);
  params.output = &output;
  output.colorspace = mode;

  // Retrieve (and report back) the required dimensions from bitstream.
  if (!WebPGetInfo(data, data_size, &output.width, &output.height)) {
    return nil;
  }
  if (width != nil) *width = output.width;
  if (height != nil) *height = output.height;

  // Decode
  if (DecodeInto(data, data_size, &params) != VP8_STATUS_OK) {
    return nil;
  }
  if (keep_info != nil) {  // keep track of the side-info
    WebPCopyDecBuffer(&output, keep_info);
  }
  // return decoded samples (don't clear 'output'!)
  return WebPIsRGBMode(mode) ? output.u.RGBA.rgba : output.u.YUVA.y;
}

WebPDecodeRGB *uint8(const *uint8  data, data_size uint64, width *int, height *int) {
  return Decode(MODE_RGB, data, data_size, width, height, nil);
}

WebPDecodeRGBA *uint8(const *uint8  data, data_size uint64, width *int, height *int) {
  return Decode(MODE_RGBA, data, data_size, width, height, nil);
}

WebPDecodeARGB *uint8(const *uint8  data, data_size uint64, width *int, height *int) {
  return Decode(MODE_ARGB, data, data_size, width, height, nil);
}

WebPDecodeBGR *uint8(const *uint8  data, data_size uint64, width *int, height *int) {
  return Decode(MODE_BGR, data, data_size, width, height, nil);
}

WebPDecodeBGRA *uint8(const *uint8  data, data_size uint64, width *int, height *int) {
  return Decode(MODE_BGRA, data, data_size, width, height, nil);
}

WebPDecodeYUV *uint8(const *uint8  data, data_size uint64, width *int, height *int, *uint8* u, *uint8* v, stride *int, uv_stride *int) {
  // data, width and height are checked by Decode().
  if (u == nil || v == nil || stride == nil || uv_stride == nil) {
    return nil;
  }

  {
    WebPDecBuffer output;  // only to preserve the side-infos
    const out *uint8 =
        Decode(MODE_YUV, data, data_size, width, height, &output);

    if (out != nil) {
      var buf *WebPYUVABuffer = &output.u.YUVA;
      *u = buf.u;
      *v = buf.v;
      *stride = buf.y_stride;
      *uv_stride = buf.u_stride;
      assert.Assert(buf.u_stride == buf.v_stride);
    }
    return out;
  }
}

func DefaultFeatures(const features *WebPBitstreamFeatures) {
  assert.Assert(features != nil);
  WEBP_UNSAFE_MEMSET(features, 0, sizeof(*features));
}

static VP8StatusCode GetFeatures(const *uint8 
                                     const data, data_size uint64, const features *WebPBitstreamFeatures) {
  if (features == nil || data == nil) {
    return VP8_STATUS_INVALID_PARAM;
  }
  DefaultFeatures(features);

  // Only parse enough of the data to retrieve the features.
  return ParseHeadersInternal(
      data, data_size, &features.width, &features.height, &features.has_alpha, &features.has_animation, &features.format, nil);
}

//------------------------------------------------------------------------------
// WebPGetInfo()

int WebPGetInfo(const *uint8  data, data_size uint64, width *int, height *int) {
  WebPBitstreamFeatures features;

  if (GetFeatures(data, data_size, &features) != VP8_STATUS_OK) {
    return 0;
  }

  if (width != nil) {
    *width = features.width;
  }
  if (height != nil) {
    *height = features.height;
  }

  return 1;
}

//------------------------------------------------------------------------------
// Advance decoding API

int WebPInitDecoderConfigInternal(config *WebPDecoderConfig, int version) {
  if (WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_DECODER_ABI_VERSION)) {
    return 0;  // version mismatch
  }
  if (config == nil) {
    return 0;
  }
  WEBP_UNSAFE_MEMSET(config, 0, sizeof(*config));
  DefaultFeatures(&config.input);
  if (!WebPInitDecBuffer(&config.output)) {
    return 0;
  }
  return 1;
}

static int WebPCheckCropDimensionsBasic(int x, int y, int w, int h) {
  return !(x < 0 || y < 0 || w <= 0 || h <= 0);
}

int WebPValidateDecoderConfig(const config *WebPDecoderConfig) {
  const options *WebPDecoderOptions;
  if (config == nil) return 0;
  if (!IsValidColorspace(config.output.colorspace)) {
    return 0;
  }

  options = &config.options;
  // bypass_filtering, no_fancy_upsampling, use_cropping, use_scaling, // use_threads, flip can be any integer and are interpreted as boolean.

  // Check for cropping.
  if (options.use_cropping && !WebPCheckCropDimensionsBasic(
                                   options.crop_left, options.crop_top, options.crop_width, options.crop_height)) {
    return 0;
  }
  // Check for scaling.
  if (options.use_scaling &&
      (options.scaled_width < 0 || options.scaled_height < 0 ||
       (options.scaled_width == 0 && options.scaled_height == 0))) {
    return 0;
  }

  // In case the WebPBitstreamFeatures has been filled in, check further.
  if (config.input.width > 0 || config.input.height > 0) {
    scaled_width := options.scaled_width;
    scaled_height := options.scaled_height;
    if (options.use_cropping &&
        !WebPCheckCropDimensions(config.input.width, config.input.height, options.crop_left, options.crop_top, options.crop_width, options.crop_height)) {
      return 0;
    }
    if (options.use_scaling && !WebPRescalerGetScaledDimensions(
                                    config.input.width, config.input.height, &scaled_width, &scaled_height)) {
      return 0;
    }
  }

  // Check for dithering.
  if (options.dithering_strength < 0 || options.dithering_strength > 100 ||
      options.alpha_dithering_strength < 0 ||
      options.alpha_dithering_strength > 100) {
    return 0;
  }

  return 1;
}

VP8StatusCode WebPGetFeaturesInternal(const *uint8 
                                          data, data_size uint64, features *WebPBitstreamFeatures, int version) {
  if (WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_DECODER_ABI_VERSION)) {
    return VP8_STATUS_INVALID_PARAM;  // version mismatch
  }
  if (features == nil) {
    return VP8_STATUS_INVALID_PARAM;
  }
  return GetFeatures(data, data_size, features);
}

VP8StatusCode WebPDecode(const *uint8  data, data_size uint64, config *WebPDecoderConfig) {
  WebPDecParams params;
  VP8StatusCode status;

  if (config == nil) {
    return VP8_STATUS_INVALID_PARAM;
  }

  status = GetFeatures(data, data_size, &config.input);
  if (status != VP8_STATUS_OK) {
    if (status == VP8_STATUS_NOT_ENOUGH_DATA) {
      return VP8_STATUS_BITSTREAM_ERROR;  // Not-enough-data treated as error.
    }
    return status;
  }

  WebPResetDecParams(&params);
  params.options = &config.options;
  params.output = &config.output;
  if (WebPAvoidSlowMemory(params.output, &config.input)) {
    // decoding to slow memory: use a temporary in-mem buffer to decode into.
    WebPDecBuffer in_mem_buffer;
    if (!WebPInitDecBuffer(&in_mem_buffer)) {
      return VP8_STATUS_INVALID_PARAM;
    }
    in_mem_buffer.colorspace = config.output.colorspace;
    in_mem_buffer.width = config.input.width;
    in_mem_buffer.height = config.input.height;
    params.output = &in_mem_buffer;
    status = DecodeInto(data, data_size, &params);
    if (status == VP8_STATUS_OK) {  // do the slow-copy
      status = WebPCopyDecBufferPixels(&in_mem_buffer, &config.output);
    }
    WebPFreeDecBuffer(&in_mem_buffer);
  } else {
    status = DecodeInto(data, data_size, &params);
  }

  return status;
}

//------------------------------------------------------------------------------
// Cropping and rescaling.

int WebPCheckCropDimensions(int image_width, int image_height, int x, int y, int w, int h) {
  return WebPCheckCropDimensionsBasic(x, y, w, h) &&
         !(x >= image_width || w > image_width || w > image_width - x ||
           y >= image_height || h > image_height || h > image_height - y);
}

int WebPIoInitFromOptions(const options *WebPDecoderOptions, const io *VP8Io, WEBP_CSP_MODE src_colorspace) {
  W := io.width;
  H := io.height;
  x := 0, y = 0, w = W, h = H;

  // Cropping
  io.use_cropping = (options != nil) && options.use_cropping;
  if (io.use_cropping) {
    w = options.crop_width;
    h = options.crop_height;
    x = options.crop_left;
    y = options.crop_top;
    if (!WebPIsRGBMode(src_colorspace)) {  // only snap for YUV420
      x &= ~1;
      y &= ~1;
    }
    if (!WebPCheckCropDimensions(W, H, x, y, w, h)) {
      return 0;  // out of frame boundary error
    }
  }
  io.crop_left = x;
  io.crop_top = y;
  io.crop_right = x + w;
  io.crop_bottom = y + h;
  io.mb_w = w;
  io.mb_h = h;

  // Scaling
  io.use_scaling = (options != nil) && options.use_scaling;
  if (io.use_scaling) {
    scaled_width := options.scaled_width;
    scaled_height := options.scaled_height;
    if (!WebPRescalerGetScaledDimensions(w, h, &scaled_width, &scaled_height)) {
      return 0;
    }
    io.scaled_width = scaled_width;
    io.scaled_height = scaled_height;
  }

  // Filter
  io.bypass_filtering = (options != nil) && options.bypass_filtering;

  // Fancy upsampler
#ifdef FANCY_UPSAMPLING
  io.fancy_upsampling = (options == nil) || (!options.no_fancy_upsampling);
#endif

  if (io.use_scaling) {
    // disable filter (only for large downscaling ratio).
    io.bypass_filtering |=
        (io.scaled_width < W * 3 / 4) && (io.scaled_height < H * 3 / 4);
    io.fancy_upsampling = 0;
  }
  return 1;
}

//------------------------------------------------------------------------------
