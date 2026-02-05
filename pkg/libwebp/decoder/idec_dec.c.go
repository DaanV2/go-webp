package decoder

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Incremental decoding
//
// Author: somnath@google.com (Somnath Banerjee)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// In append mode, buffer allocations increase as multiples of this value.
// Needs to be a power of 2.
const CHUNK_SIZE =4096
const MAX_MB_SIZE =4096

//------------------------------------------------------------------------------
// Data structures for memory and states

// Decoding states. State normally flows as:
// WEBP_HEADER.VP8_HEADER.VP8_PARTS0.VP8_DATA.DONE for a lossy image, and
// WEBP_HEADER.VP8L_HEADER.VP8L_DATA.DONE for a lossless image.
// If there is any error the decoder goes into state ERROR.
type DecState int

const (
  STATE_WEBP_HEADER DecState = iota  // All the data before that of the VP8/VP8L chunk.
  STATE_VP8_HEADER   // The VP8 Frame header (within the VP8 chunk).
  STATE_VP8_PARTS0
  STATE_VP8_DATA
  STATE_VP8L_HEADER
  STATE_VP8L_DATA
  STATE_DONE
  STATE_ERROR
)

// Operating state for the MemBuffer
type MemBufferMode int

const ( 
	MEM_MODE_NONE MemBufferMode = iota
	MEM_MODE_APPEND
	MEM_MODE_MAP 
)

// storage for partition #0 and partial data (in a rolling fashion)
type MemBuffer struct {
   mode MemBufferMode;  // Operation mode
   start uint64;        // start location of the data to be decoded
   end uint64;          // end location
   buf_size uint64;     // size of the allocated buffer
   buf *uint8;        // We don't own this buffer in case WebPIUpdate()

   part0_size uint64;         // size of partition #0
  part0_buf *uint8;  // buffer to store partition #0
}

type WebPIDecoder struct {
  state DecState;        // current decoding state
  params WebPDecParams  // Params to store output info
   is_lossless int       // for down-casting 'dec'.
  dec *void;             // either a VP8Decoder or a VP8LDecoder instance
  io VP8Io

  MemBuffer mem;         // input memory buffer.
  output WebPDecBuffer ;  // output buffer (when no external one is supplied, // or if the external one has slow-memory)
  final_output *WebPDecBuffer;  // Slow-memory output to copy to eventually.
   chunk_size uint64  // Compressed VP8/VP8L size extracted from Header.

  last_mb_y int ;  // last row reached for intra-mode decoding
}

// MB context to restore in case VP8DecodeMB() fails
type MBContext struct {
   left VP8MB;
   info VP8MB;
   token_br VP8BitReader;
}

//------------------------------------------------------------------------------
// MemBuffer: incoming data handling

func MemDataSize(/* const */ mem *MemBuffer) uint64 {
  return (mem.end - mem.start);
}

// Check if we need to preserve the compressed alpha data, as it may not have
// been decoded yet.
func NeedCompressedAlpha(/* const */ idec *WebPIDecoder) int {
  if (idec.state == STATE_WEBP_HEADER) {
    // We haven't parsed the headers yet, so we don't know whether the image is
    // lossy or lossless. This also means that we haven't parsed the ALPH chunk.
    return 0;
  }
  if (idec.is_lossless) {
    return 0;  // ALPH chunk is not present for lossless images.
  } else {
    var dec *vp8.VP8Decoder = (*vp8.VP8Decoder)idec.dec;
    assert.Assert(dec != nil);  // Must be true as idec.state != STATE_WEBP_HEADER.
    return (dec.alpha_data != nil) && !dec.is_alpha_decoded;
  }
}

