// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux


const ERROR_STR_MAX_LENGTH =100
const DELTA_INFINITY =(uint64(1) << 32)
const KEYFRAME_NONE =(-1)
const MAX_CACHED_FRAMES =30
// This value is used to match a later call to WebPReplaceTransparentPixels(),
// making it a no-op for lossless (see WebPEncode()).
const TRANSPARENT_COLOR =0x00000000


type WebPAnimEncoder struct {
	canvas_width int;                // Canvas width.
	canvas_height int;               // Canvas height.
	options WebPAnimEncoderOptions  // Global encoding options.

	last_config config.Config            // Cached in case a re-encode is needed.
	// If 'last_config' uses lossless, then
	// this config uses lossy and vice versa;
	// only valid if 'options.allow_mixed'
	// is true.
	last_config_reversed config.Config   

	curr_canvas *picture.Picture;  // Only pointer; we don't own memory.

	// Canvas buffers.
	curr_canvas_copy picture.Picture   // Possibly modified current canvas.
	// True if pixels in 'curr_canvas_copy'
	// differ from those in 'curr_canvas'.
	curr_canvas_copy_modified int  

	// Previous canvas (original animation).
	// Also used temporarily to store canvas_carryover_disposed pixel values.
	prev_canvas picture.Picture ;
	// canvas_carryover contains the previous original input frame's pixels
	// (prev_canvas) with some parts carried over from even earlier original input
	// frames, to approximate the current state of the canvas at decoding.
	// canvas_carryover is  compared to curr_canvas at encoding to see what parts
	// of the current frame are similar enough to not be explicitly encoded.
	canvas_carryover picture.Picture ;

	// Buffer of the size of a subframe, with one boolean value per pixel.
	// Used when encoding a subframe to remember the pixels that may change when
	// decoding that frame (0 means the pixel is explicitly encoded, 1 means
	// carrying over the pixel value of the previous frame).
	candidate_carryover_mask *uint8;
	// True if at least one pixel is carried over by the best candidate subframe.
	best_candidate_carries_over int ;
	// Same as candidate_carryover_mask but for the best candidate subframe.
	best_candidate_carryover_mask *uint8;

	// Encoded data.
	encoded_frames []EncodedFrame;  // Array of encoded frames.
	size uint64                   // Number of allocated frames.
	start uint64                  // Frame start index.
	count uint64                  // Number of valid frames.
	flush_count uint64            // If >0, 'flush_count' frames starting from
	// 'start' are ready to be added to mux.

	// keyframe related.
	// min(canvas size - frame size) over the frames.
	// Can be negative in certain cases due to
	// transparent pixels in a frame.
	best_delta int64         
	keyframe int               // Index of selected keyframe relative to 'start'.
	count_since_key_frame int  // Frames seen since the last keyframe.

	first_timestamp int           // Timestamp of the first frame.
	prev_timestamp int            // Timestamp of the last added frame.
	prev_candidate_undecided int  // True if it's not yet decided if previous
	// frame would be a subframe or a keyframe.

	prev_rect FrameRectangle ;  // Previous WebP frame rectangle. Only valid if
	// prev_candidate_undecided is true.

	// Misc.
	is_first_frame int  // True if first frame is yet to be added/being added.
	// True if WebPAnimEncoderAdd() has already been called
	// with a nil frame.
	got_nil_frame int  

	// Number of input frames processed so far.
	in_frame_count uint64
	// Number of frames added to mux so far. This may be
	// different from 'in_frame_count' due to merging.
	out_frame_count uint64

	mux *WebPMux;  // Muxer to assemble the WebP bitstream.
	error_str  string  // Error string. Empty if no error. used to be byte[ERROR_STR_MAX_LENGTH]
}


// Reset the counters in the WebPAnimEncoder.
func ResetCounters(/* const */ enc *WebPAnimEncoder) {
  enc.start = 0;
  enc.count = 0;
  enc.flush_count = 0;
  enc.best_delta = DELTA_INFINITY;
  enc.keyframe = KEYFRAME_NONE;
}

func DisableKeyframes(/* const */ enc_options *WebPAnimEncoderOptions) {
  enc_options.kmax = INT_MAX;
  enc_options.kmin = enc_options.kmax - 1;
}


func SanitizeEncoderOptions(/* const */ enc_options *WebPAnimEncoderOptions) {
  print_warning := enc_options.verbose;

  if (enc_options.minimize_size) {
    DisableKeyframes(enc_options);
  }

  if (enc_options.kmax == 1) {  // All frames will be keyframes.
    enc_options.kmin = 0;
    enc_options.kmax = 0;
    return;
  } else if (enc_options.kmax <= 0) {
    DisableKeyframes(enc_options);
    print_warning = 0;
  }

  if (enc_options.kmin >= enc_options.kmax) {
    enc_options.kmin = enc_options.kmax - 1;
    if (print_warning) {
      fprintf(stderr, "WARNING: Setting kmin = %d, so that kmin < kmax.\n", enc_options.kmin);
    }
  } else {
    kmin_limit := enc_options.kmax / 2 + 1;
    if (enc_options.kmin < kmin_limit && kmin_limit < enc_options.kmax) {
      // This ensures that enc.keyframe + kmin >= kmax is always true. So, we
      // can flush all the frames in the 'count_since_key_frame == kmax' case.
      enc_options.kmin = kmin_limit;
      if (print_warning) {
        fprintf(stderr, "WARNING: Setting kmin = %d, so that kmin >= kmax / 2 + 1.\n", enc_options.kmin);
      }
    }
  }
  // Limit the max number of frames that are allocated.
  if (enc_options.kmax - enc_options.kmin > MAX_CACHED_FRAMES) {
    enc_options.kmin = enc_options.kmax - MAX_CACHED_FRAMES;
    if (print_warning) {
      fprintf(stderr, "WARNING: Setting kmin = %d, so that kmax - kmin <= %d.\n", enc_options.kmin, MAX_CACHED_FRAMES);
    }
  }
  assert.Assert(enc_options.kmin < enc_options.kmax);
}

func DefaultEncoderOptions(/* const */ enc_options *WebPAnimEncoderOptions) {
  enc_options.anim_params.loop_count = 0;
  enc_options.anim_params.bgcolor = 0xffffffff;  // White.
  enc_options.minimize_size = 0;
  DisableKeyframes(enc_options);
  enc_options.allow_mixed = 0;
  enc_options.verbose = 0;
}

func WebPAnimEncoderOptionsInitInternal(enc_options *WebPAnimEncoderOptions, abi_version int) int {
  if (enc_options == nil) {
    return 0;
  }
  DefaultEncoderOptions(enc_options);
  return 1;
}



func ClearRectangle(/* const */ picture *picture.Picture, left int, top int, width, height int) {
  var j int
  for j = top; j < top + height; j++ {
    var dst *uint32 = picture.ARGB + j * picture.ARGBStride;
    var i int
    for i = left; i < left + width; i++ {
      dst[i] = TRANSPARENT_COLOR;
    }
  }
}

func WebPUtilClearPic(/* const */ picture *picture.Picture, /*const*/ rect *FrameRectangle) {
  if (rect != nil) {
    ClearRectangle(picture, rect.x_offset, rect.y_offset, rect.width, rect.height);
  } else {
    ClearRectangle(picture, 0, 0, picture.Width, picture.Height);
  }
}

func MarkNoError(/* const */ enc *WebPAnimEncoder) {
  enc.error_str = "";
}

func MarkError(/* const */ enc *WebPAnimEncoder, /*const*/ str *byte) {
  if (snprintf(enc.error_str, ERROR_STR_MAX_LENGTH, "%s.", str) < 0) {
    assert.Assert(0);  // FIX ME!
  }
}

func MarkError2(/* const */ enc *WebPAnimEncoder, /*const*/ str *byte, error_code int) {
  if (snprintf(enc.error_str, ERROR_STR_MAX_LENGTH, "%s: %d.", str, error_code) < 0) {
    assert.Assert(0);  // FIX ME!
  }
}

