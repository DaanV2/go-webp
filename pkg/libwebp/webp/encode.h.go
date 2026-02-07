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


import "github.com/daanv2/go-webp/pkg/stddef"



const WEBP_ENCODER_ABI_VERSION = 0x0210  // MAJOR(8b) + MINOR(8b)


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
// WebPConfig::exact to 1.
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
// reference (and so one can make use of picture.custom_ptr).
type WebPWriterFunction = func(/* const */ data *uint8, data_size uint64, /*const*/ picture *WebPPicture) int 

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
// The custom writer to be used with WebPMemoryWriter as custom_ptr. Upon
// completion, writer.mem and writer.size will hold the coded data.
// writer.mem must be freed by calling WebPMemoryWriterClear.
  int WebPMemoryWrite(/* const */ data *uint8, data_size uint64, /*const*/ picture *WebPPicture);

// Progress hook, called from time to time to report progress. It can return
// false to request an abort of the encoding process, or true otherwise if
// everything is OK.
typedef int (*WebPProgressHook)(int percent, /*const*/ picture *WebPPicture);



// maximum width/height allowed (inclusive), in pixels
const WEBP_MAX_DIMENSION =16383

// Main exchange structure (input samples, output bytes, statistics)
//
// Once WebPPictureInit() has been called, it's ok to make all the INPUT fields
// (use_argb, y/u/v, argb, ...) point to user-owned data, even if
// WebPPictureAlloc() has been called. Depending on the value use_argb,
// it's guaranteed that either or *argb *y/*u/content will be *v kept untouched.
type WebPPicture struct {
  //   INPUT
  //////////////
  // Main flag for encoder selecting between ARGB or YUV input.
  // It is recommended to use ARGB input (*argb, argb_stride) for lossless
  // compression, and YUV input (*y, *u, *v, etc.) for lossy compression
  // since these are the respective native colorspace for these formats.
  use_argb int

  // YUV input (mostly used for input to lossy compression)
   colorspace WebPEncCSP    // colorspace: should be YUV420 for now (=Y'CbCr).
   width, height int        // dimensions (less or equal to WEBP_MAX_DIMENSION)
  y, u, v *uint8       // pointers to luma/chroma planes.
   y_stride, uv_stride int  // luma/chroma strides.
  a *uint8;               // pointer to the alpha plane
  a_stride int;             // stride of the alpha plane
   pad1 [2]uint32;         // padding for later use

  // ARGB input (mostly used for input to lossless compression)
  argb *uint32;    // Pointer to argb (32 bit) plane.
  argb_stride int   // This is stride in pixels units, not bytes.
   pad2 [3]uint32;  // padding for later use

  //   OUTPUT
  ///////////////
  // Byte-emission hook, to store compressed bytes as they are ready.
   writer WebPWriterFunction  // can be nil
  custom_ptr *void;           // can be used by the writer.

  // map for extra information (only for lossy compression mode)
  // 1: intra type
  // 2: segment
  // 3: quant
  // 4: intra-16 prediction mode
  // 5: chroma prediction mode
  // 6: bit cost
  // 7: distortion
  extra_info_type int
  // if not nil, points to an array of size
	// ((width + 15) / 16) * ((height + 15) / 16) that
	// will be filled with a macroblock map, depending
	// on extra_info_type.
  extra_info *uint8

  // Pointer to side statistics (updated only if not nil)
  stats *WebPAuxStats;

  // Error code for the latest error encountered during encoding
   error_code WebPEncodingError

  // If not nil, report progress during encoding.
   progress_hook WebPProgressHook

   // this field is free to be set to any value and
	// used during callbacks (like progress-report e.g.).
  user_data *void

  pad3 [3]uint32  // padding for later use

  // Unused for now
	pad4, pad5 *uint8
  pad6[8] uint32  // padding for later use

  // PRIVATE FIELDS
  ////////////////////
  memory_ *void       // row chunk of memory for yuva planes
  memory_argb_ *void  // and for argb too.
  pad7 [2]*void       // padding for later use
}

