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

import (
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
	 "github.com/daanv2/go-webp/pkg/assert"
	 "github.com/daanv2/go-webp/pkg/stdlib"
	 "github.com/daanv2/go-webp/pkg/string"
	 "github.com/daanv2/go-webp/pkg/libwebp/utils"
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"  // WebPGetFeatures
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
	 "github.com/daanv2/go-webp/pkg/libwebp/webp"
)

type ParseSingleImage = func(dmux *WebPDemuxer) ParseStatus
type ParseVP8X = func(dmux *WebPDemuxer) ParseStatus
type IsValidSimpleFormat = func(dmux *WebPDemuxer) int
type IsValidExtendedFormat = func(dmux *WebPDemuxer) int


func RemapMemBuffer(/* const */ mem *MemBuffer, data *uint8,  size uint64) int {
  if (size < mem.buf_size) { return 0 }  // can't remap to a shorter buffer!

  mem.buf = data
  mem.buf_size = size
  mem.end = size

  return 1
}

func InitMemBuffer(/* const */ mem *MemBuffer, data *uint8, size uint64) int {
  stdlib.Memset(mem, 0, sizeof(*mem))
  return RemapMemBuffer(mem, data, size)
}

// Return the remaining data size available in 'mem'.
func MemDataSize(/* const */ mem *MemBuffer) uint64 {
  return (mem.end - mem.start)
}

// Return true if 'size' exceeds the end of the RIFF chunk.
func SizeIsInvalid(/* const */ mem *MemBuffer, size uint64 ) int {
  return (size > mem.riff_end - mem.start)
}

func Skip(/* const */ mem *MemBuffer, size uint64 ) {
  mem.start += size
}

func Rewind(/* const */ mem *MemBuffer, size uint64 ) {
  mem.start -= size
}

func GetBuffer(/* const */ mem *MemBuffer) /* const */ *uint8 {
  return mem.buf + mem.start
}

// Read from 'mem' and skip the read bytes.
func ReadByte(/* const */ mem *MemBuffer)   uint8 {
  v = mem.buf[mem.start]
  Skip(mem, 1)
  return v
}

func ReadLE16s(/* const */ mem *MemBuffer) int {
  data *uint8 = mem.buf + mem.start
  val = GetLE16(data)
  Skip(mem, 2)
  return val
}

func ReadLE24s(/* const */ mem *MemBuffer) int {
  data *uint8 = mem.buf + mem.start
   val = GetLE24(data)
  Skip(mem, 3)
  return val
}

func ReadLE32(/* const */ mem *MemBuffer) uint32 {
  data *uint8 = mem.buf + mem.start
  val = GetLE32(data)
  Skip(mem, 4)
  return val
}

// -----------------------------------------------------------------------------
// Secondary chunk parsing
func AddChunk(/* const */ dmux *WebPDemuxer, /* const */ chunk *Chunk) {
  *dmux.chunks_tail = chunk
  chunk.next = nil
  dmux.chunks_tail = &chunk.next
}

// Add a frame to the end of the list, ensuring the last frame is complete.
// Returns true on success, false otherwise.
func AddFrame(/* const */ dmux *WebPDemuxer, /* const */ frame *Frame) int {
  last_frame *Frame = *dmux.frames_tail
  if (last_frame != nil && !last_frame.complete) {
	return 0
  }

  *dmux.frames_tail = frame
  frame.next = nil
  dmux.frames_tail = &frame.next

  return 1
}

func SetFrameInfo(start_offset, size uint64, frame_num, complete int , features *WebPBitstreamFeatures, frame *Frame) {
  frame.img_components[0].offset = start_offset
  frame.img_components[0].size = size
  frame.width = features.width
  frame.height = features.height
  frame.has_alpha |= features.has_alpha
  frame.frame_num = frame_num
  frame.complete = complete
}

