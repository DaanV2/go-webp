// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/mux"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

const UNDEFINED_CHUNK_SIZE =((uint32)(-1))

const ChunkInfo kChunks[] = {
    {MKFOURCC('V', 'P', '8', 'X'), WEBP_CHUNK_VP8X, VP8X_CHUNK_SIZE}, {MKFOURCC('I', 'C', 'C', 'P'), WEBP_CHUNK_ICCP, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('A', 'N', 'I', 'M'), WEBP_CHUNK_ANIM, ANIM_CHUNK_SIZE}, {MKFOURCC('A', 'N', 'M', 'F'), WEBP_CHUNK_ANMF, ANMF_CHUNK_SIZE}, {MKFOURCC('A', 'L', 'P', 'H'), WEBP_CHUNK_ALPHA, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('V', 'P', '8', ' '), WEBP_CHUNK_IMAGE, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('V', 'P', '8', 'L'), WEBP_CHUNK_IMAGE, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('E', 'X', 'I', 'F'), WEBP_CHUNK_EXIF, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('X', 'M', 'P', ' '), WEBP_CHUNK_XMP, UNDEFINED_CHUNK_SIZE}, {NIL_TAG, WEBP_CHUNK_UNKNOWN, UNDEFINED_CHUNK_SIZE},

    {NIL_TAG, WEBP_CHUNK_NIL, UNDEFINED_CHUNK_SIZE}}

//------------------------------------------------------------------------------

int WebPGetMuxVersion(){
  return (MUX_MAJ_VERSION << 16) | (MUX_MIN_VERSION << 8) | MUX_REV_VERSION;
}

//------------------------------------------------------------------------------
// Life of a chunk object.

func ChunkInit(/* const */ chunk *WebPChunk) {
  assert.Assert(chunk);
  stdlib.Memset(chunk, 0, sizeof(*chunk));
  chunk.tag = NIL_TAG;
}

ChunkRelease *WebPChunk(/* const */ chunk *WebPChunk) {
  next *WebPChunk;
  if chunk == nil { return nil  }
  if (chunk.owner) {
    WebPDataClear(&chunk.data);
  }
  next = chunk.next;
  ChunkInit(chunk);
  return next;
}

//------------------------------------------------------------------------------
// Chunk misc methods.

CHUNK_INDEX ChunkGetIndexFromTag(uint32 tag) {
  var i int
  for i = 0; kChunks[i].tag != NIL_TAG; i++ {
    if tag == kChunks[i].tag { return (CHUNK_INDEX)i  }
  }
  return IDX_UNKNOWN;
}

WebPChunkId ChunkGetIdFromTag(uint32 tag) {
  var i int
  for i = 0; kChunks[i].tag != NIL_TAG; i++ {
    if tag == kChunks[i].tag { return kChunks[i].id  }
  }
  return WEBP_CHUNK_UNKNOWN;
}

uint32 ChunkGetTagFromFourCC(/* const */ byte fourcc[4]) {
  return MKFOURCC(fourcc[0], fourcc[1], fourcc[2], fourcc[3]);
}

CHUNK_INDEX ChunkGetIndexFromFourCC(/* const */ byte fourcc[4]) {
  tag := ChunkGetTagFromFourCC(fourcc);
  return ChunkGetIndexFromTag(tag);
}

//------------------------------------------------------------------------------
// Chunk search methods.

// Returns next chunk in the chunk list with the given tag.
static ChunkSearchNextInList *WebPChunk(chunk *WebPChunk, uint32 tag) {
  while (chunk != nil && chunk.tag != tag) {
    chunk = chunk.next;
  }
  return chunk;
}

ChunkSearchList *WebPChunk(first *WebPChunk, uint32 nth, uint32 tag) {
  iter := nth;
  first = ChunkSearchNextInList(first, tag);
  if first == nil { return nil  }

  while (--iter != 0) {
    next_chunk *WebPChunk = ChunkSearchNextInList(first.next, tag);
    if next_chunk == nil { break }
    first = next_chunk;
  }
  return ((nth > 0) && (iter > 0)) ? nil : first;
}

//------------------------------------------------------------------------------
// Chunk writer methods.

