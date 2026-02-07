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




// Signature for output function. Should return true if writing was successful.
// data/data_size is the segment of data to write, and 'picture' is for
// reference (and so one can make use of picture.CustomPtr).
type WebPWriterFunction = func(/* const */ data *uint8, data_size uint64, /*const*/ picture *picture.Picture) int 

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
  int picture.WebPPictureDistortion(
    const src *picture.Picture, /*const*/ ref *picture.Picture, metric_type int,  // 0 = PSNR, 1 = SSIM, 2 = LSIM
    float result[5]);

// self-crops a picture to the rectangle defined by top/left/width/height.
// Returns false in case of memory allocation error, or if the rectangle is
// outside of the source picture.
// The rectangle for the view is defined by the top-left corner pixel
// coordinates (left, top) as well as its width and height. This rectangle
// must be fully be comprised inside the 'src' source picture. If the source
// picture uses the YUV420 colorspace, the top and left coordinates will be
// snapped to even values.
  int picture.WebPPictureCrop(picture *picture.Picture, left int, top int, width, height int);

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
  int picture.WebPPictureView(/* const */ src *picture.Picture, left int, top int, width, height int, dst *picture.Picture);

// Returns true if the 'picture' is actually a view and therefore does
// not own the memory for pixels.
 int picture.WebPPictureIsView(/* const */ picture *picture.Picture);

// Rescale a picture to new dimension width x height.
// If either 'width' or 'height' (but not both) is 0 the corresponding
// dimension will be calculated preserving the aspect ratio.
// No gamma correction is applied.
// Returns false in case of error (invalid parameter or insufficient memory).
  int picture.WebPPictureRescale(picture *picture.Picture, width, height int);

// Colorspace conversion function to import RGB samples.
// Previous buffer will be free'd, if any.
// buffer should have *rgb a size of at least height * rgb_stride.
// Returns false in case of memory error.
  int picture.WebPPictureImportRGB(picture *picture.Picture, /*const*/ rgb *uint8, rgb_stride int);
// Same, but for RGBA buffer.
  int picture.WebPPictureImportRGBA(picture *picture.Picture, /*const*/ rgba *uint8, rgba_stride int);
// Same, but for RGBA buffer. Imports the RGB direct from the 32-bit format
// input buffer ignoring the alpha channel. Avoids needing to copy the data
// to a temporary 24-bit RGB buffer to import the RGB only.
  int picture.WebPPictureImportRGBX(picture *picture.Picture, /*const*/ rgbx *uint8, rgbx_stride int);

// Variants of the above, but taking BGR(A|X) input.
  int picture.WebPPictureImportBGR(picture *picture.Picture, /*const*/ bgr *uint8, bgr_stride int);
  int picture.WebPPictureImportBGRA(picture *picture.Picture, /*const*/ bgra *uint8, bgra_stride int);
  int picture.WebPPictureImportBGRX(picture *picture.Picture, /*const*/ bgrx *uint8, bgrx_stride int);

// Converts picture.ARGB data to the YUV420A format. The 'colorspace'
// parameter is deprecated and should be equal to colorspace.WEBP_YUV420.
// Upon return, picture.UseARGB is set to false. The presence of real
// non-opaque transparent values is detected, and 'colorspace' will be
// adjusted accordingly. Note that this method is lossy.
// Returns false in case of error.
  int picture.WebPPictureARGBToYUVA(
    picture *picture.Picture, colorspace.CSP /*colorspace = WEBP_*YUV420/);

// Same as picture.WebPPictureARGBToYUVA(), but the conversion is done using
// pseudo-random dithering with a strength 'dithering' between
// 0.0 (no dithering) and 1.0 (maximum dithering). This is useful
// for photographic picture.
  int picture.WebPPictureARGBToYUVADithered(
    picture *picture.Picture, colorspace.CSP colorspace, float dithering);

// Performs 'sharp' RGBA.YUVA420 downsampling and colorspace conversion
// Downsampling is handled with extra care in case of color clipping. This
// method is roughly 2x slower than picture.WebPPictureARGBToYUVA() but produces better
// and sharper YUV representation.
// Returns false in case of error.
  int picture.WebPPictureSharpARGBToYUVA(picture *picture.Picture);
// kept for backward compatibility:
  int picture.WebPPictureSmartARGBToYUVA(picture *picture.Picture);

// Converts picture.yuv to picture.ARGB and sets picture.UseARGB to true.
// The input format must be YUV_420 or YUV_420A. The conversion from YUV420 to
// ARGB incurs a small loss too.
// Note that the use of this colorspace is discouraged if one has access to the
// raw ARGB samples, since using YUV420 is comparatively lossy.
// Returns false in case of error.
  int picture.WebPPictureYUVAToARGB(picture *picture.Picture);

// Helper function: given a width x height plane of RGBA or YUV(A) samples
// clean-up or smoothen the YUV or RGB samples under fully transparent area,
// to help compressibility (no guarantee, though).
 func WebPCleanupTransparentArea(picture *picture.Picture);

// Scan the picture 'picture' for the presence of non fully opaque alpha values.
// Returns true in such case. Otherwise returns false (indicating that the
// alpha plane can be ignored altogether e.g.).
 int picture.WebPPictureHasTransparency(/* const */ picture *picture.Picture);

// Remove the transparency information (if present) by blending the color with
// the background color 'background_rgb' (specified as 24bit RGB triplet).
// After this call, all alpha values are reset to 0xff.
 func WebPBlendAlpha(picture *picture.Picture, uint32 background_rgb);

//------------------------------------------------------------------------------
// Main call

// Main encoding call, after config and picture have been initialized.
// 'picture' must be less than 16384x16384 in dimension (cf WEBP_MAX_DIMENSION),
// and the 'config' object must be a valid one.
// Returns false in case of error, true otherwise.
// In case of error, picture.ErrorCode is updated accordingly.
// 'picture' can hold the source samples in both YUV(A) or ARGB input, depending
// on the value of 'picture.UseARGB'. It is highly recommended to use
// the former for lossy encoding, and the latter for lossless encoding
// (when config.Lossless is true). Automatic conversion from one format to
// another is provided but they both incur some loss.
  int WebPEncode(/* const */ config *config.Config, picture *picture.Picture);

