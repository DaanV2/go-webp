// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import "github.com/daanv2/go-webp/pkg/constants"

type VP8Encoder struct {
	config *WebPConfig;  // user configuration and parameters
	pic *WebPPicture;          // input / output picture

	// headers
	filter_hdr VP8EncFilterHeader    // filtering information
	segment_hdr VP8EncSegmentHeader  // segment information

	profile int  // VP8's profile, deduced from Config.

	// dimension, in macroblock units.
	mb_w, mb_h int
	preds_w int  // stride of the prediction plane *preds (=4*mb_w + 1)

	// number of partitions (1, 2, 4 or 8 = MAX_NUM_PARTITIONS)
	num_parts int

	// per-partition boolean decoders.
	bw VP8BitWriter                         // part0
	parts [MAX_NUM_PARTITIONS]VP8BitWriter  // token partitions
	tokens VP8TBuffer                       // token buffer

	percent int  // for progress

	// transparency blob
	has_alpha int 
	alpha_data *uint8;  // non-nil if transparency is present
	alpha_data_size uint32
	alpha_worker WebPWorker 

	// quantization info (one set of DC/AC dequant factor per segment)
	dqm [NUM_MB_SEGMENTS]VP8SegmentInfo
	base_quant int  // nominal quantizer value. Only used
					// for relative coding of segments' quant.
	alpha int       // global susceptibility (<=> complexity)
	// U/V quantization susceptibility
	// global offset of quantizers, shared by all segments
	uv_alpha int
	dq_y1_dc int
	dq_y2_dc, dq_y2_ac int
	dq_uv_dc, dq_uv_ac int

	// probabilities and statistics
	proba VP8EncProba
	sse [4]uint64     // sum of Y/U/V/A squared errors for all macroblocks
	sse_count uint64  // pixel count for the sse[] stats
	coded_size int
	residual_bytes[3][4]int 
	block_count[3]int

	// quality/speed settings
	method int                // 0=fastest, 6=best/slowest.
	rd_opt_level VP8RDLevel   // Deduced from method.
	max_i4_header_bits int    // partition #0 safeness factor
	mb_header_limit int       // rough limit for header bits per MB
	thread_level int          // derived from config.thread_level
	do_search int             // derived from config.target_XXX
	use_tokens int            // if true, use token buffer

	// Memory
	mb_info *VP8MBInfo;  // contextual macroblock infos (mb_w + 1)
	preds *uint8;      // predictions modes: (4*mb_w+1) * (4*mb_h+1)
	nz *uint32;        // non-zero bit context: mb_w+1
	y_top *uint8;      // top luma samples.
	// top u/v samples.
	// U and V are packed into 16 bytes (8 U + 8 V)
	uv_top *uint8;     
	lf_stats *LFStats;   // autofilter stats (if nil, autofilter is off)
	top_derr *DError;    // diffusion error (nil if disabled)
}

type VP8LEncoder struct {
	config *WebPConfig  // user configuration and parameters
	pic    *WebPPicture // input picture.

	argb         *uint32                // Transformed argb image data.
	argb_content VP8LEncoderARGBContent // Content type of the argb buffer.
	// Scratch memory for argb rows
	// (used for prediction).
	argb_scratch       *uint32
	transform_data     *uint32 // Scratch memory for transform data.
	transform_mem      *uint32 // Currently allocated memory.
	transform_mem_size uint64  // Currently allocated memory size.

	current_width int // Corresponds to packed image width.

	// Encoding parameters derived from quality parameter.
	histo_bits                 int
	predictor_transform_bits   int // <= MAX_TRANSFORM_BITS
	cross_color_transform_bits int // <= MAX_TRANSFORM_BITS
	cache_bits                 int // If equal to 0, don't use color cache.

	// Encoding parameters derived from image characteristics.
	use_cross_color    int
	use_subtract_green int
	use_predict        int
	use_palette        int
	palette_size       int
	palette            [constants.MAX_PALETTE_SIZE]uint32
	// Sorted version of palette for cache purposes.
	palette_sorted [constants.MAX_PALETTE_SIZE]uint32

	// Some 'scratch' (potentially large) objects.
	// Backward Refs array for temporaries.
	refs [4]VP8LBackwardRefs
	// HashChain data for constructing
	// backward references.
	hash_chain VP8LHashChain
}
