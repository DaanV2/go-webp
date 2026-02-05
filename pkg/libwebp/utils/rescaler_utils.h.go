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



import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


const WEBP_RESCALER_RFIX =32  // fixed-point precision for multiplies
const WEBP_RESCALER_ONE =(uint64(1) << WEBP_RESCALER_RFIX)
#define WEBP_RESCALER_FRAC(x, y) \
  ((uint32)(((uint64)(x) << WEBP_RESCALER_RFIX) / (y)))

type WebPRescaler struct {
  x_expand int               // true if we're expanding in the x direction
  y_expand int               // true if we're expanding in the y direction
  num_channels int           // bytes to jump between pixels
  fx_scale uint32          // fixed-point scaling factors
  fy_scale uint32          // ''
  fxy_scale uint32         // ''
  y_accum int                // vertical accumulator
  y_add, y_sub int           // vertical increments
  x_add, x_sub int           // horizontal increments
  src_width, src_height int  // source dimensions
  dst_width, dst_height int  // destination dimensions
  src_y, dst_y int           // row counters for input and output
  dst *uint8;
  dst_stride int
  // work buffer
  irow (dst_num_channels *width) *rescaler_t
  irow (dst_num_channels *width) *rescaler_t
}

// Initialize a rescaler given scratch area 'work' and dimensions of src & dst.
// Returns false in case of error.
int WebPRescalerInit(/* const */ rescaler *WebPRescaler, src_width int, src_height int, /*const*/ dst *uint8, dst_width int, dst_height int, dst_stride int, num_channels int, rescaler_t* const (uint64(2) * dst_width *
                                                       num_channels) work);

// If either 'scaled_width' or 'scaled_height' (but not both) is 0 the value
// will be calculated preserving the aspect ratio, otherwise the values are
// left unmodified. Returns true on success, false if either value is 0 after
// performing the scaling calculation.
int WebPRescalerGetScaledDimensions(src_width int, src_height int, /*const*/ scaled_width *int, /*const*/ scaled_height *int);

// Returns the number of input lines needed next to produce one output line,
// considering that the maximum available input lines are 'max_num_lines'.
int WebPRescaleNeededLines(/* const */ rescaler *WebPRescaler, max_num_lines int);

// Import multiple rows over all channels, until at least one row is ready to
// be exported. Returns the actual number of lines that were imported.
int WebPRescalerImport(/* const */ rescaler *WebPRescaler, num_rows int , /*const*/ src *uint8, src_stride int);

// Export as many rows as possible. Return the numbers of rows written.
int WebPRescalerExport(/* const */ rescaler *WebPRescaler);

// Return true if input is finished
static  int WebPRescalerInputDone(
    const rescaler *WebPRescaler) {
  return (rescaler.src_y >= rescaler.src_height);
}
// Return true if output is finished
static  int WebPRescalerOutputDone(
    const rescaler *WebPRescaler) {
  return (rescaler.dst_y >= rescaler.dst_height);
}

// Return true if there are pending output rows ready.
static  int WebPRescalerHasPendingOutput(
    const rescaler *WebPRescaler) {
  return !WebPRescalerOutputDone(rescaler) && (rescaler.y_accum <= 0);
}

//------------------------------------------------------------------------------



#endif  // WEBP_UTILS_RESCALER_UTILS_H_
