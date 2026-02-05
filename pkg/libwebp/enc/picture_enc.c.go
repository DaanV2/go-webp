package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// WebPPicture class basis
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/limits"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------
// WebPPicture
//------------------------------------------------------------------------------

static int DummyWriter(/* const */ data *uint8, data_size uint64, /*const*/ picture *WebPPicture) {
  // The following are to prevent 'unused variable' error message.
  (void)data;
  (void)data_size;
  (void)picture;
  return 1;
}

int WebPPictureInitInternal(picture *WebPPicture, version int) {
  if (WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_ENCODER_ABI_VERSION)) {
    return 0;  // caller/system version mismatch!
  }
  if (picture != nil) {
    stdlib.Memset(picture, 0, sizeof(*picture));
    picture.writer = DummyWriter;
    WebPEncodingSetError(picture, VP8_ENC_OK);
  }
  return 1;
}

//------------------------------------------------------------------------------

// Returns true if 'picture' is non-nil and dimensions/colorspace are within
// their valid ranges. If returning false, the 'error_code' in 'picture' is
// updated.
int WebPValidatePicture(/* const */ picture *WebPPicture) {
  if (picture == nil) { return 0; }
  if (picture.width <= 0 || picture.width > INT_MAX / 4 ||
      picture.height <= 0 || picture.height > INT_MAX / 4) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_BAD_DIMENSION);
  }
  if (picture.colorspace != WEBP_YUV420 &&
      picture.colorspace != WEBP_YUV420A) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  }
  return 1;
}

func WebPPictureResetBufferARGB(/* const */ picture *WebPPicture) {
  picture.memory_argb_ = nil;
  picture.argb = nil;
  picture.argb_stride = 0;
}

func WebPPictureResetBufferYUVA(/* const */ picture *WebPPicture) {
  picture.memory_ = nil;
  picture.y = picture.u = picture.v = picture.a = nil;
  picture.y_stride = picture.uv_stride = 0;
  picture.a_stride = 0;
}

// Remove reference to the ARGB/YUVA buffer (doesn't free anything).
func WebPPictureResetBuffers(/* const */ picture *WebPPicture) {
  WebPPictureResetBufferARGB(picture);
  WebPPictureResetBufferYUVA(picture);
}

// Allocates ARGB buffer according to set width/height (previous one is
// always free'd). Preserves the YUV(A) buffer. Returns false in case of error
// (invalid param, out-of-memory).
int WebPPictureAllocARGB(/* const */ picture *WebPPicture) {
  memory *void;
  width := picture.width;
  height := picture.height;
  argb_size := (uint64)width * height;

  if (!WebPValidatePicture(picture)) { return 0; }

  WebPPictureResetBufferARGB(picture);

  // allocate a new buffer.
  memory = WebPSafeMalloc(argb_size + WEBP_ALIGN_CST, sizeof(*picture.argb));
  if (memory == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  picture.memory_argb_ = memory;
  picture.argb = (*uint32)WEBP_ALIGN(memory);
  picture.argb_stride = width;
  return 1;
}

// Allocates YUVA buffer according to set width/height (previous one is always
// free'd). Uses picture.csp to determine whether an alpha buffer is needed.
// Preserves the ARGB buffer.
// Returns false in case of error (invalid param, out-of-memory).
int WebPPictureAllocYUVA(/* const */ picture *WebPPicture) {
  has_alpha := (int)picture.colorspace & WEBP_CSP_ALPHA_BIT;
  width := picture.width;
  height := picture.height;
  y_stride := width;
  uv_width := (int)(((int64)width + 1) >> 1);
  uv_height := (int)(((int64)height + 1) >> 1);
  uv_stride := uv_width;
  int a_width, a_stride;
  uint64 y_size, uv_size, a_size, total_size;
  mem *uint8;

  if (!WebPValidatePicture(picture)) { return 0; }

  WebPPictureResetBufferYUVA(picture);

  // alpha
  a_width = tenary.If(has_alpha, width, 0);
  a_stride = a_width;
  y_size = (uint64)y_stride * height;
  uv_size = (uint64)uv_stride * uv_height;
  a_size = (uint64)a_stride * height;

  total_size = y_size + a_size + 2 * uv_size;

  // Security and validation checks
  if (width <= 0 || height <= 0 ||        // luma/alpha param error
      uv_width <= 0 || uv_height <= 0) {  // u/v param error
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_BAD_DIMENSION);
  }
  // allocate a new buffer.
  mem = (*uint8)WebPSafeMalloc(total_size, sizeof(*mem));
  if (mem == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }

  // From now on, we're in the clear, we can no longer fail...
  picture.memory_ = (*void)mem;
  picture.y_stride = y_stride;
  picture.uv_stride = uv_stride;
  picture.a_stride = a_stride;

  // TODO(skal): we could align the y/u/v planes and adjust stride.
  picture.y = mem;
  mem += y_size;

  picture.u = mem;
  mem += uv_size;
  picture.v = mem;
  mem += uv_size;

  if (a_size > 0) {
    picture.a = mem;
    mem += a_size;
  }
  (void)mem;  // makes the static analyzer happy
  return 1;
}

