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
const WEBP_RESTRICT =__restrict__
#elif defined(_MSC_VER)
const WEBP_RESTRICT =__restrict
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
typedef func (*VP8Idct)(/* const */ WEBP_RESTRICT ref *uint8, /*const*/ WEBP_RESTRICT in *int16, WEBP_RESTRICT dst *uint8, int do_two);
typedef func (*VP8Fdct)(/* const */ WEBP_RESTRICT src *uint8, /*const*/ WEBP_RESTRICT ref *uint8, WEBP_RESTRICT out *int16);
typedef func (*VP8WHT)(/* const */ WEBP_RESTRICT in *int16, WEBP_RESTRICT out *int16);
extern VP8Idct VP8ITransform;
extern VP8Fdct VP8FTransform;
extern VP8Fdct VP8FTransform2;  // performs two transforms at a time
extern VP8WHT VP8FTransformWHT;
// Predictions
// is the destination *dst block. and can be *top nil *left.
typedef func (*VP8IntraPreds)(WEBP_RESTRICT dst *uint8, /*const*/ WEBP_RESTRICT left *uint8, /*const*/ WEBP_RESTRICT top *uint8);
typedef func (*VP8Intra4Preds)(WEBP_RESTRICT dst *uint8, /*const*/ WEBP_RESTRICT top *uint8);
extern VP8Intra4Preds VP8EncPredLuma4;
extern VP8IntraPreds VP8EncPredLuma16;
extern VP8IntraPreds VP8EncPredChroma8;

typedef int (*VP8Metric)(/* const */ WEBP_RESTRICT pix *uint8, /*const*/ WEBP_RESTRICT ref *uint8);
extern VP8Metric VP8SSE16x16, VP8SSE16x8, VP8SSE8x8, VP8SSE4x4;
typedef int (*VP8WMetric)(/* const */ WEBP_RESTRICT pix *uint8, /*const*/ WEBP_RESTRICT ref *uint8, /*const*/ WEBP_RESTRICT const weights *uint16);
// The weights for VP8TDisto4x4 and VP8TDisto16x16 contain a row-major
// 4 by 4 symmetric matrix.
extern VP8WMetric VP8TDisto4x4, VP8TDisto16x16;

// Compute the average (DC) of four 4x4 blocks.
// Each sub-4x4 block #i sum is stored in dc[i].
typedef func (*VP8MeanMetric)(/* const */ WEBP_RESTRICT ref *uint8, uint32 dc[4]);
extern VP8MeanMetric VP8Mean16x4;

typedef func (*VP8BlockCopy)(/* const */ WEBP_RESTRICT src *uint8, WEBP_RESTRICT dst *uint8);
extern VP8BlockCopy VP8Copy4x4;
extern VP8BlockCopy VP8Copy16x8;
// Quantization
struct VP8Matrix;  // forward declaration
typedef int (*VP8QuantizeBlock)(
    int16 in[16], int16 out[16], /*const*/ struct WEBP_RESTRICT const mtx *VP8Matrix);
// Same as VP8QuantizeBlock, but quantizes two consecutive blocks.
typedef int (*VP8Quantize2Blocks)(
    int16 in[32], int16 out[32], /*const*/ struct WEBP_RESTRICT const mtx *VP8Matrix);

extern VP8QuantizeBlock VP8EncQuantizeBlock;
extern VP8Quantize2Blocks VP8EncQuantize2Blocks;

// specific to 2nd transform:
typedef int (*VP8QuantizeBlockWHT)(
    int16 in[16], int16 out[16], /*const*/ struct WEBP_RESTRICT const mtx *VP8Matrix);
extern VP8QuantizeBlockWHT VP8EncQuantizeBlockWHT;

extern const int VP8DspScan[16 + 4 + 4];

