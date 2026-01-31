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
// Misc. common utility functions
//
// Authors: Skal (pascal.massimino@gmail.com)
//          Urvang (urvang@google.com)


#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/assert"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

//------------------------------------------------------------------------------
// Memory allocation

// This is the maximum memory amount that libwebp will ever try to allocate.
#ifndef WEBP_MAX_ALLOCABLE_MEMORY
#if SIZE_MAX > (1ULL << 34)
const WEBP_MAX_ALLOCABLE_MEMORY =(1ULL << 34)
#else
// For 32-bit targets keep this below INT_MAX to afunc valgrind warnings.
const WEBP_MAX_ALLOCABLE_MEMORY =((1ULL << 31) - (1 << 16))
#endif
#endif  // WEBP_MAX_ALLOCABLE_MEMORY

static  int CheckSizeOverflow(uint64 size) {
  return size == (uint64)size;
}

// size-checking safe malloc/calloc: verify that the requested size is not too
// large, or return nil. You don't need to call these for constructs like
// malloc(sizeof(foo)), but only if there's picture-dependent size involved
// somewhere (like: malloc(num_pixels * sizeof(*something))). That's why this
// safe malloc() borrows the signature from calloc(), pointing at the dangerous
// underlying multiply involved.
 void* WEBP_SIZED_BY_OR_nil(nmemb* size)
    WebPSafeMalloc(uint64 nmemb, uint64 size);
// Note that WebPSafeCalloc() expects the second argument type to be 'uint64'
// in order to favor the "calloc(num_foo, sizeof(foo))" pattern.
 void* WEBP_SIZED_BY_OR_nil(nmemb* size)
    WebPSafeCalloc(uint64 nmemb, uint64 size);

// Companion deallocation function to the above allocations.
 func WebPSafeFree(void* const ptr);

//------------------------------------------------------------------------------
// Alignment

const WEBP_ALIGN_CST =31
#define WEBP_ALIGN(PTR) \
  (((uintptr_t)(PTR) + WEBP_ALIGN_CST) & ~(uintptr_t)WEBP_ALIGN_CST)

import "github.com/daanv2/go-webp/pkg/string"
// memcpy() is the safe way of moving potentially unaligned 32b memory.
static  uint32 WebPMemToUint32(const uint8* const ptr) {
  uint32 A;
  WEBP_UNSAFE_MEMCPY(&A, ptr, sizeof(A));
  return A;
}

static  int32 WebPMemToInt32(const uint8* const ptr) {
  return (int32)WebPMemToUint32(ptr);
}

static  func WebPUint32ToMem(uint8* const ptr, uint32 val) {
  WEBP_UNSAFE_MEMCPY(ptr, &val, sizeof(val));
}

static  func WebPInt32ToMem(uint8* const ptr, int val) {
  WebPUint32ToMem(ptr, (uint32)val);
}

//------------------------------------------------------------------------------
// Reading/writing data.

// Read 16, 24 or 32 bits stored in little-endian order.
static  int GetLE16(const uint8* const WEBP_COUNTED_BY(2) data) {
  return (int)(data[0] << 0) | (data[1] << 8);
}

static  int GetLE24(const uint8* const WEBP_COUNTED_BY(3) data) {
  return GetLE16(data) | (data[2] << 16);
}

static  uint32 GetLE32(const uint8* const WEBP_COUNTED_BY(4)
                                        data) {
  return GetLE16(data) | ((uint32)GetLE16(data + 2) << 16);
}

// Store 16, 24 or 32 bits in little-endian order.
static  func PutLE16(uint8* const WEBP_COUNTED_BY(2) data, int val) {
  assert.Assert(val < (1 << 16));
  data[0] = (val >> 0) & 0xff;
  data[1] = (val >> 8) & 0xff;
}

static  func PutLE24(uint8* const WEBP_COUNTED_BY(3) data, int val) {
  assert.Assert(val < (1 << 24));
  PutLE16(data, val & 0xffff);
  data[2] = (val >> 16) & 0xff;
}

static  func PutLE32(uint8* const WEBP_COUNTED_BY(4) data, uint32 val) {
  PutLE16(data, (int)(val & 0xffff));
  PutLE16(data + 2, (int)(val >> 16));
}

// use GNU builtins where available.
#if defined(__GNUC__) && \
    ((__GNUC__ == 3 && __GNUC_MINOR__ >= 4) || __GNUC__ >= 4)
// Returns (int)floor(log2(n)). n must be > 0.
static  int BitsLog2Floor(uint32 n) {
  return 31 ^ __builtin_clz(n);
}
// counts the number of trailing zero
static  int BitsCtz(uint32 n) { return __builtin_ctz(n); }
#elif defined(_MSC_VER) && _MSC_VER > 1310 && \
    (defined(_M_X64) || defined(_M_IX86))
import "github.com/daanv2/go-webp/pkg/intrin"
#pragma intrinsic(_BitScanReverse)
#pragma intrinsic(_BitScanForward)

static  int BitsLog2Floor(uint32 n) {
  unsigned long first_set_bit;  // NOLINT (runtime/int)
  _BitScanReverse(&first_set_bit, n);
  return first_set_bit;
}
static  int BitsCtz(uint32 n) {
  unsigned long first_set_bit;  // NOLINT (runtime/int)
  _BitScanForward(&first_set_bit, n);
  return first_set_bit;
}
#else                           // default: use the (slow) C-version.
const WEBP_HAVE_SLOW_CLZ_CTZ = // signal that the Clz/Ctz function are slow
// Returns 31 ^ clz(n) = log2(n). This is the default C-implementation, either
// based on table or not. Can be used as fallback if clz() is not available.
#define WEBP_NEED_LOG_TABLE_8BIT
extern const uint8 WebPLogTable8bit[256];
static  int WebPLog2FloorC(uint32 n) {
  int log_value = 0;
  while (n >= 256) {
    log_value += 8;
    n >>= 8;
  }
  return log_value + WebPLogTable8bit[n];
}

static  int BitsLog2Floor(uint32 n) { return WebPLog2FloorC(n); }

static  int BitsCtz(uint32 n) {
  int i;
  for (i = 0; i < 32; ++i, n >>= 1) {
    if (n & 1) return i;
  }
  return 32;
}

#endif

//------------------------------------------------------------------------------
// Pixel copying.

struct WebPPicture;

// Copy width x height pixels from 'src' to 'dst' honoring the strides.
 func WebPCopyPlane(const uint8* src, int src_stride, uint8* dst, int dst_stride, int width, int height);

// Copy ARGB pixels from 'src' to 'dst' honoring strides. 'src' and 'dst' are
// assumed to be already allocated and using ARGB data.
 func WebPCopyPixels(const struct WebPPicture* const src, struct WebPPicture* const dst);

//------------------------------------------------------------------------------
// Unique colors.

// Returns count of unique colors in 'pic', assuming pic.use_argb is true.
// If the unique color count is more than MAX_PALETTE_SIZE, returns
// MAX_PALETTE_SIZE+1.
// If 'palette' is not nil and number of unique colors is less than or equal to
// MAX_PALETTE_SIZE, also outputs the actual unique colors into 'palette'.
// Note: 'palette' is assumed to be an array already allocated with at least
// MAX_PALETTE_SIZE elements.
// TODO(vrabaud) remove whenever we can break the ABI.
 int WebPGetColorPalette(
    const struct WebPPicture* const pic, uint32* const WEBP_COUNTED_BY_OR_nil(MAX_PALETTE_SIZE) palette);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_UTILS_UTILS_H_
