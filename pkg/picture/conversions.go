package picture

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/color/colorspace"
	"github.com/daanv2/go-webp/pkg/color/colorspace/alpha"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
	"github.com/daanv2/go-webp/pkg/picture"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

func PictureARGBToYUVA(pict *Picture, csp colorspace.CSP, dithering float64, use_iterative_conversion int) error {
	if pict == nil {
		return nil
	}
	if pict.ARGB == nil {
		return pict.SetEncodingError(ENC_ERROR_nil_PARAMETER)
	} else if (csp & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420 {
		return pict.SetEncodingError(ENC_ERROR_INVALID_CONFIGURATION)
	}

	var argb []uint8 = pict.ARGB
	var a []uint8 = argb + CHANNEL_OFFSET(0)
	var r []uint8 = argb + CHANNEL_OFFSET(1)
	var g []uint8 = argb + CHANNEL_OFFSET(2)
	var b []uint8 = argb + CHANNEL_OFFSET(3)

	pict.ColorSpace = colorspace.WEBP_YUV420
	return ImportYUVAFromRGBA(r, g, b, a, 4, 4*pict.ARGBStride, dithering, use_iterative_conversion, pict)
}

func WebPPictureARGBToYUVADithered(picture *picture.Picture, colorspace colorspace.CSP, dithering float64) error {
	return PictureARGBToYUVA(picture, colorspace, dithering, 0)
}

func WebPPictureARGBToYUVA(picture *picture.Picture, colorspace colorspace.CSP) error {
	return PictureARGBToYUVA(picture, colorspace, 0.0, 0)
}

func WebPPictureSharpARGBToYUVA(picture *picture.Picture) error {
	return PictureARGBToYUVA(picture, colorspace.WEBP_YUV420, 0.0, 1)
}

// for backward compatibility
//go:fix inline
func WebPPictureSmartARGBToYUVA(picture *picture.Picture) int {
	return picture.WebPPictureSharpARGBToYUVA(picture)
}

func ImportYUVAFromRGBA( /* const */ r_ptr []uint8 /*const*/, g_ptr []uint8 /*const*/, b_ptr []uint8 /*const*/, a_ptr []uint8, step int, // bytes per pixel
	rgb_stride int, // bytes per scanline
	dithering float64, use_iterative_conversion int /*const*/, pict *picture.Picture) int {
	var y int
	width := pict.Width
	height := pict.Height
	has_alpha := alpha.CheckNonOpaqueUint8(a_ptr)

	pict.ColorSpace = tenary.If(has_alpha, colorspace.WEBP_YUV420A, colorspace.WEBP_YUV420)
	pict.UseARGB = false

	// disable smart conversion if source is too small (overkill).
	if width < enc.MinDimensionIterativeConversion ||
		height < enc.MinDimensionIterativeConversion {
		use_iterative_conversion = 0
	}

	if !WebPPictureAllocYUVA(pict) {
		return 0
	}
	if has_alpha {
		assert.Assert(step == 4)
	}

	if use_iterative_conversion {
		SharpYuvInit(VP8GetCPUInfo)
		if !PreprocessARGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, pict) {
			return 0
		}
		if has_alpha {
			WebPExtractAlpha(a_ptr, rgb_stride, width, height, pict.A, pict.AStride)
		}
	} else {
		uv_width := (width + 1) >> 1
		// temporary storage for accumulated R/G/B values during conversion to U/V
		// var tmp_rgb *uint16 = (*uint16)WebPSafeMalloc(4 * uv_width, sizeof(*tmp_rgb))
		tmp_rgb := make([]uint16, 4*uv_width)

		var dst_y []uint8 = pict.Y
		var dst_u []uint8 = pict.U
		var dst_v []uint8 = pict.V
		var dst_a []uint8 = pict.A

		var base_rg VP8Random
		var rg *VP8Random = nil
		if dithering > 0. {
			VP8InitRandom(&base_rg, dithering)
			rg = &base_rg
		}
		WebPInitConvertARGBToYUV()
		WebPInitGammaTables()

		// if (tmp_rgb == nil) {
		//   return picture.SetEncodingError(picture.ENC_ERROR_OUT_OF_MEMORY)
		// }

		if rg == nil {
			// Downsample Y/U/V planes, two rows at a time
			WebPImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, step, rgb_stride, has_alpha, width, height, tmp_rgb, pict.YStride, pict.UVStride, pict.AStride, dst_y, dst_u, dst_v, dst_a)
			if height & 1 {
				dst_y += (height - 1) * ptrdiff_t(pict.YStride)
				dst_u += (height >> 1) * ptrdiff_t(pict.UVStride)
				dst_v += (height >> 1) * ptrdiff_t(pict.UVStride)
				r_ptr += (height - 1) * ptrdiff_t(rgb_stride)
				b_ptr += (height - 1) * ptrdiff_t(rgb_stride)
				g_ptr += (height - 1) * ptrdiff_t(rgb_stride)
				if has_alpha {
					dst_a += (height - 1) * ptrdiff_t(pict.AStride)
					a_ptr += (height - 1) * ptrdiff_t(rgb_stride)
				}
				WebPImportYUVAFromRGBALastLine(r_ptr, g_ptr, b_ptr, a_ptr, step, has_alpha, width, tmp_rgb, dst_y, dst_u, dst_v, dst_a)
			}
		} else {
			// Copy of WebPImportYUVAFromRGBA/WebPImportYUVAFromRGBALastLine, // but with dithering.
			for y = 0; y < (height >> 1); y++ {
				rows_have_alpha := has_alpha
				ConvertRowToY(r_ptr, g_ptr, b_ptr, step, dst_y, width, rg)
				ConvertRowToY(r_ptr+rgb_stride, g_ptr+rgb_stride, b_ptr+rgb_stride, step, dst_y+picture.YStride, width, rg)
				dst_y += 2 * picture.YStride
				if has_alpha {
					rows_have_alpha &= !WebPExtractAlpha(a_ptr, rgb_stride, width, 2, dst_a, picture.AStride)
					dst_a += 2 * picture.AStride
				}
				// Collect averaged R/G/B(/A)
				if !rows_have_alpha {
					WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step, rgb_stride, tmp_rgb, width)
				} else {
					WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr, rgb_stride, tmp_rgb, width)
				}
				// Convert to U/V
				ConvertRowsToUV(tmp_rgb, dst_u, dst_v, uv_width, rg)
				dst_u += picture.UVStride
				dst_v += picture.UVStride
				r_ptr += 2 * rgb_stride
				b_ptr += 2 * rgb_stride
				g_ptr += 2 * rgb_stride
				if has_alpha {
					a_ptr += 2 * rgb_stride
				}
			}
			if height & 1 { // extra last row
				row_has_alpha := has_alpha
				ConvertRowToY(r_ptr, g_ptr, b_ptr, step, dst_y, width, rg)
				if row_has_alpha {
					row_has_alpha &= !WebPExtractAlpha(a_ptr, 0, width, 1, dst_a, 0)
				}
				// Collect averaged R/G/B(/A)
				if !row_has_alpha {
					// Collect averaged R/G/B
					WebPAccumulateRGB(r_ptr, g_ptr, b_ptr, step /*rgb_stride=*/, 0, tmp_rgb, width)
				} else {
					WebPAccumulateRGBA(r_ptr, g_ptr, b_ptr, a_ptr /*rgb_stride=*/, 0, tmp_rgb, width)
				}
				ConvertRowsToUV(tmp_rgb, dst_u, dst_v, uv_width, rg)
			}
		}

	}
	return 1
}
