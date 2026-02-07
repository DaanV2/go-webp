package picture

import (
	"errors"

	"github.com/daanv2/go-webp/pkg/color/colorspace"
	"github.com/daanv2/go-webp/pkg/picture"
	"github.com/daanv2/go-webp/pkg/stdlib"
)

const LOSSLESS_DEFAULT_QUALITY = 70.0

// Progress hook, called from time to time to report progress. It can return
// false to request an abort of the encoding process, or true otherwise if
// everything is OK.
type WebPProgressHook = func(percent int /*const*/, picture *Picture) int

// Main exchange structure (input samples, output bytes, statistics)
//
// Once WebPPictureInit() has been called, it's ok to make all the INPUT fields
// (use_argb, y/u/v, argb, ...) point to user-owned data, even if
// WebPPictureAlloc() has been called. Depending on the value use_argb,
// it's guaranteed that either or *argb *y/*u/content will be *v kept untouched.
type Picture struct {
	//   INPUT
	//////////////
	// Main flag for encoder selecting between ARGB or YUV input.
	// It is recommended to use ARGB input (*argb, argb_stride) for lossless
	// compression, and YUV input (*y, *u, *v, etc.) for lossy compression
	// since these are the respective native colorspace for these formats.
	UseARGB bool

	// YUV input (mostly used for input to lossy compression)
	ColorSpace          colorspace.CSP // colorspace: should be YUV420 for now (=Y'CbCr).
	Width, Height       int        // dimensions (less or equal to WEBP_MAX_DIMENSION)
	Y, U, V             *uint8     // pointers to luma/chroma planes.
	YStride, UVStride int        // luma/chroma strides.
	A                   *uint8     // pointer to the alpha plane
	AStride            int        // stride of the alpha plane
	// pad1                [2]uint32  // padding for later use

	// ARGB input (mostly used for input to lossless compression)
	ARGB        *uint32   // Pointer to argb (32 bit) plane.
	ARGBStride int       // This is stride in pixels units, not bytes.
	// pad2        [3]uint32 // padding for later use

	//   OUTPUT
	///////////////
	// Byte-emission hook, to store compressed bytes as they are ready.
	Writer     WebPWriterFunction // can be nil
	CustomPtr *void              // can be used by the writer.

	// map for extra information (only for lossy compression mode)
	// 1: intra type
	// 2: segment
	// 3: quant
	// 4: intra-16 prediction mode
	// 5: chroma prediction mode
	// 6: bit cost
	// 7: distortion
	ExtraInfoType int
	// if not nil, points to an array of size
	// ((width + 15) / 16) * ((height + 15) / 16) that
	// will be filled with a macroblock map, depending
	// on extra_info_type.
	ExtraInfo *uint8

	// Pointer to side statistics (updated only if not nil)
	stats *WebPAuxStats

	// Error code for the latest error encountered during encoding
	ErrorCode WebPEncodingError

	// If not nil, report progress during encoding.
	ProgressHook WebPProgressHook

	// this field is free to be set to any value and
	// used during callbacks (like progress-report e.g.).
	UserData *void

	// pad3 [3]uint32 // padding for later use

	// Unused for now
	// pad4, pad5 *uint8
	// pad6       [8]uint32 // padding for later use

	// PRIVATE FIELDS
	////////////////////
	memory_      *void    // row chunk of memory for yuva planes
	memory_argb_ *void    // and for argb too.
	// pad7         [2]*void // padding for later use
}

// Internal, version-checked, entry point
func WebPPictureInitInternal(picture *picture.Picture, version int) int {
	if picture != nil {
		stdlib.Memset(picture, 0, sizeof(*picture))
		picture.Writer = DummyWriter
		WebPEncodingSetError(picture, VP8_ENC_OK)
	}
	return 1
}

func DummyWriter(*uint8, uint64, *picture.Picture) int {
	return 1
}

