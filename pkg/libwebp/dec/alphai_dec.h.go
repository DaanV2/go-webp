package dec

// Copyright 2013 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Alpha decoder: internal header.
//
// Author: Urvang (urvang@google.com)

type ALPHDecoder struct {
	width int
	height int
	method int
	filter WEBP_FILTER_TYPE
	pre_processing int
	vp *VP8LDecoder8l_dec
	io VP8Io
	// Although alpha channel requires only 1 byte per
	// pixel, sometimes VP8LDecoder may need to allocate
	// 4 bytes per pixel internally during decode.
	use_8b_decode int  
	output *uint8
	prev_line *uint8  // last output row (or nil)
}

//------------------------------------------------------------------------------
// internal functions. Not public.

// Deallocate memory associated to dec.alpha_plane decoding
func WebPDeallocateAlphaMemory( dec *VP8Decoder);
