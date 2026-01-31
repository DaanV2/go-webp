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


//------------------------------------------------------------------------------
// Bit-writing

typedef struct VP8BitWriter VP8BitWriter;
type VP8BitWriter struct {
  int32 range;  // range-1
  int32 value;
  int run;      // number of outstanding bits
  int nb_bits;  // number of pending bits
  // internal buffer. Re-allocated regularly. Not owned.
  WEBP_SIZED_BY_OR_nil *uint8(max_pos) buf;
  uint64 pos;
  uint64 max_pos;
  int error;  // true in case of error
}

// Initialize the object. Allocates some initial memory based on expected_size.
int VP8BitWriterInit(const bw *VP8BitWriter, uint64 expected_size);
// Finalize the bitstream coding. Returns a pointer to the internal buffer.
VP *uint88BitWriterFinish(const bw *VP8BitWriter);
// Release any pending memory and zeroes the object. Not a mandatory call.
// Only useful in case of error, when the internal buffer hasn't been grabbed!
func VP8BitWriterWipeOut(const bw *VP8BitWriter);

int VP8PutBit(const bw *VP8BitWriter, int bit, int prob);
int VP8PutBitUniform(const bw *VP8BitWriter, int bit);
func VP8PutBits(const bw *VP8BitWriter, uint32 value, int nb_bits);
func VP8PutSignedBits(const bw *VP8BitWriter, int value, int nb_bits);

// Appends some bytes to the internal buffer. Data is copied.
int VP8BitWriterAppend(const bw *VP8BitWriter, const data *uint8, size uint64 );

// return approximate write position (in bits)
static  uint64 VP8BitWriterPos(const bw *VP8BitWriter) {
  nb_bits := 8 + bw.nb_bits;  // bw.nb_bits is <= 0, note
  return (bw.pos + bw.run) * 8 + nb_bits;
}

// Returns a pointer to the internal buffer.
static  VP *uint88BitWriterBuf(const bw *VP8BitWriter) {
  return bw.buf;
}
// Returns the size of the internal buffer.
static  uint64 VP8BitWriterSize(const bw *VP8BitWriter) {
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

type VP8LBitWriter struct {
  vp8l_atype_t bits;                   // bit accumulator
  int used;                            // number of bits used in accumulator
  WEBP_ENDED_BY *uint8(end) buf;     // start of buffer
  WEBP_UNSAFE_INDEXABLE cur *uint8;  // current write position
  end *uint8;                        // end of buffer

  // After all bits are written (VP8LBitWriterFinish()), the caller must observe
  // the state of 'error'. A value of 1 indicates that a memory allocation
  // failure has happened during bit writing. A value of 0 indicates successful
  // writing of bits.
  int error;
} ;

static  uint64 VP8LBitWriterNumBytes(const bw *VP8LBitWriter) {
  return (bw.cur - bw.buf) + ((bw.used + 7) >> 3);
}

// Returns false in case of memory allocation error.
int VP8LBitWriterInit(const bw *VP8LBitWriter, uint64 expected_size);
// Returns false in case of memory allocation error.
int VP8LBitWriterClone(const src *VP8LBitWriter, const dst *VP8LBitWriter);
// Finalize the bitstream coding. Returns a pointer to the internal buffer.
VP *uint88LBitWriterFinish(const bw *VP8LBitWriter);
// Release any pending memory and zeroes the object.
func VP8LBitWriterWipeOut(const bw *VP8LBitWriter);
// Resets the cursor of the BitWriter bw to when it was like in bw_init.
func VP8LBitWriterReset(const bw_init *VP8LBitWriter, const bw *VP8LBitWriter);
// Swaps the memory held by two BitWriters.
func VP8LBitWriterSwap(const src *VP8LBitWriter, const dst *VP8LBitWriter);

// Internal function for VP8LPutBits flushing VP8L_WRITER_BITS bits from the
// written state.
func VP8LPutBitsFlushBits(const bw *VP8LBitWriter, used *int, vp8l_atype_t* bits);

#if VP8L_WRITER_BITS == 16
// PutBits internal function used in the 16 bit vp8l_wtype_t case.
func VP8LPutBitsInternal(const bw *VP8LBitWriter, uint32 bits, int n_bits);
#endif

// This function writes bits into bytes in increasing addresses (little endian),
// and within a byte least-significant-bit first.
// This function can write up to VP8L_WRITER_MAX_BITS bits in one go, but
// VP8LBitReader can only read 24 bits max (VP8L_MAX_NUM_BIT_READ).
// VP8LBitWriter's 'error' flag is set in case of memory allocation error.
static  func VP8LPutBits(const bw *VP8LBitWriter, uint32 bits, int n_bits) {
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



#endif  // WEBP_UTILS_BIT_WRITER_UTILS_H_
