// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//  WebP container demux.
//

package demux

#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"  // WebPGetFeatures
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

const DMUX_MAJ_VERSION =1
const DMUX_MIN_VERSION =6
const DMUX_REV_VERSION =0

type MemBuffer struct {
  uint64 start;     // start location of the data
  uint64 end;       // end location
  uint64 riff_end;  // riff chunk end location, can be > end.
  uint64 buf_size;  // size of the buffer
  const buf *uint8;
}

type ChunkData struct {
  uint64 offset;
  uint64 size;
}

type Frame struct {
  int x_offset, y_offset;
  int width, height;
  int has_alpha;
  int duration;
  WebPMuxAnimDispose dispose_method;
  WebPMuxAnimBlend blend_method;
  int frame_num;
  int complete;                 // img_components contains a full image.
  ChunkData img_components[2];  // 0=VP8{,L} 1=ALPH
  struct next *Frame;
}

type Chunk struct {
  ChunkData data;
  struct next *Chunk;
}

type WebPDemuxer struct {
  MemBuffer mem;
  WebPDemuxState state;
  int is_ext_format;
  uint32 feature_flags;
  int canvas_width, canvas_height;
  int loop_count;
  uint32 bgcolor;
  int num_frames;
  frames *Frame;
  *Frame* frames_tail;
  chunks *Chunk;  // non-image chunks
  *Chunk* chunks_tail;
}

type <FOO> int

const ( PARSE_OK, PARSE_NEED_MORE_DATA, PARSE_ERROR } ParseStatus;

typedef type ChunkParser struct {
  uint8 id[4];
  ParseStatus (*parse)(const dmux *WebPDemuxer);
  int (*valid)(const const dmux *WebPDemuxer);
} ChunkParser;

static ParseStatus ParseSingleImage(const dmux *WebPDemuxer);
static ParseStatus ParseVP8X(const dmux *WebPDemuxer);
static int IsValidSimpleFormat(const const dmux *WebPDemuxer);
static int IsValidExtendedFormat(const const dmux *WebPDemuxer);

static const ChunkParser kMasterChunks[] = {
    {{'V', 'P', '8', ' '}, ParseSingleImage, IsValidSimpleFormat}, {{'V', 'P', '8', 'L'}, ParseSingleImage, IsValidSimpleFormat}, {{'V', 'P', '8', 'X'}, ParseVP8X, IsValidExtendedFormat}, {{'0', '0', '0', '0'}, nil, nil},
}

//------------------------------------------------------------------------------

func WebPGetDemuxVersion() int {
  return (DMUX_MAJ_VERSION << 16) | (DMUX_MIN_VERSION << 8) | DMUX_REV_VERSION;
}

// -----------------------------------------------------------------------------
// MemBuffer

func RemapMemBuffer(const mem *MemBuffer, const data *uint8, uint64 size) int {
  if (size < mem.buf_size) return 0;  // can't remap to a shorter buffer!

  mem.buf = data;
  mem.end = mem.buf_size = size;
  return 1;
}

static int InitMemBuffer(const mem *MemBuffer, const data *uint8, uint64 size) {
  WEBP_UNSAFE_MEMSET(mem, 0, sizeof(*mem));
  return RemapMemBuffer(mem, data, size);
}

// Return the remaining data size available in 'mem'.
static  uint64 MemDataSize(const const mem *MemBuffer) {
  return (mem.end - mem.start);
}

// Return true if 'size' exceeds the end of the RIFF chunk.
static  int SizeIsInvalid(const const mem *MemBuffer, uint64 size) {
  return (size > mem.riff_end - mem.start);
}

static  func Skip(const mem *MemBuffer, uint64 size) {
  mem.start += size;
}

static  func Rewind(const mem *MemBuffer, uint64 size) {
  mem.start -= size;
}

static  const GetBuffer *uint8(const mem *MemBuffer) {
  return mem.buf + mem.start;
}

// Read from 'mem' and skip the read bytes.
static  uint8 ReadByte(const mem *MemBuffer) {
  const uint8 byte = mem.buf[mem.start];
  Skip(mem, 1);
  return byte;
}

static  int ReadLE16s(const mem *MemBuffer) {
  const const data *uint8 = mem.buf + mem.start;
  const int val = GetLE16(data);
  Skip(mem, 2);
  return val;
}

static  int ReadLE24s(const mem *MemBuffer) {
  const const data *uint8 = mem.buf + mem.start;
  const int val = GetLE24(data);
  Skip(mem, 3);
  return val;
}

