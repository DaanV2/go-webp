package dec

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// VP8 decoder: internal header.
//
// Author: Skal (pascal.massimino@gmail.com)

// for memcpy()

//------------------------------------------------------------------------------
// Various defines and enums

// version numbers
const DEC_MAJ_VERSION =1
const DEC_MIN_VERSION =6
const DEC_REV_VERSION =0

// YUV-cache parameters. Cache is 32-bytes wide (= one cacheline).
// Constraints are: We need to store one 16x16 block of luma samples (y),
// and two 8x8 chroma blocks (u/v). These are better be 16-bytes aligned,
// in order to be SIMD-friendly. We also need to store the top, left and
// top-left samples (from previously decoded blocks), along with four
// extra top-right samples for luma (intra4x4 prediction only).
// One possible layout is, using 32 * (17 + 9) bytes:
//
//   .+------   <- only 1 pixel high
//   .|yyyyt.
//   .|yyyyt.
//   .|yyyyt.
//   .|yyyy..
//   .+--.+--   <- only 1 pixel high
//   .|uu.|vv
//   .|uu.|vv
//
// Every character is a 4x4 block, with legend:
//  '.' = unused
//  'y' = y-samples   'u' = u-samples     'v' = u-samples
//  '|' = left sample,   '-' = top sample,    '+' = top-left sample
//  't' = extra top-right sample for 4x4 modes
const YUV_SIZE =(BPS * 17 + BPS * 9)
const Y_OFF =(BPS * 1 + 8)
const U_OFF =(Y_OFF + BPS * 16 + BPS)
const V_OFF =(U_OFF + 16)

// minimal width under which lossy multi-threading is always disabled
const MIN_WIDTH_FOR_THREADS =512

//------------------------------------------------------------------------------
// Headers

type VP8FrameHeader struct {
   key_frame uint8
   profile uint8
   show uint8
   partition_length uint32
}

type VP8PictureHeader struct {
   width uint16
   height uint16
   xscale uint8
   yscale uint8
   colorspace uint8  // 0 = YCbCr
   clamp_type uint8
}

// segment features
type VP8SegmentHeader struct {
  use_segment int
  update_map int      // whether to update the segment map or not
  absolute_delta int  // absolute or delta values for quantizer and filter
  quantizer [NUM_MB_SEGMENTS]int8        // quantization changes
  filter_strength [NUM_MB_SEGMENTS]int8  // filter strength for segments
} 

// probas associated to one of the contexts
type VP8ProbaArray [NUM_PROBAS]uint8

type VP8BandProbas struct {  // all the probas associated to one band
  VP8ProbaArray probas[NUM_CTX]
} 

// Struct collecting all frame-persistent probabilities.
type VP8Proba struct {
	segments [MB_FEATURE_TREE_PROBS]uint8
	// Type: 0:Intra16-AC  1:Intra16-DC   2:Chroma   3:Intra4
	bands [NUM_TYPES][NUM_BANDS]VP8BandProbas
	bands_ptr [NUM_TYPES][16 + 1]*VP8BandProbas
} 

// Filter parameters
type VP8FilterHeader struct {
   simple int     // 0=complex, 1=simple
   level int      // [0..63]
   sharpness int  // [0..7]
   use_lf_delta int
   ref_lf_delta [NUM_REF_LF_DELTAS]int
   mode_lf_delta [NUM_MODE_LF_DELTAS]int
} 

//------------------------------------------------------------------------------
// Informations about the macroblocks.

type VP8FInfo struct {       // filter specs
   f_limit uint8     // filter limit in [3..189], or 0 if no filtering
   f_ilevel uint8    // inner limit in [1..63]
   f_inner uint8     // do inner filtering?
   hev_thresh uint8  // high edge variance threshold in [0..2]
} 

type VP8MB struct {  // Top/Left Contexts used for syntax-parsing
   nz uint8     // non-zero AC/DC coeffs (4bit for luma + 4bit for chroma)
   nz_dc uint8  // non-zero DC coeff (1bit)
} 

