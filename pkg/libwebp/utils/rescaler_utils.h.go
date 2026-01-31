package utils

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Rescaling functions
//
// Author: Skal (pascal.massimino@gmail.com)


#ifdef __cplusplus
extern "C" {
#endif

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

const WEBP_RESCALER_RFIX =32  // fixed-point precision for multiplies
const WEBP_RESCALER_ONE =(uint64(1) << WEBP_RESCALER_RFIX)
#define WEBP_RESCALER_FRAC(x, y) \
  ((uint32)(((uint64)(x) << WEBP_RESCALER_RFIX) / (y)))

// Structure used for on-the-fly rescaling
typedef uint32 rescaler_t;  // type for side-buffer
typedef struct WebPRescaler WebPRescaler;
type WebPRescaler struct {
  int x_expand;               // true if we're expanding in the x direction
  int y_expand;               // true if we're expanding in the y direction
  int num_channels;           // bytes to jump between pixels
  uint32 fx_scale;          // fixed-point scaling factors
  uint32 fy_scale;          // ''
  uint32 fxy_scale;         // ''
  int y_accum;                // vertical accumulator
  int y_add, y_sub;           // vertical increments
  int x_add, x_sub;           // horizontal increments
  int src_width, src_height;  // source dimensions
  int dst_width, dst_height;  // destination dimensions
  int src_y, dst_y;           // row counters for input and output
  dst *uint8;
  int dst_stride;
  // work buffer
  rescaler_t* WEBP_COUNTED_BY(dst_num_channels *width) irow;
  rescaler_t* WEBP_COUNTED_BY(dst_num_channels *width) frow;
}

// Initialize a rescaler given scratch area 'work' and dimensions of src & dst.
// Returns false in case of error.
int WebPRescalerInit(const rescaler *WebPRescaler, int src_width, int src_height, const dst *uint8, int dst_width, int dst_height, int dst_stride, int num_channels, rescaler_t* const WEBP_COUNTED_BY(uint64(2) * dst_width *
                                                       num_channels) work);

// If either 'scaled_width' or 'scaled_height' (but not both) is 0 the value
// will be calculated preserving the aspect ratio, otherwise the values are
// left unmodified. Returns true on success, false if either value is 0 after
// performing the scaling calculation.
int WebPRescalerGetScaledDimensions(int src_width, int src_height, const scaled_width *int, const scaled_height *int);

// Returns the number of input lines needed next to produce one output line,
// considering that the maximum available input lines are 'max_num_lines'.
int WebPRescaleNeededLines(const const rescaler *WebPRescaler, int max_num_lines);

// Import multiple rows over all channels, until at least one row is ready to
// be exported. Returns the actual number of lines that were imported.
int WebPRescalerImport(const rescaler *WebPRescaler, int num_rows, const src *uint8, int src_stride);

// Export as many rows as possible. Return the numbers of rows written.
int WebPRescalerExport(const rescaler *WebPRescaler);

// Return true if input is finished
static  int WebPRescalerInputDone(
    const const rescaler *WebPRescaler) {
  return (rescaler.src_y >= rescaler.src_height);
}
// Return true if output is finished
static  int WebPRescalerOutputDone(
    const const rescaler *WebPRescaler) {
  return (rescaler.dst_y >= rescaler.dst_height);
}

// Return true if there are pending output rows ready.
static  int WebPRescalerHasPendingOutput(
    const const rescaler *WebPRescaler) {
  return !WebPRescalerOutputDone(rescaler) && (rescaler.y_accum <= 0);
}

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_UTILS_RESCALER_UTILS_H_
