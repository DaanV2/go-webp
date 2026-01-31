package webp

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//  Common types + memory wrappers
//
// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/stddef"  // IWYU pragma: export for size_t



// Macro to check ABI compatibility (same major revision number)
func WEBP_ABI_IS_INCOMPATIBLE(a, b int) int {
	return (((a) >> 8) != ((b) >> 8))
}


// Allocates 'size' bytes of memory. Returns NULL upon error. Memory
// must be deallocated by calling WebPFree(). This function is made available
// by the core 'libwebp' library.
// Deprecated: Not needed in Go, use built-in memory management.
func WebPMalloc(size size_t) {
	panic("not implemented")
}

// Releases memory returned by the WebPDecode*() functions (from decode.h).
// Deprecated: Not needed in Go, use built-in memory management.
func WebPFree(void* ptr) {
	panic("not implemented")
}


import "github.com/daanv2/go-webp/pkg/string"  // For memcpy and friends

#ifdef WEBP_SUPPORT_FBOUNDS_SAFETY

// As explained in src/utils/bounds_safety.h, the below macros are defined
// somewhat delicately to handle a three-state setup:
//
// State 1: No -fbounds-safety enabled anywhere, all macros below should act
//          as-if -fbounds-safety doesn't exist.
// State 2: A file with -fbounds-safety enabled calling into files with or
//          without -fbounds-safety.
// State 3: A file without -fbounds-safety enabled calling into files with
//          -fbounds-safety. ABI breaking annotations must stay to force a
//          build failure and force us to use non-ABI breaking annotations.
//
// Currently, we only allow non-ABI changing annotations in this file to ensure
// we don't accidentally change the ABI for public functions.

import "github.com/daanv2/go-webp/pkg/ptrcheck"

const WEBP_ASSUME_UNSAFE_INDEXABLE_ABI =\
  __ptrcheck_abi_assume_unsafe_indexable()

#define WEBP_COUNTED_BY(x) __counted_by(x)
#define WEBP_COUNTED_BY_OR_NULL(x) __counted_by_or_null(x)
#define WEBP_SIZED_BY(x) __sized_by(x)
#define WEBP_SIZED_BY_OR_NULL(x) __sized_by_or_null(x)
#define WEBP_ENDED_BY(x) __ended_by(x)

const WEBP_UNSAFE_INDEXABLE =__unsafe_indexable
const WEBP_SINGLE =__single

#define WEBP_UNSAFE_FORGE_SINGLE(typ, ptr) __unsafe_forge_single(typ, ptr)

#define WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(typ, ptr, size) \
  __unsafe_forge_bidi_indexable(typ, ptr, size)

// Provide memcpy/memset/memmove wrappers to make migration easier.
#define WEBP_UNSAFE_MEMCPY(dst, src, size)                               \
  do {                                                                   \
    memcpy(WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, dst, size),        \
           WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, src, size), size); \
  } while (0)

#define WEBP_UNSAFE_MEMSET(dst, c, size)                                    \
  do {                                                                      \
    memset(WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, dst, size), c, size); \
  } while (0)

#define WEBP_UNSAFE_MEMMOVE(dst, src, size)                               \
  do {                                                                    \
    memmove(WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, dst, size),        \
            WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, src, size), size); \
  } while (0)

#define WEBP_UNSAFE_MEMCMP(s1, s2, size)                       \
  memcmp(WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, s1, size), \
         WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(uint8*, s2, size), size)

#else  // WEBP_SUPPORT_FBOUNDS_SAFETY

#define WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#define WEBP_COUNTED_BY(x)
#define WEBP_COUNTED_BY_OR_NULL(x)
#define WEBP_SIZED_BY(x)
#define WEBP_SIZED_BY_OR_NULL(x)
#define WEBP_ENDED_BY(x)

#define WEBP_UNSAFE_INDEXABLE
#define WEBP_SINGLE

#define WEBP_UNSAFE_MEMCPY(dst, src, size) memcpy(dst, src, size)
#define WEBP_UNSAFE_MEMSET(dst, c, size) memset(dst, c, size)
#define WEBP_UNSAFE_MEMMOVE(dst, src, size) memmove(dst, src, size)
#define WEBP_UNSAFE_MEMCMP(s1, s2, size) memcmp(s1, s2, size)

#define WEBP_UNSAFE_FORGE_SINGLE(typ, ptr) ((typ)(ptr))
#define WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(typ, ptr, size) ((typ)(ptr))

#endif  // WEBP_SUPPORT_FBOUNDS_SAFETY

// This macro exists to indicate intentionality with self-assignments and
// silence -Wself-assign compiler warnings.
#define WEBP_SELF_ASSIGN(x) x = x

#endif  // WEBP_WEBP_TYPES_H_