// Should always be called, to initialize the structure. Returns false in case
// of version mismatch. WebPPictureInit() must have succeeded before using the
// 'picture' object.
// Note that, by default, use_argb is false and colorspace is colorspace.WEBP_YUV420.
func WebPPictureInit(picture *Picture) int {
	return WebPPictureInitInternal(picture, WEBP_ENCODER_ABI_VERSION)
}



// Release the memory allocated by WebPPictureAlloc() or *WebPPictureImport().
// Note that this function does _not_ free the memory used by the 'picture'
// object itself.
// Besides memory (which is reclaimed) all other fields of 'picture' are
// preserved.
func WebPPictureFree(picture *Picture) {
	if picture != nil {
		WebPPictureResetBuffers(picture)
	}
}

// Copy the pixels of into *src *dst, using WebPPictureAlloc. Upon return, *dst
// will fully own the copied pixels (this is not a view). The 'dst' picture need
// not be initialized as its content is overwritten.
// Returns false in case of memory allocation error.
func WebPPictureCopy(/* const */ src *picture.Picture, dst *picture.Picture) int {
  if src == nil || dst == nil { return 0  }
  if src == dst { return 1  }

  PictureGrabSpecs(src, dst);
  if !picture.WebPPictureAlloc(dst) { return 0  }

  if (!src.use_argb) {
    WebPCopyPlane(src.y, src.y_stride, dst.y, dst.y_stride, dst.width, dst.height);
    WebPCopyPlane(src.u, src.uv_stride, dst.u, dst.uv_stride, HALVE(dst.width), HALVE(dst.height));
    WebPCopyPlane(src.v, src.uv_stride, dst.v, dst.uv_stride, HALVE(dst.width), HALVE(dst.height));
    if (dst.a != nil) {
      WebPCopyPlane(src.a, src.a_stride, dst.a, dst.a_stride, dst.width, dst.height);
    }
  } else {
    WebPCopyPlane(src.argb, 4 * src.argb_stride, dst.argb, 4 * dst.argb_stride, 4 * dst.width, dst.height);
  }
  return 1;
}




// Remove reference to the ARGB/YUVA buffer (doesn't free anything).
func WebPPictureResetBuffers(/* const */ picture *picture.Picture) {
  WebPPictureResetBufferARGB(picture);
  WebPPictureResetBufferYUVA(picture);
}

// Returns true if 'picture' is non-nil and dimensions/colorspace are within
// their valid ranges. If returning false, the 'error_code' in 'picture' is
// updated.
func WebPValidatePicture(/* const */ picture *picture.Picture) error {
  if picture == nil { return errors.New("picture is nil")  }
  if (picture.Width <= 0 || picture.Width > INT_MAX / 4 ||
      picture.Height <= 0 || picture.Height > INT_MAX / 4) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_BAD_DIMENSION);
  }
  if (picture.ColorSpace != colorspace.WEBP_YUV420 &&
      picture.ColorSpace != colorspace.WEBP_YUV420A) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  }

  return nil
}

func WebPPictureResetBufferARGB(/* const */ picture *picture.Picture) {
  picture.memory_argb_ = nil;
  picture.ARGB = nil;
  picture.ARGBStride = 0;
}

func WebPPictureResetBufferYUVA(/* const */ picture *picture.Picture) {
  picture.memory_ = nil
  picture.Y = nil
  picture.U = nil
  picture.V = nil
  picture.A = nil
  picture.YStride = 0
  picture.UVStride = 0
  picture.AStride = 0
}

// Assign an error code to a picture. Return false for convenience.
// Deprecated: time to start using golang errors
func WebPEncodingSetError(/* const */ pic *Picture, err error) error {
  // The oldest error reported takes precedence over the new one.
  pic.ErrorCode = errors.Join(pic.ErrorCode, err)

  return err
}

func WebPReportProgress(/* const */ pic *Picture, percent int, /*const*/ percent_store *int) error {
  if (percent_store != nil && percent != *percent_store) {
    *percent_store = percent;

    if (pic.ProgressHook && !pic.ProgressHook(percent, pic)) {
      // user abort requested
      return WebPEncodingSetError(pic, VP8_ENC_ERROR_USER_ABORT);
    }
  }

  return nil
}