static  uint32 ReadLE32(const mem *MemBuffer) {
  const const data *uint8 = mem.buf + mem.start;
  const uint32 val = GetLE32(data);
  Skip(mem, 4);
  return val;
}

// -----------------------------------------------------------------------------
// Secondary chunk parsing

func AddChunk(const dmux *WebPDemuxer, const chunk *Chunk) {
  *dmux.chunks_tail = chunk;
  chunk.next = nil;
  dmux.chunks_tail = &chunk.next;
}

// Add a frame to the end of the list, ensuring the last frame is complete.
// Returns true on success, false otherwise.
static int AddFrame(const dmux *WebPDemuxer, const frame *Frame) {
  const const last_frame *Frame = *dmux.frames_tail;
  if (last_frame != nil && !last_frame.complete) return 0;

  *dmux.frames_tail = frame;
  frame.next = nil;
  dmux.frames_tail = &frame.next;
  return 1;
}

func SetFrameInfo(uint64 start_offset, uint64 size, int frame_num, int complete, const const features *WebPBitstreamFeatures, const frame *Frame) {
  frame.img_components[0].offset = start_offset;
  frame.img_components[0].size = size;
  frame.width = features.width;
  frame.height = features.height;
  frame.has_alpha |= features.has_alpha;
  frame.frame_num = frame_num;
  frame.complete = complete;
}

// Store image bearing chunks to 'frame'. 'min_size' is an optional size
// requirement, it may be zero.
static ParseStatus StoreFrame(int frame_num, uint32 min_size, const mem *MemBuffer, const frame *Frame) {
  int alpha_chunks = 0;
  int image_chunks = 0;
  int done =
      (MemDataSize(mem) < CHUNK_HEADER_SIZE || MemDataSize(mem) < min_size);
  ParseStatus status = PARSE_OK;

  if (done) return PARSE_NEED_MORE_DATA;

  do {
    const uint64 chunk_start_offset = mem.start;
    const uint32 fourcc = ReadLE32(mem);
    const uint32 payload_size = ReadLE32(mem);
    uint32 payload_size_padded;
    uint64 payload_available;
    uint64 chunk_size;

    if (payload_size > MAX_CHUNK_PAYLOAD) return PARSE_ERROR;

    payload_size_padded = payload_size + (payload_size & 1);
    payload_available = (payload_size_padded > MemDataSize(mem))
                            ? MemDataSize(mem)
                            : payload_size_padded;
    chunk_size = CHUNK_HEADER_SIZE + payload_available;
    if (SizeIsInvalid(mem, payload_size_padded)) return PARSE_ERROR;
    if (payload_size_padded > MemDataSize(mem)) status = PARSE_NEED_MORE_DATA;

    switch (fourcc) {
      case MKFOURCC('A', 'L', 'P', 'H'):
        if (alpha_chunks == 0) {
          ++alpha_chunks;
          frame.img_components[1].offset = chunk_start_offset;
          frame.img_components[1].size = chunk_size;
          frame.has_alpha = 1;
          frame.frame_num = frame_num;
          Skip(mem, payload_available);
        } else {
          goto Done;
        }
        break;
      case MKFOURCC('V', 'P', '8', 'L'):
        if (alpha_chunks > 0) return PARSE_ERROR;  // VP8L has its own alpha
        // fall through
      case MKFOURCC('V', 'P', '8', ' '):
        if (image_chunks == 0) {
          // Extract the bitstream features, tolerating failures when the data
          // is incomplete.
          WebPBitstreamFeatures features;
          const VP8StatusCode vp8_status = WebPGetFeatures(
              mem.buf + chunk_start_offset, chunk_size, &features);
          if (status == PARSE_NEED_MORE_DATA &&
              vp8_status == VP8_STATUS_NOT_ENOUGH_DATA) {
            return PARSE_NEED_MORE_DATA;
          } else if (vp8_status != VP8_STATUS_OK) {
            // We have enough data, and yet WebPGetFeatures() failed.
            return PARSE_ERROR;
          }
          ++image_chunks;
          SetFrameInfo(chunk_start_offset, chunk_size, frame_num, status == PARSE_OK, &features, frame);
          Skip(mem, payload_available);
        } else {
          goto Done;
        }
        break;
      Done:
      default:
        // Restore fourcc/size when moving up one level in parsing.
        Rewind(mem, CHUNK_HEADER_SIZE);
        done = 1;
        break;
    }

    if (mem.start == mem.riff_end) {
      done = 1;
    } else if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {
      status = PARSE_NEED_MORE_DATA;
    }
  } while (!done && status == PARSE_OK);

  return status;
}

