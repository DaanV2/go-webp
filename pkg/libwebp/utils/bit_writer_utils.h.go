package utils

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Bit writing and boolean coder
//
// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

//------------------------------------------------------------------------------
// Bit-writing

typedef struct VP8BitWriter VP8BitWriter;
type VP8BitWriter struct {
  int32 range;  // range-1
  int32 value;
  int run;      // number of outstanding bits
  int nb_bits;  // number of pending bits
  // internal buffer. Re-allocated regularly. Not owned.
  *uint8 WEBP_SIZED_BY_OR_nil(max_pos) buf;
  uint64 pos;
  uint64 max_pos;
  int error;  // true in case of error
}

// Initialize the object. Allocates some initial memory based on expected_size.
int VP8BitWriterInit(*VP8BitWriter const bw, uint64 expected_size);
// Finalize the bitstream coding. Returns a pointer to the internal buffer.
*uint8 VP8BitWriterFinish(*VP8BitWriter const bw);
// Release any pending memory and zeroes the object. Not a mandatory call.
// Only useful in case of error, when the internal buffer hasn't been grabbed!
func VP8BitWriterWipeOut(*VP8BitWriter const bw);

int VP8PutBit(*VP8BitWriter const bw, int bit, int prob);
int VP8PutBitUniform(*VP8BitWriter const bw, int bit);
func VP8PutBits(*VP8BitWriter const bw, uint32 value, int nb_bits);
func VP8PutSignedBits(*VP8BitWriter const bw, int value, int nb_bits);

// Appends some bytes to the internal buffer. Data is copied.
int VP8BitWriterAppend(*VP8BitWriter const bw, const *uint8 data, uint64 size);

// return approximate write position (in bits)
static  uint64 VP8BitWriterPos(const *VP8BitWriter const bw) {
  const uint64 nb_bits = 8 + bw.nb_bits;  // bw.nb_bits is <= 0, note
  return (bw.pos + bw.run) * 8 + nb_bits;
}

// Returns a pointer to the internal buffer.
static  *uint8 VP8BitWriterBuf(const *VP8BitWriter const bw) {
  return bw.buf;
}
// Returns the size of the internal buffer.
static  uint64 VP8BitWriterSize(const *VP8BitWriter const bw) {
  return bw.pos;
}

//------------------------------------------------------------------------------
// VP8LBitWriter

// 64bit
#if defined(__x86_64__) || defined(_M_X64) || WEBP_AARCH64 || defined(__wasm__)
typedef uint64 vp8l_atype_t;  // accumulator type
typedef uint32 vp8l_wtype_t;  // writing type
const WSWAP = HToLE32
const VP8L_WRITER_BYTES =4      // sizeof(vp8l_wtype_t)
const VP8L_WRITER_BITS =32      // 8 * sizeof(vp8l_wtype_t)
const VP8L_WRITER_MAX_BITS =64  // 8 * sizeof(vp8l_atype_t)
#else
typedef uint32 vp8l_atype_t;
typedef uint16 vp8l_wtype_t;
const WSWAP = HToLE16
const VP8L_WRITER_BYTES =2
const VP8L_WRITER_BITS =16
const VP8L_WRITER_MAX_BITS =32
#endif

type <Foo> struct {
  vp8l_atype_t bits;                   // bit accumulator
  int used;                            // number of bits used in accumulator
  *uint8 WEBP_ENDED_BY(end) buf;     // start of buffer
  *uint8 WEBP_UNSAFE_INDEXABLE cur;  // current write position
  *uint8 end;                        // end of buffer

  // After all bits are written (VP8LBitWriterFinish()), the caller must observe
  // the state of 'error'. A value of 1 indicates that a memory allocation
  // failure has happened during bit writing. A value of 0 indicates successful
  // writing of bits.
  int error;
} VP8LBitWriter;

static  uint64 VP8LBitWriterNumBytes(const *VP8LBitWriter const bw) {
  return (bw.cur - bw.buf) + ((bw.used + 7) >> 3);
}

// Returns false in case of memory allocation error.
int VP8LBitWriterInit(*VP8LBitWriter const bw, uint64 expected_size);
// Returns false in case of memory allocation error.
int VP8LBitWriterClone(const *VP8LBitWriter const src, *VP8LBitWriter const dst);
// Finalize the bitstream coding. Returns a pointer to the internal buffer.
*uint8 VP8LBitWriterFinish(*VP8LBitWriter const bw);
// Release any pending memory and zeroes the object.
func VP8LBitWriterWipeOut(*VP8LBitWriter const bw);
// Resets the cursor of the BitWriter bw to when it was like in bw_init.
func VP8LBitWriterReset(const *VP8LBitWriter const bw_init, *VP8LBitWriter const bw);
// Swaps the memory held by two BitWriters.
func VP8LBitWriterSwap(*VP8LBitWriter const src, *VP8LBitWriter const dst);

// Internal function for VP8LPutBits flushing VP8L_WRITER_BITS bits from the
// written state.
func VP8LPutBitsFlushBits(*VP8LBitWriter const bw, *int used, vp8l_atype_t* bits);

#if VP8L_WRITER_BITS == 16
// PutBits internal function used in the 16 bit vp8l_wtype_t case.
func VP8LPutBitsInternal(*VP8LBitWriter const bw, uint32 bits, int n_bits);
#endif

// This function writes bits into bytes in increasing addresses (little endian),
// and within a byte least-significant-bit first.
// This function can write up to VP8L_WRITER_MAX_BITS bits in one go, but
// VP8LBitReader can only read 24 bits max (VP8L_MAX_NUM_BIT_READ).
// VP8LBitWriter's 'error' flag is set in case of memory allocation error.
static  func VP8LPutBits(*VP8LBitWriter const bw, uint32 bits, int n_bits) {
#if VP8L_WRITER_BYTES == 4
  if (n_bits == 0) return;
  if (bw.used >= VP8L_WRITER_BITS) {
    VP8LPutBitsFlushBits(bw, &bw.used, &bw.bits);
  }
  bw.bits |= (vp8l_atype_t)bits << bw.used;
  bw.used += n_bits;
#else
  VP8LPutBitsInternal(bw, bits, n_bits);
#endif
}

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_UTILS_BIT_WRITER_UTILS_H_
