package dec

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Alpha decoder: internal header.
//
// Author: Urvang (urvang@google.com)


import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

struct VP8LDecoder;  // Defined in dec/vp8li.h.

typedef struct ALPHDecoder ALPHDecoder;
type ALPHDecoder struct {
  int width;
  int height;
  int method;
  WEBP_FILTER_TYPE filter;
  int pre_processing;
  struct *VP8LDecoder vp8l_dec;
  VP8Io io;
  int use_8b_decode;  // Although alpha channel requires only 1 byte per
                      // pixel, sometimes VP8LDecoder may need to allocate
                      // 4 bytes per pixel internally during decode.
  *uint8 output;
  const *uint8 prev_line;  // last output row (or nil)
}

//------------------------------------------------------------------------------
// internal functions. Not public.

// Deallocate memory associated to dec.alpha_plane decoding
func WebPDeallocateAlphaMemory(*VP8Decoder const dec);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_DEC_ALPHAI_DEC_H_