func WebPAnimEncoder(
    width, height int, /*const*/ enc_options *WebPAnimEncoderOptions, abi_version int) *WebPAnimEncoderNewInternal {
  var enc *WebPAnimEncoder;

  if (width <= 0 || height <= 0 ||
      (width * uint64(height)) >= MAX_IMAGE_AREA) {
    return nil;
  }

//   enc = (*WebPAnimEncoder)WebPSafeCalloc(1, sizeof(*enc));
//   if enc == nil { return nil }
  enc := &WebPAnimEncoder{}
  MarkNoError(enc);

  // Dimensions and options.
  *(*int)&enc.canvas_width = width;
  *(*int)&enc.canvas_height = height;
  if (enc_options != nil) {
    *(*WebPAnimEncoderOptions)&enc.options = *enc_options;
    SanitizeEncoderOptions((*WebPAnimEncoderOptions)&enc.options);
  } else {
    DefaultEncoderOptions((*WebPAnimEncoderOptions)&enc.options);
  }

  // Canvas buffers.
  if (!picture.WebPPictureInit(&enc.curr_canvas_copy) ||
      !picture.WebPPictureInit(&enc.prev_canvas) ||
      !picture.WebPPictureInit(&enc.canvas_carryover)) {
    goto Err;
  }
  enc.curr_canvas_copy.width = width;
  enc.curr_canvas_copy.height = height;
  enc.curr_canvas_copy.use_argb = 1;
  if (!picture.WebPPictureAlloc(&enc.curr_canvas_copy) ||
      !picture.WebPPictureCopy(&enc.curr_canvas_copy, &enc.prev_canvas) ||
      !picture.WebPPictureCopy(&enc.curr_canvas_copy, &enc.canvas_carryover)) {
    goto Err;
  }
  WebPUtilClearPic(&enc.canvas_carryover, nil);
  enc.curr_canvas_copy_modified = 1;

  // Allocate for the whole canvas so that it can be reused for any subframe.
//   enc.candidate_carryover_mask = (*uint8)WebPSafeMalloc(width * (uint64)height, sizeof(*enc.candidate_carryover_mask));
//   enc.best_candidate_carryover_mask = (*uint8)WebPSafeMalloc(width * (uint64)height, sizeof(*enc.best_candidate_carryover_mask));
//   if enc.candidate_carryover_mask == nil { goto Err }
//   if enc.best_candidate_carryover_mask == nil { goto Err }

  enc.candidate_carryover_mask = make([]uint8, width * height)
  enc.best_candidate_carryover_mask = make([]uint8, width * height)

  // Encoded frames.
  ResetCounters(enc);
  // Note: one extra storage is for the previous frame.
  enc.size = enc.options.kmax - enc.options.kmin + 1;
  // We need space for at least 2 frames. But when kmin, kmax are both zero, // enc.size will be 1. So we handle that special case below.
  if enc.size < 2 { enc.size = 2 }
//   enc.encoded_frames = (*EncodedFrame)WebPSafeCalloc(enc.size, sizeof(*enc.encoded_frames));
//   if enc.encoded_frames == nil { goto Err }
  enc.encoded_frames = make([]EncodedFrame, enc.size)

  enc.mux = WebPMuxNew();
  if enc.mux == nil { goto Err }

  enc.count_since_key_frame = 0;
  enc.first_timestamp = 0;
  enc.prev_timestamp = 0;
  enc.prev_candidate_undecided = 0;
  enc.is_first_frame = 1;
  enc.got_nil_frame = 0;

  return enc;  // All OK.

Err:
  WebPAnimEncoderDelete(enc);
  return nil;
}

// Release the data contained by 'encoded_frame'.
func FrameRelease(/* const */ encoded_frame *EncodedFrame) {
  if (encoded_frame != nil) {
    WebPDataClear(&encoded_frame.sub_frame.bitstream);
    WebPDataClear(&encoded_frame.key_frame.bitstream);
    stdlib.Memset(encoded_frame, 0, sizeof(*encoded_frame));
  }
}

func WebPAnimEncoderDelete(enc *WebPAnimEncoder) {
  if (enc != nil) {
    picture.WebPPictureFree(&enc.curr_canvas_copy);
    picture.WebPPictureFree(&enc.prev_canvas);
    picture.WebPPictureFree(&enc.canvas_carryover);
    if (enc.encoded_frames != nil) {
      var i uint64
      for i = 0; i < enc.size; i++ {
        FrameRelease(&enc.encoded_frames[i]);
      }
    }
    WebPMuxDelete(enc.mux);
  }
}

// -----------------------------------------------------------------------------
// Frame addition.

// Returns cached frame at the given 'position'.
static GetFrame *EncodedFrame(/* const */ enc *WebPAnimEncoder, uint64 position) {
  assert.Assert(enc.start + position < enc.size);
  return &enc.encoded_frames[enc.start + position];
}

typedef int (*ComparePixelsFunc)(/* const */ *uint32, int, /*const*/ *uint32, int, int, int);

// Returns true if 'length' number of pixels in 'src' and 'dst' are equal,
// assuming the given step sizes between pixels.
// 'max_allowed_diff' is unused and only there to allow function pointer use.
func ComparePixelsLossless(/* const */ src *uint32, src_step int, /*const*/ dst *uint32, dst_step int, length int, max_allowed_diff int) int {
  (void)max_allowed_diff;
  assert.Assert(length > 0);
  while (length-- > 0) {
    if (*src != *dst) {
      return 0;
    }
    src += src_step;
    dst += dst_step;
  }
  return 1;
}

// Helper to check if each channel in 'src' and 'dst' is at most off by
// 'max_allowed_diff'.
func PixelsAreSimilar(uint32 src, uint32 dst, max_allowed_diff int) int {
  src_a := (src >> 24) & 0xff;
  src_r := (src >> 16) & 0xff;
  src_g := (src >> 8) & 0xff;
  src_b := (src >> 0) & 0xff;
  dst_a := (dst >> 24) & 0xff;
  dst_r := (dst >> 16) & 0xff;
  dst_g := (dst >> 8) & 0xff;
  dst_b := (dst >> 0) & 0xff;

  return (src_a == dst_a) &&
         (abs(src_r - dst_r) * dst_a <= (max_allowed_diff * 255)) &&
         (abs(src_g - dst_g) * dst_a <= (max_allowed_diff * 255)) &&
         (abs(src_b - dst_b) * dst_a <= (max_allowed_diff * 255));
}

// Returns true if 'length' number of pixels in 'src' and 'dst' are within an
// error bound, assuming the given step sizes between pixels.
func ComparePixelsLossy(/* const */ src *uint32, src_step int, /*const*/ dst *uint32, dst_step int, length int, max_allowed_diff int) int {
  assert.Assert(length > 0);
  while (length-- > 0) {
    if (!PixelsAreSimilar(*src, *dst, max_allowed_diff)) {
      return 0;
    }
    src += src_step;
    dst += dst_step;
  }
  return 1;
}

func IsEmptyRect(/* const */ rect *FrameRectangle) int {
  return (rect.width == 0) || (rect.height == 0);
}

func QualityToMaxDiff(quality float) int {
  const double val = pow(quality / 100., 0.5);
  const double max_diff = 31 * (1 - val) + 1 * val;
  return (int)(max_diff + 0.5);
}

// Assumes that an initial valid guess of change rectangle 'rect' is passed.
func MinimizeChangeRectangle(/* const */ src *picture.Picture, /*const*/ dst *picture.Picture, /*const*/ rect *FrameRectangle, is_lossless bool, quality float) {
  int i, j;
  var compare_pixels ComparePixelsFunc =
      tenary.If(is_lossless, ComparePixelsLossless, ComparePixelsLossy);
  max_allowed_diff_lossy := QualityToMaxDiff(quality);
  max_allowed_diff := tenary.If(is_lossless, 0, max_allowed_diff_lossy);

  // Assumption/correctness checks.
  assert.Assert(src.width == dst.width && src.height == dst.height);
  assert.Assert(rect.x_offset + rect.width <= dst.width);
  assert.Assert(rect.y_offset + rect.height <= dst.height);

  // Left boundary.
  for i = rect.x_offset; i < rect.x_offset + rect.width; i++ {
    const src_argb *uint32 =
        &src.argb[rect.y_offset * src.argb_stride + i];
    const dst_argb *uint32 =
        &dst.argb[rect.y_offset * dst.argb_stride + i];
    if (compare_pixels(src_argb, src.argb_stride, dst_argb, dst.argb_stride, rect.height, max_allowed_diff)) {
      --rect.width;  // Redundant column.
      ++rect.x_offset;
    } else {
      break;
    }
  }
  if rect.width == 0 { goto NoChange }

  // Right boundary.
  for i = rect.x_offset + rect.width - 1; i >= rect.x_offset; --i {
    const src_argb *uint32 =
        &src.argb[rect.y_offset * src.argb_stride + i];
    const dst_argb *uint32 =
        &dst.argb[rect.y_offset * dst.argb_stride + i];
    if (compare_pixels(src_argb, src.argb_stride, dst_argb, dst.argb_stride, rect.height, max_allowed_diff)) {
      --rect.width;  // Redundant column.
    } else {
      break;
    }
  }
  if rect.width == 0 { goto NoChange }

  // Top boundary.
  for j = rect.y_offset; j < rect.y_offset + rect.height; j++ {
    const src_argb *uint32 =
        &src.argb[j * src.argb_stride + rect.x_offset];
    const dst_argb *uint32 =
        &dst.argb[j * dst.argb_stride + rect.x_offset];
    if (compare_pixels(src_argb, 1, dst_argb, 1, rect.width, max_allowed_diff)) {
      --rect.height;  // Redundant row.
      ++rect.y_offset;
    } else {
      break;
    }
  }
  if rect.height == 0 { goto NoChange }

  // Bottom boundary.
  for j = rect.y_offset + rect.height - 1; j >= rect.y_offset; --j {
    const src_argb *uint32 =
        &src.argb[j * src.argb_stride + rect.x_offset];
    const dst_argb *uint32 =
        &dst.argb[j * dst.argb_stride + rect.x_offset];
    if (compare_pixels(src_argb, 1, dst_argb, 1, rect.width, max_allowed_diff)) {
      --rect.height;  // Redundant row.
    } else {
      break;
    }
  }
  if rect.height == 0 { goto NoChange }

  if (IsEmptyRect(rect)) {
  NoChange:
    rect.x_offset = 0;
    rect.y_offset = 0;
    rect.width = 0;
    rect.height = 0;
  }
}

// Snap rectangle to even offsets (and adjust dimensions if needed).
func SnapToEvenOffsets(/* const */ rect *FrameRectangle) {
  rect.width += (rect.x_offset & 1);
  rect.height += (rect.y_offset & 1);
  rect.x_offset &= ~1;
  rect.y_offset &= ~1;
}

