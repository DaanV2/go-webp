package dec

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


import "github.com/daanv2/go-webp/pkg/string"  // for memcpy()

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

type <FOO> int

const ( READ_DATA = 0, READ_HDR = 1, READ_DIM = 2 } VP8LDecodeState;

typedef struct VP8LTransform VP8LTransform;
type VP8LTransform struct {
  VP8LImageTransformType type;  // transform type.
  int bits;                     // subsampling bits defining transform window.
  int xsize;                    // transform window X index.
  int ysize;                    // transform window Y index.
  data *uint32;               // transform data.
}

type <Foo> struct {
  int color_cache_size;
  VP8LColorCache color_cache;
  VP8LColorCache saved_color_cache;  // for incremental

  int huffman_mask;
  int huffman_subsample_bits;
  int huffman_xsize;
  huffman_image *uint32;
  int num_htree_groups;
  htree_groups *HTreeGroup;
  HuffmanTables huffman_tables;
} VP8LMetadata;

typedef struct VP8LDecoder VP8LDecoder;
type VP8LDecoder struct {
  VP8StatusCode status;
  VP8LDecodeState state;
  io *VP8Io;

  const output *WebPDecBuffer;  // shortcut to io.opaque.output

  pixels *uint32;      // Internal data: either for alpha *uint8
                         // or for BGRA *uint32.
  argb_cache *uint32;  // Scratch buffer for temporary BGRA storage.
  accumulated_rgb_pixels *uint16;  // Scratch buffer for accumulated RGB for
                                     // YUV conversion.

  VP8LBitReader br;
  int incremental;         // if true, incremental decoding is expected
  VP8LBitReader saved_br;  // note: could be local variables too
  int saved_last_pixel;

  int width;
  int height;
  int last_row;      // last input row decoded so far.
  int last_pixel;    // last pixel decoded so far. However, it may
                     // not be transformed, scaled and
                     // color-converted yet.
  int last_out_row;  // last row output so far.

  VP8LMetadata hdr;

  int next_transform;
  VP8LTransform transforms[NUM_TRANSFORMS];
  // or'd bitset storing the transforms types.
  uint32 transforms_seen;

  rescaler_memory *uint8;  // Working memory for rescaling work.
  rescaler *WebPRescaler;    // Common rescaler for all channels.
}

//------------------------------------------------------------------------------
// internal functions. Not public.

struct ALPHDecoder;  // Defined in dec/alphai.h.

// in vp8l.c

// Decodes image header for alpha data stored using lossless compression.
// Returns false in case of error.
 int VP8LDecodeAlphaHeader(
    struct const alph_dec *ALPHDecoder, const *uint8  data, uint64 data_size);

// Decodes *at *least 'last_row' rows of alpha. If some of the initial rows are
// already decoded in previous call(s), it will resume decoding from where it
// was paused.
// Returns false in case of bitstream error.
 int VP8LDecodeAlphaImageStream(
    struct const alph_dec *ALPHDecoder, int last_row);

// Allocates and initialize a new lossless decoder instance.
 VP *VP8LDecoder8LNew(void);

// Decodes the image header. Returns false in case of error.
 int VP8LDecodeHeader(const dec *VP8LDecoder, const io *VP8Io);

// Decodes an image. It's required to decode the lossless header before calling
// this function. Returns false in case of error, with updated dec.status.
 int VP8LDecodeImage(const dec *VP8LDecoder);

// Clears and deallocate a lossless decoder instance.
func VP8LDelete(const dec *VP8LDecoder);

// Helper function for reading the different Huffman codes and storing them in
// 'huffman_tables' and 'htree_groups'.
// If mapping is nil 'num_htree_groups_max' must equal 'num_htree_groups'.
// If it is not nil, it maps 'num_htree_groups_max' indices to the
// 'num_htree_groups' groups. If 'num_htree_groups_max' > 'num_htree_groups',
// some of those indices map to -1. This is used for non-balanced codes to
// limit memory usage.
 int ReadHuffmanCodesHelper(
    int color_cache_bits, int num_htree_groups, int num_htree_groups_max, const mapping *int, const dec *VP8LDecoder, const huffman_tables *HuffmanTables, *HTreeGroup* const htree_groups);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_DEC_VP8LI_DEC_H_
