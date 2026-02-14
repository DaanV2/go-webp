// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/color/yuv"
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/stdlib"
)

//go:fix inline
func row_func(callfn func(y, u, v, rgb *uint8), XSTEP int) func(y, u, v, dst *uint8, len int) {
	return func(y, u, v, dst *uint8, len int) {
		var end *uint8 = dst + (len&~1)*(XSTEP)

		for dst != end {
			callfn(y[0], u[0], v[0], dst)
			callfn(y[1], u[0], v[0], dst+(XSTEP))
			y += 2
			u++
			v++
			dst += 2 * (XSTEP)
		}
		if len & 1 {
			callfn(y[0], u[0], v[0], dst)
		}
	}
}

// All variants implemented.
var YuvToRgbRow = row_func(yuv.YuvToRgb, 3)
var YuvToBgrRow = row_func(yuv.YuvToBgr, 3)
var YuvToRgbaRow = row_func(yuv.YuvToRgba, 4)
var YuvToBgraRow = row_func(yuv.YuvToBgra, 4)
var YuvToArgbRow = row_func(yuv.YuvToArgb, 4)
var YuvToRgba4444Row = row_func(yuv.YuvToRgba4444, 2)
var YuvToRgb565Row = row_func(yuv.YuvToRgb565, 2)

// Main call for processing a plane with a WebPSamplerRowFunc function:
func WebPSamplerProcessPlane( /* const */ y *uint8, y_stride int /*const*/, u *uint8 /*const*/, v *uint8, uv_stride int, dst *uint8, dst_stride int, width, height int, fn WebPSamplerRowFunc) {
	var j int
	for j = 0; j < height; j++ {
		fn(y, u, v, dst, width)
		y += y_stride
		if j & 1 {
			u += uv_stride
			v += uv_stride
		}
		dst += dst_stride
	}
}

func WEBP_DSP_INIT_FUNC(WebPInitSamplers) {
	WebPSamplers[MODE_RGB] = YuvToRgbRow
	WebPSamplers[MODE_RGBA] = YuvToRgbaRow
	WebPSamplers[MODE_BGR] = YuvToBgrRow
	WebPSamplers[MODE_BGRA] = YuvToBgraRow
	WebPSamplers[MODE_ARGB] = YuvToArgbRow
	WebPSamplers[MODE_RGBA_4444] = YuvToRgba4444Row
	WebPSamplers[MODE_RGB_565] = YuvToRgb565Row
	WebPSamplers[MODE_rgbA] = YuvToRgbaRow
	WebPSamplers[MODE_bgrA] = YuvToBgraRow
	WebPSamplers[MODE_Argb] = YuvToArgbRow
	WebPSamplers[MODE_rgbA_4444] = YuvToRgba4444Row
}

func WebPConvertARGBToY( /* const */ argb []uint32, y []uint8, width int) {
	var i int
	for i = 0; i < width; i++ {
		p := argb[i]
		y[i] = uint8(yuv.RGBToY((p>>16)&0xff, (p>>8)&0xff, (p>>0)&0xff, yuv.YUV_HALF))
	}
}

func WebPConvertARGBToUV( /* const */ argb []uint32, u []uint8, v []uint8, src_width int, do_store int) {
	// No rounding. Last pixel is dealt with separately.
	uv_width := src_width >> 1
	var i int
	for i = 0; i < uv_width; i++ {
		v0 := argb[2*i+0]
		v1 := argb[2*i+1]
		// yuv.RGBToU/V expects four accumulated pixels. Hence we need to
		// scale r/g/b value by a factor 2. We just shift v0/v1 one bit less.
		r := ((v0 >> 15) & 0x1fe) + ((v1 >> 15) & 0x1fe)
		g := ((v0 >> 7) & 0x1fe) + ((v1 >> 7) & 0x1fe)
		b := ((v0 << 1) & 0x1fe) + ((v1 << 1) & 0x1fe)
		tmp_u := yuv.RGBToU(r, g, b, YUV_HALF<<2)
		tmp_v := yuv.RGBToV(r, g, b, YUV_HALF<<2)
		if do_store {
			u[i] = tmp_u
			v[i] = tmp_v
		} else {
			// Approximated average-of-four. But it's an acceptable diff.
			u[i] = (u[i] + tmp_u + 1) >> 1
			v[i] = (v[i] + tmp_v + 1) >> 1
		}
	}
	if src_width & 1 { // last pixel
		v0 := argb[2*i+0]
		r := (v0 >> 14) & 0x3fc
		g := (v0 >> 6) & 0x3fc
		b := (v0 << 2) & 0x3fc
		tmp_u := yuv.RGBToU(r, g, b, YUV_HALF<<2)
		tmp_v := yuv.RGBToV(r, g, b, YUV_HALF<<2)
		if do_store {
			u[i] = tmp_u
			v[i] = tmp_v
		} else {
			u[i] = (u[i] + tmp_u + 1) >> 1
			v[i] = (v[i] + tmp_v + 1) >> 1
		}
	}
}

