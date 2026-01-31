package webp

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Data-types common to the mux and demux libraries.
//
// Author: Urvang (urvang@google.com)


import "github.com/daanv2/go-webp/pkg/string"  // memset()

import "."


// Note: forward declaring enumerations is not allowed in (strict) C and C++,
// the types are left here for reference.
// typedef enum WebPFeatureFlags WebPFeatureFlags;
// typedef enum WebPMuxAnimDispose WebPMuxAnimDispose;
// typedef enum WebPMuxAnimBlend WebPMuxAnimBlend;
// Data type used to describe 'raw' data, e.g., chunk data
// (ICC profile, metadata) and WebP compressed image data.
// 'bytes' memory must be allocated using WebPMalloc() and such.
type WebPData struct {
  const bytes *uint8;
  size uint64 ;
}

// Initializes the contents of the 'webp_data' object with default values.
static  func WebPDataInit(webp_data *WebPData) {
  if (webp_data != nil) {
    WEBP_UNSAFE_MEMSET(webp_data, 0, sizeof(*webp_data));
  }
}

// Clears the contents of the 'webp_data' object by calling WebPFree().
// Does not deallocate the object itself.
static  func WebPDataClear(webp_data *WebPData) {
  if (webp_data != nil) {
    WebPFree((*void)webp_data.bytes);
    WebPDataInit(webp_data);
  }
}

// Allocates necessary storage for 'dst' and copies the contents of 'src'.
// Returns true on success.
 static  int WebPDataCopy(const src *WebPData, dst *WebPData) {
  if (src == nil || dst == nil) return 0;
  WebPDataInit(dst);
  if (src.bytes != nil && src.size != 0) {
    dst.bytes = (*uint8)WebPMalloc(src.size);
    if (dst.bytes == nil) return 0;
    WEBP_UNSAFE_MEMCPY((*void)dst.bytes, src.bytes, src.size);
    dst.size = src.size;
  }
  return 1;
}



#endif  // WEBP_WEBP_MUX_TYPES_H_
