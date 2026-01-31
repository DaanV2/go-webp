// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//  AnimDecoder implementation.
//

package demux

import "github.com/daanv2/go-webp/pkg/libwebp/webp"

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


const NUM_CHANNELS = 4

// Channel extraction from a uint32 representation of a uint8 RGBA/BGRA
// buffer.
func CHANNEL_SHIFT(i int) int {
	if constants.WORDS_BIGENDIAN {
		return (24 - (i) * 8)
	}

	return((i) * 8)
}

type BlendRowFunc = func(/* const */ *uint32, /* const */ *uint32, int)
type BlendPixelRowNonPremult =func(/* const */ src *uint32, /* const */ dst *uint32, int num_pixels);
type BlendPixelRowPremult =func(/* const */ src *uint32, /* const */ dst *uint32, int num_pixels);

type WebPAnimDecoder struct {
	demux *WebPDemuxer        // Demuxer created from given WebP bitstream.
	// Decoder config.
	// Note: we use a pointer to a function blending multiple pixels at a time to
	// allow possible inlining of per-pixel blending function.
	config WebPDecoderConfig 
	blend_func BlendRowFunc        // Pointer to the chose blend row function.
	info WebPAnimInfo              // Global info about the animation.
	curr_frame *uint8           // Current canvas (not disposed).
	prev_frame_disposed *uint8  // Previous canvas (properly disposed).
	prev_frame_timestamp int       // Previous frame timestamp (milliseconds).
	prev_iter WebPIterator         // Iterator object for previous frame.
	prev_frame_was_keyframe int    // True if previous frame was a keyframe.
	// Index of the next frame to be decoded
	// (starting from 1).
	next_frame int 
}

func DefaultDecoderOptions(/* const */ dec_options *WebPAnimDecoderOptions) {
  dec_options.color_mode = MODE_RGBA;
  dec_options.use_threads = 0;
}

func WebPAnimDecoderOptionsInitInternal(dec_options *WebPAnimDecoderOptions, abi_version int) int {
  if (dec_options == nil ||
      WEBP_ABI_IS_INCOMPATIBLE(abi_version, WEBP_DEMUX_ABI_VERSION)) {
    return 0;
  }

  DefaultDecoderOptions(dec_options);
  return 1;
}

func ApplyDecoderOptions(/* const */ dec_options *WebPAnimDecoderOptions, /* const */ dec *WebPAnimDecoder) int {
  var mode WEBP_CSP_MODE;
  config := &dec.config;
  assert.Assert(dec_options != nil);

  mode = dec_options.color_mode;
  if (mode != MODE_RGBA && mode != MODE_BGRA && mode != MODE_rgbA && mode != MODE_bgrA) {
    return 0;
  }
  dec.blend_func = tenary.If(mode == MODE_RGBA || mode == MODE_BGRA, BlendPixelRowNonPremult, BlendPixelRowPremult)
  if (!WebPInitDecoderConfig(config)) {
    return 0;
  }
  config.output.colorspace = mode;
  config.output.is_external_memory = 1;
  config.options.use_threads = dec_options.use_threads;
  // Note: config.output.u.RGBA is set at the time of decoding each frame.
  return 1;
}

func WebPAnimDecoder(/* const */ webp_data *WebPData, /* const */ dec_options *WebPAnimDecoderOptions, abi_version int) *WebPAnimDecoderNewInternal {
  var options WebPAnimDecoderOptions;
  var features WebPBitstreamFeatures
  var dec *WebPAnimDecoder = nil;
  if (webp_data == nil || WEBP_ABI_IS_INCOMPATIBLE(abi_version, WEBP_DEMUX_ABI_VERSION)) {
    return nil;
  }

  // Validate the bitstream before doing expensive allocations. The demuxer may
  // be more tolerant than the decoder.
  if (WebPGetFeatures(webp_data.bytes, webp_data.size, &features) !=
      VP8_STATUS_OK) {
    return nil;
  }

  // Note: calloc() so that the pointer members are initialized to nil.
  dec = (*WebPAnimDecoder)WebPSafeCalloc(uint64(1), sizeof(*dec))
  if (dec == nil) goto Error;

  if (dec_options != nil) {
    options = *dec_options;
  } else {
    DefaultDecoderOptions(&options);
  }
  if (!ApplyDecoderOptions(&options, dec)) goto Error;

  dec.demux = WebPDemux(webp_data);
  if (dec.demux == nil) goto Error;

  dec.info.canvas_width = WebPDemuxGetI(dec.demux, WEBP_FF_CANVAS_WIDTH);
  dec.info.canvas_height = WebPDemuxGetI(dec.demux, WEBP_FF_CANVAS_HEIGHT);
  dec.info.loop_count = WebPDemuxGetI(dec.demux, WEBP_FF_LOOP_COUNT);
  dec.info.bgcolor = WebPDemuxGetI(dec.demux, WEBP_FF_BACKGROUND_COLOR);
  dec.info.frame_count = WebPDemuxGetI(dec.demux, WEBP_FF_FRAME_COUNT);

  // Note: calloc() because we fill frame with zeroes as well.
  dec.curr_frame = (*uint8)WebPSafeCalloc(
      dec.info.canvas_width * NUM_CHANNELS, dec.info.canvas_height);
  if (dec.curr_frame == nil) goto Error;
  dec.prev_frame_disposed = (*uint8)WebPSafeCalloc(
      dec.info.canvas_width * NUM_CHANNELS, dec.info.canvas_height);
  if (dec.prev_frame_disposed == nil) goto Error;

  WebPAnimDecoderReset(dec);
  return dec;

Error:
  WebPAnimDecoderDelete(dec);
  return nil;
}

