// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import "github.com/daanv2/go-webp/pkg/assert"

const VP8_RANDOM_DITHER_FIX = 8 // fixed-point precision for dithering
const VP8_RANDOM_TABLE_SIZE = 55

// 31b-range values
var kRandomTable = [VP8_RANDOM_TABLE_SIZE]uint32{
	0x0de15230, 0x03b31886, 0x775faccb, 0x1c88626a,
	0x68385c55, 0x14b3b828, 0x4a85fef8, 0x49ddb84b,
	0x64fcf397, 0x5c550289, 0x4a290000, 0x0d7ec1da,
	0x5940b7ab, 0x5492577d, 0x4e19ca72, 0x38d38c69,
	0x0c01ee65, 0x32a1755f, 0x5437f652, 0x5abb2c32,
	0x0faa57b1, 0x73f533e7, 0x685feeda, 0x7563cce2,
	0x6e990e83, 0x4730a7ed, 0x4fc0d9c6, 0x496b153c,
	0x4f1403fa, 0x541afb0c, 0x73990b32, 0x26d7cb1c,
	0x6fcc3706, 0x2cbb77d8, 0x75762f2a, 0x6425ccdd,
	0x24b35461, 0x0a7d8715, 0x220414a8, 0x141ebf67,
	0x56b41583, 0x73e502e3, 0x44cab16f, 0x28264d42,
	0x73baaefb, 0x0a50ebed, 0x1d6ab6fb, 0x0d3ad40b,
	0x35db3b68, 0x2b081e83, 0x77ce6b95, 0x5181e5f0,
	0x78853bbc, 0x009f9494, 0x27e5ed3c,
}

type VP8Random struct {
	index1, index2 int
	tab            [VP8_RANDOM_TABLE_SIZE]uint32
	amp            int
}

// VP8InitRandom initializes random generator with an amplitude 'dithering' in range [0..1].
func VP8InitRandom( /* const */ rg *VP8Random, dithering float64) {
	copy(rg.tab, kRandomTable)

	rg.index1 = 0
	rg.index2 = 31

	if dithering < 0.0 {
		rg.amp = 0
	} else if dithering > 1.0 {
		rg.amp = (1 << VP8_RANDOM_DITHER_FIX)
	} else {
		rg.amp = uint32((1 << VP8_RANDOM_DITHER_FIX) * dithering)
	}
}

// Returns a centered pseudo-random number with 'num_bits' amplitude.
// (uses D.Knuth's Difference-based random generator).
// 'amp' is in VP8_RANDOM_DITHER_FIX fixed-point precision.
func VP8RandomBits2( /* const  */ rg *VP8Random, num_bits, amp int) int {
	var diff int
	assert.Assert(num_bits+VP8_RANDOM_DITHER_FIX <= 31)

	diff = rg.tab[rg.index1] - rg.tab[rg.index2]
	if diff < 0 {
		diff += (uint(1) << 31)
	}
	rg.tab[rg.index1] = diff
	if rg.index1 == VP8_RANDOM_TABLE_SIZE {
		rg.index1 = 0
	} else {
		rg.index1++
	}
	if rg.index2 == VP8_RANDOM_TABLE_SIZE {
		rg.index2 = 0
	} else {
		rg.index2++
	}

	// sign-extend, 0-center
	diff = (diff << 1) >> (32 - num_bits)
	diff = (diff * amp) >> VP8_RANDOM_DITHER_FIX // restrict range
	diff += 1 << (num_bits - 1)                  // shift back to 0.5-center
	return diff
}

func VP8RandomBits( /* const  */ rg *VP8Random, num_bits int) int {
	return VP8RandomBits2(rg, num_bits, rg.amp)
}