func DoRemap(/* const */ idec *WebPIDecoder, ptrdiff_t offset) {
  /* const */ mem *MemBuffer = &idec.mem;
  var new_base *uint8 = mem.buf + mem.start;
  // note: for VP8, setting up idec.io is only really needed at the beginning
  // of the decoding, till partition #0 is complete.
  idec.io.data = new_base;
  idec.io.data_size = MemDataSize(mem);

  if (idec.dec != nil) {
    if (!idec.is_lossless) {
      var dec *VP8Decoder = (*VP8Decoder)idec.dec;
      last_part := dec.num_parts_minus_one;
      if (offset != 0) {
        var p uint32
        for p = 0; p <= last_part; p++ {
          VP8RemapBitReader(dec.parts + p, offset);
        }
        // Remap partition #0 data pointer to new offset, but only in MAP
        // mode (in APPEND mode, partition #0 is copied into a fixed memory).
        if (mem.mode == MEM_MODE_MAP) {
          VP8RemapBitReader(&dec.br, offset);
        }
      }
      {
        var last_start *uint8 = dec.parts[last_part].buf;
        // 'last_start' will be nil when 'idec.state' is < STATE_VP8_PARTS0
        // and through a portion of that state (when there isn't enough data to
        // parse the partitions). The bitreader is only used meaningfully when
        // there is enough data to begin parsing partition 0.
        if (last_start != nil) {
          part_size := mem.buf + mem.end - last_start;
          var bounded_last_start *uint8 = last_start // bidi index -> part_size
          VP8BitReaderSetBuffer(&dec.parts[last_part], bounded_last_start, part_size);
        }
      }
      if (NeedCompressedAlpha(idec)) {
        var alph_dec *ALPHDecoder = dec.alph_dec;
        dec.alpha_data += offset;
        WEBP_SELF_ASSIGN(dec.alpha_data_size);
        if (alph_dec != nil && alph_dec.vp8l_dec != nil) {
          if (alph_dec.method == ALPHA_LOSSLESS_COMPRESSION) {
            var alph_vp *VP8LDecoder8l_dec = alph_dec.vp8l_dec;
            data_size uint64;
            const bounded_alpha_data *uint8;

            assert.Assert(dec.alpha_data_size >= ALPHA_HEADER_LEN);
            data_size = dec.alpha_data_size - ALPHA_HEADER_LEN;
            bounded_alpha_data = dec.alpha_data + ALPHA_HEADER_LEN // bidi index -> data_size
            VP8LBitReaderSetBuffer(&alph_vp8l_dec.br, bounded_alpha_data, data_size);
          } else {  // alph_dec.method == ALPHA_NO_COMPRESSION
            // Nothing special to do in this case.
          }
        }
      }
    } else {  // Resize lossless bitreader
      var dec *VP8LDecoder = (*VP8LDecoder)idec.dec;
      data_size := MemDataSize(mem);
      var bounded_new_base *uint8 = new_base // bidi index -> data_size
      VP8LBitReaderSetBuffer(&dec.br, bounded_new_base, data_size);
    }
  }
}

// Appends data to the end of MemBuffer.buf. It expands the allocated memory
// size if required and also updates VP8BitReader's if new memory is allocated.
func AppendToMemBuffer(/* const */ idec *WebPIDecoder, /*const*/ data *uint8, data_size uint64) int {
  var dec *VP8Decoder = (*VP8Decoder)idec.dec;
  /* const */ mem *MemBuffer = &idec.mem;
  need_compressed_alpha := NeedCompressedAlpha(idec);
  const old_start *uint8 =
      (mem.buf == nil) ? nil : mem.buf + mem.start;
  const old_base *uint8 =
      tenary.If(need_compressed_alpha, dec.alpha_data, old_start);
  assert.Assert(mem.buf != nil || mem.start == 0);
  assert.Assert(mem.mode == MEM_MODE_APPEND);
  if (data_size > MAX_CHUNK_PAYLOAD) {
    // security safeguard: trying to allocate more than what the format
    // allows for a chunk should be considered a smoke smell.
    return 0;
  }

  if (mem.end + data_size > mem.buf_size) {  // Need some free memory
    new_mem_start := old_start - old_base;
    current_size := MemDataSize(mem) + new_mem_start;
    new_size := (uint64)current_size + data_size;
    extra_size := (new_size + CHUNK_SIZE - 1) & ~(CHUNK_SIZE - 1);

	// var new_buf *uint8 = (*uint8)WebPSafeMalloc(extra_size, sizeof(*new_buf));
    // if new_buf == nil { { return 0 } }
	new_buf := make([]uint8, extra_size)

	if old_base != nil { {stdlib.MemCpy(new_buf, old_base, current_size) }}

    mem.buf = new_buf;
    mem.buf_size = (uint64)extra_size;
    mem.start = new_mem_start;
    mem.end = current_size;
  }

  assert.Assert(mem.buf != nil);
  stdlib.MemCpy(mem.buf + mem.end, data, data_size);
  mem.end += data_size;
  assert.Assert(mem.end <= mem.buf_size);

  DoRemap(idec, mem.buf + mem.start - old_start);
  return 1;
}

func RemapMemBuffer(/* const */ idec *WebPIDecoder, /*const*/ data *uint8, data_size uint64) int {
  /* const */ mem *MemBuffer = &idec.mem;
  var old_buf *uint8 = mem.buf;
  const old_start *uint8 =
      (old_buf == nil) ? nil : old_buf + mem.start;
  assert.Assert(old_buf != nil || mem.start == 0);
  assert.Assert(mem.mode == MEM_MODE_MAP);

  if data_size < mem.buf_size {
    return 0  // can't remap to a shorter buffer!
}

  mem.buf = (*uint8)data;
  mem.end = mem.buf_size = data_size;

  DoRemap(idec, mem.buf + mem.start - old_start);
  return 1;
}

func InitMemBuffer(/* const */ mem *MemBuffer) {
  mem.mode = MEM_MODE_NONE;
  mem.buf = nil;
  mem.buf_size = 0;
  mem.part0_buf = nil;
  mem.part0_size = 0;
}

func CheckMemBufferMode(/* const */ mem *MemBuffer, MemBufferMode expected) int {
  if (mem.mode == MEM_MODE_NONE) {
    mem.mode = expected;  // switch to the expected mode
  } else if (mem.mode != expected) {
    return 0;  // we mixed the modes => error
  }
  assert.Assert(mem.mode == expected);  // mode is ok
  return 1;
}