// Creates a new Frame if 'actual_size' is within bounds and 'mem' contains
// enough data ('min_size') to parse the payload.
// Returns PARSE_OK on success with pointing to the *frame new Frame.
// Returns PARSE_NEED_MORE_DATA with insufficient data, PARSE_ERROR otherwise.
static ParseStatus NewFrame(const const mem *MemBuffer, uint32 min_size, uint32 actual_size, *Frame* frame) {
  if (SizeIsInvalid(mem, min_size)) return PARSE_ERROR;
  if (actual_size < min_size) return PARSE_ERROR;
  if (MemDataSize(mem) < min_size) return PARSE_NEED_MORE_DATA;

  *frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(**frame));
  return (*frame == nil) ? PARSE_ERROR : PARSE_OK;
}

// Parse a 'ANMF' chunk and any image bearing chunks that immediately follow.
// 'frame_chunk_size' is the previously validated, padded chunk size.
static ParseStatus ParseAnimationFrame(const dmux *WebPDemuxer, uint32 frame_chunk_size) {
  const int is_animation = !!(dmux.feature_flags & ANIMATION_FLAG);
  const uint32 anmf_payload_size = frame_chunk_size - ANMF_CHUNK_SIZE;
  int added_frame = 0;
  int bits;
  const mem *MemBuffer = &dmux.mem;
  frame *Frame;
  uint64 start_offset;
  ParseStatus status = NewFrame(mem, ANMF_CHUNK_SIZE, frame_chunk_size, &frame);
  if (status != PARSE_OK) return status;

  frame.x_offset = 2 * ReadLE24s(mem);
  frame.y_offset = 2 * ReadLE24s(mem);
  frame.width = 1 + ReadLE24s(mem);
  frame.height = 1 + ReadLE24s(mem);
  frame.duration = ReadLE24s(mem);
  bits = ReadByte(mem);
  frame.dispose_method =
      (bits & 1) ? WEBP_MUX_DISPOSE_BACKGROUND : WEBP_MUX_DISPOSE_NONE;
  frame.blend_method = (bits & 2) ? WEBP_MUX_NO_BLEND : WEBP_MUX_BLEND;
  if (frame.width * (uint64)frame.height >= MAX_IMAGE_AREA) {
    WebPSafeFree(frame);
    return PARSE_ERROR;
  }

  // Store a frame only if the animation flag is set there is some data for
  // this frame is available.
  start_offset = mem.start;
  status = StoreFrame(dmux.num_frames + 1, anmf_payload_size, mem, frame);
  if (status != PARSE_ERROR && mem.start - start_offset > anmf_payload_size) {
    status = PARSE_ERROR;
  }
  if (status != PARSE_ERROR && is_animation && frame.frame_num > 0) {
    added_frame = AddFrame(dmux, frame);
    if (added_frame) {
      ++dmux.num_frames;
    } else {
      status = PARSE_ERROR;
    }
  }

  if (!added_frame) WebPSafeFree(frame);
  return status;
}

// General chunk storage, starting with the header at 'start_offset', allowing
// the user to request the payload via a fourcc string. 'size' includes the
// header and the unpadded payload size.
// Returns true on success, false otherwise.
static int StoreChunk(const dmux *WebPDemuxer, uint64 start_offset, uint32 size) {
  const chunk *Chunk = (*Chunk)WebPSafeCalloc(uint64(1), sizeof(*chunk));
  if (chunk == nil) return 0;

  chunk.data.offset = start_offset;
  chunk.data.size = size;
  AddChunk(dmux, chunk);
  return 1;
}

// -----------------------------------------------------------------------------
// Primary chunk parsing

static ParseStatus ReadHeader(const mem *MemBuffer) {
  const uint64 min_size = RIFF_HEADER_SIZE + CHUNK_HEADER_SIZE;
  uint32 riff_size;

  // Basic file level validation.
  if (MemDataSize(mem) < min_size) return PARSE_NEED_MORE_DATA;
  if (memcmp(GetBuffer(mem), "RIFF", CHUNK_SIZE_BYTES) ||
      memcmp(GetBuffer(mem) + CHUNK_HEADER_SIZE, "WEBP", CHUNK_SIZE_BYTES)) {
    return PARSE_ERROR;
  }

  riff_size = GetLE32(GetBuffer(mem) + TAG_SIZE);
  if (riff_size < CHUNK_HEADER_SIZE) return PARSE_ERROR;
  if (riff_size > MAX_CHUNK_PAYLOAD) return PARSE_ERROR;

  // There's no point in reading past the end of the RIFF chunk
  mem.riff_end = riff_size + CHUNK_HEADER_SIZE;
  if (mem.buf_size > mem.riff_end) {
    mem.buf_size = mem.end = mem.riff_end;
  }

  Skip(mem, RIFF_HEADER_SIZE);
  return PARSE_OK;
}

