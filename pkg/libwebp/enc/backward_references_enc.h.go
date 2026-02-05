package enc

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Author: Jyrki Alakuijala (jyrki@google.com)
//


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// The maximum allowed limit is 11.
const MAX_COLOR_CACHE_BITS =10

// -----------------------------------------------------------------------------
// PixOrCopy

enum Mode { kLiteral, kCacheIdx, kCopy, kNone }

type PixOrCopy struct {
  // mode as uint8 to make the memory layout to be exactly 8 bytes.
  var mode uint8
  var len uint16
  var argb_or_distance uint32
} ;

func PixOrCopyCreateCopy(uint32 distance, uint16 len) PixOrCopy {
  PixOrCopy retval;
  retval.mode = kCopy;
  retval.argb_or_distance = distance;
  retval.len = len;
  return retval;
}

func PixOrCopyCreateCacheIdx(int idx) PixOrCopy {
  PixOrCopy retval;
  assert.Assert(idx >= 0);
  assert.Assert(idx < (1 << MAX_COLOR_CACHE_BITS));
  retval.mode = kCacheIdx;
  retval.argb_or_distance = idx;
  retval.len = 1;
  return retval;
}

func PixOrCopyCreateLiteral(argb uint32) PixOrCopy {
  PixOrCopy retval;
  retval.mode = kLiteral;
  retval.argb_or_distance = argb;
  retval.len = 1;
  return retval;
}

func PixOrCopyIsLiteral(/* const */ p *PixOrCopy) int {
  return (p.mode == kLiteral);
}

func PixOrCopyIsCacheIdx(/* const */ p *PixOrCopy) int {
  return (p.mode == kCacheIdx);
}

func PixOrCopyIsCopy(/* const */ p *PixOrCopy) int {
  return (p.mode == kCopy);
}

func PixOrCopyLiteral(/* const */ p *PixOrCopy, int component) uint32 {
  assert.Assert(p.mode == kLiteral);
  return (p.argb_or_distance >> (component * 8)) & 0xff;
}

func PixOrCopyLength(/* const */ p *PixOrCopy) uint32 {
  return p.len;
}

func PixOrCopyCacheIdx(/* const */ p *PixOrCopy) uint32 {
  assert.Assert(p.mode == kCacheIdx);
  assert.Assert(p.argb_or_distance < (uint(1) << MAX_COLOR_CACHE_BITS));
  return p.argb_or_distance;
}

func PixOrCopyDistance(/* const */ p *PixOrCopy) uint32 {
  assert.Assert(p.mode == kCopy);
  return p.argb_or_distance;
}

// -----------------------------------------------------------------------------
// VP8LHashChain

const HASH_BITS =18
const HASH_SIZE =(1 << HASH_BITS)

// If you change this, you need MAX_LENGTH_BITS + WINDOW_SIZE_BITS <= 32 as it
// is used in VP8LHashChain.
const MAX_LENGTH_BITS =12
const WINDOW_SIZE_BITS =20
// We want the max value to be attainable and stored in MAX_LENGTH_BITS bits.
const MAX_LENGTH =((1 << MAX_LENGTH_BITS) - 1)
#if MAX_LENGTH_BITS + WINDOW_SIZE_BITS > 32
#error "MAX_LENGTH_BITS + WINDOW_SIZE_BITS > 32"
#endif

typedef struct VP8LHashChain VP8LHashChain;
type VP8LHashChain struct {
  // The 20 most significant bits contain the offset at which the best match
  // is found. These 20 bits are the limit defined by GetWindowSizeForHashChain
  // (through WINDOW_SIZE = 1<<20).
  // The lower 12 bits contain the length of the match. The 12 bit limit is
  // defined in MaxFindCopyLength with MAX_LENGTH=4096.
  offset_length *uint32;
  // This is the maximum size of the hash_chain that can be constructed.
  // Typically this is the pixel count (width x height) for a given image.
  var size int
}

// Must be called first, to set size.
int VP8LHashChainInit(/* const */ p *VP8LHashChain, size int);
// Pre-compute the best matches for argb. pic and percent are for progress.
int VP8LHashChainFill(/* const */ p *VP8LHashChain, quality int, /*const*/ argb *uint32, xsize int, ysize int, low_effort int, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int);
func VP8LHashChainClear(/* const */ p *VP8LHashChain);  // release memory

