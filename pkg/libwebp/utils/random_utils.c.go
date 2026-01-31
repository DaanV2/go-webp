package utils

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Pseudo-random utilities
//
// Author: Skal (pascal.massimino@gmail.com)

// VP8InitRandom initializes random generator with an amplitude 'dithering' in range [0..1].
func VP8InitRandom(/* const */ rg *VP8Random, dithering float64) {
	copy(rg.tab, kRandomTable)

	rg.index1 = 0;
	rg.index2 = 31;

	if (dithering < 0.0) {
		rg.amp = 0;
	} else if (dithering > 1.0) {
		rg.amp = (1 << VP8_RANDOM_DITHER_FIX);
	} else {
		rg.amp = (uint32)((1 << VP8_RANDOM_DITHER_FIX) * dithering);
	}
}

