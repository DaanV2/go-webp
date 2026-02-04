package decoder

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Everything about WebPDecBuffer
//
// Author: Skal (pascal.massimino@gmail.com)

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
	"github.com/daanv2/go-webp/pkg/vp8"
)

//------------------------------------------------------------------------------
// WebPDecBuffer

// Number of bytes per pixel for the different color-spaces.
var kModeBpp = [int(webp.MODE_LAST)]uint8{
	3, 4, 3, 4, 4, 2, 2, //
	4, 4, 4, 2, // pre-multiplied modes
	1, 1,
}

// Convert to an integer to handle both the unsigned/signed enum cases
// without the need for casting to remove type limit warnings.
// Check that webp_csp_mode is within the bounds of WEBP_CSP_MODE.
func IsValidColorspace(webp_csp_mode int) bool {
	return (webp_csp_mode >= int(webp.MODE_RGB) && webp_csp_mode < int(webp.MODE_LAST))
}

// strictly speaking, the very last (or first, if flipped) row
// doesn't require padding.
func MIN_BUFFER_SIZE(WIDTH, HEIGHT, STRIDE uint64) uint64 {
	return ((uint64)(STRIDE)*((HEIGHT)-1) + (WIDTH))
}

func CheckDecBuffer( /* const */ buffer *WebPDecBuffer) vp8.VP8StatusCode {
	ok := 1
	mode := buffer.colorspace
	width := buffer.width
	height := buffer.height
	if !IsValidColorspace(mode) {
		ok = 0
	} else if !WebPIsRGBMode(mode) { // YUV checks
		var buf *WebPYUVABuffer = &buffer.u.YUVA
		uv_width := (width + 1) / 2
		uv_height := (height + 1) / 2
		y_stride := stdlib.Abs(buf.y_stride)
		u_stride := stdlib.Abs(buf.u_stride)
		v_stride := stdlib.Abs(buf.v_stride)
		a_stride := stdlib.Abs(buf.a_stride)
		y_size := MIN_BUFFER_SIZE(width, height, y_stride)
		u_size := MIN_BUFFER_SIZE(uv_width, uv_height, u_stride)
		v_size := MIN_BUFFER_SIZE(uv_width, uv_height, v_stride)
		a_size := MIN_BUFFER_SIZE(width, height, a_stride)
		ok &= (y_size <= buf.y_size)
		ok &= (u_size <= buf.u_size)
		ok &= (v_size <= buf.v_size)
		ok &= (y_stride >= width)
		ok &= (u_stride >= uv_width)
		ok &= (v_stride >= uv_width)
		ok &= (buf.y != nil)
		ok &= (buf.u != nil)
		ok &= (buf.v != nil)
		if mode == webp.MODE_YUVA {
			ok &= (a_stride >= width)
			ok &= (a_size <= buf.a_size)
			ok &= (buf.a != nil)
		}
	} else { // RGB checks
		var buf *WebPRGBABuffer = &buffer.u.RGBA
		stride := stdlib.Abs(buf.stride)
		size :=
			MIN_BUFFER_SIZE(width*kModeBpp[mode], height, stride)
		ok &= (size <= buf.size)
		ok &= (stride >= width*kModeBpp[mode])
		ok &= (buf.rgba != nil)
	}
	return tenary.If(ok, vp8.VP8_STATUS_OK, vp8.VP8_STATUS_INVALID_PARAM)
}

