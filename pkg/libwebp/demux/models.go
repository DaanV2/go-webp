// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package demux

type MemBuffer struct {
	start    uint64 // start location of the data
	end      uint64 // end location
	riff_end uint64 // riff chunk end location, can be > end.
	buf_size uint64 // size of the buffer
	buf      *uint8
}

type ChunkData struct {
	offset uint64
	size   uint64
}

type Frame struct {
	x_offset, y_offset int
	width, height      int
	has_alpha          int
	duration           int
	dispose_method     WebPMuxAnimDispose
	blend_method       WebPMuxAnimBlend
	frame_num          int
	complete           int          // img_components contains a full image
	img_components     [2]ChunkData // 0=VP8{,L} 1=ALPH
	next               *Frame
}

type Chunk struct {
	data ChunkData
	next *Chunk
}

type WebPDemuxer struct {
	canvas_width, canvas_height int
	mem MemBuffer
	state WebPDemuxState
	is_ext_format int
	feature_flags uint32
	loop_count int
	bgcolor uint32
	num_frames int
	frames *Frame
	frames_tail *Frame
	chunks *Chunk  // non-image chunks
	chunks_tail *Chunk
}

type ChunkParser struct {
  id [4]uint8
  ParseImage func(dmux *WebPDemuxer)ParseStatus
  Valid func(dmux *WebPDemuxer)int
}