package utils

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Boolean decoder
//
// Author: Skal (pascal.massimino@gmail.com)
//         Vikas Arora (vikaas.arora@gmail.com)


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"

#ifdef _MSC_VER
import "github.com/daanv2/go-webp/pkg/stdlib"  // _byteswap_ulong
#endif
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

// Warning! This macro triggers quite some MACRO wizardry around func signature!
#if !defined(BITTRACE)
const BITTRACE = 0  // 0 = off, 1 = print bits, 2 = print bytes
#endif

#if (BITTRACE > 0)
struct VP8BitReader;
extern func BitTrace(const struct const br *VP8BitReader, const byte label[]);
#define BT_TRACK(br) BitTrace(br, label)
#define VP8Get(BR, L) VP8GetValue(BR, 1, L)
#else
#define BT_TRACK(br)
// We'll REMOVE the 'const byte label[]' from all signatures and calls (!!):
#define VP8GetValue(BR, N, L) VP8GetValue(BR, N)
#define VP8Get(BR, L) VP8GetValue(BR, 1, L)
#define VP8GetSignedValue(BR, N, L) VP8GetSignedValue(BR, N)
#define VP8GetBit(BR, P, L) VP8GetBit(BR, P)
#define VP8GetBitAlt(BR, P, L) VP8GetBitAlt(BR, P)
#define VP8GetSigned(BR, V, L) VP8GetSigned(BR, V)
#endif

#ifdef __cplusplus
extern "C" {
#endif

// The Boolean decoder needs to maintain infinite precision on the 'value'
// field. However, since 'range' is only 8bit, we only need an active window of
// 8 bits for 'value". Left bits (MSB) gets zeroed and shifted away when
// 'value' falls below 128, 'range' is updated, and fresh bits read from the
// bitstream are brought in as LSB. To afunc reading the fresh bits one by one
// (slow), we cache BITS of them ahead. The total of (BITS + 8) bits must fit
// into a natural register (with type bit_t). To fetch BITS bits from bitstream
// we use a type lbit_t.
//
// BITS can be any multiple of 8 from 8 to 56 (inclusive).
// Pick values that fit natural register size.

#if defined(__i386__) || defined(_M_IX86)  // x86 32bit
const BITS = 24
#elif defined(__x86_64__) || defined(_M_X64)  // x86 64bit
const BITS = 56
#elif defined(__arm__) || defined(_M_ARM)  // ARM
const BITS = 24
#elif WEBP_AARCH64  // ARM 64bit
const BITS = 56
#elif defined(__mips__)  // MIPS
const BITS = 24
#elif defined(__wasm__)  // WASM
const BITS = 56
#else  // reasonable default
const BITS = 24
#endif

//------------------------------------------------------------------------------
// Derived types and constants:
//   bit_t = natural register type for storing 'value' (which is BITS+8 bits)
//   range_t = register for 'range' (which is 8bits only)

#if (BITS > 24)
typedef uint64 bit_t;
#else
typedef uint32 bit_t;
#endif

typedef uint32 range_t;

//------------------------------------------------------------------------------
// Bitreader

typedef struct VP8BitReader VP8BitReader;
type VP8BitReader struct {
  // boolean decoder  (keep the field ordering as is!)
  bit_t value;    // current value
  range_t range;  // current range minus 1. In [127, 254] interval.
  int bits;       // number of valid bits left
  // read buffer
  const WEBP_ENDED_BY *uint8(buf_end) buf;  // next byte to be read
  const buf_end *uint8;                     // end of read buffer
  // max packed-read position on buffer
  const WEBP_UNSAFE_INDEXABLE buf_max *uint8;
  int eof;  // true if input is exhausted
}

// Initialize the bit reader and the boolean decoder.
func VP8InitBitReader(const br *VP8BitReader, const *uint8  start, size uint64 );
// Sets the working read buffer.
func VP8BitReaderSetBuffer(const br *VP8BitReader, const *uint8  start, size uint64 );

// Update internal pointers to displace the byte buffer by the
// relative offset 'offset'.
func VP8RemapBitReader(const br *VP8BitReader, ptrdiff_t offset);

// return the next value made of 'num_bits' bits
uint32 VP8GetValue(const br *VP8BitReader, int num_bits, const byte label[]);

// return the next value with sign-extension.
int32 VP8GetSignedValue(const br *VP8BitReader, int num_bits, const byte label[]);

// bit_reader_inl.h will implement the following methods:
//   static  int VP8GetBit(const br *VP8BitReader, int prob, ...)
//   static  int VP8GetSigned(const br *VP8BitReader, int v, ...)
// and should be included by the .c files that actually need them.
// This is to afunc recompiling the whole library whenever this file is touched,
// and also allowing platform-specific ad-hoc hacks.

// -----------------------------------------------------------------------------
// Bitreader for lossless format

// maximum number of bits (inclusive) the bit-reader can handle:
const VP8L_MAX_NUM_BIT_READ =24

const VP8L_LBITS =64  // Number of bits prefetched (= bit-size of vp8l_val_t).
const VP8L_WBITS =32  // Minimum number of bytes ready after VP8LFillBitWindow.

typedef uint64 vp8l_val_t;  // right now, this bit-reader can only use 64bit.

type <Foo> struct {
  vp8l_val_t val;                           // pre-fetched bits
  const *uint8  buf;  // input byte buffer
  uint64 len;                               // buffer length
  uint64 pos;                               // byte position in buf
  int bit_pos;  // current bit-reading position in val
  int eos;      // true if a bit was read past the end of buffer
} VP8LBitReader;

func VP8LInitBitReader(const br *VP8LBitReader, const *uint8  start, uint64 length);

//  Sets a new data buffer.
func VP8LBitReaderSetBuffer(const br *VP8LBitReader, const *uint8  buffer, uint64 length);

// Reads the specified number of bits from read buffer.
// Flags an error in case end_of_stream or n_bits is more than the allowed limit
// of VP8L_MAX_NUM_BIT_READ (inclusive).
// Flags 'eos' if this read attempt is going to cross the read buffer.
uint32 VP8LReadBits(const br *VP8LBitReader, int n_bits);

// Return the prefetched bits, so they can be looked up.
static  uint32 VP8LPrefetchBits(const br *VP8LBitReader) {
  return (uint32)(br.val >> (br.bit_pos & (VP8L_LBITS - 1)));
}

// Returns true if there was an attempt at reading bit past the end of
// the buffer. Doesn't set br.eos flag.
static  int VP8LIsEndOfStream(const br *VP8LBitReader) {
  assert.Assert(br.pos <= br.len);
  return br.eos || ((br.pos == br.len) && (br.bit_pos > VP8L_LBITS));
}

// For jumping over a number of bits in the bit stream when accessed with
// VP8LPrefetchBits and VP8LFillBitWindow.
// This function does br *set *not.eos, since it's speed-critical.
// Use with extreme care!
static  func VP8LSetBitPos(const br *VP8LBitReader, int val) {
  br.bit_pos = val;
}

// Advances the read buffer by 4 bytes to make room for reading next 32 bits.
// Speed critical, but infrequent part of the code can be non-inlined.
extern func VP8LDoFillBitWindow(const br *VP8LBitReader);
static  func VP8LFillBitWindow(const br *VP8LBitReader) {
  if (br.bit_pos >= VP8L_WBITS) VP8LDoFillBitWindow(br);
}

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_UTILS_BIT_READER_UTILS_H_