// To be called last.
 static VP8StatusCode FinishDecoding(/* const */ idec *WebPIDecoder) {
  var options *WebPDecoderOptions = idec.params.options;
  var output *WebPDecBuffer = idec.params.output;

  idec.state = STATE_DONE;
  if (options != nil && options.flip) {
    const VP8StatusCode status = WebPFlipBuffer(output);
    if status != VP8_STATUS_OK { { return status } }
  }
  if (idec.final_output != nil) {
    const VP8StatusCode status = WebPCopyDecBufferPixels(
        output, idec.final_output);  // do the slow-copy
    WebPFreeDecBuffer(&idec.output);
    if status != VP8_STATUS_OK { { return status } }
    *output = *idec.final_output;
    idec.final_output = nil;
  }
  return VP8_STATUS_OK;
}

//------------------------------------------------------------------------------
// Macroblock-decoding contexts

func SaveContext(/* const */ dec *VP8Decoder, /*const*/ token_br *VP8BitReader, /*const*/ context *MBContext) {
  context.left = dec.mb_info[-1];
  context.info = dec.mb_info[dec.mb_x];
  context.token_br = *token_br;
}

func RestoreContext(/* const */ context *MBContext, /*const*/ dec *VP8Decoder, /*const*/ token_br *VP8BitReader) {
  dec.mb_info[-1] = context.left;
  dec.mb_info[dec.mb_x] = context.info;
  *token_br = context.token_br;
}

//------------------------------------------------------------------------------

static VP8StatusCode IDecError(/* const */ idec *WebPIDecoder, VP8StatusCode error) {
  if (idec.state == STATE_VP8_DATA) {
    // Synchronize the thread, clean-up and check for errors.
    (void)VP8ExitCritical((*VP8Decoder)idec.dec, &idec.io);
  }
  idec.state = STATE_ERROR;
  return error;
}

func ChangeState(/* const */ idec *WebPIDecoder, DecState new_state, uint64 consumed_bytes) {
  /* const */ mem *MemBuffer = &idec.mem;
  idec.state = new_state;
  mem.start += consumed_bytes;
  assert.Assert(mem.start <= mem.end);
  idec.io.data = mem.buf + mem.start;
  idec.io.data_size = MemDataSize(mem);
}

// Headers
func DecodeWebPHeaders(/* const */ idec *WebPIDecoder) VP8StatusCode {
  /* const */ mem *MemBuffer = &idec.mem;
  var data *uint8 = mem.buf + mem.start;
  curr_size := MemDataSize(mem);
  var status VP8StatusCode
  var headers WebPHeaderStructure

  headers.data = data // bidi index -> curr_size
  headers.data_size = curr_size;
  headers.have_all_data = 0;
  status = WebPParseHeaders(&headers);
  if (status == VP8_STATUS_NOT_ENOUGH_DATA) {
    return VP8_STATUS_SUSPENDED;  // We haven't found a VP8 chunk yet.
  } else if (status != VP8_STATUS_OK) {
    return IDecError(idec, status);
  }

  idec.chunk_size = headers.compressed_size;
  idec.is_lossless = headers.is_lossless;
  if (!idec.is_lossless) {
    var dec *VP8Decoder = VP8New();
    if (dec == nil) {
      return VP8_STATUS_OUT_OF_MEMORY;
    }
    dec.incremental = 1;
    idec.dec = dec;
    dec.alpha_data = headers.alpha_data;
    dec.alpha_data_size = headers.alpha_data_size;
    ChangeState(idec, STATE_VP8_HEADER, headers.offset);
  } else {
    var dec *VP8LDecoder = VP8LNew();
    if (dec == nil) {
      return VP8_STATUS_OUT_OF_MEMORY;
    }
    idec.dec = dec;
    ChangeState(idec, STATE_VP8L_HEADER, headers.offset);
  }
  return VP8_STATUS_OK;
}

static VP8StatusCode DecodeVP8FrameHeader(/* const */ idec *WebPIDecoder) {
  var data *uint8 = idec.mem.buf + idec.mem.start;
  curr_size := MemDataSize(&idec.mem);
  var width, height int
  var bits uint32;

  if (curr_size < VP8_FRAME_HEADER_SIZE) {
    // Not enough data bytes to extract VP8 Frame Header.
    return VP8_STATUS_SUSPENDED;
  }
  {
    var bounded_data *uint8 =
        WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(/* const */ *uint8, , curr_size);
    if (!VP8GetInfo(bounded_data, curr_size, idec.chunk_size, &width, &height)) {
      return IDecError(idec, VP8_STATUS_BITSTREAM_ERROR);
    }
  }

  bits = data[0] | (data[1] << 8) | (data[2] << 16);
  idec.mem.part0_size = (bits >> 5) + VP8_FRAME_HEADER_SIZE;

  idec.io.data = data;
  idec.io.data_size = curr_size;
  idec.state = STATE_VP8_PARTS0;
  return VP8_STATUS_OK;
}