// Store image bearing chunks to 'frame'. 'min_size' is an optional size
// requirement, it may be zero.
func StoreFrame(frame_num int, min_size uint32, /* const */ mem *MemBuffer, frame *Frame)  ParseStatus {
  alpha_chunks = 0
  image_chunks = 0
  done = (MemDataSize(mem) < CHUNK_HEADER_SIZE || MemDataSize(mem) < min_size)
  status := PARSE_OK

  if (done) {
	return PARSE_NEED_MORE_DATA
  }

  for {
    chunk_start_offset = mem.start
    fourcc = ReadLE32(mem)
    payload_size = ReadLE32(mem)
    var payload_size_padded uint32
    var payload_available uint64
    var chunk_size uint64

    if (payload_size > MAX_CHUNK_PAYLOAD) {
		return PARSE_ERROR
	}

    payload_size_padded = payload_size + (payload_size & 1)
    payload_available = tenary.If(payload_size_padded > MemDataSize(mem), MemDataSize(mem), payload_size_padded)

    chunk_size = CHUNK_HEADER_SIZE + payload_available
    if (SizeIsInvalid(mem, payload_size_padded)) {return PARSE_ERROR}
    if (payload_size_padded > MemDataSize(mem)) {status = PARSE_NEED_MORE_DATA
}
    switch (fourcc) {
      case MKFOURCC('A', 'L', 'P', 'H'):
        if (alpha_chunks == 0) {
          alpha_chunks++
          frame.img_components[1].offset = chunk_start_offset
          frame.img_components[1].size = chunk_size
          frame.has_alpha = 1
          frame.frame_num = frame_num
          Skip(mem, payload_available)
        } else {
          goto Done
        }
        break
      case MKFOURCC('V', 'P', '8', 'L'):
        if alpha_chunks > 0 {
    return PARSE_ERROR  // VP8L has its own alpha
}
        // fall through
      case MKFOURCC('V', 'P', '8', ' '):
        if (image_chunks == 0) {
          // Extract the bitstream features, tolerating failures when the data
          // is incomplete.
          var features WebPBitstreamFeatures
          const VP8StatusCode vp8_status = WebPGetFeatures(mem.buf + chunk_start_offset, chunk_size, &features)
          if (status == PARSE_NEED_MORE_DATA &&
              vp8_status == VP8_STATUS_NOT_ENOUGH_DATA) {
            return PARSE_NEED_MORE_DATA
          } else if (vp8_status != VP8_STATUS_OK) {
            // We have enough data, and yet WebPGetFeatures() failed.
            return PARSE_ERROR
          }
          image_chunks++
          SetFrameInfo(chunk_start_offset, chunk_size, frame_num, status == PARSE_OK, &features, frame)
          Skip(mem, payload_available)
        } else {
          goto Done
        }
        break
      default:
      Done:
        // Restore fourcc/size when moving up one level in parsing.
        Rewind(mem, CHUNK_HEADER_SIZE)
        done = 1
        break
    }

    if (mem.start == mem.riff_end) {
      done = 1
    } else if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {
      status = PARSE_NEED_MORE_DATA
    }

	if (!done && status == PARSE_OK) {
		continue
	} else {
		break
	}
  }

  return status
}

// Creates a new Frame if 'actual_size' is within bounds and 'mem' contains
// enough data ('min_size') to parse the payload.
// Returns PARSE_OK on success with pointing to the *frame new Frame.
// Returns PARSE_NEED_MORE_DATA with insufficient data, PARSE_ERROR otherwise.
func NewFrame(/* const */ mem *MemBuffer, min_size uint32 , actual_size uint32, frame *Frame) ParseStatus {
  if (SizeIsInvalid(mem, min_size)) {return PARSE_ERROR}
  if (actual_size < min_size) {return PARSE_ERROR}
  if (MemDataSize(mem) < min_size) {return PARSE_NEED_MORE_DATA}

//   *frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(**frame))
  frame = &Frame{}

  return PARSE_OK
}

// Parse a 'ANMF' chunk and any image bearing chunks that immediately follow.
// 'frame_chunk_size' is the previously validated, padded chunk size.
func ParseAnimationFrame(/* const */ dmux *WebPDemuxer, frame_chunk_size uint32 ) ParseStatus {
  is_animation := !!(dmux.feature_flags & ANIMATION_FLAG)
  anmf_payload_size := frame_chunk_size - ANMF_CHUNK_SIZE
  added_frame := 0
  var bits int 
  /* const */ mem *MemBuffer = &dmux.mem
  frame *Frame
  var start_offset uint64 
  var status ParseStatus = NewFrame(mem, ANMF_CHUNK_SIZE, frame_chunk_size, &frame)
  if (status != PARSE_OK) {return status}

  frame.x_offset = 2 * ReadLE24s(mem)
  frame.y_offset = 2 * ReadLE24s(mem)
  frame.width = 1 + ReadLE24s(mem)
  frame.height = 1 + ReadLE24s(mem)
  frame.duration = ReadLE24s(mem)
  bits = ReadByte(mem)
  frame.dispose_method = tenary.If(bits & 1, WEBP_MUX_DISPOSE_BACKGROUND, WEBP_MUX_DISPOSE_NONE)
  frame.blend_method = tenary.If(bits & 2, WEBP_MUX_NO_BLEND, WEBP_MUX_BLEND)
  if (frame.width * frame.height >= MAX_IMAGE_AREA) {
    return PARSE_ERROR
  }

  // Store a frame only if the animation flag is set there is some data for
  // this frame is available.
  start_offset = mem.start
  status = StoreFrame(dmux.num_frames + 1, anmf_payload_size, mem, frame)
  if (status != PARSE_ERROR && mem.start - start_offset > anmf_payload_size) {
    status = PARSE_ERROR
  }
  if (status != PARSE_ERROR && is_animation && frame.frame_num > 0) {
    added_frame = AddFrame(dmux, frame)
    if (added_frame) {
      dmux.num_frames++
    } else {
      status = PARSE_ERROR
    }
  }

  return status
}