// Collect histogram for susceptibility calculation.
const MAX_COEFF_THRESH =31  // size of histogram used by CollectHistogram.
type VP8Histogram struct {
  // We only need to store max_value and last_non_zero, not the distribution.
  int max_value;
  int last_non_zero;
} ;
typedef func (*VP8CHisto)(/* const */ WEBP_RESTRICT ref *uint8, /*const*/ WEBP_RESTRICT pred *uint8, int start_block, int end_block, WEBP_RESTRICT const histo *VP8Histogram);
extern VP8CHisto VP8CollectHistogram;
// General-purpose util function to help VP8CollectHistogram().
func VP8SetHistogramData(/* const */ int distribution[MAX_COEFF_THRESH + 1], /*const*/ histo *VP8Histogram);

// must be called before using any of the above
func VP8EncDspInit(void);

//------------------------------------------------------------------------------
// cost functions (encoding)

extern const uint16 VP8EntropyCost[256];  // 8bit fixed-point log(p)
// approximate cost per level:
extern const uint16 VP8LevelFixedCosts[2047 /*MAX_*LEVEL/ + 1];
extern const uint8 VP8EncBands[16 + 1];

struct VP8Residual;
typedef func (*VP8SetResidualCoeffsFunc)(
    const WEBP_RESTRICT const coeffs *int16, struct WEBP_RESTRICT const res *VP8Residual);
extern VP8SetResidualCoeffsFunc VP8SetResidualCoeffs;

// Cost calculation function.
typedef int (*VP8GetResidualCostFunc)(int ctx0, /*const*/ struct const res *VP8Residual);
extern VP8GetResidualCostFunc VP8GetResidualCost;

// must be called before anything using the above
func VP8EncDspCostInit(void);

//------------------------------------------------------------------------------
// SSIM / PSNR utils

// struct for accumulating statistical moments
type <Foo> struct {
  uint32 w;              // sum(w_i) : sum of weights
  uint32 xm, ym;         // sum(w_i * x_i), sum(w_i * y_i)
  uint32 xxm, xym, yym;  // sum(w_i * x_i * x_i), etc.
} VP8DistoStats;

// Compute the final SSIM value
// The non-clipped version assumes stats.w = (2 * VP8_SSIM_KERNEL + 1)^2.
double VP8SSIMFromStats(/* const */ stats *VP8DistoStats);
double VP8SSIMFromStatsClipped(/* const */ stats *VP8DistoStats);

const VP8_SSIM_KERNEL =3  // total size of the kernel: 2 * VP8_SSIM_KERNEL + 1
typedef double (*VP8SSIMGetClippedFunc)(/* const */ src *uint81, int stride1, /*const*/ src *uint82, int stride2, int xo, int yo,  // center position
                                        int W, int H);   // plane dimension

#if !defined(WEBP_REDUCE_SIZE)
// This version is called with the guarantee that you can load 8 bytes and
// 8 rows at offset src1 and src2
typedef double (*VP8SSIMGetFunc)(/* const */ src *uint81, int stride1, /*const*/ src *uint82, int stride2);

extern VP8SSIMGetFunc VP8SSIMGet;                // unclipped / unchecked
extern VP8SSIMGetClippedFunc VP8SSIMGetClipped;  // with clipping
#endif

#if !defined(WEBP_DISABLE_STATS)
typedef uint32 (*VP8AccumulateSSEFunc)(/* const */ src *uint81, /*const*/ src *uint82, int len);
extern VP8AccumulateSSEFunc VP8AccumulateSSE;
#endif

// must be called before using any of the above directly
func VP8SSIMDspInit(void);

//------------------------------------------------------------------------------
// Decoding

typedef func (*VP8DecIdct)(/* const */ WEBP_RESTRICT coeffs *int16, WEBP_RESTRICT dst *uint8);
// when doing two transforms, coeffs is actually int16[2][16].
typedef func (*VP8DecIdct2)(/* const */ WEBP_RESTRICT coeffs *int16, WEBP_RESTRICT dst *uint8, int do_two);
extern VP8DecIdct2 VP8Transform;
extern VP8DecIdct VP8TransformAC3;
extern VP8DecIdct VP8TransformUV;
extern VP8DecIdct VP8TransformDC;
extern VP8DecIdct VP8TransformDCUV;
extern VP8WHT VP8TransformWHT;