// Partition #0
static VP8StatusCode CopyParts0Data(/* const */ idec *WebPIDecoder) {
  var dec *VP8Decoder = (*VP8Decoder)idec.dec;
  var br *VP8BitReader = &dec.br;
  part_size := br.buf_end - br.buf;
  /* const */ mem *MemBuffer = &idec.mem;
  assert.Assert(!idec.is_lossless);
  assert.Assert(mem.part0_buf == nil);
  // the following is a format limitation, no need for runtime check:
  assert.Assert(part_size <= mem.part0_size);
  if (part_size == 0) {  // can't have zero-size partition #0
    return VP8_STATUS_BITSTREAM_ERROR;
  }
  if (mem.mode == MEM_MODE_APPEND) {
    // We copy and grab ownership of the partition #0 data.
    // var part0_buf *uint8 = (*uint8)WebPSafeMalloc(uint64(1), part_size);
	// if (part0_buf == nil) {
    //   return VP8_STATUS_OUT_OF_MEMORY;
    // }
	part0_buf := make([]uint8, part_size)

    stdlib.MemCpy(part0_buf, br.buf, part_size);
    mem.part0_buf = part0_buf;
    VP8BitReaderSetBuffer(br, part0_buf, part_size);
  } else {
    // Else: just keep pointers to the partition #0's data in dec.br.
  }
  mem.start += part_size;
  return VP8_STATUS_OK;
}

static VP8StatusCode DecodePartition0(/* const */ idec *WebPIDecoder) {
  var dec *VP8Decoder = (*VP8Decoder)idec.dec;
  var io *VP8Io = &idec.io;
  var params *WebPDecParams = &idec.params;
  var output *WebPDecBuffer = params.output;

  // Wait till we have enough data for the whole partition #0
  if (MemDataSize(&idec.mem) < idec.mem.part0_size) {
    return VP8_STATUS_SUSPENDED;
  }

  if (!VP8GetHeaders(dec, io)) {
    const VP8StatusCode status = dec.status;
    if (status == VP8_STATUS_SUSPENDED ||
        status == VP8_STATUS_NOT_ENOUGH_DATA) {
      // treating NOT_ENOUGH_DATA as SUSPENDED state
      return VP8_STATUS_SUSPENDED;
    }
    return IDecError(idec, status);
  }

  // Allocate/Verify output buffer now
  dec.status =
      WebPAllocateDecBuffer(io.width, io.height, params.options, output);
  if (dec.status != VP8_STATUS_OK) {
    return IDecError(idec, dec.status);
  }
  // This change must be done before calling VP8InitFrame()
  dec.mt_method =
      VP8GetThreadMethod(params.options, nil, io.width, io.height);
  VP8InitDithering(params.options, dec);

  dec.status = CopyParts0Data(idec);
  if (dec.status != VP8_STATUS_OK) {
    return IDecError(idec, dec.status);
  }

  // Finish setting up the decoding parameters. Will call io.setup().
  if (VP8EnterCritical(dec, io) != VP8_STATUS_OK) {
    return IDecError(idec, dec.status);
  }

  // Note: past this point, teardown() must always be called
  // in case of error.
  idec.state = STATE_VP8_DATA;
  // Allocate memory and prepare everything.
  if (!VP8InitFrame(dec, io)) {
    return IDecError(idec, dec.status);
  }
  return VP8_STATUS_OK;
}

// Remaining partitions
static VP8StatusCode DecodeRemaining(/* const */ idec *WebPIDecoder) {
  var dec *VP8Decoder = (*VP8Decoder)idec.dec;
  var io *VP8Io = &idec.io;

  // Make sure partition #0 has been read before, to set dec to ready.
  if (!dec.ready) {
    return IDecError(idec, VP8_STATUS_BITSTREAM_ERROR);
  }
  for ; dec.mb_y < dec.mb_h; ++dec.mb_y {
    if (idec.last_mb_y != dec.mb_y) {
      if (!VP8ParseIntraModeRow(&dec.br, dec)) {
        // note: normally, error shouldn't occur since we already have the whole
        // partition0 available here in DecodeRemaining(). Reaching EOF while
        // reading intra modes really means a BITSTREAM_ERROR.
        return IDecError(idec, VP8_STATUS_BITSTREAM_ERROR);
      }
      idec.last_mb_y = dec.mb_y;
    }
    for ; dec.mb_x < dec.mb_w; ++dec.mb_x {
      const token_br *VP8BitReader =
          &dec.parts[dec.mb_y & dec.num_parts_minus_one];
      MBContext context;
      SaveContext(dec, token_br, &context);
      if (!VP8DecodeMB(dec, token_br)) {
        // We shouldn't fail when MAX_MB data was available
        if (dec.num_parts_minus_one == 0 &&
            MemDataSize(&idec.mem) > MAX_MB_SIZE) {
          return IDecError(idec, VP8_STATUS_BITSTREAM_ERROR);
        }
        // Synchronize the threads.
        if (dec.mt_method > 0) {
          if (!WebPGetWorkerInterface().Sync(&dec.worker)) {
            return IDecError(idec, VP8_STATUS_BITSTREAM_ERROR);
          }
        }
        RestoreContext(&context, dec, token_br);
        return VP8_STATUS_SUSPENDED;
      }
      // Release buffer only if there is only one partition
      if (dec.num_parts_minus_one == 0) {
        idec.mem.start = token_br.buf - idec.mem.buf;
        assert.Assert(idec.mem.start <= idec.mem.end);
      }
    }
    VP8InitScanline(dec);  // Prepare for next scanline

    // Reconstruct, filter and emit the row.
    if (!VP8ProcessRow(dec, io)) {
      return IDecError(idec, VP8_STATUS_USER_ABORT);
    }
  }
  // Synchronize the thread and check for errors.
  if (!VP8ExitCritical(dec, io)) {
    idec.state = STATE_ERROR;  // prevent re-entry in IDecError
    return IDecError(idec, VP8_STATUS_USER_ABORT);
  }
  dec.ready = 0;
  return FinishDecoding(idec);
}