//-----------------------------------------------------------------------------

func WebPConvertRGBToY( /* const */ rgb []uint8, y []uint8, width int, step int) {
	var i int
	for i = 0; i < width; {
		y[i] = uint8(yuv.RGBToY(int(rgb[0]), int(rgb[1]), int(rgb[2]), yuv.YUV_HALF))
		i++
		rgb = rgb[step:]
	}
}

func WebPConvertBGRToY( /* const */ bgr []uint8, y []uint8, width int, step int) {
	var i int
	for i = 0; i < width; {
		y[i] = uint8(yuv.RGBToY(int(bgr[2]), int(bgr[1]), int(bgr[0]), yuv.YUV_HALF))
		i++
		bgr = bgr[step:]
	}
}

func WebPConvertRGBA32ToUV( /* const */ rgb []uint16, u []uint8, v []uint8, width int) {
	var i int
	for i = 0; i < width; {
		r := int(rgb[0])
		g := int(rgb[1])
		b := int(rgb[2])
		u[i] = uint8(yuv.RGBToU(r, g, b, yuv.YUV_HALF<<2))
		v[i] = uint8(yuv.RGBToV(r, g, b, yuv.YUV_HALF<<2))
		i += 1
		rgb = rgb[4:]
	}
}

// Gamma correction compensates loss of resolution during chroma subsampling.
const GAMMA_FIX = 12    // fixed-point precision for linear values
const GAMMA_TAB_FIX = 7 // fixed-point fractional bits precision
const GAMMA_TAB_SIZE = (1 << (GAMMA_FIX - GAMMA_TAB_FIX))
const kGamma = 0.80
const kGammaScale = ((1 << GAMMA_FIX) - 1)
const kGammaTabScale = (1 << GAMMA_TAB_FIX)
const kGammaTabRounder = (1 << GAMMA_TAB_FIX >> 1)

var kLinearToGammaTab = [GAMMA_TAB_SIZE + 1]int{}
var kGammaToLinearTab = [256]uint16{}
var kGammaTablesOk = 0

func init() {
	if kGammaTablesOk == 0 {
		var v int
		const scale = float64((1 << GAMMA_TAB_FIX) / kGammaScale)
		const norm = float64(1.0 / 255.0)
		for v = 0; v <= 255; v++ {
			kGammaToLinearTab[v] = uint16(pow(norm*v, kGamma)*kGammaScale + 0.5)
		}
		for v = 0; v <= GAMMA_TAB_SIZE; v++ {
			kLinearToGammaTab[v] = int(255.0*pow(scale*v, 1.0/kGamma) + 0.5)
		}
		kGammaTablesOk = 1
	}
}

func GammaToLinear(v uint8) uint32 {
	return uint32(kGammaToLinearTab[v])
}

func Interpolate(v int) int {
	tab_pos := v >> (GAMMA_TAB_FIX + 2)  // integer part
	x := v & ((kGammaTabScale << 2) - 1) // fractional part
	v0 := kLinearToGammaTab[tab_pos]
	v1 := kLinearToGammaTab[tab_pos+1]
	y := v1*x + v0*((kGammaTabScale<<2)-x) // interpolate
	assert.Assert(tab_pos+1 < GAMMA_TAB_SIZE+1)
	return y
}

// Convert a linear value 'v' to YUV_FIX+2 fixed-point precision
// U/V value, suitable for RGBToU/V calls.
func LinearToGamma(base_value uint32, shift int) int {
	y := Interpolate(base_value << shift)          // final uplifted value
	return (y + kGammaTabRounder) >> GAMMA_TAB_FIX // descale
}

func SUM4(ptr, step) {
	LinearToGamma(GammaToLinear((ptr)[0])+GammaToLinear((ptr)[(step)])+
		GammaToLinear((ptr)[rgb_stride])+
		GammaToLinear((ptr)[rgb_stride+(step)]),
		0)
}

