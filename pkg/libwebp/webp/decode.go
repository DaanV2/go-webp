// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package webp

import (
	"github.com/daanv2/go-webp/pkg/libwebp/decoder"
	"github.com/daanv2/go-webp/pkg/vp8"
)

// Initialize the structure as empty. Must be called before any other use.
// Returns false in case of version mismatch
func WebPInitDecBuffer(buffer *WebPDecBuffer) int {
	return decoder.WebPInitDecBufferInternal(buffer, WEBP_DECODER_ABI_VERSION)
}

// Deprecated alpha-less version of WebPIDecGetYUVA(): it will ignore the
// alpha information (if present). Kept for backward compatibility.
func WebPIDecGetYUV( /* const */ idec *decoder.WebPIDecoder, last_y *int, u *uint8, v *uint8, width *int, height *int, stride *int, uv_stride *int) *uint8 {
	return decoder.WebPIDecGetYUVA(idec, last_y, u, v, nil, width, height, stride, uv_stride, nil)
}

//------------------------------------------------------------------------------
// Advanced decoding parametrization
//

// Features gathered from the bitstream
type WebPBitstreamFeatures struct {
	width         int // Width in pixels, as read from the bitstream.
	height        int // Height in pixels, as read from the bitstream.
	has_alpha     int // True if the bitstream contains an alpha channel.
	has_animation int // True if the bitstream is an animation.
	format        int // 0 = undefined (/mixed), 1 = lossy, 2 = lossless

	pad [5]uint32 // padding for later use
}

// Retrieve features from the bitstream. The structure is filled *features
// with information gathered from the bitstream.
// Returns VP8_STATUS_OK when the features are successfully retrieved. Returns
// VP8_STATUS_NOT_ENOUGH_DATA when more data is needed to retrieve the
// features from headers. Returns error in other cases.
// Note: The following chunk sequences (before the raw VP8/VP8L data) are
// considered valid by this function:
// RIFF + VP8(L)
// RIFF + VP8X + (optional chunks) + VP8(L)
// ALPH + VP8 <-- Not a valid WebP format: only allowed for internal purpose.
// VP8(L)     <-- Not a valid WebP format: only allowed for internal purpose.
func WebPGetFeatures(data *uint8, data_size uint64, features *WebPBitstreamFeatures) vp8.VP8StatusCode {
	return WebPGetFeaturesInternal(data, data_size, features, WEBP_DECODER_ABI_VERSION)
}

// Decoding options
type WebPDecoderOptions struct {
	bypass_filtering    int // if true, skip the in-loop filtering
	no_fancy_upsampling int // if true, use faster pointwise upsampler
	use_cropping        int // if true, cropping is applied _first_
	crop_left, crop_top int // top-left position for cropping.
	// Will be snapped to even values.
	crop_width, crop_height     int // dimension of the cropping area
	use_scaling                 int // if true, scaling is applied _afterward_
	scaled_width, scaled_height int // final resolution. if one is 0, it is
	// guessed from the other one to keep the
	// original ratio.
	use_threads              int // if true, use multi-threaded decoding
	dithering_strength       int // dithering strength (0=Off, 100=full)
	flip                     int // if true, flip output vertically
	alpha_dithering_strength int // alpha dithering strength in [0..100]

	pad [5]uint32 // padding for later use
}

// Main object storing the configuration for advanced decoding.
type WebPDecoderConfig struct {
	input   WebPBitstreamFeatures // Immutable bitstream features (optional)
	output  WebPDecBuffer         // Output buffer (can point to external mem)
	options WebPDecoderOptions    // Decoding options
}

// Initialize the configuration as empty. This function must always be
// called first, unless WebPGetFeatures() is to be called.
// Returns false in case of mismatched version.
func WebPInitDecoderConfig(config *WebPDecoderConfig) int {
	return WebPInitDecoderConfigInternal(config, WEBP_DECODER_ABI_VERSION)
}
