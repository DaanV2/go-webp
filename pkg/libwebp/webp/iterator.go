// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package webp

type WebPIterator struct {
	frame_num          int
	num_frames         int                // equivalent to WEBP_FF_FRAME_COUNT.
	x_offset, y_offset int                // offset relative to the canvas.
	width, height      int                // dimensions of this frame.
	duration           int                // display duration in milliseconds.
	dispose_method     WebPMuxAnimDispose // dispose method for the frame.
	// true if 'fragment' contains a full frame. partial images
	// may still be decoded with the WebP incremental decoder.
	complete int
	// The frame given by 'frame_num'. Note for historical
	// reasons this is called a fragment.
	fragment     WebPData
	has_alpha    int              // True if the frame contains transparency.
	blend_method WebPMuxAnimBlend // Blend operation for the frame.

	pad      [2]uint32 // padding for later use.
	private_ *void     // for internal use only.
}

type WebPChunkIterator struct {
	// The current and total number of chunks with the fourcc given to
	// WebPDemuxGetChunk().
	chunk_num  int
	num_chunks int
	chunk      WebPData // The payload of the chunk.

	pad      [6]uint32 // padding for later use
	private_ *void
}
