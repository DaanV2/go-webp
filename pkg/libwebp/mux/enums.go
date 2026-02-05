// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

// CHUNK_INDEX enum: used for indexing within 'kChunks' (defined below) only.
// Note: the reason for having two enums ('WebPChunkId' and 'CHUNK_INDEX') is to
// allow two different chunks to have the same id (e.g. WebPChunkId
// 'WEBP_CHUNK_IMAGE' can correspond to CHUNK_INDEX 'IDX_VP8' or 'IDX_VP8L').
type CHUNK_INDEX int

const (
	IDX_VP8X CHUNK_INDEX = iota
	IDX_ICCP
	IDX_ANIM
	IDX_ANMF
	IDX_ALPHA
	IDX_VP8
	IDX_VP8L
	IDX_EXIF
	IDX_XMP
	IDX_UNKNOWN
	IDX_NIL 
	IDX_LAST_CHUNK
)