type SubFrameParams struct {
  should_try int               // Should try this set of parameters.
  empty_rect_allowed int       // Frame with empty rectangle can be skipped.
   rect_ll FrameRectangle       // Frame rectangle for lossless compression.
   sub_frame_ll picture.Picture     // subframe pic for lossless compression.
   // Frame rectangle for lossy compression.
                                // Could be smaller than 'rect_ll' as pixels
                                // with small diffs can be ignored.
   rect_lossy FrameRectangle    
   sub_frame_lossy picture.Picture  // subframe pic for lossy compression.
}

func SubFrameParamsInit(/* const */ params *SubFrameParams, should_try int, empty_rect_allowed int) int {
  params.should_try = should_try;
  params.empty_rect_allowed = empty_rect_allowed;
  if (!picture.WebPPictureInit(&params.sub_frame_ll) ||
      !picture.WebPPictureInit(&params.sub_frame_lossy)) {
    return 0;
  }
  return 1;
}

func SubFrameParamsFree(/* const */ params *SubFrameParams) {
  picture.WebPPictureFree(&params.sub_frame_ll);
  picture.WebPPictureFree(&params.sub_frame_lossy);
}

// Given previous and current canvas, picks the optimal rectangle for the
// current frame based on 'is_lossless' and other parameters. Assumes that the
// initial guess 'rect' is valid.
func GetSubRect(/* const */ prev_canvas *picture.Picture, /*const*/ curr_canvas *picture.Picture, is_key_frame int, is_first_frame int, empty_rect_allowed int, is_lossless bool, quality float, /*const*/ rect *FrameRectangle, /*const*/ sub_frame *picture.Picture) int {
  if (!is_key_frame || is_first_frame) {  // Optimize frame rectangle.
    // Note: This behaves as expected for first frame, as 'prev_canvas' is
    // initialized to a fully transparent canvas in the beginning.
    MinimizeChangeRectangle(prev_canvas, curr_canvas, rect, is_lossless, quality);
  }

  if (IsEmptyRect(rect)) {
    if (empty_rect_allowed) {  // No need to get 'sub_frame'.
      return 1;
    } else {  // Force a 1x1 rectangle.
      rect.width = 1;
      rect.height = 1;
      assert.Assert(rect.x_offset == 0);
      assert.Assert(rect.y_offset == 0);
    }
  }

  SnapToEvenOffsets(rect);
  return picture.WebPPictureView(curr_canvas, rect.x_offset, rect.y_offset, rect.width, rect.height, sub_frame);
}

// Picks optimal frame rectangle for both lossless and lossy compression. The
// initial guess for frame rectangles will be the full canvas.
func GetSubRects(/* const */ prev_canvas *picture.Picture, /*const*/ curr_canvas *picture.Picture, is_key_frame int, is_first_frame int, quality float, /*const*/ params *SubFrameParams) int {
  // Lossless frame rectangle.
  params.rect_ll.x_offset = 0;
  params.rect_ll.y_offset = 0;
  params.rect_ll.width = curr_canvas.width;
  params.rect_ll.height = curr_canvas.height;
  if (!GetSubRect(prev_canvas, curr_canvas, is_key_frame, is_first_frame, params.empty_rect_allowed, 1, quality, &params.rect_ll, &params.sub_frame_ll)) {
    return 0;
  }
  // Lossy frame rectangle.
  params.rect_lossy = params.rect_ll;  // seed with lossless rect.
  return GetSubRect(prev_canvas, curr_canvas, is_key_frame, is_first_frame, params.empty_rect_allowed, 0, quality, &params.rect_lossy, &params.sub_frame_lossy);
}

func clip(v, min_v,max_v int ) int {
  return (v < min_v) ? min_v : (v > max_v) ? max_v : v;
}

// Picks the optimal rectangle between two pictures, starting with initial
// values of offsets and dimensions that are passed in. The initial
// values will be clipped, if necessary, to make sure the rectangle is
// within the canvas. "use_argb" must be true for both pictures.
// Parameters:
//   prev_canvas, curr_canvas - (in) two input pictures to compare.
//   is_lossless, quality - (in) encoding settings.
//   x_offset, y_offset, width, height - (in/out) rectangle between the two
//                                                input pictures.
// Returns true on success.
func WebPAnimEncoderRefineRect(/* const */ prev_canvas *picture.Picture, /*const*/ curr_canvas *picture.Picture, is_lossless bool, quality float, /*const*/ x_offset *int, /*const*/ y_offset *int, /*const*/ width *int, /*const*/ height *int) int {
   var rect FrameRectangle
  int right, left, bottom, top;
  if (prev_canvas == nil || curr_canvas == nil ||
      prev_canvas.width != curr_canvas.width ||
      prev_canvas.height != curr_canvas.height || !prev_canvas.use_argb ||
      !curr_canvas.use_argb) {
    return 0;
  }
  right = clip(*x_offset + *width, 0, curr_canvas.width);
  left = clip(*x_offset, 0, curr_canvas.width - 1);
  bottom = clip(*y_offset + *height, 0, curr_canvas.height);
  top = clip(*y_offset, 0, curr_canvas.height - 1);
  rect.x_offset = left;
  rect.y_offset = top;
  rect.width = clip(right - left, 0, curr_canvas.width - rect.x_offset);
  rect.height = clip(bottom - top, 0, curr_canvas.height - rect.y_offset);
  MinimizeChangeRectangle(prev_canvas, curr_canvas, &rect, is_lossless, quality);
  SnapToEvenOffsets(&rect);
  *x_offset = rect.x_offset;
  *y_offset = rect.y_offset;
  *width = rect.width;
  *height = rect.height;
  return 1;
}

func DisposeFrameRectangle(dispose_method int, /*const*/ rect *FrameRectangle, /*const*/ curr_canvas *picture.Picture) {
  assert.Assert(rect != nil);
  if (dispose_method == WEBP_MUX_DISPOSE_BACKGROUND) {
    WebPUtilClearPic(curr_canvas, rect);
  }
}

func RectArea(/* const */ rect *FrameRectangle) uint32 {
  return (uint32)rect.width * rect.height;
}

func IsLosslessBlendingPossible(/* const */ src *picture.Picture, /*const*/ dst *picture.Picture, /*const*/ rect *FrameRectangle) int {
  int i, j;
  assert.Assert(src.width == dst.width && src.height == dst.height);
  assert.Assert(rect.x_offset + rect.width <= dst.width);
  assert.Assert(rect.y_offset + rect.height <= dst.height);
  for j = rect.y_offset; j < rect.y_offset + rect.height; j++ {
    for i = rect.x_offset; i < rect.x_offset + rect.width; i++ {
      src_pixel := src.argb[j * src.argb_stride + i];
      dst_pixel := dst.argb[j * dst.argb_stride + i];
      dst_alpha := dst_pixel >> 24;
      if (dst_alpha != 0xff && src_pixel != dst_pixel) {
        // In this case, if we use blending, we can't attain the desired
        // 'dst_pixel' value for this pixel. So, blending is not possible.
        return 0;
      }
    }
  }
  return 1;
}

func IsLossyBlendingPossible(/* const */ src *picture.Picture, /*const*/ dst *picture.Picture, /*const*/ rect *FrameRectangle, quality float) int {
  max_allowed_diff_lossy := QualityToMaxDiff(quality);
  int i, j;
  assert.Assert(src.width == dst.width && src.height == dst.height);
  assert.Assert(rect.x_offset + rect.width <= dst.width);
  assert.Assert(rect.y_offset + rect.height <= dst.height);
  for j = rect.y_offset; j < rect.y_offset + rect.height; j++ {
    for i = rect.x_offset; i < rect.x_offset + rect.width; i++ {
      src_pixel := src.argb[j * src.argb_stride + i];
      dst_pixel := dst.argb[j * dst.argb_stride + i];
      dst_alpha := dst_pixel >> 24;
      if (dst_alpha != 0xff &&
          !PixelsAreSimilar(src_pixel, dst_pixel, max_allowed_diff_lossy)) {
        // In this case, if we use blending, we can't attain the desired
        // 'dst_pixel' value for this pixel. So, blending is not possible.
        return 0;
      }
    }
  }
  return 1;
}

// For pixels in 'rect', replace those pixels in 'dst' that are same as 'src' by
// transparent pixels.
// Returns true if at least one pixel gets modified.
// Remember the modified pixel locations as 1s in carryover_mask.
func IncreaseTransparency(/* const */ src *picture.Picture, /*const*/ rect *FrameRectangle, /*const*/ dst *picture.Picture, /*const*/ carryover_mask *uint8) int {
  int i, j;
  modified := 0;
  // carryover_mask spans over the rect part of the canvas.
  carryover_row *uint8 = carryover_mask;
  assert.Assert(src != nil && dst != nil && rect != nil);
  assert.Assert(src.width == dst.width && src.height == dst.height);
  for j = rect.y_offset; j < rect.y_offset + rect.height; j++ {
    var psrc *uint32 = src.argb + j * src.argb_stride;
    var pdst *uint32 = dst.argb + j * dst.argb_stride;
    for i = rect.x_offset; i < rect.x_offset + rect.width; i++ {
      if (psrc[i] == pdst[i] && pdst[i] != TRANSPARENT_COLOR) {
        pdst[i] = TRANSPARENT_COLOR;
        carryover_row[i - rect.x_offset] = 1;
        modified = 1;
      }
    }
    carryover_row += rect.width;
  }
  return modified;
}