// General chunk storage, starting with the header at 'start_offset', allowing
// the user to request the payload via a fourcc string. 'size' includes the
// header and the unpadded payload size.
// Returns true on success, false otherwise.
func StoreChunk(/* const */ dmux *WebPDemuxer, start_offset uint64 , size uint32) int {
//   var chunk *Chunk = (*Chunk)WebPSafeCalloc(uint64(1), sizeof(*chunk))
//   if (chunk == nil){ return 0}
  chunk := &Chunk{
	data: ChunkData{
		offset: start_offset,
		size:   size,
	},
  }

  AddChunk(dmux, chunk)
  return 1
}

// -----------------------------------------------------------------------------
// Primary chunk parsing

func ReadHeader(/* const */ mem *MemBuffer) ParseStatus {
  min_size := RIFF_HEADER_SIZE + CHUNK_HEADER_SIZE
  var riff_size uint32 

  // Basic file level validation.
  if (MemDataSize(mem) < min_size) { return PARSE_NEED_MORE_DATA }
  if (memcmp(GetBuffer(mem), "RIFF", CHUNK_SIZE_BYTES) ||
      memcmp(GetBuffer(mem) + CHUNK_HEADER_SIZE, "WEBP", CHUNK_SIZE_BYTES)) {
    return PARSE_ERROR
  }

  riff_size = GetLE32(GetBuffer(mem) + TAG_SIZE)
  if (riff_size < CHUNK_HEADER_SIZE){ return PARSE_ERROR}
  if (riff_size > MAX_CHUNK_PAYLOAD) {return PARSE_ERROR}

  // There's no point in reading past the end of the RIFF chunk
  mem.riff_end = riff_size + CHUNK_HEADER_SIZE
  if (mem.buf_size > mem.riff_end) {
    mem.end = mem.riff_end
	mem.buf_size = mem.riff_end
  }

  Skip(mem, RIFF_HEADER_SIZE)
  return PARSE_OK
}

func ParseSingleImage(/* const */ dmux *WebPDemuxer) ParseStatus {
  min_size := CHUNK_HEADER_SIZE
  /* const */ var mem *MemBuffer = &dmux.mem
  var frame *Frame
  var status ParseStatus
  image_added := 0

  if (dmux.frames != nil) {return PARSE_ERROR}
  if (SizeIsInvalid(mem, min_size)) {return PARSE_ERROR}
  if (MemDataSize(mem) < min_size) {return PARSE_NEED_MORE_DATA}

//   frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(*frame))
//   if (frame == nil) {return PARSE_ERROR}
	frame = &Frame{}

  // For the single image case we allow parsing of a partial frame, so no
  // minimum size is imposed here.
  status = StoreFrame(1, 0, &dmux.mem, frame)
  if (status != PARSE_ERROR) {
    has_alpha := !!(dmux.feature_flags & ALPHA_FLAG)
    // Clear any alpha when the alpha flag is missing.
    if (!has_alpha && frame.img_components[1].size > 0) {
      frame.img_components[1].offset = 0
      frame.img_components[1].size = 0
      frame.has_alpha = 0
    }

    // Use the frame width/height as the canvas values for non-vp8x files.
    // Also, set ALPHA_FLAG if this is a lossless image with alpha.
    if (!dmux.is_ext_format && frame.width > 0 && frame.height > 0) {
      dmux.state = WEBP_DEMUX_PARSED_HEADER
      dmux.canvas_width = frame.width
      dmux.canvas_height = frame.height
      dmux.feature_flags |= tenary.If(frame.has_alpha, ALPHA_FLAG, 0)
    }
    if (!AddFrame(dmux, frame)) {
      status = PARSE_ERROR  // last frame was left incomplete
    } else {
      image_added = 1
      dmux.num_frames = 1
    }
  }

  return status
}

