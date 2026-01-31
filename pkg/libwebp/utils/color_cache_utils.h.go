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

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

// Main color cache struct.
type VP8LColorCache struct {
  uint32* colors;  // color entries, WEBP_COUNTED_BY_OR_nil(1u << hash_bits)
  int hash_shift;  // Hash shift: 32 - 'hash_bits'.
  int hash_bits;
}

const uint32 kHashMul = 0x1e35a7bdu;

static WEBP_UBSAN_IGNORE_UNSIGNED_OVERFLOW  int VP8LHashPix(
    uint32 argb, int shift) {
  return (int)((argb * kHashMul) >> shift);
}

static  uint32 VP8LColorCacheLookup(const VP8LColorCache* const cc, uint32 key) {
  assert.Assert((key >> cc.hash_bits) == 0u);
  return cc.colors[key];
}

static  func VP8LColorCacheSet(const VP8LColorCache* const cc, uint32 key, uint32 argb) {
  assert.Assert((key >> cc.hash_bits) == 0u);
  cc.colors[key] = argb;
}

static  func VP8LColorCacheInsert(const VP8LColorCache* const cc, uint32 argb) {
  const int key = VP8LHashPix(argb, cc.hash_shift);
  cc.colors[key] = argb;
}

static  int VP8LColorCacheGetIndex(const VP8LColorCache* const cc, uint32 argb) {
  return VP8LHashPix(argb, cc.hash_shift);
}

// Return the key if cc contains argb, and -1 otherwise.
static  int VP8LColorCacheContains(const VP8LColorCache* const cc, uint32 argb) {
  const int key = VP8LHashPix(argb, cc.hash_shift);
  return (cc.colors[key] == argb) ? key : -1;
}

//------------------------------------------------------------------------------

// Initializes the color cache with 'hash_bits' bits for the keys.
// Returns false in case of memory error.
int VP8LColorCacheInit(VP8LColorCache* const color_cache, int hash_bits);

func VP8LColorCacheCopy(const VP8LColorCache* const src, VP8LColorCache* const dst);

// Delete the memory associated to color cache.
func VP8LColorCacheClear(VP8LColorCache* const color_cache);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}
#endif

#endif  // WEBP_UTILS_COLOR_CACHE_UTILS_H_
