package mux

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Internal header for mux library.
//
// Author: Urvang (urvang@google.com)


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// Defines and constants.

const MUX_MAJ_VERSION =1
const MUX_MIN_VERSION =6
const MUX_REV_VERSION =0

// Chunk object.
typedef struct WebPChunk WebPChunk;
type WebPChunk struct {
  var tag uint32
  var owner int  // True if memory is owned *data internally.
              // VP8X, ANIM, and other internally created chunks
              // like ANMF are always owned.
  WebPData data;
  next *WebPChunk;
}

// MuxImage object. Store a full WebP image (including ANMF chunk, ALPH
// chunk and VP8/VP8L chunk),
typedef struct WebPMuxImage WebPMuxImage;
type WebPMuxImage struct {
  header *WebPChunk;   // Corresponds to WEBP_CHUNK_ANMF.
  alpha *WebPChunk;    // Corresponds to WEBP_CHUNK_ALPHA.
  img *WebPChunk;      // Corresponds to WEBP_CHUNK_IMAGE.
  unknown *WebPChunk;  // Corresponds to WEBP_CHUNK_UNKNOWN.
  var width int
  var height int
  var has_alpha int   // Through ALPH chunk or as part of VP8L.
  var is_partial int  // True if only some of the chunks are filled.
  next *WebPMuxImage;
}

// Main mux object. Stores data chunks.
type WebPMux struct {
  images *WebPMuxImage;
  iccp *WebPChunk;
  exif *WebPChunk;
  xmp *WebPChunk;
  anim *WebPChunk;
  vp *WebPChunk8x;

  unknown *WebPChunk;
  canvas_width int;
  canvas_height int;
}

// CHUNK_INDEX enum: used for indexing within 'kChunks' (defined below) only.
// Note: the reason for having two enums ('WebPChunkId' and 'CHUNK_INDEX') is to
// allow two different chunks to have the same id (e.g. WebPChunkId
// 'WEBP_CHUNK_IMAGE' can correspond to CHUNK_INDEX 'IDX_VP8' or 'IDX_VP8L').
type CHUNK_INDEX int

const (
	IDX_VP8X CHUNK_INDEX = iota
	IDX_ICCP
	IDX_ANIM
	IDX_ANMF
	IDX_ALPHA
	IDX_VP8
	IDX_VP8L
	IDX_EXIF
	IDX_XMP
	IDX_UNKNOWN
	IDX_NIL 
	IDX_LAST_CHUNK
)

const NIL_TAG =uint(0x00000000)  // To signal func chunk.

type ChunkInfo struct {
  var tag uint32
  WebPChunkId id;
  size uint32;
} ;

extern const ChunkInfo kChunks[IDX_LAST_CHUNK];

//------------------------------------------------------------------------------
// Chunk object management.

// Initialize.
func ChunkInit(/* const */ chunk *WebPChunk);

// Get chunk index from chunk tag. Returns IDX_UNKNOWN if not found.
CHUNK_INDEX ChunkGetIndexFromTag(uint32 tag);

// Get chunk id from chunk tag. Returns WEBP_CHUNK_UNKNOWN if not found.
WebPChunkId ChunkGetIdFromTag(uint32 tag);

// Convert a fourcc string to a tag.
uint32 ChunkGetTagFromFourCC(/* const */ byte fourcc[4]);

// Get chunk index from fourcc. Returns IDX_UNKNOWN if given fourcc is unknown.
CHUNK_INDEX ChunkGetIndexFromFourCC(/* const */ byte fourcc[4]);

// Search for nth chunk with given 'tag' in the chunk list.
// nth = 0 means "last of the list".
ChunkSearchList *WebPChunk(first *WebPChunk, uint32 nth, uint32 tag);

// Fill the chunk with the given data.
WebPMuxError ChunkAssignData(chunk *WebPChunk, /*const*/ data *WebPData, int copy_data, uint32 tag);

