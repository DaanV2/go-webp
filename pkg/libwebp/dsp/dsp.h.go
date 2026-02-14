package dsp

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//   Speed-critical functions.
//
// Author: Skal (pascal.massimino@gmail.com)


#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"



const BPS = 32  // this is the common stride for enc/dec

//------------------------------------------------------------------------------
// WEBP_RESTRICT

// Declares a pointer with the restrict type qualifier if available.
// This allows code to hint to the compiler that only this pointer references a
// particular object or memory region within the scope of the block in which it
// is declared. This may allow for improved optimizations due to the lack of
// pointer aliasing. See also:
// https://en.cppreference.com/w/c/language/restrict
#if defined(__GNUC__)
const =__restrict__
#elif defined(_MSC_VER)
const =__restrict
#else
#define WEBP_RESTRICT
#endif

//------------------------------------------------------------------------------
// Init stub generator

// Defines an init function stub to ensure each module exposes a symbol,
// avoiding a compiler warning.
#define WEBP_DSP_INIT_STUB(func) \
  extern func func(void);        \
  func func(){}

//------------------------------------------------------------------------------
// Encoding

// Transforms
// VP8Idct: Does one of two inverse transforms. If do_two is set, the transforms
//          will be done for (ref, in, dst) and (ref + 4, in + 16, dst + 4).
type VP8Idct = func(/* const */ ref *uint8, /*const*/ in *int16, dst *uint8, do_two int)
type VP8Fdct = func(/* const */ src *uint8, /*const*/ ref *uint8, out *int16)
type VP8WHT = func(/* const */ in *int16, out *int16)




// Predictions
// is the destination *dst block. and can be *top nil *left.
type VP8IntraPreds = func(dst *uint8, /*const*/ left *uint8, /*const*/ top *uint8)
type VP8Intra4Preds = func(dst *uint8, /*const*/ top *uint8)




typedef int (*VP8Metric)(/* const */ pix *uint8, /*const*/ ref *uint8)
extern VP8Metric VP8SSE16x16, VP8SSE16x8, VP8SSE8x8, VP8SSE4x4
typedef int (*VP8WMetric)(/* const */ pix *uint8, /*const*/ ref *uint8, /*const*/ /* const */ weights *uint16)
// The weights for VP8TDisto4x4 and VP8TDisto16x16 contain a row-major
// 4 by 4 symmetric matrix.
extern VP8WMetric VP8TDisto4x4, VP8TDisto16x16

// Compute the average (DC) of four 4x4 blocks.
// Each sub-4x4 block #i sum is stored in dc[i].
type VP8MeanMetric = func(/* const */ ref *uint8, uint32 dc[4])


type VP8BlockCopy = func(/* const */ src *uint8, dst *uint8)


// Quantization
struct VP8Matrix;  // forward declaration
typedef int (*VP8QuantizeBlock)(
    int16 in[16], int16 out[16], /*const*/ mtx *VP8Matrix)
// Same as VP8QuantizeBlock, but quantizes two consecutive blocks.
typedef int (*VP8Quantize2Blocks)(
    int16 in[32], int16 out[32], /*const*/ mtx *VP8Matrix)




// specific to 2nd transform:
typedef int (*VP8QuantizeBlockWHT)(
    int16 in[16], int16 out[16], /*const*/ mtx *VP8Matrix)


extern [16 + 4 + 4]

// Collect histogram for susceptibility calculation.
const MAX_COEFF_THRESH =31  // size of histogram used by CollectHistogram.
type VP8Histogram struct {
  // We only need to store max_value and last_non_zero, not the distribution.
  max_value int
  last_non_zero int
} 
type VP8CHisto = func(/* const */ ref *uint8, /*const*/ pred *uint8, start_block int, end_block int, /* const */ histo *VP8Histogram)

// General-purpose util function to help VP8CollectHistogram().
func VP8SetHistogramData(/* const */ int distribution[MAX_COEFF_THRESH + 1], /*const*/ histo *VP8Histogram)