static ParseStatus ParseSingleImage(const dmux *WebPDemuxer) {
  const uint64 min_size = CHUNK_HEADER_SIZE;
  const mem *MemBuffer = &dmux.mem;
  frame *Frame;
  ParseStatus status;
  int image_added = 0;

  if (dmux.frames != nil) return PARSE_ERROR;
  if (SizeIsInvalid(mem, min_size)) return PARSE_ERROR;
  if (MemDataSize(mem) < min_size) return PARSE_NEED_MORE_DATA;

  frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(*frame));
  if (frame == nil) return PARSE_ERROR;

  // For the single image case we allow parsing of a partial frame, so no
  // minimum size is imposed here.
  status = StoreFrame(1, 0, &dmux.mem, frame);
  if (status != PARSE_ERROR) {
    const int has_alpha = !!(dmux.feature_flags & ALPHA_FLAG);
    // Clear any alpha when the alpha flag is missing.
    if (!has_alpha && frame.img_components[1].size > 0) {
      frame.img_components[1].offset = 0;
      frame.img_components[1].size = 0;
      frame.has_alpha = 0;
    }

    // Use the frame width/height as the canvas values for non-vp8x files.
    // Also, set ALPHA_FLAG if this is a lossless image with alpha.
    if (!dmux.is_ext_format && frame.width > 0 && frame.height > 0) {
      dmux.state = WEBP_DEMUX_PARSED_HEADER;
      dmux.canvas_width = frame.width;
      dmux.canvas_height = frame.height;
      dmux.feature_flags |= tenary.If(frame.has_alpha, ALPHA_FLAG, 0);
    }
    if (!AddFrame(dmux, frame)) {
      status = PARSE_ERROR;  // last frame was left incomplete
    } else {
      image_added = 1;
      dmux.num_frames = 1;
    }
  }

  if (!image_added) WebPSafeFree(frame);
  return status;
}

static ParseStatus ParseVP8XChunks(const dmux *WebPDemuxer) {
  const int is_animation = !!(dmux.feature_flags & ANIMATION_FLAG);
  const mem *MemBuffer = &dmux.mem;
  int anim_chunks = 0;
  ParseStatus status = PARSE_OK;

  do {
    int store_chunk = 1;
    const uint64 chunk_start_offset = mem.start;
    const uint32 fourcc = ReadLE32(mem);
    const uint32 chunk_size = ReadLE32(mem);
    uint32 chunk_size_padded;

    if (chunk_size > MAX_CHUNK_PAYLOAD) return PARSE_ERROR;

    chunk_size_padded = chunk_size + (chunk_size & 1);
    if (SizeIsInvalid(mem, chunk_size_padded)) return PARSE_ERROR;

    switch (fourcc) {
      case MKFOURCC('V', 'P', '8', 'X'): {
        return PARSE_ERROR;
      }
      case MKFOURCC('A', 'L', 'P', 'H'):
      case MKFOURCC('V', 'P', '8', ' '):
      case MKFOURCC('V', 'P', '8', 'L'): {
        // check that this isn't an animation (all frames should be in an ANMF).
        if (anim_chunks > 0 || is_animation) return PARSE_ERROR;

        Rewind(mem, CHUNK_HEADER_SIZE);
        status = ParseSingleImage(dmux);
        break;
      }
      case MKFOURCC('A', 'N', 'I', 'M'): {
        if (chunk_size_padded < ANIM_CHUNK_SIZE) return PARSE_ERROR;

        if (MemDataSize(mem) < chunk_size_padded) {
          status = PARSE_NEED_MORE_DATA;
        } else if (anim_chunks == 0) {
          ++anim_chunks;
          dmux.bgcolor = ReadLE32(mem);
          dmux.loop_count = ReadLE16s(mem);
          Skip(mem, chunk_size_padded - ANIM_CHUNK_SIZE);
        } else {
          store_chunk = 0;
          goto Skip;
        }
        break;
      }
      case MKFOURCC('A', 'N', 'M', 'F'): {
        if (anim_chunks == 0) return PARSE_ERROR;  // 'ANIM' precedes frames.
        status = ParseAnimationFrame(dmux, chunk_size_padded);
        break;
      }
      case MKFOURCC('I', 'C', 'C', 'P'): {
        store_chunk = !!(dmux.feature_flags & ICCP_FLAG);
        goto Skip;
      }
      case MKFOURCC('E', 'X', 'I', 'F'): {
        store_chunk = !!(dmux.feature_flags & EXIF_FLAG);
        goto Skip;
      }
      case MKFOURCC('X', 'M', 'P', ' '): {
        store_chunk = !!(dmux.feature_flags & XMP_FLAG);
        goto Skip;
      }
      Skip:
      default: {
        if (chunk_size_padded <= MemDataSize(mem)) {
          if (store_chunk) {
            // Store only the chunk header and unpadded size as only the payload
            // will be returned to the user.
            if (!StoreChunk(dmux, chunk_start_offset, CHUNK_HEADER_SIZE + chunk_size)) {
              return PARSE_ERROR;
            }
          }
          Skip(mem, chunk_size_padded);
        } else {
          status = PARSE_NEED_MORE_DATA;
        }
      }
    }

    if (mem.start == mem.riff_end) {
      break;
    } else if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {
      status = PARSE_NEED_MORE_DATA;
    }
  } while (status == PARSE_OK);

  return status;
}