WebPMuxError ChunkAssignData(chunk *WebPChunk, /*const*/ data *WebPData, int copy_data, uint32 tag) {
  // For internally allocated chunks, always copy data & make it owner of data.
  if (tag == kChunks[IDX_VP8X].tag || tag == kChunks[IDX_ANIM].tag) {
    copy_data = 1;
  }

  ChunkRelease(chunk);

  if (data != nil) {
    if (copy_data) {  // Copy data.
      if !WebPDataCopy(data, &chunk.data) { return WEBP_MUX_MEMORY_ERROR  }
      chunk.owner = 1;  // Chunk is owner of data.
    } else {             // Don't copy data.
      chunk.data = *data;
    }
  }
  chunk.tag = tag;
  return WEBP_MUX_OK;
}

WebPMuxError ChunkSetHead(/* const */ chunk *WebPChunk, *WebPChunk* const chunk_list) {
  new_chunk *WebPChunk;

  assert.Assert(chunk_list != nil);
  if (*chunk_list != nil) {
    return WEBP_MUX_NOT_FOUND;
  }

//   new_chunk = (*WebPChunk)WebPSafeMalloc(uint64(1), sizeof(*new_chunk));
//   if new_chunk == nil { return WEBP_MUX_MEMORY_ERROR  }
  new_chunk = new(WebPChunk)

  *new_chunk = *chunk;
  chunk.owner = 0;
  new_chunk.next = nil;
  *chunk_list = new_chunk;
  return WEBP_MUX_OK;
}

WebPMuxError ChunkAppend(/* const */ chunk *WebPChunk, *WebPChunk** const chunk_list) {
  var err WebPMuxError 
  assert.Assert(chunk_list != nil && *chunk_list != nil);

  if (**chunk_list == nil) {
    err = ChunkSetHead(chunk, *chunk_list);
  } else {
    last_chunk *WebPChunk = **chunk_list;
    while (last_chunk.next != nil) last_chunk = last_chunk.next;
    err = ChunkSetHead(chunk, &last_chunk.next);
    if err == WEBP_MUX_OK { *chunk_list = &last_chunk.next }
  }
  return err;
}

//------------------------------------------------------------------------------
// Chunk deletion method(s).

func WebPChunk(/* const */ chunk *WebPChunk) *ChunkDelete {
  var next *WebPChunk = ChunkRelease(chunk);

  return next;
}

func ChunkListDelete(*WebPChunk* const chunk_list) {
  for *chunk_list != nil {
    *chunk_list = ChunkDelete(*chunk_list);
  }
}

//------------------------------------------------------------------------------
// Chunk serialization methods.

static ChunkEmit *uint8(/* const */ chunk *WebPChunk, dst *uint8) {
  chunk_size := chunk.data.size;
  assert.Assert(chunk);
  assert.Assert(chunk.tag != NIL_TAG);
  PutLE32(dst + 0, chunk.tag);
  PutLE32(dst + TAG_SIZE, (uint32)chunk_size);
  assert.Assert(chunk_size == (uint32)chunk_size);
  memcpy(dst + CHUNK_HEADER_SIZE, chunk.data.bytes, chunk_size);
  if chunk_size & 1 { dst[CHUNK_HEADER_SIZE + chunk_size] = 0 }  // Add padding.
  return dst + ChunkDiskSize(chunk);
}

ChunkListEmit *uint8(/* const */ chunk_list *WebPChunk, dst *uint8) {
  while (chunk_list != nil) {
    dst = ChunkEmit(chunk_list, dst);
    chunk_list = chunk_list.next;
  }
  return dst;
}

uint64 ChunkListDiskSize(/* const */ chunk_list *WebPChunk) {
  size uint64  = 0;
  while (chunk_list != nil) {
    size += ChunkDiskSize(chunk_list);
    chunk_list = chunk_list.next;
  }
  return size;
}

//------------------------------------------------------------------------------
// Life of a MuxImage object.

func MuxImageInit(/* const */ wpi *WebPMuxImage) {
  assert.Assert(wpi);
  stdlib.Memset(wpi, 0, sizeof(*wpi));
}

MuxImageRelease *WebPMuxImage(/* const */ wpi *WebPMuxImage) {
  next *WebPMuxImage;
  if wpi == nil { return nil  }
  // There should be at most one chunk of 'header', 'alpha', 'img' but we call
  // ChunkListDelete to be safe
  ChunkListDelete(&wpi.header);
  ChunkListDelete(&wpi.alpha);
  ChunkListDelete(&wpi.img);
  ChunkListDelete(&wpi.unknown);

  next = wpi.next;
  MuxImageInit(wpi);
  return next;
}