func WebPAnimDecoderGetInfo(/* const */ dec *WebPAnimDecoder, info *WebPAnimInfo) int {
  if (dec == nil || info == nil) return 0;
  *info = dec.info;
  return 1;
}

// Returns true if the frame covers the full canvas.
func IsFullFrame(width, height, canvas_width, canvas_height int) int {
  return (width == canvas_width && height == canvas_height);
}

// Clear the canvas to transparent.
func ZeroFillCanvas(buf *uint8, uint32 canvas_width, uint32 canvas_height) int {
  size = (uint64)canvas_width * canvas_height * NUM_CHANNELS * sizeof(*buf);
  
  if (!CheckSizeOverflow(size)) return 0;
 
  WEBP_UNSAFE_MEMSET(buf, 0, (uint64)size);
  return 1;
}

// Clear given frame rectangle to transparent.
func ZeroFillFrameRect(buf *uint8, buf_stride, x_offset, y_offset, width, height int) {
  assert.Assert(width * NUM_CHANNELS <= buf_stride);
  buf += y_offset * buf_stride + x_offset * NUM_CHANNELS;
  for (j := 0; j < height; ++j) {
    WEBP_UNSAFE_MEMSET(buf, 0, width * NUM_CHANNELS);
    buf += buf_stride;
  }
}

// Copy width * height pixels from 'src' to 'dst'.
func CopyCanvas(/* const */ src *uint8, dst *uint8, width, height uint32 ) int {
  size := (uint64)width * height * NUM_CHANNELS;
  if (!CheckSizeOverflow(size)) return 0;
  assert.Assert(src != nil && dst != nil);
  WEBP_UNSAFE_MEMCPY(dst, src, (uint64)size);
  return 1;
}

// Returns true if the current frame is a key-frame.
func IsKeyFrame(/* const */ curr *WebPIterator, /* const */ prev *WebPIterator, prev_frame_was_key_frame, canvas_width, canvas_height int) int {
  if (curr.frame_num == 1) {
    return 1;
  } else if ((!curr.has_alpha || curr.blend_method == WEBP_MUX_NO_BLEND) &&
             IsFullFrame(curr.width, curr.height, canvas_width, canvas_height)) {
    return 1;
  } else {
    return (prev.dispose_method == WEBP_MUX_DISPOSE_BACKGROUND) &&
           (IsFullFrame(prev.width, prev.height, canvas_width, canvas_height) ||
            prev_frame_was_key_frame);
  }
}

// Blend a single channel of 'src' over 'dst', given their alpha channel values.
// 'src' and 'dst' are assumed to be NOT pre-multiplied by alpha.
func BlendChannelNonPremult(src uint32 , src_a uint8 , dst uint32 , dst_a uint8 , scale uint32, shift int) uint8 {
  src_channel := (src >> shift) & 0xff;
  dst_channel := (dst >> shift) & 0xff;
  blend_unscaled := src_channel * src_a + dst_channel * dst_a;
  assert.Assert(blend_unscaled < (uint64(1) << 32) / scale);
  return (blend_unscaled * scale) >> CHANNEL_SHIFT(3);
}

// Blend 'src' over 'dst' assuming they are NOT pre-multiplied by alpha.
func BlendPixelNonPremult(uint32 src, uint32 dst) uint32 {
  src_a := (src >> CHANNEL_SHIFT(3)) & 0xff;

  if (src_a == 0) {
    return dst;
  } else {
    dst_a := (dst >> CHANNEL_SHIFT(3)) & 0xff;
    // This is the approximate integer arithmetic for the actual formula:
    // dst_factor_a = (dst_a * (255 - src_a)) / 255.
    dst_factor_a := (dst_a * (256 - src_a)) >> 8;
    blend_a := src_a + dst_factor_a;
    scale := (uint64(1) << 24) / blend_a;

    blend_r := BlendChannelNonPremult(
        src, src_a, dst, dst_factor_a, scale, CHANNEL_SHIFT(0));
    blend_g := BlendChannelNonPremult(
        src, src_a, dst, dst_factor_a, scale, CHANNEL_SHIFT(1));
    blend_b := BlendChannelNonPremult(
        src, src_a, dst, dst_factor_a, scale, CHANNEL_SHIFT(2));
    assert.Assert(src_a + dst_factor_a < 256);

    return ((uint32)blend_r << CHANNEL_SHIFT(0)) |
           ((uint32)blend_g << CHANNEL_SHIFT(1)) |
           ((uint32)blend_b << CHANNEL_SHIFT(2)) |
           ((uint32)blend_a << CHANNEL_SHIFT(3));
  }
}