func AllocateBuffer( /* const */ buffer *WebPDecBuffer) vp8.VP8StatusCode {
	w := buffer.width
	h := buffer.height
	mode := buffer.colorspace

	if w <= 0 || h <= 0 || !IsValidColorspace(mode) {
		return vp8.VP8_STATUS_INVALID_PARAM
	}

	if buffer.is_external_memory <= 0 && buffer.private_memory == nil {
		var output *uint8
		uv_stride := 0
		a_stride := 0
		uv_size := 0
		a_size := 0
		// We need memory and it hasn't been allocated yet.
		// => initialize output buffer, now that dimensions are known.
		var stride int
		var size uint64

		if w*kModeBpp[mode] >= (uint64(1) << 31) {
			return vp8.VP8_STATUS_INVALID_PARAM
		}
		stride = w * kModeBpp[mode]
		size = uint64(stride * h)
		if !WebPIsRGBMode(mode) {
			uv_stride = (w + 1) / 2
			uv_size = uint64(uv_stride * ((h + 1) / 2))
			if mode == webp.MODE_YUVA {
				a_stride = w
				a_size = uint64(a_stride * h)
			}
		}
		total_size = size + 2*uv_size + a_size

		// output = (*uint8)WebPSafeMalloc(total_size, sizeof(*output));
		// if (output == nil) {
		//   return VP8_STATUS_OUT_OF_MEMORY;
		// }
		ouput = make([]uint8, total_size)

		buffer.private_memory = output

		if !WebPIsRGBMode(mode) { // YUVA initialization
			var buf *WebPYUVABuffer = &buffer.u.YUVA
			buf.y = output
			buf.y_stride = stride
			buf.y_size = uint64(size)
			buf.u = output + size
			buf.u_stride = uv_stride
			buf.u_size = uint64(uv_size)
			buf.v = output + size + uv_size
			buf.v_stride = uv_stride
			buf.v_size = uint64(uv_size)
			if mode == MODE_YUVA {
				buf.a = output + size + 2*uv_size
			}
			buf.a_size = uint64(a_size)
			buf.a_stride = a_stride
		} else { // RGBA initialization
			var buf *WebPRGBABuffer = &buffer.u.RGBA
			buf.rgba = output
			buf.stride = stride
			buf.size = uint64(size)
		}
	}
	return CheckDecBuffer(buffer)
}

// Flip buffer vertically by negating the various strides.
func WebPFlipBuffer( /* const */ buffer *WebPDecBuffer) vp8.VP8StatusCode {
	if buffer == nil {
		return vp8.VP8_STATUS_INVALID_PARAM
	}
	if WebPIsRGBMode(buffer.colorspace) {
		var buf *WebPRGBABuffer = &buffer.u.RGBA
		buf.rgba += (int64)(buffer.height-1) * buf.stride
		buf.stride = -buf.stride
	} else {
		var buf *WebPYUVABuffer = &buffer.u.YUVA
		H := buffer.height
		buf.y += (H - 1) * buf.y_stride
		buf.y_stride = -buf.y_stride
		buf.u += ((H - 1) >> 1) * buf.u_stride
		buf.u_stride = -buf.u_stride
		buf.v += ((H - 1) >> 1) * buf.v_stride
		buf.v_stride = -buf.v_stride
		if buf.a != nil {
			buf.a += (H - 1) * buf.a_stride
			buf.a_stride = -buf.a_stride
		}
	}
	return vp8.VP8_STATUS_OK
}

// Prepare 'buffer' with the requested initial dimensions width/height.
// If no external storage is supplied, initializes buffer by allocating output
// memory and setting up the stride information. Validate the parameters. Return
// an error code in case of problem (no memory, or invalid stride / size /
// dimension / etc.). If is not nil *options, also verify that the options'
// parameters are valid and apply them to the width/height dimensions of the
// output buffer. This takes cropping / scaling / rotation into account.
// Also incorporates the options.flip flag to flip the buffer parameters if
// needed.
func WebPAllocateDecBuffer(width int, height int /*const*/, options *WebPDecoderOptions /*const*/, buffer *WebPDecBuffer) vp8.VP8StatusCode {
	var status vp8.VP8StatusCode
	if buffer == nil || width <= 0 || height <= 0 {
		return vp8.VP8_STATUS_INVALID_PARAM
	}
	if options != nil { // First, apply options if there is any.
		if options.use_cropping {
			cw := options.crop_width
			ch := options.crop_height
			x := options.crop_left & ~1
			y := options.crop_top & ~1
			if !WebPCheckCropDimensions(width, height, x, y, cw, ch) {
				return vp8.VP8_STATUS_INVALID_PARAM // out of frame boundary.
			}
			width = cw
			height = ch
		}

		if options.use_scaling {
			scaled_width := options.scaled_width
			scaled_height := options.scaled_height
			if !WebPRescalerGetScaledDimensions(width, height, &scaled_width, &scaled_height) {
				return vp8.VP8_STATUS_INVALID_PARAM
			}
			width = scaled_width
			height = scaled_height
		}
	}
	buffer.width = width
	buffer.height = height

	// Then, allocate buffer for real.
	status = AllocateBuffer(buffer)
	if status != vp8.VP8_STATUS_OK {
		return status
	}

	// Use the stride trick if vertical flip is needed.
	if options != nil && options.flip {
		status = WebPFlipBuffer(buffer)
	}
	return status
}

