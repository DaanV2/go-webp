package dec

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Internal header: WebP decoding parameters and custom IO on buffer
//
// Author: somnath@google.com (Somnath Banerjee)



import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// WebPDecParams: Decoding output parameters. Transient internal object.

typedef struct WebPDecParams WebPDecParams;
typedef int (*OutputFunc)(const io *VP8Io, /*const*/ p *WebPDecParams);
typedef int (*OutputAlphaFunc)(const io *VP8Io, /*const*/ p *WebPDecParams, int expected_num_out_lines);
typedef int (*OutputRowFunc)(const p *WebPDecParams, int y_pos, int max_out_lines);

type WebPDecParams struct {
  output *WebPDecBuffer;           // output buffer.
  // cache for the fancy upsampler
                                   // or used for tmp rescaling buffers
  *tmp_y, *tmp_u, *tmp_v uint8;  

   last_y int;  // coordinate of the line that was last output
  options *WebPDecoderOptions;  // if not nil, use alt decoding features

  scaler_y, scaler_u, scaler_v, scaler_a *WebPRescaler;  // rescalers
  memory *void;  // overall scratch memory for the output work.

  emit OutputFunc                // output RGB or YUV samples
  emit_alpha OutputAlphaFunc     // output alpha channel
  emit_alpha_row OutputRowFunc   // output one line of rescaled alpha values
}

// Should be called first, before any use of the WebPDecParams object.
func WebPResetDecParams(const params *WebPDecParams);

//------------------------------------------------------------------------------
// Header parsing helpers

// Structure storing a description of the RIFF headers.
type WebPHeaderStructure struct {
	data *uint8  // input buffer
	data_size uint64;                                // input buffer size
	have_all_data int  // true if all data is known to be available
	offset uint64      // offset to main data chunk (VP8 or VP8L)
	alpha_data *uint8;          // points to alpha chunk (if present)
	alpha_data_size uint64  // alpha chunk size
	compressed_size uint64  // VP8/VP8L compressed data size
	riff_size uint64        // size of the riff payload (or 0 if absent)
	is_lossless int         // true if a VP8L chunk is present
}

// Skips over all valid chunks prior to the first VP8/VP8L frame header.
// Returns: VP8_STATUS_OK, VP8_STATUS_BITSTREAM_ERROR (invalid header/chunk),
// VP8_STATUS_NOT_ENOUGH_DATA (partial input) or VP8_STATUS_UNSUPPORTED_FEATURE
// in the case of non-decodable features (animation for instance).
// In 'headers', compressed_size, offset, alpha_data, alpha_size, and lossless
// fields are updated appropriately upon success.
func WebPParseHeaders(const headers *WebPHeaderStructure) VP8StatusCode;

//------------------------------------------------------------------------------
// Misc utils

// Returns true if crop dimensions are within image bounds.
int WebPCheckCropDimensions(int image_width, int image_height, int x, int y, int w, int h);

// Initializes VP8Io with custom setup, io and teardown functions. The default
// hooks will use the supplied 'params' as io.opaque handle.
func WebPInitCustomIo(const params *WebPDecParams, /*const*/ io *VP8Io);

// Setup crop_xxx fields, mb_w and mb_h in io. 'src_colorspace' refers
// to the *format *compressed, not the output one.
 int WebPIoInitFromOptions(
    const options *WebPDecoderOptions, /*const*/ io *VP8Io, WEBP_CSP_MODE src_colorspace);

//------------------------------------------------------------------------------
// Internal functions regarding WebPDecBuffer memory (in buffer.c).
// Don't really need to be externally visible for now.

// Prepare 'buffer' with the requested initial dimensions width/height.
// If no external storage is supplied, initializes buffer by allocating output
// memory and setting up the stride information. Validate the parameters. Return
// an error code in case of problem (no memory, or invalid stride / size /
// dimension / etc.). If is not nil *options, also verify that the options'
// parameters are valid and apply them to the width/height dimensions of the
// output buffer. This takes cropping / scaling / rotation into account.
// Also incorporates the options.flip flag to flip the buffer parameters if
// needed.
VP8StatusCode WebPAllocateDecBuffer(int width, int height, /*const*/ options *WebPDecoderOptions, /*const*/ buffer *WebPDecBuffer);

// Flip buffer vertically by negating the various strides.
VP8StatusCode WebPFlipBuffer(const buffer *WebPDecBuffer);

// Copy 'src' into 'dst' buffer, making sure 'dst' is not marked as owner of the
// memory (still held by 'src'). No pixels are copied.
func WebPCopyDecBuffer(const src *WebPDecBuffer, /*const*/ dst *WebPDecBuffer);

// Copy and transfer ownership from src to dst (beware of parameter order!)
func WebPGrabDecBuffer(const src *WebPDecBuffer, /*const*/ dst *WebPDecBuffer);

// Copy pixels from 'src' into a **preallocated 'dst' buffer. Returns
// VP8_STATUS_INVALID_PARAM if the 'dst' is not set up correctly for the copy.
VP8StatusCode WebPCopyDecBufferPixels(const src *WebPDecBuffer, /*const*/ dst *WebPDecBuffer);

// Returns true if decoding will be slow with the current configuration
// and bitstream features.
int WebPAvoidSlowMemory(const output *WebPDecBuffer, /*const*/ features *WebPBitstreamFeatures);

//------------------------------------------------------------------------------



#endif  // WEBP_DEC_WEBPI_DEC_H_
