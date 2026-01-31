// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import "github.com/daanv2/go-webp/pkg/constants"

type VP8Decoder struct {
	status    VP8StatusCode
	ready     int   // true if ready to decode a picture with VP8Decode()
	error_msg *byte // set when status is not OK.

	// Main data source
	br          VP8BitReader
	incremental int // if true, incremental decoding is expected

	// headers
	frm_hdr     VP8FrameHeader
	pic_hdr     VP8PictureHeader
	filter_hdr  VP8FilterHeader
	segment_hdr VP8SegmentHeader

	// Worker
	worker WebPWorker
	// multi-thread method: 0=off, 1=[parse+recon][filter] 2=[parse][recon+filter]
	mt_method  int
	cache_id   int              // current cache row
	num_caches int              // number of cached rows of 16 pixels (1, 2 or 3)
	thread_ctx VP8ThreadContext // Thread context

	// dimension, in macroblock units.
	mb_w, mb_h int

	// Macroblock to process/filter, depending on cropping and filter_type.
	tl_mb_x, tl_mb_y int // top-left MB that must be in-loop filtered
	br_mb_x, br_mb_y int // last bottom-right MB that must be decoded

	// number of partitions minus one.
	num_parts_minus_one uint32
	// per-partition boolean decoders.
	parts [MAX_NUM_PARTITIONS]VP8BitReader

	// Dithering strength, deduced from decoding options
	dither       int       // whether to use dithering or not
	dithering_rg VP8Random // random generator for dithering

	// dequantization (one set of DC/AC dequant factor per segment)
	dqm [NUM_MB_SEGMENTS]VP8QuantMatrix

	// probabilities
	proba          VP8Proba
	use_skip_proba int
	skip_p         uint8

	// Boundary data cache and persistent buffers.
	intra_t *uint8   // top intra modes values: 4 * mb_w
	intra_l [4]uint8 // left intra modes values

	yuv_t *VP8TopSamples // top y/u/v samples

	mb_info *VP8MB    // contextual macroblock info (mb_w + 1)
	f_info  *VP8FInfo // filter strength info
	yuv_b   *uint8    // main block for Y/U/V (size = YUV_SIZE)

	cache_y         *uint8 // macroblock row for storing unfiltered samples
	cache_u         *uint8
	cache_v         *uint8
	cache_y_stride  int
	cache_uv_stride int

	// main memory chunk for the above data. Persistent.
	mem      *void
	mem_size uint64

	// Per macroblock non-persistent infos.
	mb_x, mb_y int        // current position, in macroblock units
	mb_data    *VP8MBData // parsed reconstruction data

	// Filtering side-info
	filter_type int                          // 0=off, 1=simple, 2=complex
	fstrengths  [NUM_MB_SEGMENTS][2]VP8FInfo // precalculated per-segment/type

	// Alpha
	alph_dec         *ALPHDecoder // alpha-plane decoder object
	alpha_data       *uint8       // compressed alpha data (if present)
	alpha_data_size  uint64
	is_alpha_decoded int    // true if alpha_data is decoded in alpha_plane
	alpha_plane_mem  *uint8 // memory allocated for alpha_plane
	alpha_plane      *uint8 // output. Persistent, contains the whole data.
	alpha_prev_line  *uint8 // last decoded alpha row (or nil)
	alpha_dithering  int    // derived from decoding options (0=off, 100=full)
}

type VP8LDecoder struct {
	status VP8StatusCode
	state  VP8LDecodeState
	io     *VP8Io

	output *WebPDecBuffer // shortcut to io.opaque.output

	pixels *uint32 // Internal data: either for alpha *uint8
	// or for BGRA *uint32.
	argb_cache             *uint32 // Scratch buffer for temporary BGRA storage.
	accumulated_rgb_pixels *uint16 // Scratch buffer for accumulated RGB for
	// YUV conversion.

	br               VP8LBitReader
	incremental      int           // if true, incremental decoding is expected
	saved_br         VP8LBitReader // note: could be local variables too
	saved_last_pixel int

	width    int
	height   int
	last_row int // last input row decoded so far.
	// last pixel decoded so far. However, it may
	// not be transformed, scaled and
	// color-converted yet.
	last_pixel   int
	last_out_row int // last row output so far.

	hdr VP8LMetadata

	next_transform int
	transforms     [constants.NUM_TRANSFORMS]VP8LTransform
	// or'd bitset storing the transforms types.
	transforms_seen uint32

	rescaler_memory *uint8        // Working memory for rescaling work.
	rescaler        *WebPRescaler // Common rescaler for all channels.
}