//------------------------------------------------------------------------------
// constructors / destructors

func WebPInitDecBufferInternal(buffer *WebPDecBuffer, version int) int {
	if WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_DECODER_ABI_VERSION) {
		return 0 // version mismatch
	}
	if buffer == nil {
		return 0
	}
	stdlib.Memset(buffer, 0, sizeof(*buffer))
	return 1
}

func WebPFreeDecBuffer(buffer *WebPDecBuffer) {
	if buffer != nil {
		buffer.private_memory = nil
	}
}

// Copy 'src' into 'dst' buffer, making sure 'dst' is not marked as owner of the
// memory (still held by 'src'). No pixels are copied.
func WebPCopyDecBuffer( /* const */ src *WebPDecBuffer /*const*/, dst *WebPDecBuffer) {
	if src != nil && dst != nil {
		*dst = *src
		if src.private_memory != nil {
			dst.is_external_memory = 1 // dst buffer doesn't own the memory.
			dst.private_memory = nil
		}
	}
}

// Copy and transfer ownership from src to dst (beware of parameter order!)
func WebPGrabDecBuffer( /* const */ src *WebPDecBuffer /*const*/, dst *WebPDecBuffer) {
	if src != nil && dst != nil {
		*dst = *src
		if src.private_memory != nil {
			src.is_external_memory = 1 // src relinquishes ownership
			src.private_memory = nil
		}
	}
}

// Copy pixels from 'src' into a **preallocated 'dst' buffer. Returns
// VP8_STATUS_INVALID_PARAM if the 'dst' is not set up correctly for the copy.
func WebPCopyDecBufferPixels( /* const */ src_buf *WebPDecBuffer /*const*/, dst_buf *WebPDecBuffer) VP8StatusCode {
	assert.Assert(src_buf != nil && dst_buf != nil)
	assert.Assert(src_buf.colorspace == dst_buf.colorspace)

	dst_buf.width = src_buf.width
	dst_buf.height = src_buf.height
	if CheckDecBuffer(dst_buf) != VP8_STATUS_OK {
		return VP8_STATUS_INVALID_PARAM
	}
	if WebPIsRGBMode(src_buf.colorspace) {
		var src *WebPRGBABuffer = &src_buf.u.RGBA
		var dst *WebPRGBABuffer = &dst_buf.u.RGBA
		WebPCopyPlane(src.rgba, src.stride, dst.rgba, dst.stride, src_buf.width*kModeBpp[src_buf.colorspace], src_buf.height)
	} else {
		var src *WebPYUVABuffer = &src_buf.u.YUVA
		var dst *WebPYUVABuffer = &dst_buf.u.YUVA
		WebPCopyPlane(src.y, src.y_stride, dst.y, dst.y_stride, src_buf.width, src_buf.height)
		WebPCopyPlane(src.u, src.u_stride, dst.u, dst.u_stride, (src_buf.width+1)/2, (src_buf.height+1)/2)
		WebPCopyPlane(src.v, src.v_stride, dst.v, dst.v_stride, (src_buf.width+1)/2, (src_buf.height+1)/2)
		if WebPIsAlphaMode(src_buf.colorspace) {
			WebPCopyPlane(src.a, src.a_stride, dst.a, dst.a_stride, src_buf.width, src_buf.height)
		}
	}
	return VP8_STATUS_OK
}

// Returns true if decoding will be slow with the current configuration
// and bitstream features.
func WebPAvoidSlowMemory( /* const */ output *WebPDecBuffer /*const*/, features *WebPBitstreamFeatures) int {
	assert.Assert(output != nil)
	return (output.is_external_memory >= 2) &&
		WebPIsPremultipliedMode(output.colorspace) &&
		(features != nil && features.has_alpha)
}

//------------------------------------------------------------------------------
