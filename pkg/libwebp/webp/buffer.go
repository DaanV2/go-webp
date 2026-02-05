// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package webp

type WebPRGBABuffer struct { // view as RGBA
	rgba   *uint8 // pointer to RGBA samples
	stride int    // stride in bytes from one scanline to the next.
	size   uint64 // total size of the buffer *rgba.
}

type WebPYUVABuffer struct { // view as YUVA
	y, u, v, a         *uint8 // pointer to luma, chroma U/V, alpha samples
	y_stride           int    // luma stride
	u_stride, v_stride int    // chroma strides
	a_stride           int    // alpha stride
	y_size             uint64 // luma plane size
	u_size, v_size     uint64 // chroma planes size
	a_size             uint64 // alpha-plane size
}

// Output buffer
type WebPDecBuffer struct {
	colorspace         WEBP_CSP_MODE // Colorspace.
	width, height      int           // Dimensions.
	is_external_memory int           // If non-zero, 'internal_memory' pointer is not
	// used. If value is '2' or more, the external
	// memory is considered 'slow' and multiple
	// read/write will be avoided.
	// Nameless union of buffer parameters.
	u struct {
		RGBA WebPRGBABuffer
		YUVA WebPYUVABuffer
	}
	pad [4]uint32 // padding for later use

	// Internally allocated memory (only when
	// is_external_memory is 0). Should not be used
	// externally, but accessed via the buffer union.
	private_memory *uint8
}