func ParseVP8XChunks(/* const */ dmux *WebPDemuxer) ParseStatus {
  is_animation := !!(dmux.feature_flags & ANIMATION_FLAG)
  var mem *MemBuffer = &dmux.mem
  anim_chunks := 0
  var status ParseStatus = PARSE_OK

  for {
    store_chunk := 1
    chunk_start_offset := mem.start
    fourcc := ReadLE32(mem)
    chunk_size := ReadLE32(mem)
    var chunk_size_padded uint32

    if (chunk_size > MAX_CHUNK_PAYLOAD) {return PARSE_ERROR}

    chunk_size_padded = chunk_size + (chunk_size & 1)
    if (SizeIsInvalid(mem, chunk_size_padded)) {return PARSE_ERROR}

    switch (fourcc) {
      case MKFOURCC('V', 'P', '8', 'X'): {
        return PARSE_ERROR
      }
      case MKFOURCC('A', 'L', 'P', 'H'):
      case MKFOURCC('V', 'P', '8', ' '):
      case MKFOURCC('V', 'P', '8', 'L'): {
        // check that this isn't an animation (all frames should be in an ANMF).
        if (anim_chunks > 0 || is_animation) { return PARSE_ERROR }

        Rewind(mem, CHUNK_HEADER_SIZE)
        status = ParseSingleImage(dmux)
        break
      }
      case MKFOURCC('A', 'N', 'I', 'M'): {
        if (chunk_size_padded < ANIM_CHUNK_SIZE) { return PARSE_ERROR }

        if (MemDataSize(mem) < chunk_size_padded) {
          status = PARSE_NEED_MORE_DATA
        } else if (anim_chunks == 0) {
          anim_chunks++
          dmux.bgcolor = ReadLE32(mem)
          dmux.loop_count = ReadLE16s(mem)
          Skip(mem, chunk_size_padded - ANIM_CHUNK_SIZE)
        } else {
          store_chunk = 0
          goto Skip
        }
        break
      }
      case MKFOURCC('A', 'N', 'M', 'F'): {
        if anim_chunks == 0 {
    return PARSE_ERROR  // 'ANIM' precedes frames.
}
        status = ParseAnimationFrame(dmux, chunk_size_padded)
        break
      }
      case MKFOURCC('I', 'C', 'C', 'P'): {
        store_chunk = !!(dmux.feature_flags & ICCP_FLAG)
        goto Skip
      }
      case MKFOURCC('E', 'X', 'I', 'F'): {
        store_chunk = !!(dmux.feature_flags & EXIF_FLAG)
        goto Skip
      }
      case MKFOURCC('X', 'M', 'P', ' '): {
        store_chunk = !!(dmux.feature_flags & XMP_FLAG)
        goto Skip
      }
      default: {
      Skip:
        if (chunk_size_padded <= MemDataSize(mem)) {
          if (store_chunk) {
            // Store only the chunk header and unpadded size as only the payload
            // will be returned to the user.
            if (!StoreChunk(dmux, chunk_start_offset, CHUNK_HEADER_SIZE + chunk_size)) {
              return PARSE_ERROR
            }
          }
          Skip(mem, chunk_size_padded)
        } else {
          status = PARSE_NEED_MORE_DATA
        }
      }
    }

    if (mem.start == mem.riff_end) {
      break
    } else if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {
      status = PARSE_NEED_MORE_DATA
    }

	if (status == PARSE_OK) {
		continue
	} else {
		break
	}
  }

  return status
}

func ParseVP8X(/* const */ dmux *WebPDemuxer) ParseStatus {
  /* const */var mem *MemBuffer = &dmux.mem
  var vp8x_size uint32 

  if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {return PARSE_NEED_MORE_DATA}

  dmux.is_ext_format = 1
  Skip(mem, TAG_SIZE)  // VP8X
  vp8x_size = ReadLE32(mem)
  if (vp8x_size > MAX_CHUNK_PAYLOAD) {return PARSE_ERROR}
  if (vp8x_size < VP8X_CHUNK_SIZE) {return PARSE_ERROR}
  vp8x_size += vp8x_size & 1
  if (SizeIsInvalid(mem, vp8x_size)) {return PARSE_ERROR}
  if (MemDataSize(mem) < vp8x_size) {return PARSE_NEED_MORE_DATA}

  dmux.feature_flags = ReadByte(mem)
  Skip(mem, 3)  // Reserved.
  dmux.canvas_width = 1 + ReadLE24s(mem)
  dmux.canvas_height = 1 + ReadLE24s(mem)
  if (dmux.canvas_width * dmux.canvas_height >= MAX_IMAGE_AREA) {
    return PARSE_ERROR  // image final dimension is too large
  }
  Skip(mem, vp8x_size - VP8X_CHUNK_SIZE)  // skip any trailing data.
  dmux.state = WEBP_DEMUX_PARSED_HEADER

  if (SizeIsInvalid(mem, CHUNK_HEADER_SIZE)) {return PARSE_ERROR}
  if (MemDataSize(mem) < CHUNK_HEADER_SIZE) {return PARSE_NEED_MORE_DATA}

  return ParseVP8XChunks(dmux)
}