static ParseStatus ParseVP8X(const dmux *WebPDemuxer) {
  const mem *MemBuffer = &dmux.mem;
  uint32 vp8x_size;

  if (MemDataSize(mem) < CHUNK_HEADER_SIZE) return PARSE_NEED_MORE_DATA;

  dmux.is_ext_format = 1;
  Skip(mem, TAG_SIZE);  // VP8X
  vp8x_size = ReadLE32(mem);
  if (vp8x_size > MAX_CHUNK_PAYLOAD) return PARSE_ERROR;
  if (vp8x_size < VP8X_CHUNK_SIZE) return PARSE_ERROR;
  vp8x_size += vp8x_size & 1;
  if (SizeIsInvalid(mem, vp8x_size)) return PARSE_ERROR;
  if (MemDataSize(mem) < vp8x_size) return PARSE_NEED_MORE_DATA;

  dmux.feature_flags = ReadByte(mem);
  Skip(mem, 3);  // Reserved.
  dmux.canvas_width = 1 + ReadLE24s(mem);
  dmux.canvas_height = 1 + ReadLE24s(mem);
  if (dmux.canvas_width * (uint64)dmux.canvas_height >= MAX_IMAGE_AREA) {
    return PARSE_ERROR;  // image final dimension is too large
  }
  Skip(mem, vp8x_size - VP8X_CHUNK_SIZE);  // skip any trailing data.
  dmux.state = WEBP_DEMUX_PARSED_HEADER;

  if (SizeIsInvalid(mem, CHUNK_HEADER_SIZE)) return PARSE_ERROR;
  if (MemDataSize(mem) < CHUNK_HEADER_SIZE) return PARSE_NEED_MORE_DATA;

  return ParseVP8XChunks(dmux);
}

// -----------------------------------------------------------------------------
// Format validation

static int IsValidSimpleFormat(const const dmux *WebPDemuxer) {
  const const frame *Frame = dmux.frames;
  if (dmux.state == WEBP_DEMUX_PARSING_HEADER) return 1;

  if (dmux.canvas_width <= 0 || dmux.canvas_height <= 0) return 0;
  if (dmux.state == WEBP_DEMUX_DONE && frame == nil) return 0;

  if (frame.width <= 0 || frame.height <= 0) return 0;
  return 1;
}

// If 'exact' is true, check that the image resolution matches the canvas.
// If 'exact' is false, check that the x/y offsets do not exceed the canvas.
static int CheckFrameBounds(const const frame *Frame, int exact, int canvas_width, int canvas_height) {
  if (exact) {
    if (frame.x_offset != 0 || frame.y_offset != 0) {
      return 0;
    }
    if (frame.width != canvas_width || frame.height != canvas_height) {
      return 0;
    }
  } else {
    if (frame.x_offset < 0 || frame.y_offset < 0) return 0;
    if (frame.width + frame.x_offset > canvas_width) return 0;
    if (frame.height + frame.y_offset > canvas_height) return 0;
  }
  return 1;
}

