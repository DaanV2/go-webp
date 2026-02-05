// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

import "github.com/daanv2/go-webp/pkg/assert"

// Returns size of the chunk including chunk header and padding byte (if any).
func SizeWithPadding(uint64 chunk_size) uint64 {
  assert.Assert(chunk_size <= MAX_CHUNK_PAYLOAD);
  return CHUNK_HEADER_SIZE + ((chunk_size + 1) & ~uint(1));
}

// Size of a chunk including header and padding.
func ChunkDiskSize(/* const */ chunk *WebPChunk) uint64 {
  data_size := chunk.data.size;
  return SizeWithPadding(data_size);
}


// Check if given ID corresponds to an image related chunk.
func IsWPI(id WebPChunkId) int {
  switch (id) {
    case WEBP_CHUNK_ANMF:
    case WEBP_CHUNK_ALPHA:
    case WEBP_CHUNK_IMAGE:
      return 1;
    default:
      return 0;
  }
}