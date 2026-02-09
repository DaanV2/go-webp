// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

import (
	"math"

	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
)

const (
	NIL_TAG = 0x00000000 // To signal func chunk.

	MUX_MAJ_VERSION = 1
	MUX_MIN_VERSION = 6
	MUX_REV_VERSION = 0

	UNDEFINED_CHUNK_SIZE = math.MaxUint32

	ERROR_STR_MAX_LENGTH = 100
	DELTA_INFINITY       = (uint64(1) << 32)
	KEYFRAME_NONE        = (-1)
	MAX_CACHED_FRAMES    = 30
	// This value is used to match a later call to WebPReplaceTransparentPixels(),
	// making it a no-op for lossless (see WebPEncode()).
	TRANSPARENT_COLOR = 0x00000000
)

var kChunks = [IDX_LAST_CHUNK]ChunkInfo{
	{tag: constants.MKFOURCC('V', 'P', '8', 'X'), id: webp.WEBP_CHUNK_VP8X, size: constants.VP8X_CHUNK_SIZE},
	{tag: constants.MKFOURCC('I', 'C', 'C', 'P'), id: webp.WEBP_CHUNK_ICCP, size: UNDEFINED_CHUNK_SIZE},
	{tag: constants.MKFOURCC('A', 'N', 'I', 'M'), id: webp.WEBP_CHUNK_ANIM, size: constants.ANIM_CHUNK_SIZE},
	{tag: constants.MKFOURCC('A', 'N', 'M', 'F'), id: webp.WEBP_CHUNK_ANMF, size: constants.ANMF_CHUNK_SIZE},
	{tag: constants.MKFOURCC('A', 'L', 'P', 'H'), id: webp.WEBP_CHUNK_ALPHA, size: UNDEFINED_CHUNK_SIZE},
	{tag: constants.MKFOURCC('V', 'P', '8', ' '), id: webp.WEBP_CHUNK_IMAGE, size: UNDEFINED_CHUNK_SIZE},
	{tag: constants.MKFOURCC('V', 'P', '8', 'L'), id: webp.WEBP_CHUNK_IMAGE, size: UNDEFINED_CHUNK_SIZE},
	{tag: constants.MKFOURCC('E', 'X', 'I', 'F'), id: webp.WEBP_CHUNK_EXIF, size: UNDEFINED_CHUNK_SIZE},
	{tag: constants.MKFOURCC('X', 'M', 'P', ' '), id: webp.WEBP_CHUNK_XMP, size: UNDEFINED_CHUNK_SIZE},
	{tag: NIL_TAG, id: webp.WEBP_CHUNK_UNKNOWN, size: UNDEFINED_CHUNK_SIZE},
	{tag: NIL_TAG, id: webp.WEBP_CHUNK_NIL, size: UNDEFINED_CHUNK_SIZE},
}

func WebPGetMuxVersion() int {
	return (MUX_MAJ_VERSION << 16) | (MUX_MIN_VERSION << 8) | MUX_REV_VERSION
}