static int IsValidExtendedFormat(const const dmux *WebPDemuxer) {
  const int is_animation = !!(dmux.feature_flags & ANIMATION_FLAG);
  const f *Frame = dmux.frames;

  if (dmux.state == WEBP_DEMUX_PARSING_HEADER) return 1;

  if (dmux.canvas_width <= 0 || dmux.canvas_height <= 0) return 0;
  if (dmux.loop_count < 0) return 0;
  if (dmux.state == WEBP_DEMUX_DONE && dmux.frames == nil) return 0;
  if (dmux.feature_flags & ~ALL_VALID_FLAGS) return 0;  // invalid bitstream

  while (f != nil) {
    const int cur_frame_set = f.frame_num;

    // Check frame properties.
    for (; f != nil && f.frame_num == cur_frame_set; f = f.next) {
      const const image *ChunkData = f.img_components;
      const const alpha *ChunkData = f.img_components + 1;

      if (!is_animation && f.frame_num > 1) return 0;

      if (f.complete) {
        if (alpha.size == 0 && image.size == 0) return 0;
        // Ensure alpha precedes image bitstream.
        if (alpha.size > 0 && alpha.offset > image.offset) {
          return 0;
        }

        if (f.width <= 0 || f.height <= 0) return 0;
      } else {
        // There shouldn't be a partial frame in a complete file.
        if (dmux.state == WEBP_DEMUX_DONE) return 0;

        // Ensure alpha precedes image bitstream.
        if (alpha.size > 0 && image.size > 0 &&
            alpha.offset > image.offset) {
          return 0;
        }
        // There shouldn't be any frames after an incomplete one.
        if (f.next != nil) return 0;
      }

      if (f.width > 0 && f.height > 0 &&
          !CheckFrameBounds(f, !is_animation, dmux.canvas_width, dmux.canvas_height)) {
        return 0;
      }
    }
  }
  return 1;
}

// -----------------------------------------------------------------------------
// WebPDemuxer object

func InitDemux(const dmux *WebPDemuxer, const const mem *MemBuffer) {
  dmux.state = WEBP_DEMUX_PARSING_HEADER;
  dmux.loop_count = 1;
  dmux.bgcolor = 0xFFFFFFFF;  // White background by default.
  dmux.canvas_width = -1;
  dmux.canvas_height = -1;
  dmux.frames_tail = &dmux.frames;
  dmux.chunks_tail = &dmux.chunks;
  dmux.mem = *mem;
}

