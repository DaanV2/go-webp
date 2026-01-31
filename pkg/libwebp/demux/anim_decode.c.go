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

#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

const NUM_CHANNELS =4

// Channel extraction from a uint32 representation of a uint8 RGBA/BGRA
// buffer.
#ifdef constants.WORDS_BIGENDIAN
#define CHANNEL_SHIFT(i) (24 - (i) * 8)
#else
#define CHANNEL_SHIFT(i) ((i) * 8)
#endif

typedef func (*BlendRowFunc)(*uint32 const, const *uint32 const, int);
func BlendPixelRowNonPremult(*uint32 const src, const *uint32 const dst, int num_pixels);
func BlendPixelRowPremult(*uint32 const src, const *uint32 const dst, int num_pixels);

type WebPAnimDecoder struct {
  *WebPDemuxer demux;        // Demuxer created from given WebP bitstream.
  WebPDecoderConfig config;  // Decoder config.
  // Note: we use a pointer to a function blending multiple pixels at a time to
  // allow possible inlining of per-pixel blending function.
  BlendRowFunc blend_func;       // Pointer to the chose blend row function.
  WebPAnimInfo info;             // Global info about the animation.
  *uint8 curr_frame;           // Current canvas (not disposed).
  *uint8 prev_frame_disposed;  // Previous canvas (properly disposed).
  int prev_frame_timestamp;      // Previous frame timestamp (milliseconds).
  WebPIterator prev_iter;        // Iterator object for previous frame.
  int prev_frame_was_keyframe;   // True if previous frame was a keyframe.
  int next_frame;                // Index of the next frame to be decoded
                                 // (starting from 1).
}

func DefaultDecoderOptions(*WebPAnimDecoderOptions const dec_options) {
  dec_options.color_mode = MODE_RGBA;
  dec_options.use_threads = 0;
}

int WebPAnimDecoderOptionsInitInternal(*WebPAnimDecoderOptions dec_options, int abi_version) {
  if (dec_options == nil ||
      WEBP_ABI_IS_INCOMPATIBLE(abi_version, WEBP_DEMUX_ABI_VERSION)) {
    return 0;
  }
  DefaultDecoderOptions(dec_options);
  return 1;
}

 static int ApplyDecoderOptions(
    const *WebPAnimDecoderOptions const dec_options, *WebPAnimDecoder const dec) {
  WEBP_CSP_MODE mode;
  *WebPDecoderConfig config = &dec.config;
  assert.Assert(dec_options != nil);

  mode = dec_options.color_mode;
  if (mode != MODE_RGBA && mode != MODE_BGRA && mode != MODE_rgbA &&
      mode != MODE_bgrA) {
    return 0;
  }
  dec.blend_func = (mode == MODE_RGBA || mode == MODE_BGRA)
                        ? &BlendPixelRowNonPremult
                        : &BlendPixelRowPremult;
  if (!WebPInitDecoderConfig(config)) {
    return 0;
  }
  config.output.colorspace = mode;
  config.output.is_external_memory = 1;
  config.options.use_threads = dec_options.use_threads;
  // Note: config.output.u.RGBA is set at the time of decoding each frame.
  return 1;
}