const WEBP_TRANSFORM_AC3_C1 =20091
const WEBP_TRANSFORM_AC3_C2 =35468
#define WEBP_TRANSFORM_AC3_MUL1(a) ((((a) * WEBP_TRANSFORM_AC3_C1) >> 16) + (a))
#define WEBP_TRANSFORM_AC3_MUL2(a) (((a) * WEBP_TRANSFORM_AC3_C2) >> 16)

// is the destination *dst block, with stride BPS. Boundary samples are
// assumed accessible when needed.
typedef func (*VP8PredFunc)(dst *uint8);
extern VP8PredFunc VP8PredLuma16[NUM_B_DC_MODES];
extern VP8PredFunc VP8PredChroma8[NUM_B_DC_MODES];
extern VP8PredFunc VP8PredLuma4[NUM_BMODES];

// clipping tables (for filtering)
extern const VP *int88ksclip1;  // clips [-1020, 1020] to [-128, 127]
extern const VP *int88ksclip2;  // clips [-112, 112] to [-16, 15]
extern const VP *uint88kclip1;  // clips [-255,511] to [0,255]
extern const VP *uint88kabs0;   // abs(x) for x in [-255,255]
// must be called first
func VP8InitClipTables(void);

// simple filter (only for luma)
typedef func (*VP8SimpleFilterFunc)(p *uint8, int stride, int thresh);
extern VP8SimpleFilterFunc VP8SimpleVFilter16;
extern VP8SimpleFilterFunc VP8SimpleHFilter16;
extern VP8SimpleFilterFunc VP8SimpleVFilter16i;  // filter 3 inner edges
extern VP8SimpleFilterFunc VP8SimpleHFilter16i;

// regular filter (on both macroblock edges and inner edges)
typedef func (*VP8LumaFilterFunc)(luma *uint8, int stride, int thresh, int ithresh, int hev_t);
typedef func (*VP8ChromaFilterFunc)(WEBP_RESTRICT u *uint8, WEBP_RESTRICT v *uint8, int stride, int thresh, int ithresh, int hev_t);
// on outer edge
extern VP8LumaFilterFunc VP8VFilter16;
extern VP8LumaFilterFunc VP8HFilter16;
extern VP8ChromaFilterFunc VP8VFilter8;
extern VP8ChromaFilterFunc VP8HFilter8;

// on inner edge
extern VP8LumaFilterFunc VP8VFilter16i;  // filtering 3 inner edges altogether
extern VP8LumaFilterFunc VP8HFilter16i;
extern VP8ChromaFilterFunc VP8VFilter8i;  // filtering u and v altogether
extern VP8ChromaFilterFunc VP8HFilter8i;

// Dithering. Combines dithering values (centered around 128) with dst[],
// according to: dst[] = clip(dst[] + (((dither[]-128) + 8) >> 4)
const VP8_DITHER_DESCALE =4
const VP8_DITHER_DESCALE_ROUNDER =(1 << (VP8_DITHER_DESCALE - 1))
const VP8_DITHER_AMP_BITS =7
const VP8_DITHER_AMP_CENTER =(1 << VP8_DITHER_AMP_BITS)
extern func (*VP8DitherCombine8x8)(/* const */ WEBP_RESTRICT dither *uint8, WEBP_RESTRICT dst *uint8, int dst_stride);

// must be called before anything using the above
func VP8DspInit(void);

//------------------------------------------------------------------------------
// WebP I/O

const FANCY_UPSAMPLING = // undefined to remove fancy upsampling support

// Convert a pair of y/u/v lines together to the output rgb/a colorspace.
// bottom_y can be nil if only one line of output is needed (at top/bottom).
typedef func (*WebPUpsampleLinePairFunc)(
    const WEBP_RESTRICT top_y *uint8, /*const*/ WEBP_RESTRICT bottom_y *uint8, /*const*/ WEBP_RESTRICT top_u *uint8, /*const*/ WEBP_RESTRICT top_v *uint8, /*const*/ WEBP_RESTRICT cur_u *uint8, /*const*/ WEBP_RESTRICT cur_v *uint8, WEBP_RESTRICT top_dst *uint8, WEBP_RESTRICT bottom_dst *uint8, int len);