#undef TRANSPARENT_COLOR

// Replace similar blocks of pixels by a 'see-through' transparent block
// with uniform average color.
// Assumes lossy compression is being used.
// Returns true if at least one pixel gets modified.
// Remember the modified pixel locations as 1s in carryover_mask.
func FlattenSimilarBlocks(/* const */ src *picture.Picture, /*const*/ rect *FrameRectangle, /*const*/ dst *picture.Picture, quality float, /*const*/ carryover_mask *uint8) int {
  max_allowed_diff_lossy := QualityToMaxDiff(quality);
  int i, j;
  modified := 0;
  block_size := 8;
  y_start := (rect.y_offset + block_size) & ~(block_size - 1);
  y_end := (rect.y_offset + rect.height) & ~(block_size - 1);
  x_start := (rect.x_offset + block_size) & ~(block_size - 1);
  x_end := (rect.x_offset + rect.width) & ~(block_size - 1);
  // carryover_mask spans over the rect part of the canvas.
  carryover_mask_row *uint8 = carryover_mask +
                                (y_start - rect.y_offset) * rect.width +
                                (x_start - rect.x_offset);
  assert.Assert(src != nil && dst != nil && rect != nil);
  assert.Assert(src.width == dst.width && src.height == dst.height);
  assert.Assert((block_size & (block_size - 1)) == 0);  // must be a power of 2
  // Iterate over each block and count similar pixels.
  for j = y_start; j < y_end; j += block_size {
    carryover_mask_block *uint8 = carryover_mask_row;
    for i = x_start; i < x_end; i += block_size {
      cnt := 0;
      avg_r := 0, avg_g = 0, avg_b = 0;
      var x, y int
      var psrc *uint32 = src.argb + j * src.argb_stride + i;
      var pdst *uint32 = dst.argb + j * dst.argb_stride + i;
      for y = 0; y < block_size; y++ {
        for x = 0; x < block_size; x++ {
          src_pixel := psrc[x + y * src.argb_stride];
          alpha := src_pixel >> 24;
          if (alpha == 0xff &&
              PixelsAreSimilar(src_pixel, pdst[x + y * dst.argb_stride], max_allowed_diff_lossy)) {
            cnt++
            avg_r += (src_pixel >> 16) & 0xff;
            avg_g += (src_pixel >> 8) & 0xff;
            avg_b += (src_pixel >> 0) & 0xff;
          }
        }
      }
      // If we have a fully similar block, we replace it with an
      // average transparent block. This compresses better in lossy mode.
      if (cnt == block_size * block_size) {
        color := (0x00 << 24) | ((avg_r / cnt) << 16) |
                               ((avg_g / cnt) << 8) | ((avg_b / cnt) << 0);
        for y = 0; y < block_size; y++ {
          for x = 0; x < block_size; x++ {
            pdst[x + y * dst.argb_stride] = color;
            carryover_mask_block[y * rect.width + x] = 1;
          }
        }
        modified = 1;
      }
      carryover_mask_block += block_size;
    }
    carryover_mask_row += block_size * rect.width;
  }
  return modified;
}

func EncodeFrame(/* const */ config *config.Config, /*const*/ pic *picture.Picture, /*const*/ memory *WebPMemoryWriter) int {
  pic.use_argb = 1;
  pic.writer = WebPMemoryWrite;
  pic.custom_ptr = memory;
  if (!WebPEncode(config, pic)) {
    return 0;
  }
  return 1;
}

// Struct representing a candidate encoded frame including its metadata.
type Candidate struct {
	mem WebPMemoryWriter  // Encoded bytes.
	info WebPMuxFrameInfo
	rect FrameRectangle  // Coordinates and dimensions of this candidate.
	// True if at least one pixel in rect is carried over from
	// the previous frame, meaning at least one pixel was set
	// to fully transparent and this frame is blended.
	// If this is true, such pixels are marked as 1s in
	// WebPAnimEncoder::candidate_carryover_mask.
	carries_over int  
	evaluate int      // True if this candidate should be evaluated.
}

// Generates a candidate encoded frame given a picture and metadata.
func EncodeCandidate(/* const */ sub_frame *picture.Picture, /*const*/ rect *FrameRectangle, /*const*/ encoder_config *config.Config, use_blending int, /*const*/ candidate *Candidate) WebPEncodingError {
  config.Config config = *encoder_config;
  WebPEncodingError error_code = VP8_ENC_OK;
  assert.Assert(candidate != nil);
  stdlib.Memset(candidate, 0, sizeof(*candidate));

  // Set frame rect and info.
  candidate.rect = *rect;
  candidate.info.id = WEBP_CHUNK_ANMF;
  candidate.info.x_offset = rect.x_offset;
  candidate.info.y_offset = rect.y_offset;
  candidate.info.dispose_method = WEBP_MUX_DISPOSE_NONE;  // Set later.
  candidate.info.blend_method =
      tenary.If(use_blending, WEBP_MUX_BLEND, WEBP_MUX_NO_BLEND);
  candidate.info.duration = 0;  // Set in next call to WebPAnimEncoderAdd().

  // Encode picture.
  WebPMemoryWriterInit(&candidate.mem);

  if (!config.Lossless && use_blending) {
    // Disable filtering to afunc blockiness in reconstructed frames at the
    // time of decoding.
    config.Autofilter = 0;
    config.FilterStrength = 0;
  }
  if (!EncodeFrame(&config, sub_frame, &candidate.mem)) {
    error_code = sub_frame.ErrorCode;
    goto Err;
  }

  candidate.evaluate = 1;
  return error_code;

Err:
  WebPMemoryWriterClear(&candidate.mem);
  return error_code;
}

func CopyCurrentCanvas(/* const */ enc *WebPAnimEncoder) {
  if (enc.curr_canvas_copy_modified) {
    WebPCopyPixels(enc.curr_canvas, &enc.curr_canvas_copy);
    enc.curr_canvas_copy.ProgressHook = enc.curr_canvas.ProgressHook;
    enc.curr_canvas_copy.UserData = enc.curr_canvas.UserData;
    enc.curr_canvas_copy_modified = 0;
  }
}

enum {
  LL_DISP_NONE = 0, LL_DISP_BG, LOSSY_DISP_NONE, LOSSY_DISP_BG, CANDIDATE_COUNT
}

const MIN_COLORS_LOSSY =31      // Don't try lossy below this threshold.
const MAX_COLORS_LOSSLESS =194  // Don't try lossless above this threshold.

func GetEncodedData(/* const */ memory *WebPMemoryWriter, /*const*/ encoded_data *WebPData) {
  encoded_data.bytes = memory.mem;
  encoded_data.size = memory.size;
}

// Opposite of SetPreviousDisposeMethod().
func GetPreviousDisposeMethod(/* const */ enc *WebPAnimEncoder) WebPMuxAnimDispose {
  position := enc.count - 2;
  var prev_enc_frame *EncodedFrame = GetFrame(enc, position);
  assert.Assert(enc.count >= 2);  // As current and previous frames are in enc.
  return prev_enc_frame.is_key_frame
             ? prev_enc_frame.key_frame.dispose_method
             : prev_enc_frame.sub_frame.dispose_method;
}

// Sets dispose method of the previous frame to be 'dispose_method'.
func SetPreviousDisposeMethod(/* const */ enc *WebPAnimEncoder, WebPMuxAnimDispose dispose_method) {
  position := enc.count - 2;
  var prev_enc_frame *EncodedFrame = GetFrame(enc, position);
  assert.Assert(enc.count >= 2);  // As current and previous frames are in enc.

  if (enc.prev_candidate_undecided) {
    assert.Assert(dispose_method == WEBP_MUX_DISPOSE_NONE);
    prev_enc_frame.sub_frame.dispose_method = dispose_method;
    prev_enc_frame.key_frame.dispose_method = dispose_method;
  } else {
    var prev_info *WebPMuxFrameInfo = prev_enc_frame.is_key_frame
                                            ? &prev_enc_frame.key_frame
                                            : &prev_enc_frame.sub_frame;
    prev_info.dispose_method = dispose_method;
  }
}

