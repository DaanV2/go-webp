package vp8

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Lossless decoder: internal header.
//
// Author: Skal (pascal.massimino@gmail.com)
//         Vikas Arora(vikaas.arora@gmail.com)

// for memcpy()


type VP8LDecodeState int

const ( 
	READ_DATA VP8LDecodeState = 0
	READ_HDR VP8LDecodeState = 1
	READ_DIM VP8LDecodeState = 2 
)

type VP8LTransform struct {
	vtype VP8LImageTransformType  // transform type.
	bits int                     // subsampling bits defining transform window.
	xsize int                    // transform window X index.
	ysize int                    // transform window Y index.
	data *uint32               // transform data.
}

type VP8LMetadata struct {
	color_cache_size int
	color_cache VP8LColorCache
	saved_color_cache VP8LColorCache  // for incremental

	huffman_mask int
	huffman_subsample_bits int
	huffman_xsize int
	huffman_image *uint32
	num_htree_groups int 
	htree_groups *HTreeGroup
	huffman_tables HuffmanTables
} 

type VP8LDecoder struct {
   status VP8StatusCode
   state VP8LDecodeState
  io *VP8Io

  output *WebPDecBuffer  // shortcut to io.opaque.output

  pixels *uint32      // Internal data: either for alpha *uint8
                         // or for BGRA *uint32.
  argb_cache *uint32  // Scratch buffer for temporary BGRA storage.
  accumulated_rgb_pixels *uint16  // Scratch buffer for accumulated RGB for
                                     // YUV conversion.

   br VP8LBitReader
   incremental int         // if true, incremental decoding is expected
   saved_br VP8LBitReader  // note: could be local variables too
   saved_last_pixel int

   width int
   height int
   last_row int      // last input row decoded so far.
  // last pixel decoded so far. However, it may
                     // not be transformed, scaled and
                     // color-converted yet.
   last_pixel int    
   last_out_row int  // last row output so far.

   hdr VP8LMetadata

   next_transform int
   transforms [NUM_TRANSFORMS]VP8LTransform
  // or'd bitset storing the transforms types.
   transforms_seen uint32

  rescaler_memory *uint8  // Working memory for rescaling work.
  rescaler *WebPRescaler    // Common rescaler for all channels.
}

//------------------------------------------------------------------------------
// internal functions. Not public.

// Decodes image header for alpha data stored using lossless compression.
// Returns false in case of error.
func VP8LDecodeAlphaHeader(alph_dec *ALPHDecoder,   data  *uint8, data_size uint64) int {
	// TODO: implement
	return 0
}

// Decodes *at *least 'last_row' rows of alpha. If some of the initial rows are
// already decoded in previous call(s), it will resume decoding from where it
// was paused.
// Returns false in case of bitstream error.
func VP8LDecodeAlphaImageStream(alph_dec *ALPHDecoder, last_row int) int {
	// TODO: implement
	return 0
}

// Allocates and initialize a new lossless decoder instance.
func VP8LDecoder8LNew(void)  *VP {
	// TODO: implement
	return nil
}

// Decodes the image header. Returns false in case of error.
func VP8LDecodeHeader(/* const */ dec *VP8LDecoder, /* const */ io *VP8Io) int {
	// TODO: implement
	return 0
}

// Decodes an image. It's required to decode the lossless header before calling
// this function. Returns false in case of error, with updated dec.status.
 func  VP8LDecodeImage(/* const */ dec *VP8LDecoder) int {
	// TODO: implement
	return 0
}

// Clears and deallocate a lossless decoder instance.
func VP8LDelete(/* const */ dec *VP8LDecoder) int {
	// TODO: implement
}

// Helper function for reading the different Huffman codes and storing them in
// 'huffman_tables' and 'htree_groups'.
// If mapping is nil 'num_htree_groups_max' must equal 'num_htree_groups'.
// If it is not nil, it maps 'num_htree_groups_max' indices to the
// 'num_htree_groups' groups. If 'num_htree_groups_max' > 'num_htree_groups',
// some of those indices map to -1. This is used for non-balanced codes to
// limit memory usage.
func ReadHuffmanCodesHelper(
    color_cache_bits, num_htree_groups, num_htree_groups_max int, 
	mapping *int,
	dec *VP8LDecoder,
	huffman_tables *HuffmanTables,
	htree_groups *HTreeGroup) {

}