static ParseStatus CreateRawImageDemuxer(const mem *MemBuffer, *WebPDemuxer* demuxer) {
  WebPBitstreamFeatures features;
  const VP8StatusCode status =
      WebPGetFeatures(mem.buf, mem.buf_size, &features);
  *demuxer = nil;
  if (status != VP8_STATUS_OK) {
    return (status == VP8_STATUS_NOT_ENOUGH_DATA) ? PARSE_NEED_MORE_DATA
                                                  : PARSE_ERROR;
  }

  {
    const dmux *WebPDemuxer = (*WebPDemuxer)WebPSafeCalloc(uint64(1), sizeof(*dmux));
    const frame *Frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(*frame));
    if (dmux == nil || frame == nil) goto Error;
    InitDemux(dmux, mem);
    SetFrameInfo(0, mem.buf_size, 1 /*frame_*num/, 1 /**complete/, &features, frame);
    if (!AddFrame(dmux, frame)) goto Error;
    dmux.state = WEBP_DEMUX_DONE;
    dmux.canvas_width = frame.width;
    dmux.canvas_height = frame.height;
    dmux.feature_flags |= tenary.If(frame.has_alpha, ALPHA_FLAG, 0);
    dmux.num_frames = 1;
    assert.Assert(IsValidSimpleFormat(dmux));
    *demuxer = dmux;
    return PARSE_OK;

  Error:
    WebPSafeFree(dmux);
    WebPSafeFree(frame);
    return PARSE_ERROR;
  }
}

WebPDemuxInternal *WebPDemuxer(const data *WebPData, int allow_partial, state *WebPDemuxState, int version) {
  const parser *ChunkParser;
  int partial;
  ParseStatus status = PARSE_ERROR;
  MemBuffer mem;
  dmux *WebPDemuxer;

  if (state != nil) *state = WEBP_DEMUX_PARSE_ERROR;

  if (WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_DEMUX_ABI_VERSION)) return nil;
  if (data == nil || data.bytes == nil || data.size == 0) return nil;

  if (!InitMemBuffer(&mem, data.bytes, data.size)) return nil;
  status = ReadHeader(&mem);
  if (status != PARSE_OK) {
    // If parsing of the webp file header fails attempt to handle a raw
    // VP8/VP8L frame. Note 'allow_partial' is ignored in this case.
    if (status == PARSE_ERROR) {
      status = CreateRawImageDemuxer(&mem, &dmux);
      if (status == PARSE_OK) {
        if (state != nil) *state = WEBP_DEMUX_DONE;
        return dmux;
      }
    }
    if (state != nil) {
      *state = (status == PARSE_NEED_MORE_DATA) ? WEBP_DEMUX_PARSING_HEADER
                                                : WEBP_DEMUX_PARSE_ERROR;
    }
    return nil;
  }

  partial = (mem.buf_size < mem.riff_end);
  if (!allow_partial && partial) return nil;

  dmux = (*WebPDemuxer)WebPSafeCalloc(uint64(1), sizeof(*dmux));
  if (dmux == nil) return nil;
  InitDemux(dmux, &mem);

  status = PARSE_ERROR;
  for (parser = kMasterChunks; parser.parse != nil; ++parser) {
    if (!memcmp(parser.id, GetBuffer(&dmux.mem), TAG_SIZE)) {
      status = parser.parse(dmux);
      if (status == PARSE_OK) dmux.state = WEBP_DEMUX_DONE;
      if (status == PARSE_NEED_MORE_DATA && !partial) status = PARSE_ERROR;
      if (status != PARSE_ERROR && !parser.valid(dmux)) status = PARSE_ERROR;
      if (status == PARSE_ERROR) dmux.state = WEBP_DEMUX_PARSE_ERROR;
      break;
    }
  }
  if (state != nil) *state = dmux.state;

  if (status == PARSE_ERROR) {
    WebPDemuxDelete(dmux);
    return nil;
  }
  return dmux;
}

func WebPDemuxDelete(dmux *WebPDemuxer) {
  c *Chunk;
  f *Frame;
  if (dmux == nil) return;

  for (f = dmux.frames; f != nil;) {
    const cur_frame *Frame = f;
    f = f.next;
    WebPSafeFree(cur_frame);
  }
  for (c = dmux.chunks; c != nil;) {
    const cur_chunk *Chunk = c;
    c = c.next;
    WebPSafeFree(cur_chunk);
  }
  WebPSafeFree(dmux);
}

// -----------------------------------------------------------------------------

uint32 WebPDemuxGetI(const dmux *WebPDemuxer, WebPFormatFeature feature) {
  if (dmux == nil) return 0;

  switch (feature) {
    case WEBP_FF_FORMAT_FLAGS:
      return dmux.feature_flags;
    case WEBP_FF_CANVAS_WIDTH:
      return (uint32)dmux.canvas_width;
    case WEBP_FF_CANVAS_HEIGHT:
      return (uint32)dmux.canvas_height;
    case WEBP_FF_LOOP_COUNT:
      return (uint32)dmux.loop_count;
    case WEBP_FF_BACKGROUND_COLOR:
      return dmux.bgcolor;
    case WEBP_FF_FRAME_COUNT:
      return (uint32)dmux.num_frames;
  }
  return 0;
}

// -----------------------------------------------------------------------------
// Frame iteration

static const GetFrame *Frame(const const dmux *WebPDemuxer, int frame_num) {
  const f *Frame;
  for (f = dmux.frames; f != nil; f = f.next) {
    if (frame_num == f.frame_num) break;
  }
  return f;
}

static const GetFramePayload *uint8(const const mem_buf *uint8, const const frame *Frame, const data_size *uint64) {
  *data_size = 0;
  if (frame != nil) {
    const const image *ChunkData = frame.img_components;
    const const alpha *ChunkData = frame.img_components + 1;
    uint64 start_offset = image.offset;
    *data_size = image.size;

    // if alpha exists it precedes image, update the size allowing for
    // intervening chunks.
    if (alpha.size > 0) {
      const uint64 inter_size =
          (image.offset > 0) ? image.offset - (alpha.offset + alpha.size)
                              : 0;
      start_offset = alpha.offset;
      *data_size += alpha.size + inter_size;
    }
    return mem_buf + start_offset;
  }
  return nil;
}

// Create a whole 'frame' from VP8 (+ alpha) or lossless.
static int SynthesizeFrame(const const dmux *WebPDemuxer, const const frame *Frame, const iter *WebPIterator) {
  const const mem_buf *uint8 = dmux.mem.buf;
  uint64 payload_size = 0;
  const const payload *uint8 = GetFramePayload(mem_buf, frame, &payload_size);
  if (payload == nil) return 0;
  assert.Assert(frame != nil);

  iter.frame_num = frame.frame_num;
  iter.num_frames = dmux.num_frames;
  iter.x_offset = frame.x_offset;
  iter.y_offset = frame.y_offset;
  iter.width = frame.width;
  iter.height = frame.height;
  iter.has_alpha = frame.has_alpha;
  iter.duration = frame.duration;
  iter.dispose_method = frame.dispose_method;
  iter.blend_method = frame.blend_method;
  iter.complete = frame.complete;
  iter.fragment.bytes = payload;
  iter.fragment.size = payload_size;
  return 1;
}

static int SetFrame(int frame_num, const iter *WebPIterator) {
  const frame *Frame;
  const const dmux *WebPDemuxer = (*WebPDemuxer)iter.private_;
  if (dmux == nil || frame_num < 0) return 0;
  if (frame_num > dmux.num_frames) return 0;
  if (frame_num == 0) frame_num = dmux.num_frames;

  frame = GetFrame(dmux, frame_num);
  if (frame == nil) return 0;

  return SynthesizeFrame(dmux, frame, iter);
}

int WebPDemuxGetFrame(const dmux *WebPDemuxer, int frame, iter *WebPIterator) {
  if (iter == nil) return 0;

  WEBP_UNSAFE_MEMSET(iter, 0, sizeof(*iter));
  iter.private_ = (*void)dmux;
  return SetFrame(frame, iter);
}

int WebPDemuxNextFrame(iter *WebPIterator) {
  if (iter == nil) return 0;
  return SetFrame(iter.frame_num + 1, iter);
}

int WebPDemuxPrevFrame(iter *WebPIterator) {
  if (iter == nil) return 0;
  if (iter.frame_num <= 1) return 0;
  return SetFrame(iter.frame_num - 1, iter);
}

func WebPDemuxReleaseIterator(iter *WebPIterator) { (void)iter; }

// -----------------------------------------------------------------------------
// Chunk iteration

static int ChunkCount(const const dmux *WebPDemuxer, const byte fourcc[4]) {
  const const mem_buf *uint8 = dmux.mem.buf;
  const c *Chunk;
  int count = 0;
  for (c = dmux.chunks; c != nil; c = c.next) {
    const const header *uint8 = mem_buf + c.data.offset;
    if (!memcmp(header, fourcc, TAG_SIZE)) ++count;
  }
  return count;
}

static const GetChunk *Chunk(const const dmux *WebPDemuxer, const byte fourcc[4], int chunk_num) {
  const const mem_buf *uint8 = dmux.mem.buf;
  const c *Chunk;
  int count = 0;
  for (c = dmux.chunks; c != nil; c = c.next) {
    const const header *uint8 = mem_buf + c.data.offset;
    if (!memcmp(header, fourcc, TAG_SIZE)) ++count;
    if (count == chunk_num) break;
  }
  return c;
}

static int SetChunk(const byte fourcc[4], int chunk_num, const iter *WebPChunkIterator) {
  const const dmux *WebPDemuxer = (*WebPDemuxer)iter.private_;
  int count;

  if (dmux == nil || fourcc == nil || chunk_num < 0) return 0;
  count = ChunkCount(dmux, fourcc);
  if (count == 0) return 0;
  if (chunk_num == 0) chunk_num = count;

  if (chunk_num <= count) {
    const const mem_buf *uint8 = dmux.mem.buf;
    const const chunk *Chunk = GetChunk(dmux, fourcc, chunk_num);
    iter.chunk.bytes = mem_buf + chunk.data.offset + CHUNK_HEADER_SIZE;
    iter.chunk.size = chunk.data.size - CHUNK_HEADER_SIZE;
    iter.num_chunks = count;
    iter.chunk_num = chunk_num;
    return 1;
  }
  return 0;
}

int WebPDemuxGetChunk(const dmux *WebPDemuxer, const byte fourcc[4], int chunk_num, iter *WebPChunkIterator) {
  if (iter == nil) return 0;

  WEBP_UNSAFE_MEMSET(iter, 0, sizeof(*iter));
  iter.private_ = (*void)dmux;
  return SetChunk(fourcc, chunk_num, iter);
}

func WebPDemuxNextChunk( iter *WebPChunkIterator) int {
  if (iter != nil) {
    const const fourcc *byte =
        (const *byte)iter.chunk.bytes - CHUNK_HEADER_SIZE;
    return SetChunk(fourcc, iter.chunk_num + 1, iter);
  }
  return 0;
}

func WebPDemuxPrevChunk( iter *WebPChunkIterator) int {
  if (iter != nil && iter.chunk_num > 1) {
    const const fourcc *byte =
        (const *byte)iter.chunk.bytes - CHUNK_HEADER_SIZE;
    return SetChunk(fourcc, iter.chunk_num - 1, iter);
  }
  return 0;
}

func WebPDemuxReleaseChunkIterator(iter *WebPChunkIterator) { (void)iter; }