// Pick the candidate encoded frame with smallest size and release other
// candidates.
// TODO(later): Perhaps a rough SSIM/PSNR produced by the encoder should
// also be a criteria, in addition to sizes.
func PickBestCandidate(/* const */ enc *WebPAnimEncoder, /*const*/ candidate *Candidate, WebPMuxAnimDispose dispose_method, is_key_frame int, *Candidate* const best_candidate, /*const*/ encoded_frame *EncodedFrame) {
  if (*best_candidate == nil ||
      candidate.mem.size < (*best_candidate).mem.size) {
    const dst *WebPMuxFrameInfo =
        is_key_frame ? &encoded_frame.key_frame : &encoded_frame.sub_frame;
    *dst = candidate.info;
    GetEncodedData(&candidate.mem, &dst.bitstream);
    if (!is_key_frame) {
      // Note: Previous dispose method only matters for non-keyframes.
      // Also, we don't want to modify previous dispose method that was
      // selected when a non keyframe was assumed.
      SetPreviousDisposeMethod(enc, dispose_method);

      enc.best_candidate_carries_over = candidate.carries_over;
      if (candidate.carries_over) {
        // Save the best_candidate_carryover_mask to be able to generate the
        // canvas_carryover of the next frame later in case this candidate stays
        // the best one.
        // Note: The canvas_carryover could contain the pixel values with loss
        //       due to quantization as if they were decoded, to even better
        //       estimate which areas of the canvas should be explicitly encoded
        //       at each frame. Setting config.Config::show_compressed=1 is a way
        //       to approximate this. That works poorly: in lossy mode most
        //       areas of the decoded canvas change because of quantization loss
        //       in a frame, resulting in the encoder trying to explicitly
        //       encode most of the following frame because most of the areas
        //       differ significantly between the previous decoded canvas and
        //       the current original canvas. This repeats for all frames, //       resulting in a very large encoded file for no visible benefit.
        //       The canvas carryover approach seems to fix
        //       https://issues.webmproject.org/42340478 while still staying
        //       close enough to the old behavior (only looking at the previous
        //       and current original input frames) to not break everything.
        // Save candidate_carryover_mask as best_candidate_carryover_mask by
        // swapping the two buffers.
        var tmp_carryover_mask *uint8 = enc.best_candidate_carryover_mask;
        enc.best_candidate_carryover_mask = enc.candidate_carryover_mask;
        enc.candidate_carryover_mask = tmp_carryover_mask;
      }
    }

    // Release the memory of the previous best candidate if any.
    if (*best_candidate != nil) {
      WebPMemoryWriterClear(&(*best_candidate).mem);
      (*best_candidate).evaluate = 0;
    }
    *best_candidate = candidate;
  } else {
    // Release the memory of the current candidate which is not the best one.
    WebPMemoryWriterClear(&candidate.mem);
    candidate.evaluate = 0;
  }
}

// Generates candidates for a given dispose method given pre-filled subframe
// 'params'.
static WebPEncodingError GenerateCandidates(
    const enc *WebPAnimEncoder, Candidate candidates[CANDIDATE_COUNT], WebPMuxAnimDispose dispose_method, /*const*/ canvas_carryover_disposed *picture.Picture, is_lossless bool, is_key_frame int, /*const*/ params *SubFrameParams, /*const*/ config_ll *config.Config, /*const*/ config_lossy *config.Config, *Candidate* const best_candidate, /*const*/ encoded_frame *EncodedFrame) {
  WebPEncodingError error_code = VP8_ENC_OK;
  is_dispose_none := (dispose_method == WEBP_MUX_DISPOSE_NONE);
  const candidate_ll *Candidate =
      is_dispose_none ? &candidates[LL_DISP_NONE] : &candidates[LL_DISP_BG];
  var candidate_lossy *Candidate = is_dispose_none
                                         ? &candidates[LOSSY_DISP_NONE]
                                         : &candidates[LOSSY_DISP_BG];
  var curr_canvas *picture.Picture = &enc.curr_canvas_copy;
  const canvas_carryover *picture.Picture =
      is_dispose_none ? &enc.canvas_carryover : canvas_carryover_disposed;
  int use_blending_ll, use_blending_lossy;
  int evaluate_ll, evaluate_lossy;

  CopyCurrentCanvas(enc);
  use_blending_ll =
      !is_key_frame && IsLosslessBlendingPossible(canvas_carryover, curr_canvas, &params.rect_ll);
  use_blending_lossy =
      !is_key_frame &&
      IsLossyBlendingPossible(canvas_carryover, curr_canvas, &params.rect_lossy, config_lossy.quality);

  // Pick candidates to be tried.
  if (!enc.options.allow_mixed) {
    evaluate_ll = is_lossless;
    evaluate_lossy = !is_lossless;
  } else if (enc.options.minimize_size) {
    evaluate_ll = 1;
    evaluate_lossy = 1;
  } else {  // Use a heuristic for trying lossless and/or lossy compression.
    num_colors := WebPGetColorPalette(&params.sub_frame_ll, nil);
    evaluate_ll = (num_colors < MAX_COLORS_LOSSLESS);
    evaluate_lossy = (num_colors >= MIN_COLORS_LOSSY);
  }

  // Generate candidates.
  if (evaluate_ll) {
    CopyCurrentCanvas(enc);
    if (use_blending_ll) {
      // Reset the whole carryover mask to "all pixels are explicitly encoded in
      // this current frame".
      stdlib.Memset(enc.candidate_carryover_mask, 0, params.rect_ll.width * params.rect_ll.height);
      enc.curr_canvas_copy_modified =
          IncreaseTransparency(canvas_carryover, &params.rect_ll, curr_canvas, enc.candidate_carryover_mask);
    }
    error_code = EncodeCandidate(&params.sub_frame_ll, &params.rect_ll, config_ll, use_blending_ll, candidate_ll);
    if error_code != VP8_ENC_OK { return error_code  }
    candidate_ll.carries_over = enc.curr_canvas_copy_modified;
    PickBestCandidate(enc, candidate_ll, dispose_method, is_key_frame, best_candidate, encoded_frame);
  }
  if (evaluate_lossy) {
    CopyCurrentCanvas(enc);
    if (use_blending_lossy) {
      // Reset the whole carryover mask to "all pixels are explicitly encoded in
      // this current frame".
      stdlib.Memset(enc.candidate_carryover_mask, 0, params.rect_lossy.width * params.rect_lossy.height);
      enc.curr_canvas_copy_modified = FlattenSimilarBlocks(
          canvas_carryover, &params.rect_lossy, curr_canvas, config_lossy.quality, enc.candidate_carryover_mask);
    }
    error_code =
        EncodeCandidate(&params.sub_frame_lossy, &params.rect_lossy, config_lossy, use_blending_lossy, candidate_lossy);
    if error_code != VP8_ENC_OK { return error_code  }
    candidate_lossy.carries_over = enc.curr_canvas_copy_modified;
    enc.curr_canvas_copy_modified = 1;
    PickBestCandidate(enc, candidate_lossy, dispose_method, is_key_frame, best_candidate, encoded_frame);
  }
  return error_code;
}

#undef MIN_COLORS_LOSSY
#undef MAX_COLORS_LOSSLESS

func IncreasePreviousDuration(/* const */ enc *WebPAnimEncoder, duration int) int {
  position := enc.count - 1;
  var prev_enc_frame *EncodedFrame = GetFrame(enc, position);
  var new_duration int

  assert.Assert(enc.count >= 1);
  assert.Assert(!prev_enc_frame.is_key_frame ||
         prev_enc_frame.sub_frame.duration ==
             prev_enc_frame.key_frame.duration);
  assert.Assert(prev_enc_frame.sub_frame.duration ==
         (prev_enc_frame.sub_frame.duration & (MAX_DURATION - 1)));
  assert.Assert(duration == (duration & (MAX_DURATION - 1)));

  new_duration = prev_enc_frame.sub_frame.duration + duration;
  if (new_duration >= MAX_DURATION) {  // Special case.
    // Separate out previous frame from earlier merged frames to afunc overflow.
    // We add a 1x1 transparent frame for the previous frame, with blending on.
    var rect FrameRectangle  = {0, 0, 1, 1}
    lossless_1x1_bytes[] := {
        0x52, 0x49, 0x46, 0x46, 0x14, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x4c, 0x08, 0x00, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00, 0x10, 0x88, 0x88, 0x08}
    var lossless_1x1 WebPData  = {lossless_1x1_bytes, sizeof(lossless_1x1_bytes)}
    lossy_1x1_bytes[] := {
        0x52, 0x49, 0x46, 0x46, 0x40, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x58, 0x0a, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, 0x4c, 0x50, 0x48, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x56, 0x50, 0x38, 0x20, 0x18, 0x00, 0x00, 0x00, 0x30, 0x01, 0x00, 0x9d, 0x01, 0x2a, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x34, 0x25, 0xa4, 0x00, 0x03, 0x70, 0x00, 0xfe, 0xfb, 0xfd, 0x50, 0x00}
    var lossy_1x1 WebPData  = {lossy_1x1_bytes, sizeof(lossy_1x1_bytes)}
    can_use_lossless :=
        (enc.last_config.Lossless || enc.options.allow_mixed);
    var curr_enc_frame *EncodedFrame = GetFrame(enc, enc.count);
    curr_enc_frame.is_key_frame = 0;
    curr_enc_frame.sub_frame.id = WEBP_CHUNK_ANMF;
    curr_enc_frame.sub_frame.x_offset = 0;
    curr_enc_frame.sub_frame.y_offset = 0;
    curr_enc_frame.sub_frame.dispose_method = WEBP_MUX_DISPOSE_NONE;
    curr_enc_frame.sub_frame.blend_method = WEBP_MUX_BLEND;
    curr_enc_frame.sub_frame.duration = duration;
    if (!WebPDataCopy(can_use_lossless ? &lossless_1x1 : &lossy_1x1, &curr_enc_frame.sub_frame.bitstream)) {
      return 0;
    }
    ++enc.count;
    ++enc.count_since_key_frame;
    enc.flush_count = enc.count - 1;
    enc.prev_candidate_undecided = 0;
    enc.prev_rect = rect;
  } else {  // Regular case.
    // Increase duration of the previous frame by 'duration'.
    prev_enc_frame.sub_frame.duration = new_duration;
    prev_enc_frame.key_frame.duration = new_duration;
  }
  return 1;
}