// -----------------------------------------------------------------------------
// Format validation

func IsValidSimpleFormat(/* const */ dmux *WebPDemuxer) int {
  var frame *Frame = dmux.frames
  if (dmux.state == WEBP_DEMUX_PARSING_HEADER) {return 1}

  if (dmux.canvas_width <= 0 || dmux.canvas_height <= 0) {return 0}
  if (dmux.state == WEBP_DEMUX_DONE && frame == nil) {return 0}

  if (frame.width <= 0 || frame.height <= 0) {return 0}
  return 1
}

// If 'exact' is true, check that the image resolution matches the canvas.
// If 'exact' is false, check that the x/y offsets do not exceed the canvas.
func CheckFrameBounds(/* const */ frame *Frame, exact int, canvas_width int, canvas_height int) int {
  if (exact) {
    if (frame.x_offset != 0 || frame.y_offset != 0) {
      return 0
    }
    if (frame.width != canvas_width || frame.height != canvas_height) {
      return 0
    }
  } else {
    if (frame.x_offset < 0 || frame.y_offset < 0) {return 0}
    if (frame.width + frame.x_offset > canvas_width) {return 0}
    if (frame.height + frame.y_offset > canvas_height) {return 0}
  }
  return 1
}

func IsValidExtendedFormat(/* const */ dmux *WebPDemuxer) int {
  is_animation := !!(dmux.feature_flags & ANIMATION_FLAG)
  var f *Frame = dmux.frames

  if (dmux.state == WEBP_DEMUX_PARSING_HEADER) {return 1}

  if (dmux.canvas_width <= 0 || dmux.canvas_height <= 0) {return 0}
  if (dmux.loop_count < 0) {return 0}
  if (dmux.state == WEBP_DEMUX_DONE && dmux.frames == nil) {return 0}
  if dmux.feature_flags & ~ALL_VALID_FLAGS {
    return 0  // invalid bitstream
	}

  for (f != nil) {
    cur_frame_set := f.frame_num

    // Check frame properties.
    for  f != nil && f.frame_num == cur_frame_set  {
		f = f.next
      var image *ChunkData = f.img_components
      var alpha *ChunkData = f.img_components + 1

      if (!is_animation && f.frame_num > 1) {return 0}

      if (f.complete) {
        if (alpha.size == 0 && image.size == 0) {return 0}
        // Ensure alpha precedes image bitstream.
        if (alpha.size > 0 && alpha.offset > image.offset) {
          return 0
        }

        if (f.width <= 0 || f.height <= 0) {return 0}
      } else {
        // There shouldn't be a partial frame in a complete file.
        if (dmux.state == WEBP_DEMUX_DONE) {return 0}

        // Ensure alpha precedes image bitstream.
        if (alpha.size > 0 && image.size > 0 &&
            alpha.offset > image.offset) {
          return 0
        }
        // There shouldn't be any frames after an incomplete one.
        if (f.next != nil) {return 0}
      }

      if (f.width > 0 && f.height > 0 &&
          !CheckFrameBounds(f, !is_animation, dmux.canvas_width, dmux.canvas_height)) {
        return 0
      }
    }
  }
  return 1
}

// -----------------------------------------------------------------------------
// WebPDemuxer object

func InitDemux(/* const */ dmux *WebPDemuxer, /*const*/ mem *MemBuffer) {
  dmux.state = WEBP_DEMUX_PARSING_HEADER
  dmux.loop_count = 1
  dmux.bgcolor = 0xFFFFFFFF  // White background by default.
  dmux.canvas_width = -1
  dmux.canvas_height = -1
  dmux.frames_tail = &dmux.frames
  dmux.chunks_tail = &dmux.chunks
  dmux.mem = *mem
}

