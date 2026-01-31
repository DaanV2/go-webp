package utils

// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Endian related functions.


#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#if defined(WORDS_BIGENDIAN)
const HToLE32 = BSwap32
const HToLE16 = BSwap16
#else
#define HToLE32(x) (x)
#define HToLE16(x) (x)
#endif

#if !defined(HAVE_CONFIG_H)
#if LOCAL_GCC_PREREQ(4, 8) || __has_builtin(__builtin_bswap16)
#define HAVE_BUILTIN_BSWAP16
#endif
#if LOCAL_GCC_PREREQ(4, 3) || __has_builtin(__builtin_bswap32)
#define HAVE_BUILTIN_BSWAP32
#endif
#if LOCAL_GCC_PREREQ(4, 3) || __has_builtin(__builtin_bswap64)
#define HAVE_BUILTIN_BSWAP64
#endif
#endif  // !HAVE_CONFIG_H

static  uint16 BSwap16(uint16 x) {
#if defined(HAVE_BUILTIN_BSWAP16)
  return __builtin_bswap16(x);
#elif defined(_MSC_VER)
  return _byteswap_ushort(x);
#else
  // gcc will recognize a 'rorw $8, ...' here:
  return (x >> 8) | ((x & 0xff) << 8);
#endif  // HAVE_BUILTIN_BSWAP16
}

static  uint32 BSwap32(uint32 x) {
#if defined(WEBP_USE_MIPS32_R2)
  uint32 ret;
  __asm__ volatile(
      "wsbh   %[ret], %[x]          \n\t"
      "rotr   %[ret], %[ret],  16   \n\t"
      : [ret] "=r"(ret)
      : [x] "r"(x));
  return ret;
#elif defined(HAVE_BUILTIN_BSWAP32)
  return __builtin_bswap32(x);
#elif defined(__i386__) || defined(__x86_64__)
  uint32 swapped_bytes;
  __asm__ volatile("bswap %0" : "=r"(swapped_bytes) : "0"(x));
  return swapped_bytes;
#elif defined(_MSC_VER)
  return (uint32)_byteswap_ulong(x);
#else
  return (x >> 24) | ((x >> 8) & 0xff00) | ((x << 8) & 0xff0000) | (x << 24);
#endif  // HAVE_BUILTIN_BSWAP32
}

static  uint64 BSwap64(uint64 x) {
#if defined(HAVE_BUILTIN_BSWAP64)
  return __builtin_bswap64(x);
#elif defined(__x86_64__)
  uint64 swapped_bytes;
  __asm__ volatile("bswapq %0" : "=r"(swapped_bytes) : "0"(x));
  return swapped_bytes;
#elif defined(_MSC_VER)
  return (uint64)_byteswap_uint64(x);
#else   // generic code for swapping 64-bit values (suggested by bdb@)
  x = ((x & 0xffffffff00000000ull) >> 32) | ((x & 0x00000000ffffffffull) << 32);
  x = ((x & 0xffff0000ffff0000ull) >> 16) | ((x & 0x0000ffff0000ffffull) << 16);
  x = ((x & 0xff00ff00ff00ff00ull) >> 8) | ((x & 0x00ff00ff00ff00ffull) << 8);
  return x;
#endif  // HAVE_BUILTIN_BSWAP64
}

#endif  // WEBP_UTILS_ENDIAN_INL_UTILS_H_