static VP8StatusCode ErrorStatusLossless(/* const */ idec *WebPIDecoder, VP8StatusCode status) {
  if (status == VP8_STATUS_SUSPENDED || status == VP8_STATUS_NOT_ENOUGH_DATA) {
    return VP8_STATUS_SUSPENDED;
  }
  return IDecError(idec, status);
}

static VP8StatusCode DecodeVP8LHeader(/* const */ idec *WebPIDecoder) {
  var io *VP8Io = &idec.io;
  var dec *VP8LDecoder = (*VP8LDecoder)idec.dec;
  var params *WebPDecParams = &idec.params;
  var output *WebPDecBuffer = params.output;
  curr_size := MemDataSize(&idec.mem);
  assert.Assert(idec.is_lossless);

  // Wait until there's enough data for decoding header.
  if (curr_size < (idec.chunk_size >> 3)) {
    dec.status = VP8_STATUS_SUSPENDED;
    return ErrorStatusLossless(idec, dec.status);
  }

  if (!VP8LDecodeHeader(dec, io)) {
    if (dec.status == VP8_STATUS_BITSTREAM_ERROR &&
        curr_size < idec.chunk_size) {
      dec.status = VP8_STATUS_SUSPENDED;
    }
    return ErrorStatusLossless(idec, dec.status);
  }
  // Allocate/verify output buffer now.
  dec.status =
      WebPAllocateDecBuffer(io.width, io.height, params.options, output);
  if (dec.status != VP8_STATUS_OK) {
    return IDecError(idec, dec.status);
  }

  idec.state = STATE_VP8L_DATA;
  return VP8_STATUS_OK;
}

static VP8StatusCode DecodeVP8LData(/* const */ idec *WebPIDecoder) {
  var dec *VP8LDecoder = (*VP8LDecoder)idec.dec;
  curr_size := MemDataSize(&idec.mem);
  assert.Assert(idec.is_lossless);

  // Switch to incremental decoding if we don't have all the bytes available.
  dec.incremental = (curr_size < idec.chunk_size);

  if (!VP8LDecodeImage(dec)) {
    return ErrorStatusLossless(idec, dec.status);
  }
  assert.Assert(dec.status == VP8_STATUS_OK || dec.status == VP8_STATUS_SUSPENDED);
  return (dec.status == VP8_STATUS_SUSPENDED) ? dec.status
                                               : FinishDecoding(idec);
}

// Main decoding loop
static VP8StatusCode IDecode(idec *WebPIDecoder) {
  VP8StatusCode status = VP8_STATUS_SUSPENDED;

  if (idec.state == STATE_WEBP_HEADER) {
    status = DecodeWebPHeaders(idec);
  } else {
    if (idec.dec == nil) {
      return VP8_STATUS_SUSPENDED;  // can't continue if we have no decoder.
    }
  }
  if (idec.state == STATE_VP8_HEADER) {
    status = DecodeVP8FrameHeader(idec);
  }
  if (idec.state == STATE_VP8_PARTS0) {
    status = DecodePartition0(idec);
  }
  if (idec.state == STATE_VP8_DATA) {
    var dec *VP8Decoder = (*VP8Decoder)idec.dec;
    if (dec == nil) {
      return VP8_STATUS_SUSPENDED;  // can't continue if we have no decoder.
    }
    status = DecodeRemaining(idec);
  }
  if (idec.state == STATE_VP8L_HEADER) {
    status = DecodeVP8LHeader(idec);
  }
  if (idec.state == STATE_VP8L_DATA) {
    status = DecodeVP8LData(idec);
  }
  return status;
}

//------------------------------------------------------------------------------
// Internal constructor