//------------------------------------------------------------------------------
// MuxImage search methods.

// Get a reference to appropriate chunk list within an image given chunk tag.
static *WebPChunk* GetChunkListFromId(/* const */ wpi *WebPMuxImage, WebPChunkId id) {
  assert.Assert(wpi != nil);
  switch (id) {
    case WEBP_CHUNK_ANMF:
      return (*WebPChunk*)&wpi.header;
    case WEBP_CHUNK_ALPHA:
      return (*WebPChunk*)&wpi.alpha;
    case WEBP_CHUNK_IMAGE:
      return (*WebPChunk*)&wpi.img;
    default:
      return nil;
  }
}

func MuxImageCount(/* const */ wpi_list *WebPMuxImage, WebPChunkId id) int {
  count := 0;
  const current *WebPMuxImage;
  for current = wpi_list; current != nil; current = current.next {
    if (id == WEBP_CHUNK_NIL) {
      count++  // Special case: count all images.
    } else {
      var wpi_chunk *WebPChunk = *GetChunkListFromId(current, id);
      if (wpi_chunk != nil) {
        const WebPChunkId wpi_chunk_id = ChunkGetIdFromTag(wpi_chunk.tag);
        if (wpi_chunk_id == id) count++  // Count images with a matching 'id'.
      }
    }
  }
  return count;
}

// Outputs a pointer to 'prev_wpi.next',
//   where 'prev_wpi' is the pointer to the image at position (nth - 1).
// Returns true if nth image was found.
func SearchImageToGetOrDelete(*WebPMuxImage* wpi_list, uint32 nth, *WebPMuxImage** const location) int {
  count := 0;
  assert.Assert(wpi_list);
  *location = wpi_list;

  if (nth == 0) {
    nth = MuxImageCount(*wpi_list, WEBP_CHUNK_NIL);
    if nth == 0 {
    return 0  // Not found.
}
  }

  while (*wpi_list != nil) {
    var cur_wpi *WebPMuxImage = *wpi_list;
    count++
    if count == nth {
    return 1  // Found.
}
    wpi_list = &cur_wpi.next;
    *location = wpi_list;
  }
  return 0;  // Not found.
}

//------------------------------------------------------------------------------
// MuxImage writer methods.

WebPMuxError MuxImagePush(/* const */ wpi *WebPMuxImage, *WebPMuxImage* wpi_list) {
  new_wpi *WebPMuxImage;

  while (*wpi_list != nil) {
    var cur_wpi *WebPMuxImage = *wpi_list;
    if cur_wpi.next == nil { break }
    wpi_list = &cur_wpi.next;
  }

//   new_wpi = (*WebPMuxImage)WebPSafeMalloc(uint64(1), sizeof(*new_wpi));
//   if new_wpi == nil { return WEBP_MUX_MEMORY_ERROR  }
  new_wpi := new(WebPMuxImage)

  *new_wpi = *wpi;
  new_wpi.next = nil;

  if (*wpi_list != nil) {
    (*wpi_list).next = new_wpi;
  } else {
    *wpi_list = new_wpi;
  }
  return WEBP_MUX_OK;
}

//------------------------------------------------------------------------------
// MuxImage deletion methods.

func WebPMuxImage(/* const */ wpi *WebPMuxImage) *MuxImageDelete {
  // Delete the components of wpi. If wpi is nil this is a noop.
  var next *WebPMuxImage = MuxImageRelease(wpi);
//   WebPSafeFree(wpi);
  return next;
}

WebPMuxError MuxImageDeleteNth(*WebPMuxImage* wpi_list, uint32 nth) {
  assert.Assert(wpi_list);
  if (!SearchImageToGetOrDelete(wpi_list, nth, &wpi_list)) {
    return WEBP_MUX_NOT_FOUND;
  }
  *wpi_list = MuxImageDelete(*wpi_list);
  return WEBP_MUX_OK;
}

//------------------------------------------------------------------------------
// MuxImage reader methods.

WebPMuxError MuxImageGetNth(/* const */ *WebPMuxImage* wpi_list, uint32 nth, *WebPMuxImage* wpi) {
  assert.Assert(wpi_list);
  assert.Assert(wpi);
  if (!SearchImageToGetOrDelete((*WebPMuxImage*)wpi_list, nth, (*WebPMuxImage**)&wpi_list)) {
    return WEBP_MUX_NOT_FOUND;
  }
  *wpi = (*WebPMuxImage)*wpi_list;
  return WEBP_MUX_OK;
}

