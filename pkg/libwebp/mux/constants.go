// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

const (
	NIL_TAG = uint(0x00000000) // To signal func chunk.

	MUX_MAJ_VERSION = 1
	MUX_MIN_VERSION = 6
	MUX_REV_VERSION = 0

	UNDEFINED_CHUNK_SIZE = uint32(-1)
)

var kChunks [IDX_LAST_CHUNK]ChunkInfo


var kChunks []ChunkInfo{
    {MKFOURCC('V', 'P', '8', 'X'), WEBP_CHUNK_VP8X, VP8X_CHUNK_SIZE}, {MKFOURCC('I', 'C', 'C', 'P'), WEBP_CHUNK_ICCP, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('A', 'N', 'I', 'M'), WEBP_CHUNK_ANIM, ANIM_CHUNK_SIZE}, {MKFOURCC('A', 'N', 'M', 'F'), WEBP_CHUNK_ANMF, ANMF_CHUNK_SIZE}, {MKFOURCC('A', 'L', 'P', 'H'), WEBP_CHUNK_ALPHA, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('V', 'P', '8', ' '), WEBP_CHUNK_IMAGE, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('V', 'P', '8', 'L'), WEBP_CHUNK_IMAGE, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('E', 'X', 'I', 'F'), WEBP_CHUNK_EXIF, UNDEFINED_CHUNK_SIZE}, {MKFOURCC('X', 'M', 'P', ' '), WEBP_CHUNK_XMP, UNDEFINED_CHUNK_SIZE}, {NIL_TAG, WEBP_CHUNK_UNKNOWN, UNDEFINED_CHUNK_SIZE},

    {NIL_TAG, WEBP_CHUNK_NIL, UNDEFINED_CHUNK_SIZE}}

func WebPGetMuxVersion() int {
  return (MUX_MAJ_VERSION << 16) | (MUX_MIN_VERSION << 8) | MUX_REV_VERSION;
}