// Internal, version-checked, entry point
  int WebPPictureInitInternal(*WebPPicture, int);

// Should always be called, to initialize the structure. Returns false in case
// of version mismatch. WebPPictureInit() must have succeeded before using the
// 'picture' object.
// Note that, by default, use_argb is false and colorspace is WEBP_YUV420.
func WebPPictureInit(picture *WebPPicture) int {
  return WebPPictureInitInternal(picture, WEBP_ENCODER_ABI_VERSION);
}

//------------------------------------------------------------------------------
// WebPPicture utils

// Convenience allocation / deallocation based on picture.width/height:
// Allocate y/u/v buffers as per colorspace/width/height specification.
// Note! This function will free the previous buffer if needed.
// Returns false in case of memory error.
  int WebPPictureAlloc(picture *WebPPicture);

// Release the memory allocated by WebPPictureAlloc() or *WebPPictureImport().
// Note that this function does _not_ free the memory used by the 'picture'
// object itself.
// Besides memory (which is reclaimed) all other fields of 'picture' are
// preserved.
 func WebPPictureFree(picture *WebPPicture);

// Copy the pixels of into *src *dst, using WebPPictureAlloc. Upon return, *dst
// will fully own the copied pixels (this is not a view). The 'dst' picture need
// not be initialized as its content is overwritten.
// Returns false in case of memory allocation error.
  int WebPPictureCopy(/* const */ src *WebPPicture, dst *WebPPicture);

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
  int WebPPictureDistortion(
    const src *WebPPicture, /*const*/ ref *WebPPicture, metric_type int,  // 0 = PSNR, 1 = SSIM, 2 = LSIM
    float result[5]);

// self-crops a picture to the rectangle defined by top/left/width/height.
// Returns false in case of memory allocation error, or if the rectangle is
// outside of the source picture.
// The rectangle for the view is defined by the top-left corner pixel
// coordinates (left, top) as well as its width and height. This rectangle
// must be fully be comprised inside the 'src' source picture. If the source
// picture uses the YUV420 colorspace, the top and left coordinates will be
// snapped to even values.
  int WebPPictureCrop(picture *WebPPicture, left int, top int, width, height int);

// Extracts a view from 'src' picture into 'dst'. The rectangle for the view
// is defined by the top-left corner pixel coordinates (left, top) as well
// as its width and height. This rectangle must be fully be comprised inside
// the 'src' source picture. If the source picture uses the YUV420 colorspace,
// the top and left coordinates will be snapped to even values.
// Picture 'src' must out-live 'dst' picture. Self-extraction of view is allowed
// ('src' equal to 'dst') as a mean of fast-cropping (but note that doing so,
// the original dimension will be lost). Picture 'dst' need not be initialized
// with WebPPictureInit() if it is different from 'src', since its content will
// be overwritten.
// Returns false in case of invalid parameters.
  int WebPPictureView(/* const */ src *WebPPicture, left int, top int, width, height int, dst *WebPPicture);

// Returns true if the 'picture' is actually a view and therefore does
// not own the memory for pixels.
 int WebPPictureIsView(/* const */ picture *WebPPicture);

// Rescale a picture to new dimension width x height.
// If either 'width' or 'height' (but not both) is 0 the corresponding
// dimension will be calculated preserving the aspect ratio.
// No gamma correction is applied.
// Returns false in case of error (invalid parameter or insufficient memory).
  int WebPPictureRescale(picture *WebPPicture, width, height int);

// Colorspace conversion function to import RGB samples.
// Previous buffer will be free'd, if any.
// buffer should have *rgb a size of at least height * rgb_stride.
// Returns false in case of memory error.
  int WebPPictureImportRGB(picture *WebPPicture, /*const*/ rgb *uint8, rgb_stride int);
// Same, but for RGBA buffer.
  int WebPPictureImportRGBA(picture *WebPPicture, /*const*/ rgba *uint8, rgba_stride int);