func SUM2(ptr) {
	LinearToGamma(GammaToLinear((ptr)[0])+GammaToLinear((ptr)[rgb_stride]), 1)
}

//------------------------------------------------------------------------------
// "Fast" regular RGB.YUV

func SUM4(ptr, step) {
	LinearToGamma(GammaToLinear((ptr)[0])+GammaToLinear((ptr)[(step)])+
		GammaToLinear((ptr)[rgb_stride])+
		GammaToLinear((ptr)[rgb_stride+(step)]),
		0)
}

func SUM2(ptr) {
	LinearToGamma(GammaToLinear((ptr)[0])+GammaToLinear((ptr)[rgb_stride]), 1)
}
func SUM2ALPHA(ptr) { return ((ptr)[0] + (ptr)[rgb_stride]) }
func SUM4ALPHA(ptr) { return (SUM2ALPHA(ptr) + SUM2ALPHA((ptr)+4)) }

func DIVIDE_BY_ALPHA(sum, a int) int { return (4 * (sum) / (a)) }

func LinearToGammaWeighted( /* const */ src *uint8 /*const*/, a_ptr *uint8, uint32 total_a, step int, rgb_stride int) int {
	sum := a_ptr[0]*GammaToLinear(src[0]) +
		a_ptr[step]*GammaToLinear(src[step]) +
		a_ptr[rgb_stride]*GammaToLinear(src[rgb_stride]) +
		a_ptr[rgb_stride+step]*GammaToLinear(src[rgb_stride+step])
	assert.Assert(total_a > 0 && total_a <= 4*0xff)

	return LinearToGamma(DIVIDE_BY_ALPHA(sum, total_a), 0)
}

func WebPAccumulateRGBA( /* const */ r_ptr *uint8 /*const*/, g_ptr *uint8 /*const*/, b_ptr *uint8 /*const*/, a_ptr *uint8, rgb_stride int, dst *uint16, width int) {
	var i, j int = 0, 0
	// we loop over 2x2 blocks and produce one R/G/B/A value for each.
	for i < (width >> 1) {
		a := SUM4ALPHA(a_ptr + j)
		var r, g, b int
		if a == 4*0xff || a == 0 {
			r = SUM4(r_ptr+j, 4)
			g = SUM4(g_ptr+j, 4)
			b = SUM4(b_ptr+j, 4)
		} else {
			r = LinearToGammaWeighted(r_ptr+j, a_ptr+j, a, 4, rgb_stride)
			g = LinearToGammaWeighted(g_ptr+j, a_ptr+j, a, 4, rgb_stride)
			b = LinearToGammaWeighted(b_ptr+j, a_ptr+j, a, 4, rgb_stride)
		}
		dst[0] = r
		dst[1] = g
		dst[2] = b
		dst[3] = a

		i += 1
		j += 2 * 4
		dst += 4
	}
	if width & 1 {
		a := uint(2) * SUM2ALPHA(a_ptr+j)
		var r, g, b int
		if a == 4*0xff || a == 0 {
			r = SUM2(r_ptr + j)
			g = SUM2(g_ptr + j)
			b = SUM2(b_ptr + j)
		} else {
			r = LinearToGammaWeighted(r_ptr+j, a_ptr+j, a, 0, rgb_stride)
			g = LinearToGammaWeighted(g_ptr+j, a_ptr+j, a, 0, rgb_stride)
			b = LinearToGammaWeighted(b_ptr+j, a_ptr+j, a, 0, rgb_stride)
		}
		dst[0] = r
		dst[1] = g
		dst[2] = b
		dst[3] = a
	}
}

func WebPAccumulateRGB( /* const */ r_ptr *uint8 /*const*/, g_ptr *uint8 /*const*/, b_ptr *uint8, step int, rgb_stride int, dst *uint16, width int) {
	var i, j int = 0, 0
	for i < (width >> 1) {
		dst[0] = SUM4(r_ptr+j, step)
		dst[1] = SUM4(g_ptr+j, step)
		dst[2] = SUM4(b_ptr+j, step)
		i += 1
		j += 2 * step
		dst += 4
	}
	if width & 1 {
		dst[0] = SUM2(r_ptr + j)
		dst[1] = SUM2(g_ptr + j)
		dst[2] = SUM2(b_ptr + j)

	}
}

