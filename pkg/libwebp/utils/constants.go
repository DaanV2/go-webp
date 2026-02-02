package utils

// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

const (
	// This is the maximum memory amount that libwebp will ever try to allocate.
	WEBP_MAX_ALLOCABLE_MEMORY = (uint64(1) << 34)

	// For 32-bit targets keep this below INT_MAX to afunc valgrind warnings.
	// WEBP_MAX_ALLOCABLE_MEMORY =((uint64(1) << 31) - (1 << 16))

	WEBP_ALIGN_CST =31
)