// Dequantization matrices
type quant_t [2]int  // [DC / AC].  Can be 'uint16[2]' too (~slower).

type VP8QuantMatrix struct {
  y1_mat, y2_mat, uv_mat quant_t

   uv_quant int  // U/V quantizer value
   dither int    // dithering amplitude (0 = off, max=255)
} 

// Data needed to reconstruct a macroblock
type VP8MBData struct {
	coeffs [384]int16  // 384 coeffs = (16+4+4) * 4*4
	is_i4x4 uint8      // true if intra4x4
	imodes [16]uint8   // one 16x16 mode (#0) or sixteen 4x4 modes
	// chroma prediction mode
	// bit-wise info about the content of each sub-4x4 blocks (in decoding order).
	// Each of the 4x4 blocks for y/u/v is associated with a 2b code according to:
	//   code=0 . no coefficient
	//   code=1 . only DC
	//   code=2 . first three coefficients are non-zero
	//   code=3 . more than three coefficients are non-zero
	// This allows to call specialized transform functions.
	uvmode uint8
	non_zero_y uint32
	non_zero_uv uint32
	dither uint8  // local dithering strength (deduced from non_*zero)
	skip uint8
	segment uint8
} 

// Persistent information needed by the parallel processing
type VP8ThreadContext struct {
	id int              // cache row to process (in [0..2])
	mb_y int            // macroblock position of the row
	filter_row int      // true if row-filtering is needed
	f_info *VP8FInfo    // filter strengths (swapped with dec.f_info)
	mb_data *VP8MBData  // reconstruction data (swapped with dec.mb_data)
	io VP8Io            // copy of the VP8Io to pass to put()
} 

// Saved top samples, per macroblock. Fits into a cache-line.
type VP8TopSamples struct {
  y [16]uint8
  u [8]uint8
  v [8]uint8
} 

//------------------------------------------------------------------------------
// VP8Decoder: the main opaque structure handed over to user

type VP8Decoder struct {
  status VP8StatusCode
  ready int              // true if ready to decode a picture with VP8Decode()
  error_msg *byte  // set when status is not OK.

  // Main data source
   br VP8BitReader
   incremental int  // if true, incremental decoding is expected

  // headers
   frm_hdr VP8FrameHeader
   pic_hdr VP8PictureHeader
   filter_hdr VP8FilterHeader
   segment_hdr VP8SegmentHeader

  // Worker
   worker WebPWorker
      // multi-thread method: 0=off, 1=[parse+recon][filter] 2=[parse][recon+filter]
   mt_method int
   cache_id int    // current cache row
   num_caches int  // number of cached rows of 16 pixels (1, 2 or 3)
   thread_ctx VP8ThreadContext  // Thread context

  // dimension, in macroblock units.
   mb_w, mb_h int

  // Macroblock to process/filter, depending on cropping and filter_type.
   tl_mb_x, tl_mb_y int  // top-left MB that must be in-loop filtered
   br_mb_x, br_mb_y int  // last bottom-right MB that must be decoded

  // number of partitions minus one.
   num_parts_minus_one uint32
  // per-partition boolean decoders.
   parts [MAX_NUM_PARTITIONS]VP8BitReader

  // Dithering strength, deduced from decoding options
   dither int              // whether to use dithering or not
   dithering_rg VP8Random  // random generator for dithering

  // dequantization (one set of DC/AC dequant factor per segment)
   dqm [NUM_MB_SEGMENTS]VP8QuantMatrix

  // probabilities
   proba VP8Proba
   use_skip_proba int
   skip_p uint8

  // Boundary data cache and persistent buffers.
  intra_t *uint8    // top intra modes values: 4 * mb_w
   intra_l [4]uint8  // left intra modes values

  yuv_t *VP8TopSamples  // top y/u/v samples

  mb_info *VP8MB    // contextual macroblock info (mb_w + 1)
  f_info *VP8FInfo  // filter strength info
  yuv_b *uint8    // main block for Y/U/V (size = YUV_SIZE)

  cache_y *uint8  // macroblock row for storing unfiltered samples
  cache_u *uint8
  cache_v *uint8
   cache_y_stride int
   cache_uv_stride int

  // main memory chunk for the above data. Persistent.
  mem *void
   mem_size uint64

  // Per macroblock non-persistent infos.
   mb_x, mb_y int      // current position, in macroblock units
  mb_data *VP8MBData  // parsed reconstruction data

  // Filtering side-info
   filter_type int                          // 0=off, 1=simple, 2=complex
   fstrengths [NUM_MB_SEGMENTS][2]VP8FInfo  // precalculated per-segment/type

  // Alpha
  alph_dec *ALPHDecoder  // alpha-plane decoder object
  alpha_data *uint8  // compressed alpha data (if present)
   alpha_data_size uint64
   is_alpha_decoded int      // true if alpha_data is decoded in alpha_plane
  alpha_plane_mem *uint8  // memory allocated for alpha_plane
  alpha_plane *uint8      // output. Persistent, contains the whole data.
  alpha_prev_line *uint8  // last decoded alpha row (or nil)
   alpha_dithering int  // derived from decoding options (0=off, 100=full)
}