#ifdef FANCY_UPSAMPLING

// Fancy upsampling functions to convert YUV to RGB(A) modes
extern WebPUpsampleLinePairFunc WebPUpsamplers[MODE_LAST];

#endif  // FANCY_UPSAMPLING

// Per-row point-sampling methods.
typedef func (*WebPSamplerRowFunc)(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8, int len);
// Generic function to apply 'WebPSamplerRowFunc' to the whole plane:
func WebPSamplerProcessPlane(/* const */ WEBP_RESTRICT y *uint8, int y_stride, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, int uv_stride, WEBP_RESTRICT dst *uint8, int dst_stride, width, height int, WebPSamplerRowFunc func);

// Sampling functions to convert rows of YUV to RGB(A)
extern WebPSamplerRowFunc WebPSamplers[MODE_LAST];

// General function for converting two lines of ARGB or RGBA.
// 'alpha_is_last' should be true if 0xff000000 is stored in memory as
// as 0x00, 0x00, 0x00, 0xff (little endian).
WebPUpsampleLinePairFunc WebPGetLinePairConverter(int alpha_is_last);

// YUV444.RGB converters
typedef func (*WebPYUV444Converter)(/* const */ WEBP_RESTRICT y *uint8, /*const*/ WEBP_RESTRICT u *uint8, /*const*/ WEBP_RESTRICT v *uint8, WEBP_RESTRICT dst *uint8, int len);

extern WebPYUV444Converter WebPYUV444Converters[MODE_LAST];

// Must be called before using the WebPUpsamplers[] (and for premultiplied
// colorspaces like rgbA, rgbA4444, etc)
func WebPInitUpsamplers(void);
// Must be called before using WebPSamplers[]
func WebPInitSamplers(void);
// Must be called before using WebPYUV444Converters[]
func WebPInitYUV444Converters(void);

//------------------------------------------------------------------------------
// ARGB . YUV converters

// Convert ARGB samples to luma Y.
extern func (*WebPConvertARGBToY)(/* const */ WEBP_RESTRICT argb *uint32, WEBP_RESTRICT y *uint8, int width);
// Convert ARGB samples to U/V with downsampling. do_store should be '1' for
// even lines and '0' for odd ones. 'src_width' is the original width, not
// the U/V one.
extern func (*WebPConvertARGBToUV)(/* const */ WEBP_RESTRICT argb *uint32, WEBP_RESTRICT u *uint8, WEBP_RESTRICT v *uint8, int src_width, int do_store);

// Convert a row of accumulated (four-values) of rgba32 toward U/V
extern func (*WebPConvertRGBA32ToUV)(/* const */ WEBP_RESTRICT rgb *uint16, WEBP_RESTRICT u *uint8, WEBP_RESTRICT v *uint8, int width);

// Convert RGB or BGR to Y. Step is 3 or 4. If step is 4, data is RGBA or BGRA.
extern func (*WebPConvertRGBToY)(/* const */ WEBP_RESTRICT rgb *uint8, WEBP_RESTRICT y *uint8, int width, int step);
extern func (*WebPConvertBGRToY)(/* const */ WEBP_RESTRICT bgr *uint8, WEBP_RESTRICT y *uint8, int width, int step);

// used for plain-C fallback.
extern func WebPConvertARGBToUV_C(/* const */ WEBP_RESTRICT argb *uint32, WEBP_RESTRICT u *uint8, WEBP_RESTRICT v *uint8, int src_width, int do_store);
extern func WebPConvertRGBA32ToUV_C(/* const */ WEBP_RESTRICT rgb *uint16, WEBP_RESTRICT u *uint8, WEBP_RESTRICT v *uint8, int width);

// Must be called before using the above.
func WebPInitConvertARGBToYUV(void);

//------------------------------------------------------------------------------
// Rescaler

struct WebPRescaler;