//------------------------------------------------------------------------------
// cost functions (encoding)

extern const uint16 VP8EntropyCost[256];  // 8bit fixed-point log(p)
// approximate cost per level:
extern const uint16 VP8LevelFixedCosts[2047 /*MAX_*LEVEL/ + 1]
extern const uint8 VP8EncBands[16 + 1]

struct VP8Residual
type VP8SetResidualCoeffsFunc = func(/*const*//* const */ coeffs *int16, res *VP8Residual)


// Cost calculation function.
typedef int (*VP8GetResidualCostFunc)(int ctx0, /*const*/ struct const res *VP8Residual)


// must be called before anything using the above
func VP8EncDspCostInit(void)

//------------------------------------------------------------------------------
// SSIM / PSNR utils

// struct for accumulating statistical moments
type VP8DistoStats struct {
	w uint32              // sum(w_i) : sum of weights
	xm, ym uint32         // sum(w_i * x_i), sum(w_i * y_i)
	xxm, xym, yym uint32  // sum(w_i * x_i * x_i), etc.
}

const VP8_SSIM_KERNEL =3  // total size of the kernel: 2 * VP8_SSIM_KERNEL + 1

type VP8SSIMGetClippedFunc = func(/* const */ src *uint81, int stride1, /*const*/ src *uint82,  stride2 int, xo int, yo int, W int, H  int) float64    // plane dimensio
type VP8DecIdct = func(/* const */ coeffs *int16, dst *uint8)
// when doing two transforms, coeffs is actually int16[2][16].
type VP8DecIdct2 = func(/* const */ coeffs *int16, dst *uint8, do_two int)







const WEBP_TRANSFORM_AC3_C1 =20091
const WEBP_TRANSFORM_AC3_C2 =35468
#define WEBP_TRANSFORM_AC3_MUL1(a) ((((a) * WEBP_TRANSFORM_AC3_C1) >> 16) + (a))
#define WEBP_TRANSFORM_AC3_MUL2(a) (((a) * WEBP_TRANSFORM_AC3_C2) >> 16)

// is the destination *dst block, with stride BPS. Boundary samples are
// assumed accessible when needed.
type VP8PredFunc = func(dst *uint8)
extern VP8PredFunc VP8PredLuma16[NUM_B_DC_MODES]
extern VP8PredFunc VP8PredChroma8[NUM_B_DC_MODES]
extern VP8PredFunc VP8PredLuma4[NUM_BMODES]

// clipping tables (for filtering)
const VP8ksclip1 []int8  // clips [-1020, 1020] to [-128, 127]
const VP8ksclip2 []int8  // clips [-112, 112] to [-16, 15]
const VP8kclip1 []uint8  // clips [-255,511] to [0,255]
const VP8kabs0 []uint8   // abs(x) for x in [-255,255]
// must be called first
func VP8InitClipTables(void)

// simple filter (only for luma)
type VP8SimpleFilterFunc = func(p *uint8, stride int, thresh int)
// regular filter (on both macroblock edges and inner edges)
type VP8LumaFilterFunc = func(luma *uint8, stride int, thresh int, ithresh int, hev_t int)
type VP8ChromaFilterFunc = func(u *uint8, v *uint8, stride int, thresh int, ithresh int, hev_t int)

// Dithering. Combines dithering values (centered around 128) with dst[],
// according to: dst[] = clip(dst[] + (((dither[]-128) + 8) >> 4)
const VP8_DITHER_DESCALE =4
const VP8_DITHER_DESCALE_ROUNDER =(1 << (VP8_DITHER_DESCALE - 1))
const VP8_DITHER_AMP_BITS =7
const VP8_DITHER_AMP_CENTER =(1 << VP8_DITHER_AMP_BITS)
type VP8DitherCombine8x8 = func(/* const */ dither *uint8, dst *uint8, dst_stride int)


//------------------------------------------------------------------------------
// WebP I/O