//------------------------------------------------------------------------------
// internal functions. Not public.

// in vp8.c
func VP8SetError(/* const */ dec *VP8Decoder, error VP8StatusCode, /* const */ msg *byte) int {
	// TODO: implement function body
	return 0
}

// in tree.c
func VP8ResetProba(/* const */ proba *VP8Proba) {
	// TODO: implement function body
}
func VP8ParseProba(/* const */ br *VP8BitReader, /* const */ dec *VP8Decoder) {
	// TODO: implement function body
}
// parses one row of intra mode data in partition 0, returns !eof
func VP8ParseIntraModeRow(/* const */ br *VP8BitReader, /* const */ dec *VP8Decoder) int {
	// TODO: implement function body
	return 0
}

// in quant.c
func VP8ParseQuant(/* const */ dec *VP8Decoder) {
	// TODO: implement function body
}

// in frame.c
func VP8InitFrame(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) int {
	// TODO: implement function body
	return 0
}
// After this call returns, one must always call VP8ExitCritical() with the
// same parameters. Both functions should be used in pair. Returns VP8_STATUS_OK
// if ok, otherwise sets and returns the error status on *dec.
func VP8EnterCritical(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) VP8StatusCode {
	// TODO: implement function body
	return 0
}
// Must always be called in pair with VP8EnterCritical().
// Returns false in case of error.
 func VP8ExitCritical(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) int {
	// TODO: implement function body
	return 0
}
// Return the multi-threading method to use (0=off), depending
// on options and bitstream size. Only for lossy decoding.
func VP8GetThreadMethod(/* const */ options *WebPDecoderOptions, /* const */ headers *WebPHeaderStructure, width, height int) int {
	// TODO: implement function body
	return 0
}
// Initialize dithering post-process if needed.
func VP8InitDithering(/* const */ options *WebPDecoderOptions, /* const */ dec *VP8Decoder){
	// TODO: implement function body
}
// Process the last decoded row (filtering + output).
func VP8ProcessRow(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) int {
	// TODO: implement function body
	return 0
}
// To be called at the start of a new scanline, to initialize predictors.
func VP8InitScanline(/* const */ dec *VP8Decoder){
	// TODO: implement function body
}
// Decode one macroblock. Returns false if there is not enough data.
 func VP8DecodeMB(/* const */ dec *VP8Decoder, /* const */ token_br *VP8BitReader) int {
	// TODO: implement function body
	return 0
}

// in alpha.c
func uint88DecompressAlphaRows(/* const */ dec *VP8Decoder, /* const */ io *VP8Io, row, num_rows int) *VP {
	// TODO: implement function body
	return nil
}