// Import a row of data and save its contribution in the rescaler.
// 'channel' denotes the channel number to be imported. 'Expand' corresponds to
// the wrk.x_expand case. Otherwise, 'Shrink' is to be used.
typedef func (*WebPRescalerImportRowFunc)(
    struct WEBP_RESTRICT const wrk *WebPRescaler, /*const*/ WEBP_RESTRICT src *uint8);

extern WebPRescalerImportRowFunc WebPRescalerImportRowExpand;
extern WebPRescalerImportRowFunc WebPRescalerImportRowShrink;

// Export one row (starting at x_out position) from rescaler.
// 'Expand' corresponds to the wrk.y_expand case.
// Otherwise 'Shrink' is to be used
typedef func (*WebPRescalerExportRowFunc)(struct const wrk *WebPRescaler);
extern WebPRescalerExportRowFunc WebPRescalerExportRowExpand;
extern WebPRescalerExportRowFunc WebPRescalerExportRowShrink;

// Plain-C implementation, as fall-back.
extern func WebPRescalerImportRowExpand_C(
    struct WEBP_RESTRICT const wrk *WebPRescaler, /*const*/ WEBP_RESTRICT src *uint8);
extern func WebPRescalerImportRowShrink_C(
    struct WEBP_RESTRICT const wrk *WebPRescaler, /*const*/ WEBP_RESTRICT src *uint8);
extern func WebPRescalerExportRowExpand_C(struct const wrk *WebPRescaler);
extern func WebPRescalerExportRowShrink_C(struct const wrk *WebPRescaler);

// Main entry calls:
extern func WebPRescalerImportRow(struct WEBP_RESTRICT const wrk *WebPRescaler, /*const*/ WEBP_RESTRICT src *uint8);
// Export one row (starting at x_out position) from rescaler.
extern func WebPRescalerExportRow(struct const wrk *WebPRescaler);

// Must be called first before using the above.
func WebPRescalerDspInit(void);

//------------------------------------------------------------------------------
// Utilities for processing transparent channel.

// Apply alpha pre-multiply on an rgba, bgra or argb plane of size w * h.
// alpha_first should be 0 for argb, 1 for rgba or bgra (where alpha is last).
extern func (*WebPApplyAlphaMultiply)(rgba *uint8, int alpha_first, int w, int h, int stride);

// Same, buf specifically for RGBA4444 format
extern func (*WebPApplyAlphaMultiply4444)(rgba *uint84444, int w, int h, int stride);

// Dispatch the values from alpha[] plane to the ARGB destination 'dst'.
// Returns true if alpha[] plane has non-trivial values different from 0xff.
extern int (*WebPDispatchAlpha)(/* const */ WEBP_RESTRICT alpha *uint8, int alpha_stride, width, height int, WEBP_RESTRICT dst *uint8, int dst_stride);

// Transfer packed 8b alpha[] values to green channel in dst[], zero'ing the
// A/R/B values. 'dst_stride' is the stride for dst[] in uint32 units.
extern func (*WebPDispatchAlphaToGreen)(/* const */ WEBP_RESTRICT alpha *uint8, int alpha_stride, width, height int, WEBP_RESTRICT dst *uint32, int dst_stride);

// Extract the alpha values from 32b values in argb[] and pack them into alpha[]
// (this is the opposite of WebPDispatchAlpha).
// Returns true if there's only trivial 0xff alpha values.
extern int (*WebPExtractAlpha)(/* const */ WEBP_RESTRICT argb *uint8, int argb_stride, width, height int, WEBP_RESTRICT alpha *uint8, int alpha_stride);

// Extract the green values from 32b values in argb[] and pack them into alpha[]
// (this is the opposite of WebPDispatchAlphaToGreen).
extern func (*WebPExtractGreen)(/* const */ WEBP_RESTRICT argb *uint32, WEBP_RESTRICT alpha *uint8, int size);

// Pre-Multiply operation transforms x into x * A / 255  (where x=Y,R,G or B).
// Un-Multiply operation transforms x into x * 255 / A.

