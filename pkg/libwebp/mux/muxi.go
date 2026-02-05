// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
)

// Returns size of the chunk including chunk header and padding byte (if any).
func SizeWithPadding(chunk_size uint64) uint64 {
	assert.Assert(chunk_size <= uint64(constants.MAX_CHUNK_PAYLOAD))

	return uint64(constants.CHUNK_HEADER_SIZE) + ((chunk_size + 1) & ^uint64(1))
}

// Size of a chunk including header and padding.
func ChunkDiskSize( /* const */ chunk *WebPChunk) uint64 {
	data_size := chunk.data.size
	return SizeWithPadding(data_size)
}

// Check if given ID corresponds to an image related chunk.
func IsWPI(id webp.WebPChunkId) int {
	switch id {
	case webp.WEBP_CHUNK_ANMF, webp.WEBP_CHUNK_ALPHA, webp.WEBP_CHUNK_IMAGE:
		return 1
	default:
		return 0
	}
}