//------------------------------------------------------------------------------
// MuxImage serialization methods.

// Size of an image.
uint64 MuxImageDiskSize(/* const */ wpi *WebPMuxImage) {
  size uint64  = 0;
  if wpi.header != nil { size += ChunkDiskSize(wpi.header) }
  if wpi.alpha != nil { size += ChunkDiskSize(wpi.alpha) }
  if wpi.img != nil { size += ChunkDiskSize(wpi.img) }
  if wpi.unknown != nil { size += ChunkListDiskSize(wpi.unknown) }
  return size;
}

// Special case as ANMF chunk encapsulates other image chunks.
static ChunkEmitSpecial *uint8(/* const */ header *WebPChunk, uint64 total_size, dst *uint8) {
  header_size := header.data.size;
  offset_to_next := total_size - CHUNK_HEADER_SIZE;
  assert.Assert(header.tag == kChunks[IDX_ANMF].tag);
  PutLE32(dst + 0, header.tag);
  PutLE32(dst + TAG_SIZE, (uint32)offset_to_next);
  assert.Assert(header_size == (uint32)header_size);
  memcpy(dst + CHUNK_HEADER_SIZE, header.data.bytes, header_size);
  if (header_size & 1) {
    dst[CHUNK_HEADER_SIZE + header_size] = 0;  // Add padding.
  }
  return dst + ChunkDiskSize(header);
}

MuxImageEmit *uint8(/* const */ wpi *WebPMuxImage, dst *uint8) {
  // Ordering of chunks to be emitted is strictly as follows:
  // 1. ANMF chunk (if present).
  // 2. ALPH chunk (if present).
  // 3. VP8/VP8L chunk.
  assert.Assert(wpi);
  if (wpi.header != nil) {
    dst = ChunkEmitSpecial(wpi.header, MuxImageDiskSize(wpi), dst);
  }
  if wpi.alpha != nil { dst = ChunkEmit(wpi.alpha, dst) }
  if wpi.img != nil { dst = ChunkEmit(wpi.img, dst) }
  if wpi.unknown != nil { dst = ChunkListEmit(wpi.unknown, dst) }
  return dst;
}

//------------------------------------------------------------------------------
// Helper methods for mux.

func MuxHasAlpha(/* const */ images *WebPMuxImage) int {
  while (images != nil) {
    if images.has_alpha { return 1  }
    images = images.next;
  }
  return 0;
}

MuxEmitRiffHeader *uint8(/* const */ data *uint8, size uint64 ) {
  PutLE32(data + 0, MKFOURCC('R', 'I', 'F', 'F'));
  PutLE32(data + TAG_SIZE, (uint32)size - CHUNK_HEADER_SIZE);
  assert.Assert(size == (uint32)size);
  PutLE32(data + TAG_SIZE + CHUNK_SIZE_BYTES, MKFOURCC('W', 'E', 'B', 'P'));
  return data + RIFF_HEADER_SIZE;
}

*WebPChunk* MuxGetChunkListFromId(/* const */ mux *WebPMux, WebPChunkId id) {
  assert.Assert(mux != nil);
  switch (id) {
    case WEBP_CHUNK_VP8X:
      return (*WebPChunk*)&mux.vp8x;
    case WEBP_CHUNK_ICCP:
      return (*WebPChunk*)&mux.iccp;
    case WEBP_CHUNK_ANIM:
      return (*WebPChunk*)&mux.anim;
    case WEBP_CHUNK_EXIF:
      return (*WebPChunk*)&mux.exif;
    case WEBP_CHUNK_XMP:
      return (*WebPChunk*)&mux.xmp;
    default:
      return (*WebPChunk*)&mux.unknown;
  }
}

func IsNotCompatible(int feature, int num_items) int {
  return (feature != 0) != (num_items > 0);
}

const NO_FLAG =((WebPFeatureFlags)0)

// Test basic constraints:
// retrieval, maximum number of chunks by index (use -1 to skip)
// and feature incompatibility (use NO_FLAG to skip).
// On success returns WEBP_MUX_OK and stores the chunk count in *num.
func ValidateChunk(/* const */ mux *WebPMux, CHUNK_INDEX idx, WebPFeatureFlags feature, uint32 vp8x_flags, int max, num *int) WebPMuxError {
  var err WebPMuxError  = WebPMuxNumChunks(mux, kChunks[idx].id, num);
  if err != WEBP_MUX_OK { return err  }
  if max > -1 && *num > max { return WEBP_MUX_INVALID_ARGUMENT  }
  if (feature != NO_FLAG && IsNotCompatible(vp8x_flags & feature, *num)) {
    return WEBP_MUX_INVALID_ARGUMENT;
  }
  return WEBP_MUX_OK;
}

