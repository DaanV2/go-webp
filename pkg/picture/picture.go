package picture

import (
	"github.com/daanv2/go-webp/pkg/stdlib"
)

const LOSSLESS_DEFAULT_QUALITY = 70.0

// Progress hook, called from time to time to report progress. It can return
// false to request an abort of the encoding process, or true otherwise if
// everything is OK.
type WebPProgressHook = func(percent int /*const*/, picture *WebPPicture) int

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
	colorspace          WebPEncCSP // colorspace: should be YUV420 for now (=Y'CbCr).
	width, height       int        // dimensions (less or equal to WEBP_MAX_DIMENSION)
	y, u, v             *uint8     // pointers to luma/chroma planes.
	y_stride, uv_stride int        // luma/chroma strides.
	a                   *uint8     // pointer to the alpha plane
	a_stride            int        // stride of the alpha plane
	pad1                [2]uint32  // padding for later use

	// ARGB input (mostly used for input to lossless compression)
	argb        *uint32   // Pointer to argb (32 bit) plane.
	argb_stride int       // This is stride in pixels units, not bytes.
	pad2        [3]uint32 // padding for later use

	//   OUTPUT
	///////////////
	// Byte-emission hook, to store compressed bytes as they are ready.
	writer     WebPWriterFunction // can be nil
	custom_ptr *void              // can be used by the writer.

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
	stats *WebPAuxStats

	// Error code for the latest error encountered during encoding
	error_code WebPEncodingError

	// If not nil, report progress during encoding.
	progress_hook WebPProgressHook

	// this field is free to be set to any value and
	// used during callbacks (like progress-report e.g.).
	user_data *void

	pad3 [3]uint32 // padding for later use

	// Unused for now
	pad4, pad5 *uint8
	pad6       [8]uint32 // padding for later use

	// PRIVATE FIELDS
	////////////////////
	memory_      *void    // row chunk of memory for yuva planes
	memory_argb_ *void    // and for argb too.
	pad7         [2]*void // padding for later use
}

// Internal, version-checked, entry point
func WebPPictureInitInternal(picture *picture.WebPPicture, version int) int {
	if picture != nil {
		stdlib.Memset(picture, 0, sizeof(*picture))
		picture.writer = DummyWriter
		WebPEncodingSetError(picture, VP8_ENC_OK)
	}
	return 1
}

func DummyWriter(*uint8, uint64, *picture.WebPPicture) int {
	return 1
}

// Should always be called, to initialize the structure. Returns false in case
// of version mismatch. WebPPictureInit() must have succeeded before using the
// 'picture' object.
// Note that, by default, use_argb is false and colorspace is WEBP_YUV420.
func WebPPictureInit(picture *WebPPicture) int {
	return WebPPictureInitInternal(picture, WEBP_ENCODER_ABI_VERSION)
}

// Convenience allocation / deallocation based on picture.width/height:
// Allocate y/u/v buffers as per colorspace/width/height specification.
// Note! This function will free the previous buffer if needed.
// Returns false in case of memory error.
func WebPPictureAlloc(pict *WebPPicture) int {
	if pict != nil {
		WebPPictureFree(pict) // erase previous buffer

		if !pict.use_argb {
			return WebPPictureAllocYUVA(pict)
		} else {
			return WebPPictureAllocARGB(pict)
		}
	}
	return 1
}

// Release the memory allocated by WebPPictureAlloc() or *WebPPictureImport().
// Note that this function does _not_ free the memory used by the 'picture'
// object itself.
// Besides memory (which is reclaimed) all other fields of 'picture' are
// preserved.
func WebPPictureFree(picture *WebPPicture) {
	if picture != nil {
		WebPPictureResetBuffers(picture)
	}
}

