// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

type VP8FrameHeader struct {
	key_frame        uint8
	profile          uint8
	show             uint8
	partition_length uint32
}

type VP8PictureHeader struct {
	width      uint16
	height     uint16
	xscale     uint8
	yscale     uint8
	colorspace uint8 // 0 = YCbCr
	clamp_type uint8
}

// segment features
type VP8SegmentHeader struct {
	use_segment     int
	update_map      int                   // whether to update the segment map or not
	absolute_delta  int                   // absolute or delta values for quantizer and filter
	quantizer       [NUM_MB_SEGMENTS]int8 // quantization changes
	filter_strength [NUM_MB_SEGMENTS]int8 // filter strength for segments
}

// probas associated to one of the contexts
type VP8ProbaArray [NUM_PROBAS]uint8

type VP8BandProbas struct { // all the probas associated to one band
	probas [NUM_CTX]VP8ProbaArray
}

// Struct collecting all frame-persistent probabilities.
type VP8Proba struct {
	segments [MB_FEATURE_TREE_PROBS]uint8
	// Type: 0:Intra16-AC  1:Intra16-DC   2:Chroma   3:Intra4
	bands     [NUM_TYPES][NUM_BANDS]VP8BandProbas
	bands_ptr [NUM_TYPES][16 + 1]*VP8BandProbas
}

// Filter parameters
type VP8FilterHeader struct {
	simple        int // 0=complex, 1=simple
	level         int // [0..63]
	sharpness     int // [0..7]
	use_lf_delta  int
	ref_lf_delta  [NUM_REF_LF_DELTAS]int
	mode_lf_delta [NUM_MODE_LF_DELTAS]int
}

//------------------------------------------------------------------------------
// Informations about the macroblocks.

type VP8FInfo struct { // filter specs
	f_limit    uint8 // filter limit in [3..189], or 0 if no filtering
	f_ilevel   uint8 // inner limit in [1..63]
	f_inner    uint8 // do inner filtering?
	hev_thresh uint8 // high edge variance threshold in [0..2]
}

type VP8MB struct { // Top/Left Contexts used for syntax-parsing
	nz    uint8 // non-zero AC/DC coeffs (4bit for luma + 4bit for chroma)
	nz_dc uint8 // non-zero DC coeff (1bit)
}

// Dequantization matrices
type quant_t [2]int // [DC / AC].  Can be 'uint16[2]' too (~slower).

type VP8QuantMatrix struct {
	y1_mat, y2_mat, uv_mat quant_t

	uv_quant int // U/V quantizer value
	dither   int // dithering amplitude (0 = off, max=255)
}

// Data needed to reconstruct a macroblock
type VP8MBData struct {
	coeffs  [384]int16 // 384 coeffs = (16+4+4) * 4*4
	is_i4x4 uint8      // true if intra4x4
	imodes  [16]uint8  // one 16x16 mode (#0) or sixteen 4x4 modes
	// chroma prediction mode
	// bit-wise info about the content of each sub-4x4 blocks (in decoding order).
	// Each of the 4x4 blocks for y/u/v is associated with a 2b code according to:
	//   code=0 . no coefficient
	//   code=1 . only DC
	//   code=2 . first three coefficients are non-zero
	//   code=3 . more than three coefficients are non-zero
	// This allows to call specialized transform functions.
	uvmode      uint8
	non_zero_y  uint32
	non_zero_uv uint32
	dither      uint8 // local dithering strength (deduced from non_*zero)
	skip        uint8
	segment     uint8
}

// Persistent information needed by the parallel processing
type VP8ThreadContext struct {
	id         int        // cache row to process (in [0..2])
	mb_y       int        // macroblock position of the row
	filter_row int        // true if row-filtering is needed
	f_info     *VP8FInfo  // filter strengths (swapped with dec.f_info)
	mb_data    *VP8MBData // reconstruction data (swapped with dec.mb_data)
	io         VP8Io      // copy of the VP8Io to pass to put()
}

// Saved top samples, per macroblock. Fits into a cache-line.
type VP8TopSamples struct {
	y [16]uint8
	u [8]uint8
	v [8]uint8
}

type VP8LTransform struct {
	vtype VP8LImageTransformType // transform type.
	bits  int                    // subsampling bits defining transform window.
	xsize int                    // transform window X index.
	ysize int                    // transform window Y index.
	data  *uint32                // transform data.
}

type VP8LMetadata struct {
	color_cache_size  int
	color_cache       VP8LColorCache
	saved_color_cache VP8LColorCache // for incremental

	huffman_mask           int
	huffman_subsample_bits int
	huffman_xsize          int
	huffman_image          *uint32
	num_htree_groups       int
	htree_groups           *HTreeGroup
	huffman_tables         HuffmanTables
}

// type used for scores, rate, distortion
// Note that MAX_COST is not the maximum allowed by sizeof(score_t),
// in order to allow overflowing computations.
type score_t int64  