func MuxValidate(/* const */ mux *WebPMux) WebPMuxError {
  var num_iccp int
  var num_exif int
  var num_xmp int
  var num_anim int
  var num_frames int
  var num_vp8x int
  var num_images int
  var num_alpha int
  var flags uint32
  var err WebPMuxError 

  // Verify mux is not nil.
  if mux == nil { return WEBP_MUX_INVALID_ARGUMENT  }

  // Verify mux has at least one image.
  if mux.images == nil { return WEBP_MUX_INVALID_ARGUMENT  }

  err = WebPMuxGetFeatures(mux, &flags);
  if err != WEBP_MUX_OK { return err  }

  // At most one color profile chunk.
  err = ValidateChunk(mux, IDX_ICCP, ICCP_FLAG, flags, 1, &num_iccp);
  if err != WEBP_MUX_OK { return err  }

  // At most one EXIF metadata.
  err = ValidateChunk(mux, IDX_EXIF, EXIF_FLAG, flags, 1, &num_exif);
  if err != WEBP_MUX_OK { return err  }

  // At most one XMP metadata.
  err = ValidateChunk(mux, IDX_XMP, XMP_FLAG, flags, 1, &num_xmp);
  if err != WEBP_MUX_OK { return err  }

  // Animation: ANIMATION_FLAG, ANIM chunk and ANMF chunk(s) are consistent.
  // At most one ANIM chunk.
  err = ValidateChunk(mux, IDX_ANIM, NO_FLAG, flags, 1, &num_anim);
  if err != WEBP_MUX_OK { return err  }
  err = ValidateChunk(mux, IDX_ANMF, NO_FLAG, flags, -1, &num_frames);
  if err != WEBP_MUX_OK { return err  }

  {
    has_animation := !!(flags & ANIMATION_FLAG);
    if (has_animation && (num_anim == 0 || num_frames == 0)) {
      return WEBP_MUX_INVALID_ARGUMENT;
    }
    if (!has_animation && (num_anim == 1 || num_frames > 0)) {
      return WEBP_MUX_INVALID_ARGUMENT;
    }
    if (!has_animation) {
      var images *WebPMuxImage = mux.images;
      // There can be only one image.
      if (images == nil || images.next != nil) {
        return WEBP_MUX_INVALID_ARGUMENT;
      }
      // Size must match.
      if (mux.canvas_width > 0) {
        if (images.width != mux.canvas_width ||
            images.height != mux.canvas_height) {
          return WEBP_MUX_INVALID_ARGUMENT;
        }
      }
    }
  }

  // Verify either VP8X chunk is present OR there is only one elem in
  // mux.images.
  err = ValidateChunk(mux, IDX_VP8X, NO_FLAG, flags, 1, &num_vp8x);
  if err != WEBP_MUX_OK { return err  }
  err = ValidateChunk(mux, IDX_VP8, NO_FLAG, flags, -1, &num_images);
  if err != WEBP_MUX_OK { return err  }
  if num_vp8x == 0 && num_images != 1 { return WEBP_MUX_INVALID_ARGUMENT  }

  // ALPHA_FLAG & alpha chunk(s) are consistent.
  // Note: ALPHA_FLAG can be set when there is actually no Alpha data present.
  if (MuxHasAlpha(mux.images)) {
    if (num_vp8x > 0) {
      // VP8X chunk is present, so it should contain ALPHA_FLAG.
      if !(flags & ALPHA_FLAG) { return WEBP_MUX_INVALID_ARGUMENT  }
    } else {
      // VP8X chunk is not present, so ALPH chunks should NOT be present either.
      err = WebPMuxNumChunks(mux, WEBP_CHUNK_ALPHA, &num_alpha);
      if err != WEBP_MUX_OK { return err  }
      if num_alpha > 0 { return WEBP_MUX_INVALID_ARGUMENT  }
    }
  }

  return WEBP_MUX_OK;
}

#undef NO_FLAG

//------------------------------------------------------------------------------