// Copies the pixels that are identical in 'a' and 'b' to 'dst'.
func CopyIdenticalPixels(/* const */ a *picture.Picture, /*const*/ b *picture.Picture, /*const*/ dst *picture.Picture) {
  int y, x;
  var row_a *uint32 = a.argb;
  var row_b *uint32 = b.argb;
  row_dst *uint32 = dst.argb;
  assert.Assert(a.width == b.width && a.height == b.height);
  assert.Assert(a.width == dst.width && a.height == dst.height);
  assert.Assert(a.use_argb && b.use_argb && dst.use_argb);

  for y = 0; y < a.height; y++ {
    for x = 0; x < a.width; x++ {
      if (row_a[x] == row_b[x]) {
        row_dst[x] = row_a[x];
      }
    }
    row_a += a.argb_stride;
    row_b += b.argb_stride;
    row_dst += dst.argb_stride;
  }
}

// Copies the pixels where 'mask' is 0 from 'src' to 'dst'.
func CopyMaskedPixels(/* const */ src *picture.Picture, /*const*/ mask *uint8, /*const*/ dst *picture.Picture) {
  int y, x;
  var row_src *uint32 = src.argb;
  var row_mask *uint8 = mask;
  row_dst *uint32 = dst.argb;
  assert.Assert(src.width == dst.width && src.height == dst.height);
  assert.Assert(src.use_argb && dst.use_argb);

  for y = 0; y < src.height; y++ {
    for x = 0; x < src.width; x++ {
      if (row_mask[x] == 0) {
        row_dst[x] = row_src[x];
      }
    }
    row_src += src.argb_stride;
    row_mask += src.width;
    row_dst += dst.argb_stride;
  }
}

// Depending on the configuration, tries different compressions
// (lossy/lossless), dispose methods, blending methods etc to encode the current
// frame and outputs the best one in 'encoded_frame'.
// 'frame_skipped' will be set to true if this frame should actually be skipped.
func SetFrame(/* const */ enc *WebPAnimEncoder, /*const*/ config *config.Config, is_key_frame int , /*const*/ best_candidate_rect *FrameRectangle, /*const*/ encoded_frame *EncodedFrame, /*const*/ frame_skipped *int) WebPEncodingError {
  var i int
  WebPEncodingError error_code = VP8_ENC_OK;
  var curr_canvas *picture.Picture = &enc.curr_canvas_copy;
  var canvas_carryover *picture.Picture = &enc.canvas_carryover;
  // canvas_carryover with the area corresponding to the previous frame disposed
  // to background color.
  canvas_carryover_disposed *picture.Picture = nil;
  Candidate candidates[CANDIDATE_COUNT];
  best_candidate *Candidate = nil;
  is_lossless := config.Lossless;
  consider_lossless := is_lossless || enc.options.allow_mixed;
  consider_lossy := !is_lossless || enc.options.allow_mixed;
  is_first_frame := enc.is_first_frame;

  // First frame cannot be skipped as there is no 'previous frame' to merge it
  // to. So, empty rectangle is not allowed for the first frame.
  empty_rect_allowed_none := !is_first_frame;

  // Even if there is exact pixel match between 'disposed previous canvas' and
  // 'current canvas', we can't skip current frame, as there may not be exact
  // pixel match between 'previous canvas' and 'current canvas'. So, we don't
  // allow empty rectangle in this case.
  empty_rect_allowed_bg := 0;

  // If current frame is a keyframe, dispose method of previous frame doesn't
  // matter, so we don't try dispose to background.
  // Also, if keyframe insertion is on, and previous frame could be picked as
  // either a subframe or a keyframe, then we can't be sure about what frame
  // rectangle would be disposed. In that case too, we don't try dispose to
  // background.
  dispose_bg_possible :=
      !is_key_frame && !enc.prev_candidate_undecided;

   var dispose_none_params SubFrameParams
   var dispose_bg_params SubFrameParams

  config.Config config_ll = *config;
  config.Config config_lossy = *config;
  config_ll.Lossless = 1;
  config_lossy.Lossless = 0;
  enc.last_config = *config;
  enc.last_config_reversed = tenary.If(config.Lossless, config_lossy, config_ll);
  *frame_skipped = 0;

  if (!SubFrameParamsInit(&dispose_none_params, 1, empty_rect_allowed_none) ||
      !SubFrameParamsInit(&dispose_bg_params, 0, empty_rect_allowed_bg)) {
    return VP8_ENC_ERROR_INVALID_CONFIGURATION;
  }

  stdlib.Memset(candidates, 0, sizeof(candidates));

  // Change-rectangle assuming previous frame was DISPOSE_NONE.
  if (!GetSubRects(canvas_carryover, curr_canvas, is_key_frame, is_first_frame, config_lossy.quality, &dispose_none_params)) {
    error_code = VP8_ENC_ERROR_INVALID_CONFIGURATION;
    goto Err;
  }

  if ((consider_lossless && IsEmptyRect(&dispose_none_params.rect_ll)) ||
      (consider_lossy && IsEmptyRect(&dispose_none_params.rect_lossy))) {
    // Don't encode the frame at all. Instead, the duration of the previous
    // frame will be increased later.
    assert.Assert(empty_rect_allowed_none);
    *frame_skipped = 1;
    goto End;
  }

  if (dispose_bg_possible) {
    // For memory optimization, canvas_carryover_disposed reuses the buffer from
    // enc.prev_canvas. This is safe because prev_canvas is not read again
    // before its contents are updated at the end of the CacheFrame() function.
    canvas_carryover_disposed = &enc.prev_canvas;
    // Change-rectangle assuming previous frame was DISPOSE_BACKGROUND.
    WebPCopyPixels(canvas_carryover, canvas_carryover_disposed);
    DisposeFrameRectangle(WEBP_MUX_DISPOSE_BACKGROUND, &enc.prev_rect, canvas_carryover_disposed);

    if (!GetSubRects(canvas_carryover_disposed, curr_canvas, is_key_frame, is_first_frame, config_lossy.quality, &dispose_bg_params)) {
      error_code = VP8_ENC_ERROR_INVALID_CONFIGURATION;
      goto Err;
    }
    assert.Assert(!IsEmptyRect(&dispose_bg_params.rect_ll));
    assert.Assert(!IsEmptyRect(&dispose_bg_params.rect_lossy));

    if (enc.options.minimize_size) {  // Try both dispose methods.
      dispose_bg_params.should_try = 1;
      dispose_none_params.should_try = 1;
    } else if ((is_lossless && RectArea(&dispose_bg_params.rect_ll) <
                                   RectArea(&dispose_none_params.rect_ll)) ||
               (!is_lossless &&
                RectArea(&dispose_bg_params.rect_lossy) <
                    RectArea(&dispose_none_params.rect_lossy))) {
      dispose_bg_params.should_try = 1;  // Pick DISPOSE_BACKGROUND.
      dispose_none_params.should_try = 0;
    }
  }

  if (dispose_none_params.should_try) {
    error_code =
        GenerateCandidates(enc, candidates, WEBP_MUX_DISPOSE_NONE, /*canvas_carryover_disposed=*/nil, is_lossless, is_key_frame, &dispose_none_params, &config_ll, &config_lossy, &best_candidate, encoded_frame);
    if error_code != VP8_ENC_OK { goto Err }
  }

  if (dispose_bg_params.should_try) {
    assert.Assert(!enc.is_first_frame);
    assert.Assert(dispose_bg_possible);
    error_code = GenerateCandidates(
        enc, candidates, WEBP_MUX_DISPOSE_BACKGROUND, canvas_carryover_disposed, is_lossless, is_key_frame, &dispose_bg_params, &config_ll, &config_lossy, &best_candidate, encoded_frame);
    if error_code != VP8_ENC_OK { goto Err }
  }

  assert.Assert(best_candidate != nil);
  *best_candidate_rect = best_candidate.rect;
  goto End;

Err:
  for i = 0; i < CANDIDATE_COUNT; i++ {
    if (candidates[i].evaluate) {
      WebPMemoryWriterClear(&candidates[i].mem);
    }
  }

End:
  SubFrameParamsFree(&dispose_none_params);
  SubFrameParamsFree(&dispose_bg_params);
  return error_code;
}

// Calculate the penalty incurred if we encode given frame as a keyframe
// instead of a subframe.
func KeyFramePenalty(/* const */ encoded_frame *EncodedFrame) int64 {
  return ((int64)encoded_frame.key_frame.bitstream.size - encoded_frame.sub_frame.bitstream.size);
}

