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