func NewDecoder (/* const */ output_buffer *WebPDecBuffer, /*const*/ features *WebPBitstreamFeatures) *WebPIDecoder {
	//   idec *WebPIDecoder = (*WebPIDecoder)WebPSafeCalloc(uint64(1), sizeof(*idec));
	//   if (idec == nil) {
	//     return nil;
	//   }
  idec := &WebPIDecoder{}

  idec.state = STATE_WEBP_HEADER;
  idec.chunk_size = 0;

  idec.last_mb_y = -1;

  InitMemBuffer(&idec.mem);
  if (!WebPInitDecBuffer(&idec.output) || !VP8InitIo(&idec.io)) {
    return nil;
  }

  WebPResetDecParams(&idec.params);
  if (output_buffer == nil || WebPAvoidSlowMemory(output_buffer, features)) {
    idec.params.output = &idec.output;
    idec.final_output = output_buffer;
    if (output_buffer != nil) {
      idec.params.output.colorspace = output_buffer.colorspace;
    }
  } else {
    idec.params.output = output_buffer;
    idec.final_output = nil;
  }
  WebPInitCustomIo(&idec.params, &idec.io);  // Plug the I/O functions.

  return idec;
}

// Creates a new incremental decoder with the supplied buffer parameter.
// This output_buffer can be passed nil, in which case a default output buffer
// is used (with MODE_RGB). Otherwise, an internal reference to 'output_buffer'
// is kept, which means that the lifespan of 'output_buffer' must be larger than
// that of the returned WebPIDecoder object.
// The supplied 'output_buffer' content MUST NOT be changed between calls to
// WebPIAppend() or WebPIUpdate() unless 'output_buffer.is_external_memory' is
// not set to 0. In such a case, it is allowed to modify the pointers, size and
// stride of output_buffer.u.RGBA or output_buffer.u.YUVA, provided they remain
// within valid bounds.
// All other fields of WebPDecBuffer MUST remain constant between calls.
// Returns nil if the allocation failed.
func WebPINewDecoder (output_buffer *WebPDecBuffer) *WebPIDecoder {
  return NewDecoder(output_buffer, nil);
}

WebPIDecode *WebPIDecoder(/* const */ data *uint8, data_size uint64, config *WebPDecoderConfig) {
  idec *WebPIDecoder;
  WebPBitstreamFeatures tmp_features;
  const features *WebPBitstreamFeatures =
      (config == nil) ? &tmp_features : &config.input;
  stdlib.Memset(&tmp_features, 0, sizeof(tmp_features));

  // Parse the bitstream's features, if requested:
  if (data != nil && data_size > 0) {
    if (WebPGetFeatures(data, data_size, features) != VP8_STATUS_OK) {
      return nil;
    }
  }

  // Create an instance of the incremental decoder
  idec = (config != nil) ? NewDecoder(&config.output, features)
                          : NewDecoder(nil, features);
  if (idec == nil) {
    return nil;
  }
  // Finish initialization
  if (config != nil) {
    idec.params.options = &config.options;
  }
  return idec;
}

// Deletes the WebPIDecoder object and associated memory. Must always be called
// if WebPINewDecoder, WebPINewRGB or WebPINewYUV succeeded.
func WebPIDelete(idec *WebPIDecoder) {
  if idec == nil { return }
  if (idec.dec != nil) {
    if (!idec.is_lossless) {
      if (idec.state == STATE_VP8_DATA) {
        // Synchronize the thread, clean-up and check for errors.
        // TODO(vrabaud) do we care about the return result?
        (void)VP8ExitCritical((*VP8Decoder)idec.dec, &idec.io);
      }
      VP8Delete((*VP8Decoder)idec.dec);
    }
  }
  idec.mem = nil
  WebPFreeDecBuffer(&idec.output);
}

//------------------------------------------------------------------------------
// Wrapper toward WebPINewDecoder

// This function allocates and initializes an incremental-decoder object, which
// will output the RGB/A samples specified by 'csp' into a preallocated
// buffer 'output_buffer'. The size of this buffer is at least
// 'output_buffer_size' and the stride (distance in bytes between two scanlines)
// is specified by 'output_stride'.
// Additionally, output_buffer can be passed nil in which case the output
// buffer will be allocated automatically when the decoding starts. The
// colorspace 'csp' is taken into account for allocating this buffer. All other
// parameters are ignored.
// Returns nil if the allocation failed, or if some parameters are invalid.
func WebPINewRGB(csp WEBP_CSP_MODE , output_buffer *uint8, output_buffer_size uint64 , output_stride int ) *WebPIDecoder {
  is_external_memory := (output_buffer != nil) ? 1 : 0;
  idec *WebPIDecoder;

  if csp >= MODE_YUV { return nil }
  if (is_external_memory == 0) {  // Overwrite parameters to sane values.
    output_buffer = nil;
    output_buffer_size = 0;
    output_stride = 0;
  } else {  // A buffer was passed. Validate the other params.
    if (output_stride == 0 || output_buffer_size == 0) {
      return nil;  // invalid parameter.
    }
  }
  idec = WebPINewDecoder(nil);
  if idec == nil { { return nil } }
  idec.output.colorspace = csp;
  idec.output.is_external_memory = is_external_memory;
  idec.output.u.RGBA.rgba = output_buffer;
  idec.output.u.RGBA.stride = output_stride;
  idec.output.u.RGBA.size = output_buffer_size;
  return idec;
}


