package picture

import (
	"unsafe"

	"github.com/daanv2/go-webp/pkg/color/colorspace"
	"github.com/daanv2/go-webp/pkg/picture"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

// Convenience allocation / deallocation based on picture.Width/height:
// Allocate y/u/v buffers as per colorspace/width/height specification.
// Note! This function will free the previous buffer if needed.
// Returns false in case of memory error.
func WebPPictureAlloc(pict *Picture) error {
	if pict != nil {
		WebPPictureFree(pict) // erase previous buffer

		if !pict.UseARGB {
			return WebPPictureAllocYUVA(pict)
		} else {
			return WebPPictureAllocARGB(pict)
		}
	}
	return nil
}

// Allocates YUVA buffer according to set width/height (previous one is always free'd).
// Uses picture.csp to determine whether an alpha buffer is needed.
// Preserves the ARGB buffer.
// Returns false in case of error (invalid param, out-of-memory).
func WebPPictureAllocYUVA( /* const */ picture *picture.Picture) error {
	has_alpha := int(picture.ColorSpace) & colorspace.WEBP_CSP_ALPHA_BIT
	width := picture.Width
	height := picture.Height
	y_stride := width
	uv_width := int(int64(width+1) >> 1)
	uv_height := int(int64(height+1) >> 1)
	uv_stride := uv_width
	var a_width, a_stride int
	var y_size, uv_size, a_size, total_size uint64
	var mem *uint8

	err := WebPValidatePicture(picture)
	if err != nil {
		return err
	}

	WebPPictureResetBufferYUVA(picture)

	// alpha
	a_width = tenary.If(has_alpha, width, 0)
	a_stride = a_width
	y_size = uint64(y_stride * height)
	uv_size = uint64(uv_stride * uv_height)
	a_size = uint64(a_stride * height)

	total_size = y_size + a_size + 2*uv_size

	// Security and validation checks
	if width <= 0 || height <= 0 || // luma/alpha param error
		uv_width <= 0 || uv_height <= 0 { // u/v param error
		return picture.SetEncodingError(picture.ENC_ERROR_BAD_DIMENSION)
	}
	// allocate a new buffer.

	//   mem = (*uint8)WebPSafeMalloc(total_size, sizeof(*mem));
	//   if (mem == nil) {
	//     return picture.SetEncodingError(picture.ENC_ERROR_OUT_OF_MEMORY)
	//   }
	mem := make([]uint8, total_size)

	// From now on, we're in the clear, we can no longer fail...
	// C: picture.memory_ = (void*)mem
	picture.memory_ = unsafe.Pointer(&mem[0])
	picture.YStride = y_stride
	picture.UVStride = uv_stride
	picture.AStride = a_stride

	// TODO(skal): we could align the y/u/v planes and adjust stride.
	// C: picture.Y = mem
	// C: mem += y_size
	// C: picture.U = mem
	// C: mem += uv_size
	// C: picture.V = mem
	// C: mem += uv_size
	// C: if (a_size > 0) {
	// C:   picture.A = mem
	// C:   mem += a_size
	// C: }
	// C: (void)mem  // makes the static analyzer happy
	return 1
}

// Allocates ARGB buffer according to set width/height (previous one is
// always free'd). Preserves the YUV(A) buffer. Returns false in case of error
// (invalid param, out-of-memory).
func WebPPictureAllocARGB( /* const */ picture *picture.Picture) error {
	width := picture.Width
	height := picture.Height
	argb_size := uint64(width * height)

	err := WebPValidatePicture(picture)
	if err != nil {
		return err
	}

	WebPPictureResetBufferARGB(picture)

	// allocate a new buffer.
	//   memory = WebPSafeMalloc(argb_size + WEBP_ALIGN_CST, sizeof(*picture.ARGB));
	//   if (memory == nil) {
	//     return picture.SetEncodingError(picture.ENC_ERROR_OUT_OF_MEMORY)
	//   }
	memory := make([]uint8, argb_size+WEBP_ALIGN_CST)

	picture.memory_argb_ = unsafe.Pointer(&memory[0])
	// C: picture.ARGB = (uint32*)WEBP_ALIGN(memory)
	picture.ARGBStride = width

	return nil
}