// Copy the pixels of into *src *dst, using WebPPictureAlloc. Upon return, *dst
// will fully own the copied pixels (this is not a view). The 'dst' picture need
// not be initialized as its content is overwritten.
// Returns false in case of memory allocation error.
func WebPPictureCopy(/* const */ src *picture.WebPPicture, dst *picture.WebPPicture) int {
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

// Allocates YUVA buffer according to set width/height (previous one is always
// free'd). Uses picture.csp to determine whether an alpha buffer is needed.
// Preserves the ARGB buffer.
// Returns false in case of error (invalid param, out-of-memory).
func WebPPictureAllocYUVA(/* const */ picture *picture.WebPPicture) int {
  has_alpha := int(picture.colorspace) & WEBP_CSP_ALPHA_BIT;
  width := picture.width;
  height := picture.height;
  y_stride := width;
  uv_width := int(int64(width + 1) >> 1)
  uv_height := int(int64(height + 1) >> 1)
  uv_stride := uv_width;
  var a_width, a_stride int
  var y_size, uv_size, a_size, total_size uint64
  var mem *uint8;

  if !WebPValidatePicture(picture) { return 0  }

  WebPPictureResetBufferYUVA(picture);

  // alpha
  a_width = tenary.If(has_alpha, width, 0);
  a_stride = a_width;
  y_size = uint64(y_stride * height)
  uv_size = uint64(uv_stride * uv_height)
  a_size = uint64(a_stride * height)

  total_size = y_size + a_size + 2 * uv_size;

  // Security and validation checks
  if (width <= 0 || height <= 0 ||        // luma/alpha param error
      uv_width <= 0 || uv_height <= 0) {  // u/v param error
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_BAD_DIMENSION);
  }
  // allocate a new buffer.

  //   mem = (*uint8)WebPSafeMalloc(total_size, sizeof(*mem));
//   if (mem == nil) {
//     return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
//   }
  mem := make([]uint8, total_size)

  // From now on, we're in the clear, we can no longer fail...
  picture.memory_ = (*void)mem;
  picture.y_stride = y_stride;
  picture.uv_stride = uv_stride;
  picture.a_stride = a_stride;

  // TODO(skal): we could align the y/u/v planes and adjust stride.
  picture.y = mem;
  mem += y_size;

  picture.u = mem;
  mem += uv_size;
  picture.v = mem;
  mem += uv_size;

  if (a_size > 0) {
    picture.a = mem;
    mem += a_size;
  }
  (void)mem;  // makes the static analyzer happy
  return 1;
}

// Allocates ARGB buffer according to set width/height (previous one is
// always free'd). Preserves the YUV(A) buffer. Returns false in case of error
// (invalid param, out-of-memory).
func WebPPictureAllocARGB(/* const */ picture *picture.WebPPicture) int {
  width := picture.width
  height := picture.height
  argb_size := uint64(width * height)

  if !WebPValidatePicture(picture) { return 0  }

  WebPPictureResetBufferARGB(picture);

  // allocate a new buffer.
//   memory = WebPSafeMalloc(argb_size + WEBP_ALIGN_CST, sizeof(*picture.argb));
//   if (memory == nil) {
//     return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
//   }
  memory := make([]uint8, argb_size + WEBP_ALIGN_CST)

  picture.memory_argb_ = memory;
  picture.argb = (*uint32)WEBP_ALIGN(memory);
  picture.argb_stride = width;
  return 1;
}

// Remove reference to the ARGB/YUVA buffer (doesn't free anything).
func WebPPictureResetBuffers(/* const */ picture *picture.WebPPicture) {
  WebPPictureResetBufferARGB(picture);
  WebPPictureResetBufferYUVA(picture);
}

// Returns true if 'picture' is non-nil and dimensions/colorspace are within
// their valid ranges. If returning false, the 'error_code' in 'picture' is
// updated.
func WebPValidatePicture(/* const */ picture *picture.WebPPicture) int {
  if picture == nil { return 0  }
  if (picture.width <= 0 || picture.width > INT_MAX / 4 ||
      picture.height <= 0 || picture.height > INT_MAX / 4) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_BAD_DIMENSION);
  }
  if (picture.colorspace != WEBP_YUV420 &&
      picture.colorspace != WEBP_YUV420A) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  }
  return 1;
}

func WebPPictureResetBufferARGB(/* const */ picture *picture.WebPPicture) {
  picture.memory_argb_ = nil;
  picture.argb = nil;
  picture.argb_stride = 0;
}

func WebPPictureResetBufferYUVA(/* const */ picture *picture.WebPPicture) {
  picture.memory_ = nil
  picture.y = nil
  picture.u = nil
  picture.v = nil
  picture.a = nil
  picture.y_stride = 0
  picture.uv_stride = 0
  picture.a_stride = 0
}