const TRUE = // undefined to remove fancy upsampling support

// Convert a pair of y/u/v lines together to the output rgb/a colorspace.
// bottom_y can be nil if only one line of output is needed (at top/bottom).
type WebPUpsampleLinePairFunc = func(/*const*/top_y *uint8, /*const*/ bottom_y *uint8, /*const*/ top_u *uint8, /*const*/ top_v *uint8, /*const*/ cur_u *uint8, /*const*/ cur_v *uint8, top_dst *uint8, bottom_dst *uint8, len int)

#ifdef TRUE

// Fancy upsampling functions to convert YUV to RGB(A) modes
extern WebPUpsampleLinePairFunc WebPUpsamplers[MODE_LAST]

#endif  // TRUE

// Per-row point-sampling methods.
type WebPSamplerRowFunc = func(/* const */ y *uint8, /*const*/ u *uint8, /*const*/ v *uint8, dst *uint8, len int)
// Generic function to apply 'WebPSamplerRowFunc' to the whole plane:
func WebPSamplerProcessPlane(/* const */ y *uint8, y_stride int, /*const*/ u *uint8, /*const*/ v *uint8, uv_stride int, dst *uint8, dst_stride int, width, height int, WebPSamplerRowFunc func)

// Sampling functions to convert rows of YUV to RGB(A)
extern WebPSamplerRowFunc WebPSamplers[MODE_LAST]

// General function for converting two lines of ARGB or RGBA.
// 'alpha_is_last' should be true if 0xff000000 is stored in memory as
// as 0x00, 0x00, 0x00, 0xff (little endian).
WebPUpsampleLinePairFunc WebPGetLinePairConverter(int alpha_is_last)

// YUV444.RGB converters
type WebPYUV444Converter = func(/* const */ y *uint8, /*const*/ u *uint8, /*const*/ v *uint8, dst *uint8, len int)

// Must be called before using the WebPUpsamplers[] (and for premultiplied
// colorspaces like rgbA, rgbA4444, etc)
func WebPInitUpsamplers(void)
// Must be called before using WebPSamplers[]
func WebPInitSamplers(void)
// Must be called before using WebPYUV444Converters[]
func WebPInitYUV444Converters(void)

//------------------------------------------------------------------------------
// ARGB . YUV converters

// Convert ARGB samples to luma Y.
type WebPConvertARGBToY = func(/* const */ argb *uint32, y *uint8, width int)
// Convert ARGB samples to U/V with downsampling. do_store should be '1' for
// even lines and '0' for odd ones. 'src_width' is the original width, not
// the U/V one.
type WebPConvertARGBToUV = func(/* const */ argb *uint32, u *uint8, v *uint8, src_width int, do_store int)

// Convert a row of accumulated (four-values) of rgba32 toward U/V
type WebPConvertRGBA32ToUV = func(/* const */ rgb *uint16, u *uint8, v *uint8, width int)

// Convert RGB or BGR to Y. Step is 3 or 4. If step is 4, data is RGBA or BGRA.
type WebPConvertRGBToY = func(/* const */ rgb *uint8, y *uint8, width int, step int)
type WebPConvertBGRToY = func(/* const */ bgr *uint8, y *uint8, width int, step int)

// Must be called before using the above.
func WebPInitConvertARGBToYUV(void)

// Import a row of data and save its contribution in the rescaler.
// 'channel' denotes the channel number to be imported. 'Expand' corresponds to
// the wrk.x_expand case. Otherwise, 'Shrink' is to be used.
type WebPRescalerImportRowFunc = func(wrk *WebPRescaler, /*const*/ src *uint8)




// Export one row (starting at x_out position) from rescaler.
// 'Expand' corresponds to the wrk.y_expand case.
// Otherwise 'Shrink' is to be used
type WebPRescalerExportRowFunc = func(struct const wrk *WebPRescaler)