// Same, but for RGBA buffer. Imports the RGB direct from the 32-bit format
// input buffer ignoring the alpha channel. Avoids needing to copy the data
// to a temporary 24-bit RGB buffer to import the RGB only.
  int WebPPictureImportRGBX(picture *WebPPicture, /*const*/ rgbx *uint8, rgbx_stride int);

// Variants of the above, but taking BGR(A|X) input.
  int WebPPictureImportBGR(picture *WebPPicture, /*const*/ bgr *uint8, bgr_stride int);
  int WebPPictureImportBGRA(picture *WebPPicture, /*const*/ bgra *uint8, bgra_stride int);
  int WebPPictureImportBGRX(picture *WebPPicture, /*const*/ bgrx *uint8, bgrx_stride int);

// Converts picture.argb data to the YUV420A format. The 'colorspace'
// parameter is deprecated and should be equal to WEBP_YUV420.
// Upon return, picture.use_argb is set to false. The presence of real
// non-opaque transparent values is detected, and 'colorspace' will be
// adjusted accordingly. Note that this method is lossy.
// Returns false in case of error.
  int WebPPictureARGBToYUVA(
    picture *WebPPicture, WebPEncCSP /*colorspace = WEBP_*YUV420/);

// Same as WebPPictureARGBToYUVA(), but the conversion is done using
// pseudo-random dithering with a strength 'dithering' between
// 0.0 (no dithering) and 1.0 (maximum dithering). This is useful
// for photographic picture.
  int WebPPictureARGBToYUVADithered(
    picture *WebPPicture, WebPEncCSP colorspace, float dithering);

// Performs 'sharp' RGBA.YUVA420 downsampling and colorspace conversion
// Downsampling is handled with extra care in case of color clipping. This
// method is roughly 2x slower than WebPPictureARGBToYUVA() but produces better
// and sharper YUV representation.
// Returns false in case of error.
  int WebPPictureSharpARGBToYUVA(picture *WebPPicture);
// kept for backward compatibility:
  int WebPPictureSmartARGBToYUVA(picture *WebPPicture);

// Converts picture.yuv to picture.argb and sets picture.use_argb to true.
// The input format must be YUV_420 or YUV_420A. The conversion from YUV420 to
// ARGB incurs a small loss too.
// Note that the use of this colorspace is discouraged if one has access to the
// raw ARGB samples, since using YUV420 is comparatively lossy.
// Returns false in case of error.
  int WebPPictureYUVAToARGB(picture *WebPPicture);

// Helper function: given a width x height plane of RGBA or YUV(A) samples
// clean-up or smoothen the YUV or RGB samples under fully transparent area,
// to help compressibility (no guarantee, though).
 func WebPCleanupTransparentArea(picture *WebPPicture);

// Scan the picture 'picture' for the presence of non fully opaque alpha values.
// Returns true in such case. Otherwise returns false (indicating that the
// alpha plane can be ignored altogether e.g.).
 int WebPPictureHasTransparency(/* const */ picture *WebPPicture);

// Remove the transparency information (if present) by blending the color with
// the background color 'background_rgb' (specified as 24bit RGB triplet).
// After this call, all alpha values are reset to 0xff.
 func WebPBlendAlpha(picture *WebPPicture, uint32 background_rgb);

//------------------------------------------------------------------------------
// Main call

// Main encoding call, after config and picture have been initialized.
// 'picture' must be less than 16384x16384 in dimension (cf WEBP_MAX_DIMENSION),
// and the 'config' object must be a valid one.
// Returns false in case of error, true otherwise.
// In case of error, picture.error_code is updated accordingly.
// 'picture' can hold the source samples in both YUV(A) or ARGB input, depending
// on the value of 'picture.use_argb'. It is highly recommended to use
// the former for lossy encoding, and the latter for lossless encoding
// (when config.lossless is true). Automatic conversion from one format to
// another is provided but they both incur some loss.
  int WebPEncode(/* const */ config *WebPConfig, picture *WebPPicture);