// This function allocates and initializes an incremental-decoder object, which
// will output the raw luma/chroma samples into a preallocated planes if
// supplied. The luma plane is specified by its pointer 'luma', its size
// 'luma_size' and its stride 'luma_stride'. Similarly, the chroma-u plane
// is specified by the 'u', 'u_size' and 'u_stride' parameters, and the chroma-v
// plane by 'v' and 'v_size'. And same for the alpha-plane. The 'a' pointer
// can be pass nil in case one is not interested in the transparency plane.
// Conversely, 'luma' can be passed nil if no preallocated planes are supplied.
// In this case, the output buffer will be automatically allocated (using
// MODE_YUVA) when decoding starts. All parameters are then ignored.
// Returns nil if the allocation failed or if a parameter is invalid.
func WebPINewYUVA(luma *uint8, luma_size uint64, luma_stride int, u *uint8, u_size uint64, u_stride int, v *uint8, v_size uint64, v_stride int, a *uint8, a_size uint64, a_stride int) *WebPIDecoder {
  is_external_memory := (luma != nil) ? 1 : 0;
  idec *WebPIDecoder;
  WEBP_CSP_MODE colorspace;

  if (is_external_memory == 0) {  // Overwrite parameters to sane values.
    luma = nil;
    luma_size = 0;
    u = nil;
    u_size = 0;
    v = nil;
    v_size = 0;
    a = nil;
    a_size = 0;
    luma_stride = u_stride = v_stride = a_stride = 0;
    colorspace = MODE_YUVA;
  } else {  // A luma buffer was passed. Validate the other parameters.
    if u == nil || v == nil { { return nil } }
    if luma_size == 0 || u_size == 0 || v_size == 0 { { return nil } }
    if luma_stride == 0 || u_stride == 0 || v_stride == 0 { { return nil } }
    if (a != nil) {
      if a_size == 0 || a_stride == 0 { { return nil } }
    }
    colorspace = (a == nil) ? MODE_YUV : MODE_YUVA;
  }

  idec = WebPINewDecoder(nil);
  if idec == nil { { return nil } }

  idec.output.colorspace = colorspace;
  idec.output.is_external_memory = is_external_memory;
  idec.output.u.YUVA.y = luma;
  idec.output.u.YUVA.y_stride = luma_stride;
  idec.output.u.YUVA.y_size = luma_size;
  idec.output.u.YUVA.u = u;
  idec.output.u.YUVA.u_stride = u_stride;
  idec.output.u.YUVA.u_size = u_size;
  idec.output.u.YUVA.v = v;
  idec.output.u.YUVA.v_stride = v_stride;
  idec.output.u.YUVA.v_size = v_size;
  idec.output.u.YUVA.a = a;
  idec.output.u.YUVA.a_stride = a_stride;
  idec.output.u.YUVA.a_size = a_size;
  return idec;
}

// Deprecated version of the above, without the alpha plane.
// Kept for backward compatibility.
func WebPINewYUV(luma *uint8, luma_size uint64, luma_stride int, u *uint8, u_size uint64, u_stride int, v *uint8, v_size uint64, v_stride int) *WebPIDecoder {
  return WebPINewYUVA(luma, luma_size, luma_stride, u, u_size, u_stride, v, v_size, v_stride, nil, 0, 0);
}

//------------------------------------------------------------------------------

func IDecCheckStatus(/* const */ idec *WebPIDecoder) vp8.VP8StatusCode {
  assert.Assert(idec);
  if (idec.state == STATE_ERROR) {
    return VP8_STATUS_BITSTREAM_ERROR;
  }
  if (idec.state == STATE_DONE) {
    return VP8_STATUS_OK;
  }
  return VP8_STATUS_SUSPENDED;
}

// Copies and decodes the next available data. Returns VP8_STATUS_OK when
// the image is successfully decoded. Returns VP8_STATUS_SUSPENDED when more
// data is expected. Returns error in other cases.
func WebPIAppend(idec *WebPIDecoder, /*const*/ data *uint8, data_size uint64) vp8.VP8StatusCode {
  var status VP8StatusCode
  if (idec == nil || data == nil) {
    return VP8_STATUS_INVALID_PARAM;
  }
  status = IDecCheckStatus(idec);
  if (status != VP8_STATUS_SUSPENDED) {
    return status;
  }
  // Check mixed calls between RemapMemBuffer and AppendToMemBuffer.
  if (!CheckMemBufferMode(&idec.mem, MEM_MODE_APPEND)) {
    return VP8_STATUS_INVALID_PARAM;
  }
  // Append data to memory buffer
  if (!AppendToMemBuffer(idec, data, data_size)) {
    return VP8_STATUS_OUT_OF_MEMORY;
  }
  return IDecode(idec);
}

