package webp

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//   WebP encoder: main interface
//
// Author: Skal (pascal.massimino@gmail.com)

import (
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/stddef"
)



const WEBP_ENCODER_ABI_VERSION = constants.WEBP_ENCODER_ABI_VERSION


//------------------------------------------------------------------------------
// One-stop-shop call! No questions asked:

// Returns the size of the compressed data (pointed to by *output), or 0 if
// an error occurred. The compressed data must be released by the caller
// using the call 'WebPFree(*output)'.
// These functions compress using the lossy format, and the quality_factor
// can go from 0 (smaller output, lower quality) to 100 (best quality,
// larger output).
func WebPEncodeRGB(/* const */ rgb *uint8, width, height int, stride int, quality_factor float64 , output *uint8) uint64 {
	// TODO:
}
func WebPEncodeBGR(/* const */ bgr *uint8, width, height int, stride int, quality_factor float64 , output *uint8) uint64 {
	// TODO:
}
func WebPEncodeRGBA(/* const */ rgba *uint8, width, height int, stride int, quality_factor float64 , output *uint8) uint64 {
	// TODO:
}
func WebPEncodeBGRA(/* const */ bgra *uint8, width, height int, stride int, quality_factor float64 , output *uint8) uint64 {
	// TODO:
}

// These functions are the equivalent of the above, but compressing in a
// lossless manner. Files are usually larger than lossy format, but will
// not suffer any compression loss.
// Note these functions, like the lossy versions, use the library's default
// settings. For lossless this means 'exact' is disabled. RGB values in fully
// transparent areas (that is, areas with alpha values equal to 0) will be
// modified to improve compression. To afunc this, use WebPEncode() and set
// config.Config::exact to 1.
func WebPEncodeLosslessRGB(/* const */ rgb *uint8, width, height int, stride int, output *uint8) uint64 {
	// TODO:
}
func WebPEncodeLosslessBGR(/* const */ bgr *uint8, width, height int, stride int, output *uint8) uint64 {
	// TODO:
}
func WebPEncodeLosslessRGBA(/* const */ rgba *uint8, width, height int, stride int, output *uint8) uint64 {
	// TODO:
}
func WebPEncodeLosslessBGRA(/* const */ bgra *uint8, width, height int, stride int, output *uint8) uint64 {
	// TODO:
}






// WebPMemoryWrite: a special WebPWriterFunction that writes to memory using
// the following WebPMemoryWriter object (to be set as a custom_ptr).
type WebPMemoryWriter struct {
  mem *uint8     // final buffer (of size 'max_size', larger than 'size').
  size uint64      // final size
  max_size uint64  // total capacity
  pad [1]uint32  // padding for later use
}

// The following must be called first before any use.
 func WebPMemoryWriterInit(writer *WebPMemoryWriter);

// The following must be called to deallocate writer.mem memory. The 'writer'
// object itself is not deallocated.
 func WebPMemoryWriterClear(writer *WebPMemoryWriter);





// maximum width/height allowed (inclusive), in pixels
const WEBP_MAX_DIMENSION =16383




// Compute the single distortion for packed planes of samples.
// 'src' will be compared to 'ref', and the raw distortion stored into
// '*distortion'. The refined metric (log(MSE), log(1 - ssim),...' will be
// stored in '*result'.
// 'x_step' is the horizontal stride (in bytes) between samples.
// 'src/ref_stride' is the byte distance between rows.
// Returns false in case of error (bad parameter, memory allocation error, ...).
int WebPPlaneDistortion(
/* const */ src *uint8, src_stride uint64 , /*const*/ ref *uint8, ref_stride uint64 , width, height int, x_step uint64 , vtype int,  // 0 = PSNR, 1 = SSIM, 2 = LSIM
distortion *float, result *float);

// Compute PSNR, SSIM or LSIM distortion metric between two pictures. Results
// are in dB, stored in result[] in the B/G/R/A/All order. The distortion is
// always performed using ARGB samples. Hence if the input is YUV(A), the
// picture will be internally converted to ARGB (just for the measurement).
// Warning: this function is rather CPU-intensive.
int WebPPictureDistortion(/* const */ src *picture.Picture, /*const*/ ref *picture.Picture, metric_type int,  // 0 = PSNR, 1 = SSIM, 2 = LSIM
    float result[5]);

// self-crops a picture to the rectangle defined by top/left/width/height.
// Returns false in case of memory allocation error, or if the rectangle is
// outside of the source picture.
// The rectangle for the view is defined by the top-left corner pixel
// coordinates (left, top) as well as its width and height. This rectangle
// must be fully be comprised inside the 'src' source picture. If the source
// picture uses the YUV420 colorspace, the top and left coordinates will be
// snapped to even values.
int WebPPictureCrop(picture *picture.Picture, left int, top int, width, height int);