func CreateRawImageDemuxer(/* const */ mem *MemBuffer, demuxer *WebPDemuxer) ParseStatus {
  var features WebPBitstreamFeatures
  var status VP8StatusCode  = WebPGetFeatures(mem.buf, mem.buf_size, &features)
  *demuxer = nil
  if (status != VP8_STATUS_OK) {
    return tenary.If(status == VP8_STATUS_NOT_ENOUGH_DATA, PARSE_NEED_MORE_DATA, PARSE_ERROR)
  }

  {
    // var dmux *WebPDemuxer = (*WebPDemuxer)WebPSafeCalloc(uint64(1), sizeof(*dmux))
    // var frame *Frame = (*Frame)WebPSafeCalloc(uint64(1), sizeof(*frame))
    // if (dmux == nil || frame == nil) goto Error
	dmux := &WebPDemuxer{}
	frame := &Frame{}

    InitDemux(dmux, mem)
    SetFrameInfo(0, mem.buf_size, 1 /*frame_num*/, 1 /*complete*/, &features, frame);
    if (!AddFrame(dmux, frame)) {goto Error}
    dmux.state = WEBP_DEMUX_DONE
    dmux.canvas_width = frame.width
    dmux.canvas_height = frame.height
    dmux.feature_flags |= tenary.If(frame.has_alpha, ALPHA_FLAG, 0)
    dmux.num_frames = 1
    assert.Assert(IsValidSimpleFormat(dmux))
    *demuxer = dmux
    return PARSE_OK

  Error:
    return PARSE_ERROR
  }
}

func WebPDemuxerFn(/* const */ data *WebPData, allow_partial int, state *WebPDemuxState, version int) *WebPDemuxInternal {
  var parser *ChunkParser
  var partial int 
  var status ParseStatus = PARSE_ERROR
  var mem MemBuffer
   var dmux *WebPDemuxer

  if (state != nil) {*state = WEBP_DEMUX_PARSE_ERROR}

  if (data == nil || data.bytes == nil || data.size == 0) { return nil }

  if (!InitMemBuffer(&mem, data.bytes, data.size)) { return nil }
  status = ReadHeader(&mem)
  if (status != PARSE_OK) {
    // If parsing of the webp file header fails attempt to handle a raw
    // VP8/VP8L frame. Note 'allow_partial' is ignored in this case.
    if (status == PARSE_ERROR) {
      status = CreateRawImageDemuxer(&mem, &dmux)
      if (status == PARSE_OK) {
        if (state != nil) {*state = WEBP_DEMUX_DONE}
        return dmux
      }
    }
    if (state != nil) {
      *state = tenary.If(status == PARSE_NEED_MORE_DATA, constants.WEBP_DEMUX_PARSING_HEADER, constants.WEBP_DEMUX_PARSE_ERROR)
    }
    return nil
  }

  partial = (mem.buf_size < mem.riff_end)
  if (!allow_partial && partial) { return nil }

//   dmux = (*WebPDemuxer)WebPSafeCalloc(uint64(1), sizeof(*dmux))
//   if (dmux == nil) { return nil }
  	dmux = &WebPDemuxer{}
	InitDemux(dmux, &mem)

  status = PARSE_ERROR
  for parser = kMasterChunks; parser.parse != nil; parser++ {
    if (!memcmp(parser.id, GetBuffer(&dmux.mem), TAG_SIZE)) {
      status = parser.parse(dmux)
      if (status == PARSE_OK) {dmux.state = WEBP_DEMUX_DONE}
      if (status == PARSE_NEED_MORE_DATA && !partial) {status = PARSE_ERROR}
      if (status != PARSE_ERROR && !parser.valid(dmux)) {status = PARSE_ERROR}
      if (status == PARSE_ERROR) {dmux.state = WEBP_DEMUX_PARSE_ERROR}
      break
    }
  }
  if (state != nil) {*state = dmux.state}

  if (status == PARSE_ERROR) {
    WebPDemuxDelete(dmux)
    return nil
  }
  return dmux
}

// Frees memory associated with 'dmux'.
func WebPDemuxDelete(dmux *WebPDemuxer) {
  c *Chunk
  f *Frame
  if (dmux == nil) {return}

  for f = dmux.frames; f != nil; {
    var cur_frame *Frame = f
    f = f.next
  }
  for c = dmux.chunks; c != nil; {
    var cur_chunk *Chunk = c
    c = c.next
  }
}