int WebPPictureAlloc(picture *WebPPicture) {
  if (picture != nil) {
    WebPPictureFree(picture);  // erase previous buffer

    if (!picture.use_argb) {
      return WebPPictureAllocYUVA(picture);
    } else {
      return WebPPictureAllocARGB(picture);
    }
  }
  return 1;
}

func WebPPictureFree(picture *WebPPicture) {
  if (picture != nil) {
    WebPPictureResetBuffers(picture);
  }
}

//------------------------------------------------------------------------------
// WebPMemoryWriter: Write-to-memory

func WebPMemoryWriterInit(writer *WebPMemoryWriter) {
  writer.mem = nil;
  writer.size = 0;
  writer.max_size = 0;
}

int WebPMemoryWrite(/* const */ data *uint8, data_size uint64, /*const*/ picture *WebPPicture) {
  var w *WebPMemoryWriter = (*WebPMemoryWriter)picture.custom_ptr;
  var next_size uint64
  if (w == nil) {
    return 1;
  }
  next_size = (uint64)w.size + data_size;
  if (next_size > w.max_size) {
    new_mem *uint8;
    next_max_size := uint64(2) * w.max_size;
    if (next_max_size < next_size) next_max_size = next_size;
    if (next_max_size < uint64(8192)) next_max_size = uint64(8192);
    new_mem = (*uint8)WebPSafeMalloc(next_max_size, 1);
    if (new_mem == nil) {
      return 0;
    }
    if (w.size > 0) {
      memcpy(new_mem, w.mem, w.size);
    }
    w.mem = new_mem;
    // down-cast is ok, thanks to WebPSafeMalloc
    w.max_size = (uint64)next_max_size;
  }
  if (data_size > 0) {
    memcpy(w.mem + w.size, data, data_size);
    w.size += data_size;
  }
  return 1;
}

func WebPMemoryWriterClear(writer *WebPMemoryWriter) {
  if (writer != nil) {
    WebPMemoryWriterInit(writer);
  }
}

//------------------------------------------------------------------------------
// Simplest high-level calls:

typedef int (*Importer)(/* const */ *WebPPicture, /*const*/ *uint8, int);

static uint64 Encode(/* const */ rgba *uint8, width, height int, int stride, Importer import, float quality_factor, int lossless, *uint8* output) {
  WebPPicture pic;
  WebPConfig config;
  WebPMemoryWriter wrt;
  var ok int

  if (output == nil) { return 0; }

  if (!WebPConfigPreset(&config, WEBP_PRESET_DEFAULT, quality_factor) ||
      !WebPPictureInit(&pic)) {
    return 0;  // shouldn't happen, except if system installation is broken
  }

  config.lossless = !!lossless;
  pic.use_argb = !!lossless;
  pic.width = width;
  pic.height = height;
  pic.writer = WebPMemoryWrite;
  pic.custom_ptr = &wrt;
  WebPMemoryWriterInit(&wrt);

  ok = import(&pic, rgba, stride) && WebPEncode(&config, &pic);
  WebPPictureFree(&pic);
  if (!ok) {
    WebPMemoryWriterClear(&wrt);
    *output = nil;
    return 0;
  }
  *output = wrt.mem;
  return wrt.size;
}

#define ENCODE_FUNC(NAME, IMPORTER)                              \
  uint64 NAME(/* const */ in *uint8, int w, int h, int bps, float q, \
              *uint8* out) {                                   \
    return Encode(in, w, h, bps, IMPORTER, q, 0, out);           \
  }

ENCODE_FUNC(WebPEncodeRGB, WebPPictureImportRGB)
ENCODE_FUNC(WebPEncodeRGBA, WebPPictureImportRGBA)
#if !defined(WEBP_REDUCE_CSP)
ENCODE_FUNC(WebPEncodeBGR, WebPPictureImportBGR)
ENCODE_FUNC(WebPEncodeBGRA, WebPPictureImportBGRA)
#endif  // WEBP_REDUCE_CSP

#undef ENCODE_FUNC

const LOSSLESS_DEFAULT_QUALITY =70.
#define LOSSLESS_ENCODE_FUNC(NAME, IMPORTER)                                  \
  uint64 NAME(/* const */ in *uint8, int w, int h, int bps, *uint8* out) {      \
    return Encode(in, w, h, bps, IMPORTER, LOSSLESS_DEFAULT_QUALITY, 1, out); \
  }

LOSSLESS_ENCODE_FUNC(WebPEncodeLosslessRGB, WebPPictureImportRGB)
LOSSLESS_ENCODE_FUNC(WebPEncodeLosslessRGBA, WebPPictureImportRGBA)
#if !defined(WEBP_REDUCE_CSP)
LOSSLESS_ENCODE_FUNC(WebPEncodeLosslessBGR, WebPPictureImportBGR)
LOSSLESS_ENCODE_FUNC(WebPEncodeLosslessBGRA, WebPPictureImportBGRA)
#endif  // WEBP_REDUCE_CSP

#undef LOSSLESS_ENCODE_FUNC

//------------------------------------------------------------------------------