func CacheFrame(/* const */ enc *WebPAnimEncoder, /*const*/ config *config.Config) int {
  ok := 0;
  frame_skipped := 0;
  WebPEncodingError error_code = VP8_ENC_OK;
  position := enc.count;
  var encoded_frame *EncodedFrame = GetFrame(enc, position);
  var best_key_candidate_rect, best_sub_candidate_rect FrameRectangle
  var candidate_undecided int

  enc.count++

  if (enc.is_first_frame) {  // Add this as a keyframe.
    error_code = SetFrame(enc, config, 1, &best_key_candidate_rect, encoded_frame, &frame_skipped);
    if error_code != VP8_ENC_OK { goto End }
    assert.Assert(frame_skipped == 0);  // First frame can't be skipped, even if empty.
    assert.Assert(position == 0 && enc.count == 1);
    encoded_frame.is_key_frame = 1;
    enc.flush_count = 0;
    enc.count_since_key_frame = 0;
    candidate_undecided = 0;
  } else {
    ++enc.count_since_key_frame;

    // When some parts of the current original input frame (curr_canvas) did not
    // change since the previous original input frame (prev_canvas), consider
    // the encoder already did its best job at encoding these parts, and there
    // is no need to explicitly encode these parts again. To afunc that, copy
    // these pixels from curr_canvas (or prev_canvas) to canvas_carryover so
    // that they are detected as unchanged in the SetFrame() implementation
    // below. If all parts are identical, the whole frame may be skipped.
    // TODO: Only allocate and use canvas_carryover for lossy and near-lossless.
    CopyIdenticalPixels(&enc.prev_canvas, enc.curr_canvas, &enc.canvas_carryover);

    if (enc.count_since_key_frame <= enc.options.kmin) {
      // Add this as a frame rectangle.
      error_code = SetFrame(enc, config, 0, &best_sub_candidate_rect, encoded_frame, &frame_skipped);
      if (error_code != VP8_ENC_OK){ goto End;}
      if frame_skipped { {goto Skip }}
      encoded_frame.is_key_frame = 0;
      enc.flush_count = enc.count - 1;
      candidate_undecided = 0;
    } else {
      var curr_delta int64

      // Add this as a frame rectangle to enc.
      // TODO: Only try to encode a subframe when it can be used (for example
      //       only when enc.count_since_key_frame < enc.options.kmax ||
      //       enc.best_delta < DELTA_INFINITY).
      //       frame_skipped should still be tested to keep exact same behavior.
      error_code = SetFrame(enc, config, 0, &best_sub_candidate_rect, encoded_frame, &frame_skipped);
      if error_code != VP8_ENC_OK { goto End }
      if frame_skipped { goto Skip }

      // Add this as a keyframe to enc, too.
      error_code = SetFrame(enc, config, 1, &best_key_candidate_rect, encoded_frame, &frame_skipped);
      if error_code != VP8_ENC_OK { goto End }
      assert.Assert(frame_skipped == 0);  // keyframe cannot be an empty rectangle.

      // Analyze size difference of the two variants.
      curr_delta = KeyFramePenalty(encoded_frame);
      if (curr_delta <= enc.best_delta) {  // Pick this as the keyframe.
        if (enc.keyframe != KEYFRAME_NONE) {
          var old_keyframe *EncodedFrame = GetFrame(enc, enc.keyframe);
          assert.Assert(old_keyframe.is_key_frame);
          old_keyframe.is_key_frame = 0;
        }
        encoded_frame.is_key_frame = 1;
        candidate_undecided = 1;
        enc.keyframe = (int)position;
        enc.best_delta = curr_delta;
        enc.flush_count = enc.count - 1;  // We can flush previous frames.
      } else {
        encoded_frame.is_key_frame = 0;
        candidate_undecided = 0;
      }
      // Note: We need '>=' below because when kmin and kmax are both zero, // count_since_key_frame will always be > kmax.
      if (enc.count_since_key_frame >= enc.options.kmax) {
        // Start a new sequence of kmin subframes, followed by (kmax-kmin)
        // candidate frames. Exactly one of these candidate frames will end up
        // as a keyframe in the output encoded animation.
        enc.count_since_key_frame = 0;
        enc.best_delta = DELTA_INFINITY;
        // Freeze the previous candidate, whether it is a keyframe or not.
        candidate_undecided = 0;
        enc.keyframe = KEYFRAME_NONE;
        // Flush all previous frames.
        enc.flush_count = enc.count - 1;
      }
    }
  }

  if (encoded_frame.is_key_frame) {
    // A keyframe does not carry any pixels over from previous frames.
    WebPCopyPixels(enc.curr_canvas, &enc.canvas_carryover);
  } else {
    var curr_rect *FrameRectangle = &best_sub_candidate_rect;
     var curr_canvas_in_curr_rect picture.Picture
     var canvas_carryover_in_curr_rect picture.Picture

    // There is no carried over pixel in the disposed rectangle, if any.
    // Note that this could not have been done earlier because the decision to
    // dispose a frame is taken when encoding a next frame's candidate.
    var prev_dispose_method WebPMuxAnimDispose =
        GetPreviousDisposeMethod(enc);
    assert.Assert(prev_dispose_method == WEBP_MUX_DISPOSE_NONE ||
           !enc.prev_candidate_undecided);
    DisposeFrameRectangle(prev_dispose_method, &enc.prev_rect, &enc.canvas_carryover);

    // The pixels outside the current frame rectangle are not carried over.
    // Focus on the current frame rectangle.
    if (!picture.WebPPictureView(enc.curr_canvas, curr_rect.x_offset, curr_rect.y_offset, curr_rect.width, curr_rect.height, &curr_canvas_in_curr_rect) ||
        !picture.WebPPictureView(&enc.canvas_carryover, curr_rect.x_offset, curr_rect.y_offset, curr_rect.width, curr_rect.height, &canvas_carryover_in_curr_rect)) {
      error_code = VP8_ENC_ERROR_INVALID_CONFIGURATION;
      goto End;
    }

    if (enc.best_candidate_carries_over) {
      // Carry over the pixels that were set to fully transparent in the current
      // frame (meaning they are left untouched in canvas_carryover). Copy the
      // other pixels (the explicitly encoded ones) from the original input
      // canvas (curr_canvas) to next frame's canvas_carryover.
      CopyMaskedPixels(&curr_canvas_in_curr_rect, enc.best_candidate_carryover_mask, &canvas_carryover_in_curr_rect);
    } else {
      // No pixel is carried over from previous frames, either because the
      // current subframe is not blended, or because no pixel was set to
      // TRANSPARENT_COLOR.
      WebPCopyPixels(&curr_canvas_in_curr_rect, &canvas_carryover_in_curr_rect);
    }
  }

  // Save the current frame environment as the previous frame environment for
  // the next call to this function.
  WebPCopyPixels(enc.curr_canvas, &enc.prev_canvas);
  enc.prev_candidate_undecided = candidate_undecided;
  if (candidate_undecided) {
    // The previous frame rectangle is not known for sure. Do not save it.
  } else {
    enc.prev_rect = encoded_frame.is_key_frame ? best_key_candidate_rect
                                                 : best_sub_candidate_rect;
  }

  enc.is_first_frame = 0;

Skip:
  ok = 1;
  ++enc.in_frame_count;

End:
  if (!ok || frame_skipped) {
    FrameRelease(encoded_frame);
    // We reset some counters, as the frame addition failed/was skipped.
    --enc.count;
    if !enc.is_first_frame { --enc.count_since_key_frame }
    if (!ok) {
      MarkError2(enc, "ERROR adding frame. WebPEncodingError", error_code);
    }
  }
  enc.curr_canvas.ErrorCode = error_code;  // report error_code
  assert.Assert(ok || error_code != VP8_ENC_OK);
  return ok;
}

func FlushFrames(/* const */ enc *WebPAnimEncoder) int {
  while (enc.flush_count > 0) {
    var err WebPMuxError 
    var curr *EncodedFrame = GetFrame(enc, 0);
    const info *WebPMuxFrameInfo =
        curr.is_key_frame ? &curr.key_frame : &curr.sub_frame;
    assert.Assert(enc.mux != nil);
    err = WebPMuxPushFrame(enc.mux, info, 1);
    if (err != WEBP_MUX_OK) {
      MarkError2(enc, "ERROR adding frame. WebPMuxError", err);
      return 0;
    }
    if (enc.options.verbose) {
      fprintf(stderr, "INFO: Added frame. offset:%d,%d dispose:%d blend:%d\n", info.x_offset, info.y_offset, info.dispose_method, info.blend_method);
    }
    ++enc.out_frame_count;
    FrameRelease(curr);
    ++enc.start;
    --enc.flush_count;
    --enc.count;
    if enc.keyframe != KEYFRAME_NONE { --enc.keyframe }
  }

  if (enc.count == 1 && enc.start != 0) {
    // Move enc.start to index 0.
    enc_start_tmp := (int)enc.start;
    EncodedFrame temp = enc.encoded_frames[0];
    enc.encoded_frames[0] = enc.encoded_frames[enc_start_tmp];
    enc.encoded_frames[enc_start_tmp] = temp;
    FrameRelease(&enc.encoded_frames[enc_start_tmp]);
    enc.start = 0;
  }
  return 1;
}

#undef DELTA_INFINITY
#undef KEYFRAME_NONE

