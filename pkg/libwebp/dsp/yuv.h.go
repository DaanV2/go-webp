package dsp

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

import (
	"github.com/daanv2/go-webp/pkg/color/yuv"
	"github.com/daanv2/go-webp/pkg/constants"
)

// Macros to give the offset of each channel in a uint32 containing ARGB.
func CHANNEL_OFFSET(i int) int {
	if constants.WORDS_BIGENDIAN {
		// uint32 0xff000000 is 0xff,00,00,00 in memory
		return i
	}

	// uint32 0xff000000 is 0x00,00,00,ff in memory
	return 3 - i
}
//------------------------------------------------------------------------------
// slower on x86 by ~7-8%, but bit-exact with the SSE2/NEON version

//go:fix inline
func VP8YUVToR(y int, v int) int {
	return yuv.YUVToR(y, v)
}

//go:fix inline
func VP8YUVToG(y int, u int, v int) int {
	return yuv.YUVToG(y, u, v)
}

//go:fix inline
func VP8YUVToB(y int, u int) int {
	return yuv.YUVToB(y, v)
}