// Extracts a view from 'src' picture into 'dst'. The rectangle for the view
// is defined by the top-left corner pixel coordinates (left, top) as well
// as its width and height. This rectangle must be fully be comprised inside
// the 'src' source picture. If the source picture uses the YUV420 colorspace,
// the top and left coordinates will be snapped to even values.
// Picture 'src' must out-live 'dst' picture. Self-extraction of view is allowed
// ('src' equal to 'dst') as a mean of fast-cropping (but note that doing so,
// the original dimension will be lost). Picture 'dst' need not be initialized
// with picture.WebPPictureInit() if it is different from 'src', since its content will
// be overwritten.
// Returns false in case of invalid parameters.
int WebPPictureView(/* const */ src *picture.Picture, left int, top int, width, height int, dst *picture.Picture);

// Returns true if the 'picture' is actually a view and therefore does
// not own the memory for pixels.
int WebPPictureIsView(/* const */ picture *picture.Picture);

// Rescale a picture to new dimension width x height.
// If either 'width' or 'height' (but not both) is 0 the corresponding
// dimension will be calculated preserving the aspect ratio.
// No gamma correction is applied.
// Returns false in case of error (invalid parameter or insufficient memory).
int WebPPictureRescale(picture *picture.Picture, width, height int);

// Colorspace conversion function to import RGB samples.
// Previous buffer will be free'd, if any.
// buffer should have *rgb a size of at least height * rgb_stride.
// Returns false in case of memory error.
int WebPPictureImportRGB(picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int);
// Same, but for RGBA buffer.
int WebPPictureImportRGBA(picture *picture.Picture, /*const*/ rgba *uint8, rgba_stride int);
// Same, but for RGBA buffer. Imports the RGB direct from the 32-bit format
// input buffer ignoring the alpha channel. Avoids needing to copy the data
// to a temporary 24-bit RGB buffer to import the RGB only.
int WebPPictureImportRGBX(picture *picture.Picture, /*const*/ rgbx *uint8, rgbx_stride int);

// Variants of the above, but taking BGR(A|X) input.
int WebPPictureImportBGR(picture *picture.Picture, /*const*/ bgr *uint8, bgr_stride int);
int WebPPictureImportBGRA(picture *picture.Picture, /*const*/ bgra *uint8, bgra_stride int);
int WebPPictureImportBGRX(picture *picture.Picture, /*const*/ bgrx *uint8, bgrx_stride int);

// Converts picture.ARGB data to the YUV420A format. The 'colorspace'
// parameter is deprecated and should be equal to colorspace.WEBP_YUV420.
// Upon return, picture.UseARGB is set to false. The presence of real
// non-opaque transparent values is detected, and 'colorspace' will be
// adjusted accordingly. Note that this method is lossy.
// Returns false in case of error.
int WebPPictureARGBToYUVA(picture *picture.Picture, colorspace.CSP /*colorspace = WEBP_*YUV420*/);

// Same as picture.WebPPictureARGBToYUVA(), but the conversion is done using
// pseudo-random dithering with a strength 'dithering' between
// 0.0 (no dithering) and 1.0 (maximum dithering). This is useful
// for photographic picture.
int WebPPictureARGBToYUVADithered(picture *picture.Picture, colorspace.CSP colorspace, float dithering);

// Performs 'sharp' RGBA.YUVA420 downsampling and colorspace conversion
// Downsampling is handled with extra care in case of color clipping. This
// method is roughly 2x slower than picture.WebPPictureARGBToYUVA() but produces better
// and sharper YUV representation.
// Returns false in case of error.
int WebPPictureSharpARGBToYUVA(picture *picture.Picture);
// kept for backward compatibility:
int WebPPictureSmartARGBToYUVA(picture *picture.Picture);

// Converts picture.yuv to picture.ARGB and sets picture.UseARGB to true.
// The input format must be YUV_420 or YUV_420A. The conversion from YUV420 to
// ARGB incurs a small loss too.
// Note that the use of this colorspace is discouraged if one has access to the
// raw ARGB samples, since using YUV420 is comparatively lossy.
// Returns false in case of error.
int WebPPictureYUVAToARGB(picture *picture.Picture);

// Helper function: given a width x height plane of RGBA or YUV(A) samples
// clean-up or smoothen the YUV or RGB samples under fully transparent area,
// to help compressibility (no guarantee, though).
func WebPCleanupTransparentArea(picture *picture.Picture);

// Scan the picture 'picture' for the presence of non fully opaque alpha values.
// Returns true in such case. Otherwise returns false (indicating that the
// alpha plane can be ignored altogether e.g.).
int WebPPictureHasTransparency(/* const */ picture *picture.Picture);


