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

import (
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/gcc"
)

func HToLE32(x int) int {
	if constants.WORDS_BIGENDIAN {
		return BSwap32(x)
	} else {
		return x
	}
}
func HToLE16(x int) int {
	if constants.WORDS_BIGENDIAN {
		return BSwap16(x)
	} else {
		return x
	}
}

func BSwap16(x uint16) uint16 {
	return gcc.Builtin_bswap16(x)
}

func BSwap32(x uint32) uint32 {
	return gcc.Builtin_bswap32(x)
}

func BSwap64(x uint64) uint64 {
	return gcc.Builtin_bswap64(x)
}