// Blend 'num_pixels' in 'src' over 'dst' assuming they are NOT pre-multiplied
// by alpha.
func BlendPixelRowNonPremult(const src *uint32, const dst *uint32, int num_pixels) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    src_alpha := (src[i] >> CHANNEL_SHIFT(3)) & 0xff;
    if (src_alpha != 0xff) {
      src[i] = BlendPixelNonPremult(src[i], dst[i]);
    }
  }
}

// Individually multiply each channel in 'pix' by 'scale'.
func ChannelwiseMultiply(pix uint32, scale uint32) uint32 {
  mask := 0x00FF00FF;
  rb := ((pix & mask) * scale) >> 8;
  ag := ((pix >> 8) & mask) * scale;
  return (rb & mask) | (ag & ~mask);
}

// Blend 'src' over 'dst' assuming they are pre-multiplied by alpha.
func BlendPixelPremult(src, dst uint32) uint32 {
  src_a := (src >> CHANNEL_SHIFT(3)) & 0xff;
  return src + ChannelwiseMultiply(dst, 256 - src_a);
}

// Blend 'num_pixels' in 'src' over 'dst' assuming they are pre-multiplied by
// alpha.
func BlendPixelRowPremult(/* const */ src *uint32, /* const */ dst *uint32, num_pixels int) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    src_alpha := (src[i] >> CHANNEL_SHIFT(3)) & 0xff;
    if (src_alpha != 0xff) {
      src[i] = BlendPixelPremult(src[i], dst[i]);
    }
  }
}

// Returns two ranges (<left, width> pairs) at row 'canvas_y', that belong to
// 'src' but not 'dst'. A point range is empty if the corresponding width is 0.
func FindBlendRangeAtRow(/* const */ src *WebPIterator, /* const */ dst *WebPIterator, canvas_y int, left1, width1, left2, width2 *int) {
  src_max_x := src.x_offset + src.width;
  dst_max_x := dst.x_offset + dst.width;
  dst_max_y := dst.y_offset + dst.height;
  assert.Assert(canvas_y >= src.y_offset && canvas_y < (src.y_offset + src.height));
  *left1 = -1;
  *width1 = 0;
  *left2 = -1;
  *width2 = 0;

  if (canvas_y < dst.y_offset || canvas_y >= dst_max_y ||
      src.x_offset >= dst_max_x || src_max_x <= dst.x_offset) {
    *left1 = src.x_offset;
    *width1 = src.width;
    return;
  }

  if (src.x_offset < dst.x_offset) {
    *left1 = src.x_offset;
    *width1 = dst.x_offset - src.x_offset;
  }

  if (src_max_x > dst_max_x) {
    *left2 = dst_max_x;
    *width2 = src_max_x - dst_max_x;
  }
}