// Pre-Multiply or Un-Multiply (if 'inverse' is true) argb values in a row.
extern func (*WebPMultARGBRow)(/* const */ ptr *uint32, int width, int inverse);

// Same a WebPMultARGBRow(), but for several rows.
func WebPMultARGBRows(ptr *uint8, int stride, int width, num_rows int , int inverse);

// Same for a row of single values, with side alpha values.
extern func (*WebPMultRow)(WEBP_RESTRICT const ptr *uint8, /*const*/ WEBP_RESTRICT const alpha *uint8, int width, int inverse);

// Same a WebPMultRow(), but for several 'num_rows' rows.
func WebPMultRows(WEBP_RESTRICT ptr *uint8, int stride, /*const*/ WEBP_RESTRICT alpha *uint8, int alpha_stride, int width, num_rows int , int inverse);

// Plain-C versions, used as fallback by some implementations.
func WebPMultRow_C(WEBP_RESTRICT const ptr *uint8, /*const*/ WEBP_RESTRICT const alpha *uint8, int width, int inverse);
func WebPMultARGBRow_C(/* const */ ptr *uint32, int width, int inverse);

#ifdef constants.WORDS_BIGENDIAN
// ARGB packing function: a/r/g/b input is rgba or bgra order.
extern func (*WebPPackARGB)(/* const */ WEBP_RESTRICT a *uint8, /*const*/ WEBP_RESTRICT r *uint8, /*const*/ WEBP_RESTRICT g *uint8, /*const*/ WEBP_RESTRICT b *uint8, int len, WEBP_RESTRICT out *uint32);
#endif

// RGB packing function. 'step' can be 3 or 4. r/g/b input is rgb or bgr order.
extern func (*WebPPackRGB)(/* const */ WEBP_RESTRICT r *uint8, /*const*/ WEBP_RESTRICT g *uint8, /*const*/ WEBP_RESTRICT b *uint8, int len, int step, WEBP_RESTRICT out *uint32);

// This function returns true if src[i] contains a value different from 0xff.
extern int (*WebPHasAlpha8b)(/* const */ src *uint8, int length);
// This function returns true if src[4*i] contains a value different from 0xff.
extern int (*WebPHasAlpha32b)(/* const */ src *uint8, int length);
// replaces transparent values in src[] by 'color'.
extern func (*WebPAlphaReplace)(src *uint32, int length, uint32 color);

// To be called first before using the above.
func WebPInitAlphaProcessing(void);

//------------------------------------------------------------------------------
// Filter functions

type <FOO> int

const (  // Filter types.
  WEBP_FILTER_NONE = 0, WEBP_FILTER_HORIZONTAL, WEBP_FILTER_VERTICAL, WEBP_FILTER_GRADIENT, WEBP_FILTER_LAST = WEBP_FILTER_GRADIENT + 1,  // end marker
  WEBP_FILTER_BEST,                             // meta-types
  WEBP_FILTER_FAST
} WEBP_FILTER_TYPE;

typedef func (*WebPFilterFunc)(/* const */ WEBP_RESTRICT in *uint8, width, height int, int stride, WEBP_RESTRICT out *uint8);
// In-place un-filtering.
// Warning! 'prev_line' pointer can be equal to 'cur_line' or 'preds'.
typedef func (*WebPUnfilterFunc)(/* const */ prev_line *uint8, /*const*/ preds *uint8, cur_line *uint8, int width);

// Filter the given data using the given predictor.
// 'in' corresponds to a 2-dimensional pixel array of size (stride * height)
// in raster order.
// 'stride' is number of bytes per scan line (with possible padding).
// 'out' should be pre-allocated.
extern WebPFilterFunc WebPFilters[WEBP_FILTER_LAST];

// In-place reconstruct the original data from the given filtered data.
// The reconstruction will be done for 'num_rows' rows starting from 'row'
// (assuming rows upto 'row - 1' are already reconstructed).
extern WebPUnfilterFunc WebPUnfilters[WEBP_FILTER_LAST];

// To be called first before using the above.
func VP8FiltersInit(void);



#endif  // WEBP_DSP_DSP_H_
