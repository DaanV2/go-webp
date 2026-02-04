// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package decoder

import "github.com/daanv2/go-webp/pkg/constants"

const (
	CHUNK_HEADER_SIZE = constants.CHUNK_HEADER_SIZE
	MAX_CHUNK_PAYLOAD = constants.MAX_CHUNK_PAYLOAD
	RIFF_HEADER_SIZE  = constants.RIFF_HEADER_SIZE
	TAG_SIZE          = constants.TAG_SIZE

	// minimal amp that will provide a non-zero dithering effect
	MIN_DITHER_AMP      = 4
	DITHER_AMP_TAB_SIZE = 12

	MT_CACHE_LINES =3
	ST_CACHE_LINES =1  // 1 cache row only for single-threaded case
)

// kFilterExtraRows[] = How many extra lines are needed on the MB boundary
// for caching, given a filtering level.
// Simple filter:  up to 2 luma samples are read and 1 is written.
// Complex filter: up to 4 luma samples are read and 3 are written. Same for
//                 U/V, so it's 8 samples total (because of the 2x upsampling).
var kFilterExtraRows = [3]int{0, 2, 8}

var kScan = [16]uint16{
	0 + 0*constants.BPS, 4 + 0*constants.BPS, 8 + 0*constants.BPS, 12 + 0*constants.BPS, 0 + 4*constants.BPS, 4 + 4*constants.BPS, 8 + 4*constants.BPS, 12 + 4*constants.BPS, 0 + 8*constants.BPS, 4 + 8*constants.BPS, 8 + 8*constants.BPS, 12 + 8*constants.BPS, 0 + 12*constants.BPS, 4 + 12*constants.BPS, 8 + 12*constants.BPS, 12 + 12*constants.BPS}

var kQuantToDitherAmp = [DITHER_AMP_TAB_SIZE]uint8{
	// roughly, it's dqm.uv_mat[1]
	8, 7, 6, 4, 4, 2, 2, 2, 1, 1, 1, 1}
