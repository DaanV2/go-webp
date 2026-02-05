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