*WebPAnimDecoder WebPAnimDecoderNewInternal(
    const *WebPData webp_data, const *WebPAnimDecoderOptions dec_options, int abi_version) {
  WebPAnimDecoderOptions options;
  *WebPAnimDecoder dec = nil;
  WebPBitstreamFeatures features;
  if (webp_data == nil ||
      WEBP_ABI_IS_INCOMPATIBLE(abi_version, WEBP_DEMUX_ABI_VERSION)) {
    return nil;
  }

  // Validate the bitstream before doing expensive allocations. The demuxer may
  // be more tolerant than the decoder.
  if (WebPGetFeatures(webp_data.bytes, webp_data.size, &features) !=
      VP8_STATUS_OK) {
    return nil;
  }

  // Note: calloc() so that the pointer members are initialized to nil.
  dec = (*WebPAnimDecoder)WebPSafeCalloc(uint64(1), sizeof(*dec));
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

int WebPAnimDecoderGetInfo(const *WebPAnimDecoder dec, *WebPAnimInfo info) {
  if (dec == nil || info == nil) return 0;
  *info = dec.info;
  return 1;
}

// Returns true if the frame covers the full canvas.
static int IsFullFrame(int width, int height, int canvas_width, int canvas_height) {
  return (width == canvas_width && height == canvas_height);
}

// Clear the canvas to transparent.
 static int ZeroFillCanvas(*uint8 buf, uint32 canvas_width, uint32 canvas_height) {
  const uint64 size =
      (uint64)canvas_width * canvas_height * NUM_CHANNELS * sizeof(*buf);
  if (!CheckSizeOverflow(size)) return 0;
  WEBP_UNSAFE_MEMSET(buf, 0, (uint64)size);
  return 1;
}

// Clear given frame rectangle to transparent.
func ZeroFillFrameRect(*uint8 buf, int buf_stride, int x_offset, int y_offset, int width, int height) {
  int j;
  assert.Assert(width * NUM_CHANNELS <= buf_stride);
  buf += y_offset * buf_stride + x_offset * NUM_CHANNELS;
  for (j = 0; j < height; ++j) {
    WEBP_UNSAFE_MEMSET(buf, 0, width * NUM_CHANNELS);
    buf += buf_stride;
  }
}

// Copy width * height pixels from 'src' to 'dst'.
 static int CopyCanvas(const *uint8 src, *uint8 dst, uint32 width, uint32 height) {
  const uint64 size = (uint64)width * height * NUM_CHANNELS;
  if (!CheckSizeOverflow(size)) return 0;
  assert.Assert(src != nil && dst != nil);
  WEBP_UNSAFE_MEMCPY(dst, src, (uint64)size);
  return 1;
}

// Returns true if the current frame is a key-frame.
static int IsKeyFrame(const *WebPIterator const curr, const *WebPIterator const prev, int prev_frame_was_key_frame, int canvas_width, int canvas_height) {
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
static uint8 BlendChannelNonPremult(uint32 src, uint8 src_a, uint32 dst, uint8 dst_a, uint32 scale, int shift) {
  const uint8 src_channel = (src >> shift) & 0xff;
  const uint8 dst_channel = (dst >> shift) & 0xff;
  const uint32 blend_unscaled = src_channel * src_a + dst_channel * dst_a;
  assert.Assert(blend_unscaled < (uint64(1) << 32) / scale);
  return (blend_unscaled * scale) >> CHANNEL_SHIFT(3);
}

// Blend 'src' over 'dst' assuming they are NOT pre-multiplied by alpha.
static uint32 BlendPixelNonPremult(uint32 src, uint32 dst) {
  const uint8 src_a = (src >> CHANNEL_SHIFT(3)) & 0xff;

  if (src_a == 0) {
    return dst;
  } else {
    const uint8 dst_a = (dst >> CHANNEL_SHIFT(3)) & 0xff;
    // This is the approximate integer arithmetic for the actual formula:
    // dst_factor_a = (dst_a * (255 - src_a)) / 255.
    const uint8 dst_factor_a = (dst_a * (256 - src_a)) >> 8;
    const uint8 blend_a = src_a + dst_factor_a;
    const uint32 scale = (1UL << 24) / blend_a;

    const uint8 blend_r = BlendChannelNonPremult(
        src, src_a, dst, dst_factor_a, scale, CHANNEL_SHIFT(0));
    const uint8 blend_g = BlendChannelNonPremult(
        src, src_a, dst, dst_factor_a, scale, CHANNEL_SHIFT(1));
    const uint8 blend_b = BlendChannelNonPremult(
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
func BlendPixelRowNonPremult(*uint32 const src, const *uint32 const dst, int num_pixels) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    const uint8 src_alpha = (src[i] >> CHANNEL_SHIFT(3)) & 0xff;
    if (src_alpha != 0xff) {
      src[i] = BlendPixelNonPremult(src[i], dst[i]);
    }
  }
}

// Individually multiply each channel in 'pix' by 'scale'.
static  uint32 ChannelwiseMultiply(uint32 pix, uint32 scale) {
  uint32 mask = 0x00FF00FF;
  uint32 rb = ((pix & mask) * scale) >> 8;
  uint32 ag = ((pix >> 8) & mask) * scale;
  return (rb & mask) | (ag & ~mask);
}

// Blend 'src' over 'dst' assuming they are pre-multiplied by alpha.
static uint32 BlendPixelPremult(uint32 src, uint32 dst) {
  const uint8 src_a = (src >> CHANNEL_SHIFT(3)) & 0xff;
  return src + ChannelwiseMultiply(dst, 256 - src_a);
}

// Blend 'num_pixels' in 'src' over 'dst' assuming they are pre-multiplied by
// alpha.
func BlendPixelRowPremult(*uint32 const src, const *uint32 const dst, int num_pixels) {
  int i;
  for (i = 0; i < num_pixels; ++i) {
    const uint8 src_alpha = (src[i] >> CHANNEL_SHIFT(3)) & 0xff;
    if (src_alpha != 0xff) {
      src[i] = BlendPixelPremult(src[i], dst[i]);
    }
  }
}

// Returns two ranges (<left, width> pairs) at row 'canvas_y', that belong to
// 'src' but not 'dst'. A point range is empty if the corresponding width is 0.
func FindBlendRangeAtRow(const *WebPIterator const src, const *WebPIterator const dst, int canvas_y, *int const left1, *int const width1, *int const left2, *int const width2) {
  const int src_max_x = src.x_offset + src.width;
  const int dst_max_x = dst.x_offset + dst.width;
  const int dst_max_y = dst.y_offset + dst.height;
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

int WebPAnimDecoderGetNext(*WebPAnimDecoder dec, *uint8* buf_ptr, *int timestamp_ptr) {
  WebPIterator iter;
  uint32 width;
  uint32 height;
  int is_key_frame;
  int timestamp;
  BlendRowFunc blend_row;

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
    const *uint8 in = iter.fragment.bytes;
    const uint64 in_size = iter.fragment.size;
    const uint32 stride = width * NUM_CHANNELS;  // at most 25 + 2 bits
    const uint64 out_offset = (uint64)iter.y_offset * stride +
                                (uint64)iter.x_offset * NUM_CHANNELS;  // 53b
    const uint64 size = (uint64)iter.height * stride;  // at most 25 + 27b
    *WebPDecoderConfig const config = &dec.config;
    *WebPRGBABuffer const buf = &config.output.u.RGBA;
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
        const uint64 offset = (iter.y_offset + y) * width + iter.x_offset;
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
        const int canvas_y = iter.y_offset + y;
        int left1, width1, left2, width2;
        FindBlendRangeAtRow(&iter, &dec.prev_iter, canvas_y, &left1, &width1, &left2, &width2);
        if (width1 > 0) {
          const uint64 offset1 = canvas_y * width + left1;
          blend_row((*uint32)dec.curr_frame + offset1, (*uint32)dec.prev_frame_disposed + offset1, width1);
        }
        if (width2 > 0) {
          const uint64 offset2 = canvas_y * width + left2;
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

int WebPAnimDecoderHasMoreFrames(const *WebPAnimDecoder dec) {
  if (dec == nil) return 0;
  return (dec.next_frame <= (int)dec.info.frame_count);
}

func WebPAnimDecoderReset(*WebPAnimDecoder dec) {
  if (dec != nil) {
    dec.prev_frame_timestamp = 0;
    WebPDemuxReleaseIterator(&dec.prev_iter);
    WEBP_UNSAFE_MEMSET(&dec.prev_iter, 0, sizeof(dec.prev_iter));
    dec.prev_frame_was_keyframe = 0;
    dec.next_frame = 1;
  }
}

const *WebPDemuxer WebPAnimDecoderGetDemuxer(const *WebPAnimDecoder dec) {
  if (dec == nil) return nil;
  return dec.demux;
}

func WebPAnimDecoderDelete(*WebPAnimDecoder dec) {
  if (dec != nil) {
    WebPDemuxReleaseIterator(&dec.prev_iter);
    WebPDemuxDelete(dec.demux);
    WebPSafeFree(dec.curr_frame);
    WebPSafeFree(dec.prev_frame_disposed);
    WebPSafeFree(dec);
  }
}