// Get the 'feature' value from the 'dmux'.
// NOTE: values are only valid if WebPDemux() was used or WebPDemuxPartial()
// returned a state > WEBP_DEMUX_PARSING_HEADER.
// If 'feature' is WEBP_FF_FORMAT_FLAGS, the returned value is a bit-wise
// combination of WebPFeatureFlags values.
// If 'feature' is WEBP_FF_LOOP_COUNT, WEBP_FF_BACKGROUND_COLOR, the returned
// value is only meaningful if the bitstream is animated.
func WebPDemuxGetI(/* const */ dmux *WebPDemuxer, WebPFormatFeature feature) uint32 {
  if (dmux == nil) { return 0 }

  switch (feature) {
    case constants.WEBP_FF_FORMAT_FLAGS:
      return dmux.feature_flags
    case constants.WEBP_FF_CANVAS_WIDTH:
      return (uint32)dmux.canvas_width
    case constants.WEBP_FF_CANVAS_HEIGHT:
      return (uint32)dmux.canvas_height
    case constants.WEBP_FF_LOOP_COUNT:
      return (uint32)dmux.loop_count
    case constants.WEBP_FF_BACKGROUND_COLOR:
      return dmux.bgcolor
    case constants.WEBP_FF_FRAME_COUNT:
      return (uint32)dmux.num_frames
  }
  return 0
}

// -----------------------------------------------------------------------------
// Frame iteration

func GetFrame(/* const */ dmux *WebPDemuxer, frame_num int) *Frame {
  const f *Frame
  for f = dmux.frames f != nil f = f.next {
    if (frame_num == f.frame_num) break
  }
  return f
}

func GetFramePayload(/* const */ mem_buf *uint8, /*const*/ frame *Frame, /*const*/ data_size *uint64) *uint8 {
  *data_size = 0
  if (frame != nil) {
    var image *ChunkData = frame.img_components
    var alpha *ChunkData = frame.img_components + 1
    start_offset := image.offset
    *data_size = image.size

    // if alpha exists it precedes image, update the size allowing for
    // intervening chunks.
    if (alpha.size > 0) {
      inter_size := tenary.If((image.offset > 0), image.offset - (alpha.offset + alpha.size), 0)
      start_offset = alpha.offset
      *data_size += alpha.size + inter_size
    }
    return mem_buf + start_offset
  }
  return nil
}

// Create a whole 'frame' from VP8 (+ alpha) or lossless.
func SynthesizeFrame(/* const */ dmux *WebPDemuxer, /*const*/ frame *Frame, /*const*/ iter *WebPIterator) int {
  var mem_buf *uint8 = dmux.mem.buf
  payload_size := 0
  var payload *uint8 = GetFramePayload(mem_buf, frame, &payload_size)
  if (payload == nil) { return 0 }
  assert.Assert(frame != nil)

  iter.frame_num = frame.frame_num
  iter.num_frames = dmux.num_frames
  iter.x_offset = frame.x_offset
  iter.y_offset = frame.y_offset
  iter.width = frame.width
  iter.height = frame.height
  iter.has_alpha = frame.has_alpha
  iter.duration = frame.duration
  iter.dispose_method = frame.dispose_method
  iter.blend_method = frame.blend_method
  iter.complete = frame.complete
  iter.fragment.bytes = payload
  iter.fragment.size = payload_size
  return 1
}

func SetFrame(frame_num int, /*const*/ iter *WebPIterator) int {
  const frame *Frame
  var dmux *WebPDemuxer = (*WebPDemuxer)iter.private_
  if (dmux == nil || frame_num < 0) { return 0 }
  if (frame_num > dmux.num_frames) { return 0 }
  if (frame_num == 0) frame_num = dmux.num_frames

  frame = GetFrame(dmux, frame_num)
  if (frame == nil) { return 0 }

  return SynthesizeFrame(dmux, frame, iter)
}

// Retrieves frame 'frame_number' from 'dmux'.
// 'iter.fragment' points to the frame on return from this function.
// Setting 'frame_number' equal to 0 will return the last frame of the image.
// Returns false if 'dmux' is nil or frame 'frame_number' is not present.
// Call WebPDemuxReleaseIterator() when use of the iterator is complete.
// NOTE: 'dmux' must persist for the lifetime of 'iter'.
func WebPDemuxGetFrame(/* const */ dmux *WebPDemuxer, frame int , iter *WebPIterator) int {
  if (iter == nil) { return 0 }

  stdlib.Memset(iter, 0, sizeof(*iter))
  iter.private_ = (*void)dmux
  return SetFrame(frame, iter)
}

// Sets 'iter.fragment' to point to the next ('iter.frame_num' + 1) or
// previous ('iter.frame_num' - 1) frame. These functions do not loop.
// Returns true on success, false otherwise.
func WebPDemuxNextFrame(iter *WebPIterator) int {
  if (iter == nil) { return 0 }
  return SetFrame(iter.frame_num + 1, iter)
}