// Main entry calls:
extern func WebPRescalerImportRow(wrk *WebPRescaler, /*const*/ src *uint8)
// Export one row (starting at x_out position) from rescaler.
extern func WebPRescalerExportRow(struct const wrk *WebPRescaler)

// Must be called first before using the above.
func WebPRescalerDspInit(void)

//------------------------------------------------------------------------------
// Utilities for processing transparent channel.

// Apply alpha pre-multiply on an rgba, bgra or argb plane of size w * h.
// alpha_first should be 0 for argb, 1 for rgba or bgra (where alpha is last).
type WebPApplyAlphaMultiply = func(rgba *uint8, alpha_first int, w int, h int, stride int)

// Same, buf specifically for RGBA4444 format
type WebPApplyAlphaMultiply4444 = func(rgba *uint84444, w int, h int, stride int)

// Dispatch the values from alpha[] plane to the ARGB destination 'dst'.
// Returns true if alpha[] plane has non-trivial values different from 0xff.
type WebPDispatchAlpha = func(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint8, dst_stride int)

// Transfer packed 8b alpha[] values to green channel in dst[], zero'ing the
// A/R/B values. 'dst_stride' is the stride for dst[] in uint32 units.
type WebPDispatchAlphaToGreen = func(/* const */ alpha *uint8, alpha_stride int, width, height int, dst *uint32, dst_stride int)

// Extract the alpha values from 32b values in argb[] and pack them into alpha[]
// (this is the opposite of WebPDispatchAlpha).
// Returns true if there's only trivial 0xff alpha values.
type WebPExtractAlpha = func(/* const */ argb *uint8, argb_stride int, width, height int, alpha *uint8, alpha_stride int)

// Extract the green values from 32b values in argb[] and pack them into alpha[]
// (this is the opposite of WebPDispatchAlphaToGreen).
type WebPExtractGreen = func(/* const */ argb *uint32, alpha *uint8, size int)

// Pre-Multiply operation transforms x into x * A / 255  (where x=Y,R,G or B).
// Un-Multiply operation transforms x into x * 255 / A.

// Pre-Multiply or Un-Multiply (if 'inverse' is true) argb values in a row.
type WebPMultARGBRow = func(/* const */ ptr *uint32, width int, inverse int)

// Same a WebPMultARGBRow(), but for several rows.
func WebPMultARGBRows(ptr *uint8, stride int, width int, num_rows int , inverse int)

// Same for a row of single values, with side alpha values.
type WebPMultRow = func(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int)

// Same a WebPMultRow(), but for several 'num_rows' rows.
func WebPMultRows(ptr *uint8, stride int, /*const*/ alpha *uint8, alpha_stride int, width int, num_rows int , inverse int)

// Plain-C versions, used as fallback by some implementations.
func WebPMultRow_C(/* const */ ptr *uint8, /*const*/ /* const */ alpha *uint8, width int, inverse int)
func WebPMultARGBRow_C(/* const */ ptr *uint32, width int, inverse int)

// RGB packing function. 'step' can be 3 or 4. r/g/b input is rgb or bgr order.
type WebPPackRGB = func(/* const */ r *uint8, /*const*/ g *uint8, /*const*/ b *uint8, len int, step int, out *uint32)

// This function returns true if src[i] contains a value different from 0xff.
type WebPHasAlpha8b = func(/* const */ src *uint8, length int) int
// This function returns true if src[4*i] contains a value different from 0xff.
type WebPHasAlpha32b = func(/* const */ src *uint8, length int) int
// replaces transparent values in src[] by 'color'.
type WebPAlphaReplace = func(src *uint32, length int, color uint32)

// To be called first before using the above.
func WebPInitAlphaProcessing(void)

type WebPFilterFunc = func(/* const */ in *uint8, width, height int, stride int, out *uint8)
// In-place un-filtering.
// Warning! 'prev_line' pointer can be equal to 'cur_line' or 'preds'.
type WebPUnfilterFunc = func(/* const */ prev_line *uint8, /*const*/ preds *uint8, cur_line *uint8, width int)