func WebPAnimDecoderGetNext(dec *WebPAnimDecoder, *uint8* buf_ptr, timestamp_ptr *int) int {
  var iter WebPIterator
  var width uint32
  var height uint32
  var is_key_frame int
  var timestamp int
  var blend_row BlendRowFunc

  if (dec == nil || buf_ptr == nil || timestamp_ptr == nil) return 0;
  if (!WebPAnimDecoderHasMoreFrames(dec)) return 0;

  width = dec.info.canvas_width;
  height = dec.info.canvas_height;
  blend_row = dec.blend_func;

  // Get compressed frame.
  if (!WebPDemuxGetFrame(dec.demux, dec.next_frame, &iter)) {
    return 0;
  }
  timestamp = dec.prev_frame_timestamp + iter.duration;

  // Initialize.
  is_key_frame = IsKeyFrame(&iter, &dec.prev_iter, dec.prev_frame_was_keyframe, width, height);
  if (is_key_frame) {
    if (!ZeroFillCanvas(dec.curr_frame, width, height)) {
      goto Error;
    }
  } else {
    if (!CopyCanvas(dec.prev_frame_disposed, dec.curr_frame, width, height)) {
      goto Error;
    }
  }

  // Decode.
  {
    var in *uint8 = iter.fragment.bytes;
    in_size := iter.fragment.size;
    stride := width * NUM_CHANNELS;  // at most 25 + 2 bits
    out_offset := (uint64)iter.y_offset * stride +
                                (uint64)iter.x_offset * NUM_CHANNELS;  // 53b
    const size uint64  = (uint64)iter.height * stride;  // at most 25 + 27b
    var config *WebPDecoderConfig = &dec.config;
    var buf *WebPRGBABuffer = &config.output.u.RGBA;
    if ((uint64)size != size) goto Error;
    buf.stride = (int)stride;
    buf.size = (uint64)size;
    buf.rgba = dec.curr_frame + out_offset;

    if (WebPDecode(in, in_size, config) != VP8_STATUS_OK) {
      goto Error;
    }
  }

  // During the decoding of current frame, we may have set some pixels to be
  // transparent (i.e. alpha < 255). However, the value of each of these
  // pixels should have been determined by blending it against the value of
  // that pixel in the previous frame if blending method of is WEBP_MUX_BLEND.
  if (iter.frame_num > 1 && iter.blend_method == WEBP_MUX_BLEND &&
      !is_key_frame) {
    if (dec.prev_iter.dispose_method == WEBP_MUX_DISPOSE_NONE) {
      int y;
      // Blend transparent pixels with pixels in previous canvas.
      for (y = 0; y < iter.height; ++y) {
        offset := (iter.y_offset + y) * width + iter.x_offset;
        blend_row((*uint32)dec.curr_frame + offset, (*uint32)dec.prev_frame_disposed + offset, iter.width);
      }
    } else {
      int y;
      assert.Assert(dec.prev_iter.dispose_method == WEBP_MUX_DISPOSE_BACKGROUND);
      // We need to blend a transparent pixel with its value just after
      // initialization. That is, blend it with:
      // * Fully transparent pixel if it belongs to prevRect <-- No-op.
      // * The pixel in the previous canvas otherwise <-- Need alpha-blending.
      for (y = 0; y < iter.height; ++y) {
        canvas_y := iter.y_offset + y;
        int left1, width1, left2, width2;
        FindBlendRangeAtRow(&iter, &dec.prev_iter, canvas_y, &left1, &width1, &left2, &width2);
        if (width1 > 0) {
          offset1 := canvas_y * width + left1;
          blend_row((*uint32)dec.curr_frame + offset1, (*uint32)dec.prev_frame_disposed + offset1, width1);
        }
        if (width2 > 0) {
          offset2 := canvas_y * width + left2;
          blend_row((*uint32)dec.curr_frame + offset2, (*uint32)dec.prev_frame_disposed + offset2, width2);
        }
      }
    }
  }

  // Update info of the previous frame and dispose it for the next iteration.
  dec.prev_frame_timestamp = timestamp;
  WebPDemuxReleaseIterator(&dec.prev_iter);
  dec.prev_iter = iter;
  dec.prev_frame_was_keyframe = is_key_frame;
  if (!CopyCanvas(dec.curr_frame, dec.prev_frame_disposed, width, height)) {
    goto Error;
  }
  if (dec.prev_iter.dispose_method == WEBP_MUX_DISPOSE_BACKGROUND) {
    ZeroFillFrameRect(dec.prev_frame_disposed, width * NUM_CHANNELS, dec.prev_iter.x_offset, dec.prev_iter.y_offset, dec.prev_iter.width, dec.prev_iter.height);
  }
  ++dec.next_frame;

  // All OK, fill in the values.
  *buf_ptr = dec.curr_frame;
  *timestamp_ptr = timestamp;
  return 1;

Error:
  WebPDemuxReleaseIterator(&iter);
  return 0;
}

func WebPAnimDecoderHasMoreFrames(/* const  */dec *WebPAnimDecoder) int {
  if (dec == nil) return 0;
  return (dec.next_frame <= (int)dec.info.frame_count);
}

func WebPAnimDecoderReset(dec *WebPAnimDecoder) {
  if (dec != nil) {
    dec.prev_frame_timestamp = 0;
    WebPDemuxReleaseIterator(&dec.prev_iter);
    WEBP_UNSAFE_MEMSET(&dec.prev_iter, 0, sizeof(dec.prev_iter));
    dec.prev_frame_was_keyframe = 0;
    dec.next_frame = 1;
  }
}

func WebPDemuxer(const dec *WebPAnimDecoder) *WebPAnimDecoderGetDemuxer {
  if (dec == nil) { return nil }

  return dec.demux;
}

func WebPAnimDecoderDelete(dec *WebPAnimDecoder) {
  if (dec != nil) {
    WebPDemuxReleaseIterator(&dec.prev_iter);
    WebPDemuxDelete(dec.demux);
    WebPSafeFree(dec.curr_frame);
    WebPSafeFree(dec.prev_frame_disposed);
    WebPSafeFree(dec);
  }
}