func WebPDemuxPrevFrame(iter *WebPIterator) int {
  if (iter == nil) { return 0 }
  if (iter.frame_num <= 1) { return 0 }
  return SetFrame(iter.frame_num - 1, iter)
}

// Releases any memory associated with 'iter'.
// Must be called before any subsequent calls to WebPDemuxGetChunk() on the same
// iter. Also, must be called before destroying the associated WebPDemuxer with
// WebPDemuxDelete().
func WebPDemuxReleaseIterator(iter *WebPIterator) { (void)iter }

// -----------------------------------------------------------------------------
// Chunk iteration

func ChunkCount(/* const */ dmux *WebPDemuxer, /*const*/ fourcc [4]byte) int {
  var mem_buf *uint8 = dmux.mem.buf
  const c *Chunk
  count := 0
  for c = dmux.chunks c != nil c = c.next {
    var header *uint8 = mem_buf + c.data.offset
    if (!memcmp(header, fourcc, TAG_SIZE)) ++count
  }
  return count
}

static GetChunk(/* const */ dmux *WebPDemuxer, /*const*/ fourcc [4]byte, chunk_num int) *Chunk {
  var mem_buf *uint8 = dmux.mem.buf
  const c *Chunk
  count := 0
  for c = dmux.chunks c != nil c = c.next {
    var header *uint8 = mem_buf + c.data.offset
    if (!memcmp(header, fourcc, TAG_SIZE)) ++count
    if (count == chunk_num) break
  }
  return c
}

func SetChunk(/* const */ fourcc [4]byte, chunk_num int, /*const*/ iter *WebPChunkIterator) int {
  var dmux *WebPDemuxer = (*WebPDemuxer)iter.private_
  int count

  if (dmux == nil || fourcc == nil || chunk_num < 0) { return 0 }
  count = ChunkCount(dmux, fourcc)
  if (count == 0) { return 0 }
  if (chunk_num == 0) chunk_num = count

  if (chunk_num <= count) {
    var mem_buf *uint8 = dmux.mem.buf
    var chunk *Chunk = GetChunk(dmux, fourcc, chunk_num)
    iter.chunk.bytes = mem_buf + chunk.data.offset + CHUNK_HEADER_SIZE
    iter.chunk.size = chunk.data.size - CHUNK_HEADER_SIZE
    iter.num_chunks = count
    iter.chunk_num = chunk_num
    return 1
  }
  return 0
}

// Retrieves the 'chunk_number' instance of the chunk with id 'fourcc' from
// 'dmux'.
// 'fourcc' is a character array containing the fourcc of the chunk to return,
// e.g., "ICCP", "XMP ", "EXIF", etc.
// Setting 'chunk_number' equal to 0 will return the last chunk in a set.
// Returns true if the chunk is found, false otherwise. Image related chunk
// payloads are accessed through WebPDemuxGetFrame() and related functions.
// Call WebPDemuxReleaseChunkIterator() when use of the iterator is complete.
// NOTE: 'dmux' must persist for the lifetime of the iterator.
func WebPDemuxGetChunk(/* const */ dmux *WebPDemuxer, /*const*/ fourcc [4]byte, chunk_num int, iter *WebPChunkIterator) int {
  if (iter == nil) { return 0 }

  stdlib.Memset(iter, 0, sizeof(*iter))
  iter.private_ = (*void)dmux
  return SetChunk(fourcc, chunk_num, iter)
}

// Sets 'iter.chunk' to point to the next ('iter.chunk_num' + 1) or previous
// ('iter.chunk_num' - 1) chunk. These functions do not loop.
// Returns true on success, false otherwise.
func WebPDemuxNextChunk( iter *WebPChunkIterator) int {
  if (iter != nil) {
    const fourcc *byte =
        (/* const */ *byte)iter.chunk.bytes - CHUNK_HEADER_SIZE
    return SetChunk(fourcc, iter.chunk_num + 1, iter)
  }
  return 0
}

func WebPDemuxPrevChunk( iter *WebPChunkIterator) int {
  if (iter != nil && iter.chunk_num > 1) {
    const fourcc *byte =
        (/* const */ *byte)iter.chunk.bytes - CHUNK_HEADER_SIZE
    return SetChunk(fourcc, iter.chunk_num - 1, iter)
  }
  return 0
}

// Releases any memory associated with 'iter'.
// Must be called before destroying the associated WebPDemuxer with
// WebPDemuxDelete().
func WebPDemuxReleaseChunkIterator(iter *WebPChunkIterator) { (void)iter }
