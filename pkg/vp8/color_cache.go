// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

const kHashMul = uint32(0x1e35a7bd)

// Initializes the color cache with 'hash_bits' bits for the keys.
// Returns false in case of memory error.
func VP8LColorCacheInit( /* const */ color_cache *VP8LColorCache, hash_bits int) int {
	hash_size := 1 << hash_bits
	assert.Assert(color_cache != nil)
	assert.Assert(hash_bits > 0)

	//   colors *uint32 = (*uint32)WebPSafeCalloc((uint64)hash_size, sizeof(*color_cache.colors));
	colors := make([]uint32, hash_size)
	if colors == nil {
		color_cache.colors = nil
		// WEBP_SELF_ASSIGN(color_cache.hash_bits)
		return 0
	}
	color_cache.hash_shift = 32 - hash_bits
	color_cache.hash_bits = hash_bits
	color_cache.colors = colors
	return 1
}

// Delete the memory associated to color cache.
func VP8LColorCacheClear( /* const */ color_cache *VP8LColorCache) {
	if color_cache != nil {
		color_cache.colors = nil
		// WEBP_SELF_ASSIGN(color_cache.hash_bits)
	}
}

func VP8LHashPix(argb uint32, shift int) int {
	return int((argb * kHashMul) >> shift)
}

func VP8LColorCacheLookup( /* const */ cc *VP8LColorCache, key uint32) uint32 {
	assert.Assert((key >> cc.hash_bits) == uint(0))
	return cc.colors[key]
}

func VP8LColorCacheSet( /* const */ cc *VP8LColorCache, key uint32, argb uint32) {
	assert.Assert((key >> cc.hash_bits) == uint(0))
	cc.colors[key] = argb
}

func VP8LColorCacheInsert( /* const */ cc *VP8LColorCache, argb uint32) {
	key := VP8LHashPix(argb, cc.hash_shift)
	cc.colors[key] = argb
}

func VP8LColorCacheGetIndex( /* const */ cc *VP8LColorCache, argb uint32) int {
	return VP8LHashPix(argb, cc.hash_shift)
}

// Return the key if cc contains argb, and -1 otherwise.
func VP8LColorCacheContains( /* const */ cc *VP8LColorCache, argb uint32) int {
	key := VP8LHashPix(argb, cc.hash_shift)
	return tenary.If(cc.colors[key] == argb, key, -1)
}

func VP8LColorCacheCopy( /* const */ src *VP8LColorCache /*const*/, dst *VP8LColorCache) {
	assert.Assert(src != nil)
	assert.Assert(dst != nil)
	assert.Assert(src.hash_bits == dst.hash_bits)
	// C: stdlib.MemCpy(dst.colors, src.colors, (uint64(1)<<dst.hash_bits)*sizeof(*dst.colors))
}