// Sets 'chunk' as the only element in 'chunk_list' if it is empty.
// On success ownership is transferred from 'chunk' to the 'chunk_list'.
WebPMuxError ChunkSetHead(/* const */ chunk *WebPChunk, *WebPChunk* const chunk_list);
// Sets 'chunk' at last position in the 'chunk_list'.
// On success ownership is transferred from 'chunk' to the 'chunk_list'.
// also points towards *chunk_list the last valid element of the initial
// *chunk_list.
WebPMuxError ChunkAppend(/* const */ chunk *WebPChunk, *WebPChunk** const chunk_list);

// Releases chunk and returns chunk.next.
ChunkRelease *WebPChunk(/* const */ chunk *WebPChunk);

// Deletes given chunk & returns chunk.next.
ChunkDelete *WebPChunk(/* const */ chunk *WebPChunk);

// Deletes all chunks in the given chunk list.
func ChunkListDelete(*WebPChunk* const chunk_list);

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

// Total size of a list of chunks.
uint64 ChunkListDiskSize(/* const */ chunk_list *WebPChunk);

// Write out the given list of chunks into 'dst'.
ChunkListEmit *uint8(/* const */ chunk_list *WebPChunk, dst *uint8);

//------------------------------------------------------------------------------
// MuxImage object management.

// Initialize.
func MuxImageInit(/* const */ wpi *WebPMuxImage);

// Releases image 'wpi' and returns wpi.next.
MuxImageRelease *WebPMuxImage(/* const */ wpi *WebPMuxImage);

// Delete image 'wpi' and return the next image in the list or nil.
// 'wpi' can be nil.
MuxImageDelete *WebPMuxImage(/* const */ wpi *WebPMuxImage);

// Count number of images matching the given tag id in the 'wpi_list'.
// If id == WEBP_CHUNK_NIL, all images will be matched.
int MuxImageCount(/* const */ wpi_list *WebPMuxImage, WebPChunkId id);

// Update width/height/has_alpha info from chunks within wpi.
// Also remove ALPH chunk if not needed.
int MuxImageFinalize(/* const */ wpi *WebPMuxImage);

// Check if given ID corresponds to an image related chunk.
func IsWPI(WebPChunkId id) int {
  switch (id) {
    case WEBP_CHUNK_ANMF:
    case WEBP_CHUNK_ALPHA:
    case WEBP_CHUNK_IMAGE:
      return 1;
    default:
      return 0;
  }
}

// Pushes 'wpi' at the end of 'wpi_list'.
WebPMuxError MuxImagePush(/* const */ wpi *WebPMuxImage, *WebPMuxImage* wpi_list);

// Delete nth image in the image list.
WebPMuxError MuxImageDeleteNth(*WebPMuxImage* wpi_list, uint32 nth);

// Get nth image in the image list.
WebPMuxError MuxImageGetNth(/* const */ *WebPMuxImage* wpi_list, uint32 nth, *WebPMuxImage* wpi);

// Total size of the given image.
uint64 MuxImageDiskSize(/* const */ wpi *WebPMuxImage);

// Write out the given image into 'dst'.
MuxImageEmit *uint8(/* const */ wpi *WebPMuxImage, dst *uint8);

//------------------------------------------------------------------------------
// Helper methods for mux.

// Checks if the given image list contains at least one image with alpha.
int MuxHasAlpha(/* const */ images *WebPMuxImage);

// Write out RIFF header into 'data', given total data size 'size'.
MuxEmitRiffHeader *uint8(/* const */ data *uint8, size uint64 );

// Returns the list where chunk with given ID is to be inserted in mux.
*WebPChunk* MuxGetChunkListFromId(/* const */ mux *WebPMux, WebPChunkId id);

// Validates the given mux object.
WebPMuxError MuxValidate(/* const */ mux *WebPMux);

//------------------------------------------------------------------------------



#endif  // WEBP_MUX_MUXI_H_