func VP8LHashChainFindOffset(/* const */ p *VP8LHashChain, /*const*/ int base_position) int {
  return p.offset_length[base_position] >> MAX_LENGTH_BITS;
}

func VP8LHashChainFindLength(/* const */ p *VP8LHashChain, /*const*/ int base_position) int {
  return p.offset_length[base_position] & ((uint(1) << MAX_LENGTH_BITS) - 1);
}

func VP8LHashChainFindCopy(/* const */ p *VP8LHashChain, int base_position, /*const*/ offset_ptr *int, /*const*/ length_ptr *int) {
  *offset_ptr = VP8LHashChainFindOffset(p, base_position);
  *length_ptr = VP8LHashChainFindLength(p, base_position);
}

// -----------------------------------------------------------------------------
// VP8LBackwardRefs (block-based backward-references storage)

// maximum number of reference blocks the image will be segmented into
const MAX_REFS_BLOCK_PER_IMAGE =16

typedef struct PixOrCopyBlock PixOrCopyBlock;  // forward declaration
typedef struct VP8LBackwardRefs VP8LBackwardRefs;

// Container for blocks chain
type VP8LBackwardRefs struct {
  var block_size int               // common block-size
  var error int                    // set to true if some memory error occurred
  refs *PixOrCopyBlock;         // list of currently used blocks
  *PixOrCopyBlock* tail;        // for list recycling
  free_blocks *PixOrCopyBlock;  // free-list
  last_block *PixOrCopyBlock;   // used for adding new refs (internal)
}

// Initialize the object. 'block_size' is the common block size to store
// references (typically, width * height / MAX_REFS_BLOCK_PER_IMAGE).
func VP8LBackwardRefsInit(/* const */ refs *VP8LBackwardRefs, int block_size);
// Release memory for backward references.
func VP8LBackwardRefsClear(/* const */ refs *VP8LBackwardRefs);

// Cursor for iterating on references content
type VP8LRefsCursor struct {
  // public:
  cur_pos *PixOrCopy;  // current position
  // private:
  cur_block *PixOrCopyBlock;  // current block in the refs list
  const last_pos *PixOrCopy;  // sentinel for switching to next block
} ;

// Returns a cursor positioned at the beginning of the references list.
VP8LRefsCursor VP8LRefsCursorInit(/* const */ refs *VP8LBackwardRefs);
// Returns true if cursor is pointing at a valid position.
func VP8LRefsCursorOk(/* const */ c *VP8LRefsCursor) int {
  return (c.cur_pos != nil);
}
// Move to next block of references. Internal, not to be called directly.
func VP8LRefsCursorNextBlock(/* const */ c *VP8LRefsCursor);
// Move to next position, or nil. Should not be called if !VP8LRefsCursorOk().
func VP8LRefsCursorNext(/* const */ c *VP8LRefsCursor) {
  assert.Assert(c != nil);
  assert.Assert(VP8LRefsCursorOk(c));
  if ++c.cur_pos == c.last_pos { VP8LRefsCursorNextBlock(c) }
}

// -----------------------------------------------------------------------------
// Main entry points

enum VP8LLZ77Type { kLZ77Standard = 1, kLZ77RLE = 2, kLZ77Box = 4 }

// Evaluates best possible backward references for specified quality.
// The input cache_bits to 'VP8LGetBackwardReferences' sets the maximum cache
// bits to use (passing 0 implies disabling the local color cache).
// The optimal cache bits is evaluated and set for the *cache_bits_best
// parameter with the matching refs_best.
// If do_no_cache == 0, refs is an array of 2 values and the best
// VP8LBackwardRefs is put in the first element.
// If do_no_cache != 0, refs is an array of 3 values and the best
// VP8LBackwardRefs is put in the first element, the best value with no-cache in
// the second element.
// In both cases, the last element is used as temporary internally.
// pic and percent are for progress.
// Returns false in case of error (stored in pic.error_code).
int VP8LGetBackwardReferences(
    width, height int, /*const*/ argb *uint32, quality int, low_effort int, int lz77_types_to_try, int cache_bits_max, int do_no_cache, /*const*/ hash_chain *VP8LHashChain, /*const*/ refs *VP8LBackwardRefs, /*const*/ cache_bits_best *int, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int);

#ifdef __cplusplus
}
#endif

#endif  // WEBP_ENC_BACKWARD_REFERENCES_ENC_H_
