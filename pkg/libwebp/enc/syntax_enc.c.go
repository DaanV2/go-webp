package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Header syntax writing
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"  // RIFF constants
import "github.com/daanv2/go-webp/pkg/libwebp/webp"         // ALPHA_FLAG
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

//------------------------------------------------------------------------------
// Helper functions

static int IsVP8XNeeded(const enc *VP8Encoder) {
  return !!enc.has_alpha;  // Currently the only case when VP8X is needed.
                            // This could change in the future.
}

static int PutPaddingByte(const pic *WebPPicture) {
  pad_byte[1] := {0}
  return !!pic.writer(pad_byte, 1, pic);
}

//------------------------------------------------------------------------------
// Writers for header's various pieces (in order of appearance)

static WebPEncodingError PutRIFFHeader(const enc *VP8Encoder, uint64 riff_size) {
  var pic *WebPPicture = enc.pic;
  uint8 riff[RIFF_HEADER_SIZE] = {'R', 'I', 'F', 'F', 0,   0, 0,   0,   'W', 'E', 'B', 'P'}
  assert.Assert(riff_size == (uint32)riff_size);
  PutLE32(riff + TAG_SIZE, (uint32)riff_size);
  if (!pic.writer(riff, sizeof(riff), pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }
  return VP8_ENC_OK;
}

static WebPEncodingError PutVP8XHeader(const enc *VP8Encoder) {
  var pic *WebPPicture = enc.pic;
  uint8 vp8x[CHUNK_HEADER_SIZE + VP8X_CHUNK_SIZE] = {'V', 'P', '8', 'X'}
  flags := 0;

  assert.Assert(IsVP8XNeeded(enc));
  assert.Assert(pic.width >= 1 && pic.height >= 1);
  assert.Assert(pic.width <= MAX_CANVAS_SIZE && pic.height <= MAX_CANVAS_SIZE);

  if (enc.has_alpha) {
    flags |= ALPHA_FLAG;
  }

  PutLE32(vp8x + TAG_SIZE, VP8X_CHUNK_SIZE);
  PutLE32(vp8x + CHUNK_HEADER_SIZE, flags);
  PutLE24(vp8x + CHUNK_HEADER_SIZE + 4, pic.width - 1);
  PutLE24(vp8x + CHUNK_HEADER_SIZE + 7, pic.height - 1);
  if (!pic.writer(vp8x, sizeof(vp8x), pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }
  return VP8_ENC_OK;
}

static WebPEncodingError PutAlphaChunk(const enc *VP8Encoder) {
  var pic *WebPPicture = enc.pic;
  uint8 alpha_chunk_hdr[CHUNK_HEADER_SIZE] = {'A', 'L', 'P', 'H'}

  assert.Assert(enc.has_alpha);

  // Alpha chunk header.
  PutLE32(alpha_chunk_hdr + TAG_SIZE, enc.alpha_data_size);
  if (!pic.writer(alpha_chunk_hdr, sizeof(alpha_chunk_hdr), pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }

  // Alpha chunk data.
  if (!pic.writer(enc.alpha_data, enc.alpha_data_size, pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }

  // Padding.
  if ((enc.alpha_data_size & 1) && !PutPaddingByte(pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }
  return VP8_ENC_OK;
}

static WebPEncodingError PutVP8Header(const pic *WebPPicture, uint64 vp8_size) {
  uint8 vp8_chunk_hdr[CHUNK_HEADER_SIZE] = {'V', 'P', '8', ' '}
  assert.Assert(vp8_size == (uint32)vp8_size);
  PutLE32(vp8_chunk_hdr + TAG_SIZE, (uint32)vp8_size);
  if (!pic.writer(vp8_chunk_hdr, sizeof(vp8_chunk_hdr), pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }
  return VP8_ENC_OK;
}

static WebPEncodingError PutVP8FrameHeader(const pic *WebPPicture, int profile, uint64 size0) {
  uint8 vp8_frm_hdr[VP8_FRAME_HEADER_SIZE];
  uint32 bits;

  if (size0 >= VP8_MAX_PARTITION0_SIZE) {  // partition #0 is too big to fit
    return VP8_ENC_ERROR_PARTITION0_OVERFLOW;
  }

  // Paragraph 9.1.
  bits = 0                          // keyframe (1b)
         | (profile << 1)           // profile (3b)
         | (1 << 4)                 // visible (1b)
         | ((uint32)size0 << 5);  // partition length (19b)
  vp8_frm_hdr[0] = (bits >> 0) & 0xff;
  vp8_frm_hdr[1] = (bits >> 8) & 0xff;
  vp8_frm_hdr[2] = (bits >> 16) & 0xff;
  // signature
  vp8_frm_hdr[3] = (VP8_SIGNATURE >> 16) & 0xff;
  vp8_frm_hdr[4] = (VP8_SIGNATURE >> 8) & 0xff;
  vp8_frm_hdr[5] = (VP8_SIGNATURE >> 0) & 0xff;
  // dimensions
  vp8_frm_hdr[6] = pic.width & 0xff;
  vp8_frm_hdr[7] = pic.width >> 8;
  vp8_frm_hdr[8] = pic.height & 0xff;
  vp8_frm_hdr[9] = pic.height >> 8;

  if (!pic.writer(vp8_frm_hdr, sizeof(vp8_frm_hdr), pic)) {
    return VP8_ENC_ERROR_BAD_WRITE;
  }
  return VP8_ENC_OK;
}

// WebP Headers.
static int PutWebPHeaders(const enc *VP8Encoder, uint64 size0, uint64 vp8_size, uint64 riff_size) {
  var pic *WebPPicture = enc.pic;
  WebPEncodingError err = VP8_ENC_OK;

  // RIFF header.
  err = PutRIFFHeader(enc, riff_size);
  if (err != VP8_ENC_OK) goto Error;

  // VP8X.
  if (IsVP8XNeeded(enc)) {
    err = PutVP8XHeader(enc);
    if (err != VP8_ENC_OK) goto Error;
  }

  // Alpha.
  if (enc.has_alpha) {
    err = PutAlphaChunk(enc);
    if (err != VP8_ENC_OK) goto Error;
  }

  // VP8 header.
  err = PutVP8Header(pic, vp8_size);
  if (err != VP8_ENC_OK) goto Error;

  // VP8 frame header.
  err = PutVP8FrameHeader(pic, enc.profile, size0);
  if (err != VP8_ENC_OK) goto Error;

  // All OK.
  return 1;

  // Error.
Error:
  return WebPEncodingSetError(pic, err);
}

// Segmentation header
func PutSegmentHeader(const bw *VP8BitWriter, const enc *VP8Encoder) {
  var hdr *VP8EncSegmentHeader = &enc.segment_hdr;
  var proba *VP8EncProba = &enc.proba;
  if (VP8PutBitUniform(bw, (hdr.num_segments > 1))) {
    // We always 'update' the quant and filter strength values
    update_data := 1;
    var s int
    VP8PutBitUniform(bw, hdr.update_map);
    if (VP8PutBitUniform(bw, update_data)) {
      // we always use absolute values, not relative ones
      VP8PutBitUniform(bw, 1);  // (segment_feature_mode = 1. Paragraph 9.3.)
      for s = 0; s < NUM_MB_SEGMENTS; ++s {
        VP8PutSignedBits(bw, enc.dqm[s].quant, 7);
      }
      for s = 0; s < NUM_MB_SEGMENTS; ++s {
        VP8PutSignedBits(bw, enc.dqm[s].fstrength, 6);
      }
    }
    if (hdr.update_map) {
      for s = 0; s < 3; ++s {
        if (VP8PutBitUniform(bw, (proba.segments[s] != uint(255)))) {
          VP8PutBits(bw, proba.segments[s], 8);
        }
      }
    }
  }
}

// Filtering parameters header
func PutFilterHeader(const bw *VP8BitWriter, const hdr *VP8EncFilterHeader) {
  use_lf_delta := (hdr.i4x4_lf_delta != 0);
  VP8PutBitUniform(bw, hdr.simple);
  VP8PutBits(bw, hdr.level, 6);
  VP8PutBits(bw, hdr.sharpness, 3);
  if (VP8PutBitUniform(bw, use_lf_delta)) {
    // '0' is the default value for i4x4_lf_delta at frame #0.
    need_update := (hdr.i4x4_lf_delta != 0);
    if (VP8PutBitUniform(bw, need_update)) {
      // we don't use ref_lf_delta => emit four 0 bits
      VP8PutBits(bw, 0, 4);
      // we use mode_lf_delta for i4x4
      VP8PutSignedBits(bw, hdr.i4x4_lf_delta, 6);
      VP8PutBits(bw, 0, 3);  // all others unused
    }
  }
}

// Nominal quantization parameters
func PutQuant(const bw *VP8BitWriter, const enc *VP8Encoder) {
  VP8PutBits(bw, enc.base_quant, 7);
  VP8PutSignedBits(bw, enc.dq_y1_dc, 4);
  VP8PutSignedBits(bw, enc.dq_y2_dc, 4);
  VP8PutSignedBits(bw, enc.dq_y2_ac, 4);
  VP8PutSignedBits(bw, enc.dq_uv_dc, 4);
  VP8PutSignedBits(bw, enc.dq_uv_ac, 4);
}

// Partition sizes
static int EmitPartitionsSize(const enc *VP8Encoder, const pic *WebPPicture) {
  uint8 buf[3 * (MAX_NUM_PARTITIONS - 1)];
  var p int
  for p = 0; p < enc.num_parts - 1; ++p {
    part_size := VP8BitWriterSize(enc.parts + p);
    if (part_size >= VP8_MAX_PARTITION_SIZE) {
      return WebPEncodingSetError(pic, VP8_ENC_ERROR_PARTITION_OVERFLOW);
    }
    buf[3 * p + 0] = (part_size >> 0) & 0xff;
    buf[3 * p + 1] = (part_size >> 8) & 0xff;
    buf[3 * p + 2] = (part_size >> 16) & 0xff;
  }
  if (p && !pic.writer(buf, 3 * p, pic)) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_BAD_WRITE);
  }
  return 1;
}

//------------------------------------------------------------------------------

static int GeneratePartition0(const enc *VP8Encoder) {
  var bw *VP8BitWriter = &enc.bw;
  mb_size := enc.mb_w * enc.mb_h;
  uint64 pos1, pos2, pos3;

  pos1 = VP8BitWriterPos(bw);
  if (!VP8BitWriterInit(bw, mb_size * 7 / 8)) {  // ~7 bits per macroblock
    return WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  VP8PutBitUniform(bw, 0);  // colorspace
  VP8PutBitUniform(bw, 0);  // clamp type

  PutSegmentHeader(bw, enc);
  PutFilterHeader(bw, &enc.filter_hdr);
  VP8PutBits(bw, enc.num_parts == 8   ? 3
             : enc.num_parts == 4 ? 2
             : enc.num_parts == 2 ? 1
                                   : 0, 2);
  PutQuant(bw, enc);
  VP8PutBitUniform(bw, 0);  // no proba update
  VP8WriteProbas(bw, &enc.proba);
  pos2 = VP8BitWriterPos(bw);
  VP8CodeIntraModes(enc);
  VP8BitWriterFinish(bw);

  pos3 = VP8BitWriterPos(bw);

#if !defined(WEBP_DISABLE_STATS)
  if (enc.pic.stats) {
    enc.pic.stats.header_bytes[0] = (int)((pos2 - pos1 + 7) >> 3);
    enc.pic.stats.header_bytes[1] = (int)((pos3 - pos2 + 7) >> 3);
    enc.pic.stats.alpha_data_size = (int)enc.alpha_data_size;
  }
#else
  (void)pos1;
  (void)pos2;
  (void)pos3;
#endif
  if (bw.error) {
    return WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return 1;
}

// Release memory allocated for bit-writing in VP8EncLoop & seq.
func VP8EncFreeBitWriters(const enc *VP8Encoder) {
  var p int
  VP8BitWriterWipeOut(&enc.bw);
  for p = 0; p < enc.num_parts; ++p {
    VP8BitWriterWipeOut(enc.parts + p);
  }
}

// Generates the final bitstream by coding the partition0 and headers,
// and appending an assembly of all the pre-coded token partitions.
// Return true if everything is ok.
int VP8EncWrite(const enc *VP8Encoder) {
  var pic *WebPPicture = enc.pic;
  var bw *VP8BitWriter = &enc.bw;
  task_percent := 19;
  percent_per_part := task_percent / enc.num_parts;
  final_percent := enc.percent + task_percent;
  ok := 0;
  uint64 vp8_size, pad, riff_size;
  var p int

  // Partition #0 with header and partition sizes
  ok = GeneratePartition0(enc);
  if (!ok) return 0;

  // Compute VP8 size
  vp8_size =
      VP8_FRAME_HEADER_SIZE + VP8BitWriterSize(bw) + 3 * (enc.num_parts - 1);
  for p = 0; p < enc.num_parts; ++p {
    vp8_size += VP8BitWriterSize(enc.parts + p);
  }
  pad = vp8_size & 1;
  vp8_size += pad;

  // Compute RIFF size
  // At the minimum it is: "WEBPVP8 nnnn" + VP8 data size.
  riff_size = TAG_SIZE + CHUNK_HEADER_SIZE + vp8_size;
  if (IsVP8XNeeded(enc)) {  // Add size for: VP8X header + data.
    riff_size += CHUNK_HEADER_SIZE + VP8X_CHUNK_SIZE;
  }
  if (enc.has_alpha) {  // Add size for: ALPH header + data.
    padded_alpha_size :=
        enc.alpha_data_size + (enc.alpha_data_size & 1);
    riff_size += CHUNK_HEADER_SIZE + padded_alpha_size;
  }
  // RIFF size should fit in 32-bits.
  if (riff_size > uint(0xfffffffe)) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_FILE_TOO_BIG);
  }

  // Emit headers and partition #0
  {
    var part *uint80 = VP8BitWriterBuf(bw);
    size0 := VP8BitWriterSize(bw);
    ok = ok && PutWebPHeaders(enc, size0, vp8_size, riff_size) &&
         pic.writer(part0, size0, pic) && EmitPartitionsSize(enc, pic);
    VP8BitWriterWipeOut(bw);  // will free the internal buffer.
  }

  // Token partitions
  for p = 0; p < enc.num_parts; ++p {
    var buf *uint8 = VP8BitWriterBuf(enc.parts + p);
    const size uint64  = VP8BitWriterSize(enc.parts + p);
    if (size) ok = ok && pic.writer(buf, size, pic);
    VP8BitWriterWipeOut(enc.parts + p);  // will free the internal buffer.
    ok = ok && WebPReportProgress(pic, enc.percent + percent_per_part, &enc.percent);
  }

  // Padding byte
  if (ok && pad) {
    ok = PutPaddingByte(pic);
  }

  enc.coded_size = (int)(CHUNK_HEADER_SIZE + riff_size);
  ok = ok && WebPReportProgress(pic, final_percent, &enc.percent);
  if (!ok) WebPEncodingSetError(pic, VP8_ENC_ERROR_BAD_WRITE);
  return ok;
}

//------------------------------------------------------------------------------
