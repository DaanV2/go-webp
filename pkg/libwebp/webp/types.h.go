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

// IWYU pragma: export for uint64

// For memcpy and friends

// Macro to check ABI compatibility (same major revision number)
func WEBP_ABI_IS_INCOMPATIBLE(a, b int) bool {
	return (((a) >> 8) != ((b) >> 8))
}


// Allocates 'size' bytes of memory. Returns nil upon error. Memory
// must be deallocated by calling WebPFree(). This function is made available
// by the core 'libwebp' library.
// Deprecated: Not needed in Go, use built-in memory management.
func WebPMalloc(size uint64) {
	panic("not implemented")
}

// Releases memory returned by the *WebPDecode() functions (from decode.h).
// Deprecated: Not needed in Go, use built-in memory management.
func WebPFree(ptr any) {
	panic("not implemented")
}


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

 // Deprecated: Does nothing.
const WEBP_ASSUME_UNSAFE_INDEXABLE_ABI = 0

// Deprecated: Use golang std code
const WEBP_UNSAFE_INDEXABLE = 0

// Deprecated: Use golang std code
const WEBP_SINGLE = 0


// Deprecated: Use golang std code
func WEBP_COUNTED_BY(x any) {}
// Deprecated: Use golang std code
func WEBP_COUNTED_BY_OR_nil(x any) {}
// Deprecated: Use golang std code
func WEBP_SIZED_BY(x any) {}
// Deprecated: Use golang std code
func WEBP_SIZED_BY_OR_nil(x any) {}
// Deprecated: Use golang std code
func WEBP_ENDED_BY(x any) {}

// Deprecated: Use golang std code
func WEBP_UNSAFE_MEMCPY(dst, src, size any) {}

// Deprecated: Use golang std code
func WEBP_UNSAFE_MEMSET(dst, c, size any) {/* memset(dst, c, size) */}

// Deprecated: Use golang std code
func  WEBP_UNSAFE_MEMMOVE(dst, src, size any) {/* memmove(dst, src, size) */}

// Deprecated: Use golang std code
func WEBP_UNSAFE_MEMCMP(s1, s2, size any) {/* memcmp(s1, s2, size) */}

// Deprecated: Use golang std code
func WEBP_UNSAFE_FORGE_SINGLE(typ, ptr any) {/* ((typ)(ptr)) */}

// Deprecated: Use golang std code
func WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(typ, ptr, size any) {/* ((typ)(ptr)) */}

// This macro exists to indicate intentionality with self-assignments and
// silence -Wself-assign compiler warnings.
// Deprecated: Not needed in Go.
func WEBP_SELF_ASSIGN(x any) {/*x = x*/}
