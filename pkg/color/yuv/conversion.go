package yuv

import "github.com/daanv2/go-webp/pkg/util/tenary"

// inline YUV<.RGB conversion function
//
// The exact naming is Y'CbCr, following the ITU-R BT.601 standard.
// More information at: https://en.wikipedia.org/wiki/YCbCr
// Y = 0.2568 * R + 0.5041 * G + 0.0979 * B + 16
// U = -0.1482 * R - 0.2910 * G + 0.4392 * B + 128
// V = 0.4392 * R - 0.3678 * G - 0.0714 * B + 128
// We use 16bit fixed point operations for RGB.YUV conversion (YUV_FIX).
//
// For the Y'CbCr to RGB conversion, the BT.601 specification reads:
//   R = 1.164 * (Y-16) + 1.596 * (V-128)
//   G = 1.164 * (Y-16) - 0.813 * (V-128) - 0.392 * (U-128)
//   B = 1.164 * (Y-16)                   + 2.017 * (U-128)
// where Y is in the [16,235] range, and U/V in the [16,240] range.
//
// The fixed-point implementation used here is:
//  R = (19077 . y             + 26149 . v - 14234) >> 6
//  G = (19077 . y -  6419 . u - 13320 . v +  8708) >> 6
//  B = (19077 . y + 33050 . u             - 17685) >> 6
// where the '.' operator is the mulhi_epu16 variant:
//   a . b = ((a << 8) * b) >> 16
// that preserves 8 bits of fractional precision before final descaling.

func YUVToR(y int, v int) int {
	return Clip8(MultHi(y, 19077) + MultHi(v, 26149) - 14234)
}

func YUVToG(y, u, v int) int {
	return Clip8(MultHi(y, 19077) - MultHi(u, 6419) - MultHi(v, 13320) + 8708)
}

func YUVToB(y, u int) int {
	return Clip8(MultHi(y, 19077) + MultHi(u, 33050) - 17685)
}

// TODO move to a emulation package
func MultHi(v, coeff int) int { // _mm_mulhi_epu16 emulation
	return (v * coeff) >> 8
}

// TODO move to a emulation package
func Clip8(v int) int {
	return tenary.If((v&(^YUV_MASK2)) == 0, (v >> YUV_FIX2), tenary.If(v < 0, 0, 255))
}

// Stub functions that can be called with various rounding values:
func ClipUV(uv int, rounding int) int {
	uv = (uv + rounding + (128 << (YUV_FIX + 2))) >> (YUV_FIX + 2)
	return tenary.If((uv&^0xff) == 0, uv, tenary.If(uv < 0, 0, 255))
}

func RGBToY(r int, g int, b int, rounding int) int {
	luma := 16839*r + 33059*g + 6420*b
	return (luma + rounding + (16 << YUV_FIX)) >> YUV_FIX // no need to clip
}

func RGBToU(r int, g int, b int, rounding int) int {
	u := -9719*r - 19081*g + 28800*b
	return ClipUV(u, rounding)
}

func RGBToV(r int, g int, b int, rounding int) int {
	v := +28800*r - 24116*g - 4684*b
	return ClipUV(v, rounding)
}
