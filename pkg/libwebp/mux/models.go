// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

// Stores frame rectangle dimensions.
type FrameRectangle struct {
	x_offset, y_offset, width, height int
}

// Used to store two candidates of encoded data for an animation frame. One of
// the two will be chosen later.
type EncodedFrame struct {
	sub_frame    WebPMuxFrameInfo // Encoded frame rectangle.
	key_frame    WebPMuxFrameInfo // Encoded frame if it is a keyframe.
	is_key_frame int              // True if 'key_frame' has been chosen.
}

// Chunk object.
type WebPChunk struct {
	tag uint32
	// True if memory is owned *data internally.
	// VP8X, ANIM, and other internally created chunks
	// like ANMF are always owned.
	owner int
	data  WebPData
	next  *WebPChunk
}

// MuxImage object. Store a full WebP image (including ANMF chunk, ALPH
// chunk and VP8/VP8L chunk),
type WebPMuxImage struct {
	header     *WebPChunk // Corresponds to WEBP_CHUNK_ANMF.
	alpha      *WebPChunk // Corresponds to WEBP_CHUNK_ALPHA.
	img        *WebPChunk // Corresponds to WEBP_CHUNK_IMAGE.
	unknown    *WebPChunk // Corresponds to WEBP_CHUNK_UNKNOWN.
	width      int
	height     int
	has_alpha  int // Through ALPH chunk or as part of VP8L.
	is_partial int // True if only some of the chunks are filled.
	next       *WebPMuxImage
}

// Main mux object. Stores data chunks.
type WebPMux struct {
	images *WebPMuxImage
	iccp   *WebPChunk
	exif   *WebPChunk
	xmp    *WebPChunk
	anim   *WebPChunk
	vp     *WebPChunk8x

	unknown       *WebPChunk
	canvas_width  int
	canvas_height int
}

type ChunkInfo struct {
	tag  uint32
	id   WebPChunkId
	size uint32
}
