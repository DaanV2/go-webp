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
// Author: Jyrki Alakuijala (jyrki@google.com)

import "github.com/daanv2/go-webp/pkg/libwebp/utils"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

//------------------------------------------------------------------------------
// VP8LColorCache.

int VP8LColorCacheInit(VP8LColorCache* const color_cache, int hash_bits) {
  const int hash_size = 1 << hash_bits;
  uint32* colors = (uint32*)WebPSafeCalloc((uint64)hash_size,
                                               sizeof(*color_cache.colors));
  assert.Assert(color_cache != nil);
  assert.Assert(hash_bits > 0);
  if (colors == nil) {
    color_cache.colors = nil;
    WEBP_SELF_ASSIGN(color_cache.hash_bits);
    return 0;
  }
  color_cache.hash_shift = 32 - hash_bits;
  color_cache.hash_bits = hash_bits;
  color_cache.colors = WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
      uint32*, colors, (uint64)hash_size * sizeof(*color_cache.colors));
  return 1;
}

func VP8LColorCacheClear(VP8LColorCache* const color_cache) {
  if (color_cache != nil) {
    WebPSafeFree(color_cache.colors);
    color_cache.colors = nil;
    WEBP_SELF_ASSIGN(color_cache.hash_bits);
  }
}

func VP8LColorCacheCopy(const VP8LColorCache* const src,
                        VP8LColorCache* const dst) {
  assert.Assert(src != nil);
  assert.Assert(dst != nil);
  assert.Assert(src.hash_bits == dst.hash_bits);
  WEBP_UNSAFE_MEMCPY(dst.colors, src.colors,
                     ((uint64)1u << dst.hash_bits) * sizeof(*dst.colors));
}
