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
// Color Cache for WebP Lossless
//
// Authors: Jyrki Alakuijala (jyrki@google.com)
//          Urvang Joshi (urvang@google.com)


import "github.com/daanv2/go-webp/pkg/assert"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"






const kHashMul  = uint32(0x1e35a7bd);

static WEBP_UBSAN_IGNORE_UNSIGNED_OVERFLOW  int VP8LHashPix(
    uint32 argb, int shift) {
  return (int)((argb * kHashMul) >> shift);
}

static  uint32 VP8LColorCacheLookup(/* const */ cc *VP8LColorCache, uint32 key) {
  assert.Assert((key >> cc.hash_bits) == uint(0));
  return cc.colors[key];
}

static  func VP8LColorCacheSet(/* const */ cc *VP8LColorCache, uint32 key, uint32 argb) {
  assert.Assert((key >> cc.hash_bits) == uint(0));
  cc.colors[key] = argb;
}

static  func VP8LColorCacheInsert(/* const */ cc *VP8LColorCache, uint32 argb) {
  key := VP8LHashPix(argb, cc.hash_shift);
  cc.colors[key] = argb;
}

static  int VP8LColorCacheGetIndex(/* const */ cc *VP8LColorCache, uint32 argb) {
  return VP8LHashPix(argb, cc.hash_shift);
}

// Return the key if cc contains argb, and -1 otherwise.
static  int VP8LColorCacheContains(/* const */ cc *VP8LColorCache, uint32 argb) {
  key := VP8LHashPix(argb, cc.hash_shift);
  return (cc.colors[key] == argb) ? key : -1;
}

//------------------------------------------------------------------------------

// Initializes the color cache with 'hash_bits' bits for the keys.
// Returns false in case of memory error.
int VP8LColorCacheInit(/* const */ color_cache *VP8LColorCache, int hash_bits);

func VP8LColorCacheCopy(/* const */ src *VP8LColorCache, /*const*/ dst *VP8LColorCache);

// Delete the memory associated to color cache.
func VP8LColorCacheClear(/* const */ color_cache *VP8LColorCache);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}
#endif

#endif  // WEBP_UTILS_COLOR_CACHE_UTILS_H_