func WebPImportYUVAFromRGBA( /* const */ r_ptr []uint8 /*const*/, g_ptr []uint8 /*const*/, b_ptr []uint8 /*const*/, a_ptr []uint8, step int, // bytes per pixel
	rgb_stride int, // bytes per scanline
	has_alpha bool, width, height int, tmp_rgb *uint16, y_stride int, uv_stride int, a_stride int, dst_y *uint8, dst_u *uint8, dst_v *uint8, dst_a *uint8) {
	var y int
	is_rgb := (r_ptr < b_ptr) // otherwise it's bgr
	uv_width := (width + 1) >> 1

	has_alpha &= dst_a != nil

	WebPInitGammaTables()

	// Downsample Y/U/V planes, two rows at a time
	for y = 0; y < (height >> 1); y++ {
		rows_have_alpha := has_alpha
		if is_rgb {
			WebPConvertRGBToY(r_ptr, dst_y, width, step)
			WebPConvertRGBToY(r_ptr+rgb_stride, dst_y+y_stride, width, step)
		} else {
			WebPConvertBGRToY(b_ptr, dst_y, width, step)
			WebPConvertBGRToY(b_ptr+rgb_stride, dst_y+y_stride, width, step)
		}
		dst_y += 2 * y_stride
		if has_alpha {
			rows_have_alpha &= !WebPExtractAlpha(a_ptr, rgb_stride, width, 2, dst_a, a_stride)
			dst_a += 2 * a_stride
		} else if dst_a != nil {
			var i int
			for i = 0; i < 2; {
				stdlib.Memset(dst_a, 0xff, width)
				i++
				dst_a += a_stride
			}
		}

		// Collect averaged R/G/B(/A)
		if !rows_have_alpha {
			WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, tmp_rgb, width)
		} else {
			WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr, rgb_stride, tmp_rgb, width)
		}
		// Convert to U/V
		WebPConvertRGBA32ToUV(tmp_rgb, dst_u, dst_v, uv_width)
		dst_u += uv_stride
		dst_v += uv_stride
		r_ptr += 2 * rgb_stride
		b_ptr += 2 * rgb_stride
		g_ptr += 2 * rgb_stride
		if has_alpha {
			a_ptr += 2 * rgb_stride
		}
	}
}

func WebPImportYUVAFromRGBALastLine( /* const */ r_ptr *uint8 /*const*/, g_ptr *uint8 /*const*/, b_ptr *uint8 /*const*/, a_ptr *uint8, step int, // bytes per pixel
	has_alpha bool, width int, tmp_rgb *uint16, dst_y *uint8, dst_u *uint8, dst_v *uint8, dst_a *uint8) {
	is_rgb := (r_ptr < b_ptr) // otherwise it's bgr
	uv_width := (width + 1) >> 1
	row_has_alpha := has_alpha && dst_a != nil

	if is_rgb {
		WebPConvertRGBToY(r_ptr, dst_y, width, step)
	} else {
		WebPConvertBGRToY(b_ptr, dst_y, width, step)
	}
	if row_has_alpha {
		row_has_alpha &= !WebPExtractAlpha(a_ptr, 0, width, 1, dst_a, 0)
	} else if dst_a != nil {
		stdlib.Memset(dst_a, 0xff, width)
	}

	// Collect averaged R/G/B(/A)
	if !row_has_alpha {
		// Collect averaged R/G/B
		WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step /*rgb_stride=*/, 0, tmp_rgb, width)
	} else {
		WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr /*rgb_stride=*/, 0, tmp_rgb, width)
	}
	WebPConvertRGBA32ToUV(tmp_rgb, dst_u, dst_v, uv_width)
}

// Macros to give the offset of each channel in a uint32 containing ARGB.
func CHANNEL_OFFSET(i int) int {
	if constants.FALSE {
		// uint32 0xff000000 is 0xff,00,00,00 in memory
		return i
	}

	// uint32 0xff000000 is 0x00,00,00,ff in memory
	return 3 - i
}

//------------------------------------------------------------------------------
// slower on x86 by ~7-8%, but bit-exact with the SSE2/NEON version

//go:fix inline
func VP8YUVToR(y int, v int) int {
	return yuv.YUVToR(y, v)
}

//go:fix inline
func VP8YUVToG(y int, u int, v int) int {
	return yuv.YUVToG(y, u, v)
}

//go:fix inline
func VP8YUVToB(y int, u int) int {
	return yuv.YUVToB(y, v)
}