// A variant of the above function to be used when data buffer contains
// partial data from the beginning. In this case data buffer is not copied
// to the internal memory.
// Note that the value of the 'data' pointer can change between calls to
// WebPIUpdate, for instance when the data buffer is resized to fit larger data.
func WebPIUpdate(idec *WebPIDecoder, /*const*/ data *uint8, data_size uint64) VP8StatusCode {
  var status VP8StatusCode
  if (idec == nil || data == nil) {
    return VP8_STATUS_INVALID_PARAM;
  }
  status = IDecCheckStatus(idec);
  if (status != VP8_STATUS_SUSPENDED) {
    return status;
  }
  // Check mixed calls between RemapMemBuffer and AppendToMemBuffer.
  if (!CheckMemBufferMode(&idec.mem, MEM_MODE_MAP)) {
    return VP8_STATUS_INVALID_PARAM;
  }
  // Make the memory buffer point to the new buffer
  if (!RemapMemBuffer(idec, data, data_size)) {
    return VP8_STATUS_INVALID_PARAM;
  }
  return IDecode(idec);
}

//------------------------------------------------------------------------------

func GetOutputBuffer(/* const */ idec *WebPIDecoder)  *WebPDecBuffer {
  if (idec == nil || idec.dec == nil) {
    return nil;
  }
  if (idec.state <= STATE_VP8_PARTS0) {
    return nil;
  }
  if (idec.final_output != nil) {
    return nil;  // not yet slow-copied
  }
  return idec.params.output;
}

// Generic call to retrieve information about the displayable area.
// If non nil, the left/right/width/height pointers are filled with the visible
// rectangular area so far.
// Returns nil in case the incremental decoder object is in an invalid state.
// Otherwise returns the pointer to the internal representation. This structure
// is read-only, tied to WebPIDecoder's lifespan and should not be modified.
func WebPIDecodedArea(/* const */ idec *WebPIDecoder, left *int, top *int, width *int, height *int) *WebPDecBuffer {
  var src *WebPDecBuffer = GetOutputBuffer(idec);
  if left != nil { *left = 0 }
  if top != nil { *top = 0 }
  if (src != nil) {
    if width != nil { *width = src.width }
    if height != nil { *height = idec.params.last_y }
  } else {
    if width != nil { *width = 0 }
    if height != nil { *height = 0 }
  }
  return src;
}

// Returns the RGB/A image decoded so far. Returns nil if output params
// are not initialized yet. The RGB/A output type corresponds to the colorspace
// specified during call to WebPINewDecoder() or WebPINewRGB().
// is the index *last_y of last decoded row in raster scan order. Some pointers
// (*last_y, etc *width.) can be nil if corresponding information is not
// needed. The values in these pointers are only valid on successful (non-nil)
// return.
func WebPIDecGetRGB(/* const */ idec *WebPIDecoder, last_y *int, width *int, height *int, stride *int) *uint8 {
  var src *WebPDecBuffer = GetOutputBuffer(idec);
  if src == nil { { return nil } }
  if (src.colorspace >= MODE_YUV) {
    return nil;
  }

  if last_y != nil { *last_y = idec.params.last_y }
  if width != nil { *width = src.width }
  if height != nil { *height = src.height }
  if stride != nil { *stride = src.u.RGBA.stride }

  return src.u.RGBA.rgba;
}

// Same as above function to get a YUVA image. Returns pointer to the luma
// plane or nil in case of error. If there is no alpha information
// the alpha pointer '*a' will be returned nil.
func WebPIDecGetYUVA(/* const */ idec *WebPIDecoder, last_y *int, u \*uint8, v *uint8, a *uint8, width *int, height *int, stride *int, uv_stride *int, a_stride *int) *uint8 {
  var src *WebPDecBuffer = GetOutputBuffer(idec);
  if src == nil { { return nil } }
  if (src.colorspace < MODE_YUV) {
    return nil;
  }

  if last_y != nil { *last_y = idec.params.last_y }
  if u != nil { *u = src.u.YUVA.u }
  if v != nil { *v = src.u.YUVA.v }
  if a != nil { *a = src.u.YUVA.a }
  if width != nil { *width = src.width }
  if height != nil { *height = src.height }
  if stride != nil { *stride = src.u.YUVA.y_stride }
  if uv_stride != nil { *uv_stride = src.u.YUVA.u_stride }
  if a_stride != nil { *a_stride = src.u.YUVA.a_stride }

  return src.u.YUVA.y;
}

// Set the custom IO function pointers and user-data. The setter for IO hooks
// should be called before initiating incremental decoding. Returns true if
// WebPIDecoder object is successfully modified, false otherwise.
func WebPISetIOHooks(/* const */ idec *WebPIDecoder, VP8IoPutHook put, VP8IoSetupHook setup, VP8IoTeardownHook teardown, user_data *void) int {
  if (idec == nil || idec.state > STATE_WEBP_HEADER) {
    return 0;
  }

  idec.io.put = put;
  idec.io.setup = setup;
  idec.io.teardown = teardown;
  idec.io.opaque = user_data;

  return 1;
}