func WebPAnimEncoderAdd(enc *WebPAnimEncoder, frame *picture.Picture, timestamp int, /*const*/ encoder_config *config.Config) int {
   var config config.Config
  var ok int

  if (enc == nil) {
    return 0;
  }
  MarkNoError(enc);

  if (!enc.is_first_frame) {
    // Make sure timestamps are non-decreasing (integer wrap-around is OK).
    prev_frame_duration :=
        (uint32)timestamp - enc.prev_timestamp;
    if (prev_frame_duration >= MAX_DURATION) {
      if (frame != nil) {
        frame.ErrorCode = VP8_ENC_ERROR_INVALID_CONFIGURATION;
      }
      MarkError(enc, "ERROR adding frame: timestamps must be non-decreasing");
      return 0;
    }
    if (!IncreasePreviousDuration(enc, (int)prev_frame_duration)) {
      return 0;
    }
    // IncreasePreviousDuration() may add a frame to afunc exceeding
    // MAX_DURATION which could cause CacheFrame() to over read 'encoded_frames'
    // before the next flush.
    if (enc.count == enc.size && !FlushFrames(enc)) {
      return 0;
    }
  } else {
    enc.first_timestamp = timestamp;
  }

  if (frame == nil) {  // Special: last call.
    enc.got_nil_frame = 1;
    enc.prev_timestamp = timestamp;
    return 1;
  }

  if (frame.width != enc.canvas_width ||
      frame.height != enc.canvas_height) {
    frame.ErrorCode = VP8_ENC_ERROR_INVALID_CONFIGURATION;
    MarkError(enc, "ERROR adding frame: Invalid frame dimensions");
    return 0;
  }

  if (!frame.use_argb) {  // Convert frame from YUV(A) to ARGB.
    if (enc.options.verbose) {
      fprintf(stderr, "WARNING: Converting frame from YUV(A) to ARGB format; "
              "this incurs a small loss.\n");
    }
    if (!picture.WebPPictureYUVAToARGB(frame)) {
      MarkError(enc, "ERROR converting frame from YUV(A) to ARGB");
      return 0;
    }
  }

  if (encoder_config != nil) {
	err := config.Validate()
    if (err != nil) { // TODO: return err
      MarkError(enc, "ERROR adding frame: Invalid config.Config");
      return 0;
    }
    config = *encoder_config;
  } else {
    if (!config.ConfigInit(&config)) {
      MarkError(enc, "Cannot Init config");
      return 0;
    }
    config.Lossless = 1;
  }
  assert.Assert(enc.curr_canvas == nil);
  enc.curr_canvas = frame;  // Store reference.
  assert.Assert(enc.curr_canvas_copy_modified == 1);
  CopyCurrentCanvas(enc);

  ok = CacheFrame(enc, &config) && FlushFrames(enc);

  enc.curr_canvas = nil;
  enc.curr_canvas_copy_modified = 1;
  if (ok) {
    enc.prev_timestamp = timestamp;
  }
  return ok;
}

// -----------------------------------------------------------------------------
// Bitstream assembly.

 static int DecodeFrameOntoCanvas(
    const frame *WebPMuxFrameInfo, /*const*/ canvas *picture.Picture) {
  var image *WebPData = &frame.bitstream;
   var sub_image picture.Picture
   var config WebPDecoderConfig
  if (!WebPInitDecoderConfig(&config)) {
    return 0;
  }
  WebPUtilClearPic(canvas, nil);
  if (WebPGetFeatures(image.bytes, image.size, &config.input) !=
      VP8_STATUS_OK) {
    return 0;
  }
  if (!picture.WebPPictureView(canvas, frame.x_offset, frame.y_offset, config.input.width, config.input.height, &sub_image)) {
    return 0;
  }
  config.output.is_external_memory = 1;
  config.output.colorspace = MODE_BGRA;
  config.output.u.RGBA.rgba = (*uint8)sub_image.argb;
  config.output.u.RGBA.stride = sub_image.argb_stride * 4;
  config.output.u.RGBA.size = config.output.u.RGBA.stride * sub_image.height;

  if (WebPDecode(image.bytes, image.size, &config) != VP8_STATUS_OK) {
    return 0;
  }
  return 1;
}

func FrameToFullCanvas(/* const */ enc *WebPAnimEncoder, /*const*/ frame *WebPMuxFrameInfo, /*const*/ full_image *WebPData) int {
  var canvas_buf *picture.Picture = &enc.curr_canvas_copy;
  WebPMemoryWriter mem1, mem2;
  WebPMemoryWriterInit(&mem1);
  WebPMemoryWriterInit(&mem2);

  if !DecodeFrameOntoCanvas(frame, canvas_buf) { goto Err }
  if !EncodeFrame(&enc.last_config, canvas_buf, &mem1) { goto Err }
  GetEncodedData(&mem1, full_image);

  if (enc.options.allow_mixed) {
    if !EncodeFrame(&enc.last_config_reversed, canvas_buf, &mem2) { goto Err }
    if (mem2.size < mem1.size) {
      GetEncodedData(&mem2, full_image);
      WebPMemoryWriterClear(&mem1);
    } else {
      WebPMemoryWriterClear(&mem2);
    }
  }
  return 1;

Err:
  WebPMemoryWriterClear(&mem1);
  WebPMemoryWriterClear(&mem2);
  return 0;
}

// Convert a single-frame animation to a non-animated image if appropriate.
// TODO(urvang): Can we pick one of the two heuristically (based on frame
// rectangle and/or presence of alpha)?
func OptimizeSingleFrame(/* const */ enc *WebPAnimEncoder, /*const*/ webp_data *WebPData) WebPMuxError {
  var err WebPMuxError = WEBP_MUX_OK;
  var canvas_width , canvas_height int
  var  frame WebPMuxFrameInfo
  var  full_image WebPData
  var  webp_data2 WebPData
  var mux *WebPMux = WebPMuxCreate(webp_data, 0);
  if mux == nil { return WEBP_MUX_BAD_DATA  }
  assert.Assert(enc.out_frame_count == 1);
  WebPDataInit(&frame.bitstream);
  WebPDataInit(&full_image);
  WebPDataInit(&webp_data2);

  err = WebPMuxGetFrame(mux, 1, &frame);
  if err != WEBP_MUX_OK { goto End }
  if frame.id != WEBP_CHUNK_ANMF { goto End }  // Non-animation: nothing to do.
  err = WebPMuxGetCanvasSize(mux, &canvas_width, &canvas_height);
  if err != WEBP_MUX_OK { goto End }
  if (!FrameToFullCanvas(enc, &frame, &full_image)) {
    err = WEBP_MUX_BAD_DATA;
    goto End;
  }
  err = WebPMuxSetImage(mux, &full_image, 1);
  if err != WEBP_MUX_OK { goto End }
  err = WebPMuxAssemble(mux, &webp_data2);
  if err != WEBP_MUX_OK { goto End }

  if (webp_data2.size < webp_data.size) {  // Pick 'webp_data2' if smaller.
    WebPDataClear(webp_data);
    *webp_data = webp_data2;
    WebPDataInit(&webp_data2);
  }

End:
  WebPDataClear(&frame.bitstream);
  WebPDataClear(&full_image);
  WebPMuxDelete(mux);
  WebPDataClear(&webp_data2);
  return err;
}

func WebPAnimEncoderAssemble(enc *WebPAnimEncoder, webp_data *WebPData) int {
  mux *WebPMux;
  var err WebPMuxError 

  if (enc == nil) {
    return 0;
  }
  MarkNoError(enc);

  if (webp_data == nil) {
    MarkError(enc, "ERROR assembling: nil input");
    return 0;
  }

  if (enc.in_frame_count == 0) {
    MarkError(enc, "ERROR: No frames to assemble");
    return 0;
  }

  if (!enc.got_nil_frame && enc.in_frame_count > 1 && enc.count > 0) {
    // set duration of the last frame to be avg of durations of previous frames.
    const double delta_time =
        (uint32)enc.prev_timestamp - enc.first_timestamp;
    average_duration := (int)(delta_time / (enc.in_frame_count - 1));
    if (!IncreasePreviousDuration(enc, average_duration)) {
      return 0;
    }
  }

  // Flush any remaining frames.
  enc.flush_count = enc.count;
  if (!FlushFrames(enc)) {
    return 0;
  }

  // Set definitive canvas size.
  mux = enc.mux;
  err = WebPMuxSetCanvasSize(mux, enc.canvas_width, enc.canvas_height);
  if err != WEBP_MUX_OK { goto Err }

  err = WebPMuxSetAnimationParams(mux, &enc.options.anim_params);
  if err != WEBP_MUX_OK { goto Err }

  // Assemble into a WebP bitstream.
  err = WebPMuxAssemble(mux, webp_data);
  if err != WEBP_MUX_OK { goto Err }

  if (enc.out_frame_count == 1) {
    err = OptimizeSingleFrame(enc, webp_data);
    if err != WEBP_MUX_OK { goto Err }
  }
  return 1;

Err:
  MarkError2(enc, "ERROR assembling WebP", err);
  return 0;
}

func WebPAnimEncoderGetError(enc *WebPAnimEncoder)  *byte {
  if enc == nil { return nil  }
  return enc.error_str;
}

func WebPAnimEncoderSetChunk(enc *WebPAnimEncoder, /*const*/ fourcc [4]byte, /*const*/ chunk_data *WebPData, copy_data int) WebPMuxError {
  if enc == nil { return WEBP_MUX_INVALID_ARGUMENT  }
  return WebPMuxSetChunk(enc.mux, fourcc, chunk_data, copy_data);
}

func WebPAnimEncoderGetChunk(/* const */ enc *WebPAnimEncoder, /*const*/ fourcc [4]byte, chunk_data *WebPData) WebPMuxError {
  if enc == nil { return WEBP_MUX_INVALID_ARGUMENT  }
  return WebPMuxGetChunk(enc.mux, fourcc, chunk_data);
}

func WebPAnimEncoderDeleteChunk(enc *WebPAnimEncoder, /*const*/ fourcc [4]byte) WebPMuxError {
  if enc == nil { return WEBP_MUX_INVALID_ARGUMENT  }
  return WebPMuxDeleteChunk(enc.mux, fourcc);
}
