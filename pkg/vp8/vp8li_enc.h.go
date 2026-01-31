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
// Lossless encoder: internal header.
//
// Author: Vikas Arora (vikaas.arora@gmail.com)


import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// maximum value of 'transform_bits' in VP8LEncoder.
const MAX_TRANSFORM_BITS =(MIN_TRANSFORM_BITS + (1 << NUM_TRANSFORM_BITS) - 1)

type VP8LEncoderARGBContent int

const (
  kEncoderNone  VP8LEncoderARGBContent = iota
  kEncoderARGB
  kEncoderNearLossless
  kEncoderPalette
)

type VP8LEncoder struct {
  config *WebPConfig;  // user configuration and parameters
  pic *WebPPicture;    // input picture.

  argb *uint32;                       // Transformed argb image data.
  argb_content VP8LEncoderARGBContent;  // Content type of the argb buffer.
  argb_scratch *uint32;               // Scratch memory for argb rows
                                        // (used for prediction).
  transform_data *uint32;             // Scratch memory for transform data.
  transform_mem *uint32;              // Currently allocated memory.
  uint64 transform_mem_size;            // Currently allocated memory size.

  int current_width;  // Corresponds to packed image width.

  // Encoding parameters derived from quality parameter.
  int histo_bits;
  int predictor_transform_bits;    // <= MAX_TRANSFORM_BITS
  int cross_color_transform_bits;  // <= MAX_TRANSFORM_BITS
  int cache_bits;                  // If equal to 0, don't use color cache.

  // Encoding parameters derived from image characteristics.
  int use_cross_color;
  int use_subtract_green;
  int use_predict;
  int use_palette;
  int palette_size;
  uint32 palette[MAX_PALETTE_SIZE];
  // Sorted version of palette for cache purposes.
  uint32 palette_sorted[MAX_PALETTE_SIZE];

  // Some 'scratch' (potentially large) objects.
  struct VP8LBackwardRefs refs[4];  // Backward Refs array for temporaries.
  VP8LHashChain hash_chain;         // HashChain data for constructing
                                    // backward references.
} ;

//------------------------------------------------------------------------------
// internal functions. Not public.

// Encodes the picture.
// Returns 0 if config or picture is nil or picture doesn't have valid argb
// input.
int VP8LEncodeImage(const config *WebPConfig, const picture *WebPPicture);

// Encodes the main image stream using the supplied bit writer.
// Returns false in case of error (stored in picture.error_code).
int VP8LEncodeStream(const config *WebPConfig, const picture *WebPPicture, const bw *VP8LBitWriter);

#if (WEBP_NEAR_LOSSLESS == 1)
// in near_lossless.c
// Near lossless preprocessing in RGB color-space.
int VP8ApplyNearLossless(const picture *WebPPicture, int quality, const argb_dst *uint32);
#endif

//------------------------------------------------------------------------------
// Image transforms in predictor.c.

// pic and percent are for progress.
// Returns false in case of error (stored in pic.error_code).
int VP8LResidualImage(int width, int height, int min_bits, int max_bits, int low_effort, const argb *uint32, const argb_scratch *uint32, const image *uint32, int near_lossless, int exact, int used_subtract_green, const pic *WebPPicture, int percent_range, const percent *int, const best_bits *int);

int VP8LColorSpaceTransform(int width, int height, int bits, int quality, const argb *uint32, image *uint32, const pic *WebPPicture, int percent_range, const percent *int, const best_bits *int);

func VP8LOptimizeSampling(const image *uint32, int full_width, int full_height, int bits, int max_bits, best_bits_out *int);

//------------------------------------------------------------------------------



#endif  // WEBP_ENC_VP8LI_ENC_H_
