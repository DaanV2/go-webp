
// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

const (
	kCat3 = []uint8{173, 148, 140, 0}
	kCat4 = []uint8{176, 155, 140, 135, 0}
	kCat5 = []uint8{180, 157, 141, 134, 130, 0}
	kCat6 = []uint8{254, 254, 243, 230, 196, 177, 153, 140, 133, 130, 129, 0}
	kCat3456 = [][]uint8{kCat3, kCat4, kCat5, kCat6}
	kZigzag = [16]uint8{0, 1,  4,  8,  5, 2,  3,  6, 9, 12, 13, 10, 7, 11, 14, 15}
)

// Input / Output
type VP8IoPutHook func(/* const */ io *VP8Io)int 
type VP8IoSetupHook func(io *VP8Io)int 
type VP8IoTeardownHook func(/* const */ io *VP8Io) 

type VP8Io struct {
	// set by VP8GetHeaders()
	// picture dimensions, in pixels (invariable).
	// These are the original, uncropped dimensions.
	// The actual area passed to put() is stored
	// in mb_w / mb_h fields.

	// set before calling put()
	width, height int  
	mb_y int                  // position of the current rows (in pixels)
	mb_w int                  // number of columns in the sample
	mb_h int                  // number of rows in the sample
  y, u, v *uint8  // rows to copy (in yuv420 format)
   y_stride int              // row stride for luma
   uv_stride int             // row stride for chroma

  opaque *void;  // user data

  // called when fresh samples are available. Currently, samples are in
  // YUV420 format, and can be up to width x 24 in size (depending on the
  // in-loop filtering level, e.g.). Should return false in case of error
  // or abort request. The actual size of the area to update is mb_w x mb_h
  // in size, taking cropping into account.
  put VP8IoPutHook;

  // called just before starting to decode the blocks.
  // Must return false in case of setup error, true otherwise. If false is
  // returned, teardown() will NOT be called. But if the setup succeeded
  // and true is returned, then teardown() will always be called afterward.
   setup VP8IoSetupHook;

  // Called just after block decoding is finished (or when an error occurred
  // during put()). Is NOT called if setup() failed.
   teardown VP8IoTeardownHook;

  // this is a recommendation for the user-side yuv.rgb converter. This flag
  // is set when calling setup() hook and can be overwritten by it. It then
  // can be taken into consideration during the put() method.
   fancy_upsampling int

  // Input buffer.
   data_size uint64
  data *uint8;

  // If true, in-loop filtering will not be performed even if present in the
  // bitstream. Switching off filtering may speed up decoding at the expense
  // of more visible blocking. Note that output will also be non-compliant
  // with the VP8 specifications.
  bypass_filtering int ;

  // Cropping parameters.
  use_cropping int ;
  crop_left, crop_right, crop_top, crop_bottom int

  // Scaling parameters.
   use_scaling int
   scaled_width, scaled_height int

  // If non nil, pointer to the alpha data (if present) corresponding to the
  // start of the current row (That is: it is pre-offset by mb_y and takes
  // cropping into account).
  a *uint8;
}

//------------------------------------------------------------------------------
// Signature and pointer-to-function for GetCoeffs() variants below.

type GetCoeffsFunc = func(br *VP8BitReader, /* const */ prob []*VP8BandProbas,  ctx int, dq quant_t, n int, out *int16) int;

//------------------------------------------------------------------------------
// VP8Decoder

func SetOk(/* const */ dec *VP8Decoder) {
  dec.status = VP8_STATUS_OK;
  dec.error_msg = "OK";
}

// Internal, version-checked, entry point
func VP8InitIoInternal(/* const */ io *VP8Io, version int) int {
  if (WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_DECODER_ABI_VERSION)) {
    return 0;  // mismatch error
  }
  if (io != nil) {
    stdlib.Memset(io, 0, sizeof(*io));
  }
  return 1;
}

// Create a new decoder object.
func VP8Decoder8New() *VP {
  var dec *VP8Decoder = (*VP8Decoder)WebPSafeCalloc(uint64(1), sizeof(*dec));
  if (dec != nil) {
    SetOk(dec);
    WebPGetWorkerInterface().Init(&dec.worker);
    dec.ready = 0;
    dec.num_parts_minus_one = 0;
    InitGetCoeffs();
  }
  return dec;
}

// Return current status of the decoder:
func VP8Status(/* const */ dec *VP8Decoder) VP8StatusCode {
  if (!dec) { return VP8_STATUS_INVALID_PARAM }
  return dec.status;
}

// return readable string corresponding to the last status.
func byte8StatusMessage(/* const */ dec *VP8Decoder) *VP {
  if (dec == nil) { return "no object"; }
  if (!dec.error_msg) { return "OK"; }
  return dec.error_msg;
}

// Destroy the decoder object.
func VP8Delete(/* const */ dec *VP8Decoder) {
  if (dec != nil) {
    VP8Clear(dec);
  }
}

func VP8SetError(/* const */ dec *VP8Decoder, error VP8StatusCode, /* const */ msg *byte) int {
  // VP8_STATUS_SUSPENDED is only meaningful in incremental decoding.
  assert.Assert(dec.incremental || error != VP8_STATUS_SUSPENDED);
  // The oldest error reported takes precedence over the new one.
  if (dec.status == VP8_STATUS_OK) {
    dec.status = error;
    dec.error_msg = msg;
    dec.ready = 0;
  }
  return 0;
}

// Returns true if the next 3 bytes in data contain the VP8 signature.
func VP8CheckSignature(data []uint8, data_size uint64) int {
  return (data_size >= 3 && data[0] == 0x9d && data[1] == 0x01 && data[2] == 0x2a);
}

// Validates the VP8 data-header and retrieves basic header information viz
// width and height. Returns 0 in case of formatting error. *width/*height
// can be passed nil.
func VP8GetInfo(data *uint8, data_size uint64, chunk_size uint64, width *int, height *int) int {
  if (data == nil || data_size < VP8_FRAME_HEADER_SIZE) {
    return 0;  // not enough data
  }
  // check signature
  if (!VP8CheckSignature(data + 3, data_size - 3)) {
    return 0;  // Wrong signature.
  } else {
    bits := data[0] | (data[1] << 8) | (data[2] << 16);
    key_frame := !(bits & 1);
    w := ((data[7] << 8) | data[6]) & 0x3fff;
    h := ((data[9] << 8) | data[8]) & 0x3fff;

    if (!key_frame) {  // Not a keyframe.
      return 0;
    }

    if (((bits >> 1) & 7) > 3) {
      return 0;  // unknown profile
    }
    if (!((bits >> 4) & 1)) {
      return 0;  // first frame is invisible!
    }
    if (((bits >> 5)) >= chunk_size) {  // partition_length
      return 0;                         // inconsistent size information.
    }
    if (w == 0 || h == 0) {
      return 0;  // We don't support both width and height to be zero.
    }

    if (width) {
      *width = w;
    }
    if (height) {
      *height = h;
    }

    return 1;
  }
}

// Header parsing
func ResetSegmentHeader(/* const */ hdr *VP8SegmentHeader) {
  assert.Assert(hdr != nil);
  hdr.use_segment = 0;
  hdr.update_map = 0;
  hdr.absolute_delta = 1;
  stdlib.Memset(hdr.quantizer, 0, sizeof(hdr.quantizer));
  stdlib.Memset(hdr.filter_strength, 0, sizeof(hdr.filter_strength));
}

// Paragraph 9.3
func ParseSegmentHeader(br *VP8BitReader, hdr *VP8SegmentHeader, proba *VP8Proba) int {
  assert.Assert(br != nil);
  assert.Assert(hdr != nil);
  hdr.use_segment = VP8Get(br, "global-header");
  if (hdr.use_segment) {
    hdr.update_map = VP8Get(br, "global-header");
    if (VP8Get(br, "global-header")) {  // update data
      var s int
      hdr.absolute_delta = VP8Get(br, "global-header");
      for s = 0; s < NUM_MB_SEGMENTS; s++ {
        hdr.quantizer[s] = VP8Get(br, "global-header")
                                ? VP8GetSignedValue(br, 7, "global-header")
                                : 0;
      }
      for s = 0; s < NUM_MB_SEGMENTS; s++ {
        hdr.filter_strength[s] =
            VP8Get(br, "global-header")
                ? VP8GetSignedValue(br, 6, "global-header")
                : 0;
      }
    }
    if (hdr.update_map) {
      var s int
      for s = 0; s < MB_FEATURE_TREE_PROBS; s++ {
        proba.segments[s] = VP8Get(br, "global-header")
                                 ? VP8GetValue(br, 8, "global-header")
                                 : uint(255);
      }
    }
  } else {
    hdr.update_map = 0;
  }
  return !br.eof;
}

// Paragraph 9.5
// If we don't have all the necessary data in 'buf', this function returns
// VP8_STATUS_SUSPENDED in incremental decoding, VP8_STATUS_NOT_ENOUGH_DATA
// otherwise.
// In incremental decoding, this case is not necessarily an error. Still, no
// bitreader is ever initialized to make it possible to read unavailable memory.
// If we don't even have the partitions' sizes, then VP8_STATUS_NOT_ENOUGH_DATA
// is returned, and this is an unrecoverable error.
// If the partitions were positioned ok, VP8_STATUS_OK is returned.
func ParsePartitions(dec *VP8Decoder, buf *uint8, size uint64 ) VP8StatusCode {
  var br *VP8BitReader = &dec.br;
  const sz *uint8 = buf;
  var buf_end *uint8 = buf + size;
  const part_start *uint8;
  size_left := size;
  var last_part uint64
  var p uint64

  dec.num_parts_minus_one = (1 << VP8GetValue(br, 2, "global-header")) - 1;
  last_part = dec.num_parts_minus_one;
  if (size < 3 * last_part) {
    // we can't even read the sizes with sz[]! That's a failure.
    return VP8_STATUS_NOT_ENOUGH_DATA;
  }
  part_start = buf + last_part * 3;
  size_left -= last_part * 3;
  for p = 0; p < last_part; p++ {
    psize := sz[0] | (sz[1] << 8) | (sz[2] << 16);
    if (psize > size_left) psize = size_left;
    VP8InitBitReader(dec.parts + p, part_start, psize);
    part_start += psize;
    size_left -= psize;
    sz += 3;
  }
  VP8InitBitReader(dec.parts + last_part, part_start, size_left);
  if (part_start < buf_end) {return VP8_STATUS_OK}

  return tenary.If(dec.incremental, 
		VP8_STATUS_SUSPENDED,  // Init is ok, but there's not enough data
        VP8_STATUS_NOT_ENOUGH_DATA)
}

// Paragraph 9.4
func ParseFilterHeader(br *VP8BitReader, /* const */ dec *VP8Decoder) int {
  var hdr *VP8FilterHeader = &dec.filter_hdr;
  hdr.simple = VP8Get(br, "global-header");
  hdr.level = VP8GetValue(br, 6, "global-header");
  hdr.sharpness = VP8GetValue(br, 3, "global-header");
  hdr.use_lf_delta = VP8Get(br, "global-header");
  if (hdr.use_lf_delta) {
    if (VP8Get(br, "global-header")) {  // update lf-delta?
      var i int
      for i = 0; i < NUM_REF_LF_DELTAS; i++ {
        if (VP8Get(br, "global-header")) {
          hdr.ref_lf_delta[i] = VP8GetSignedValue(br, 6, "global-header");
        }
      }
      for i = 0; i < NUM_MODE_LF_DELTAS; i++ {
        if (VP8Get(br, "global-header")) {
          hdr.mode_lf_delta[i] = VP8GetSignedValue(br, 6, "global-header");
        }
      }
    }
  }
  dec.filter_type = (hdr.level == 0) ? 0 : tenary.If(hdr.simple, 1, 2);
  return !br.eof;
}

// Decode the VP8 frame header. Returns true if ok.
// Note: 'io.data' must be pointing to the start of the VP8 frame header.
func VP8GetHeaders(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) int {
  var buf_size uint64;
  var buf *uint8 ;
  var frm_hdr *VP8FrameHeader;
  var pic_hdr *VP8PictureHeader;
  var br *VP8BitReader;
  var status VP8StatusCode;

  if (dec == nil) {
    return 0;
  }
  SetOk(dec);
  if (io == nil) {
    return VP8SetError(dec, VP8_STATUS_INVALID_PARAM, "nil VP8Io passed to VP8GetHeaders()");
  }
  buf_size = io.data_size;
  buf = io.data // bidi index -> io.data_size;
  if (buf_size < 4) {
    return VP8SetError(dec, VP8_STATUS_NOT_ENOUGH_DATA, "Truncated header.");
  }

  // Paragraph 9.1
  {
    bits := buf[0] | (buf[1] << 8) | (buf[2] << 16);
    frm_hdr = &dec.frm_hdr;
    frm_hdr.key_frame = !(bits & 1);
    frm_hdr.profile = (bits >> 1) & 7;
    frm_hdr.show = (bits >> 4) & 1;
    frm_hdr.partition_length = (bits >> 5);
    if (frm_hdr.profile > 3) {
      return VP8SetError(dec, VP8_STATUS_BITSTREAM_ERROR, "Incorrect keyframe parameters.");
    }
    if (!frm_hdr.show) {
      return VP8SetError(dec, VP8_STATUS_UNSUPPORTED_FEATURE, "Frame not displayable.");
    }
    buf += 3;
    buf_size -= 3;
  }

  pic_hdr = &dec.pic_hdr;
  if (frm_hdr.key_frame) {
    // Paragraph 9.2
    if (buf_size < 7) {
      return VP8SetError(dec, VP8_STATUS_NOT_ENOUGH_DATA, "cannot parse picture header");
    }
    if (!VP8CheckSignature(buf, buf_size)) {
      return VP8SetError(dec, VP8_STATUS_BITSTREAM_ERROR, "Bad code word");
    }
    pic_hdr.width = ((buf[4] << 8) | buf[3]) & 0x3fff;
    pic_hdr.xscale = buf[4] >> 6;  // ratio: 1, 5/4 5/3 or 2
    pic_hdr.height = ((buf[6] << 8) | buf[5]) & 0x3fff;
    pic_hdr.yscale = buf[6] >> 6;
    buf += 7;
    buf_size -= 7;

    dec.mb_w = (pic_hdr.width + 15) >> 4;
    dec.mb_h = (pic_hdr.height + 15) >> 4;

    // Setup default output area (can be later modified during io.setup())
    io.width = pic_hdr.width;
    io.height = pic_hdr.height;
    // IMPORTANT! use some sane dimensions in and fields *crop *scaled.
    // So they can be used interchangeably without always testing for
    // 'use_cropping'.
    io.use_cropping = 0;
    io.crop_top = 0;
    io.crop_left = 0;
    io.crop_right = io.width;
    io.crop_bottom = io.height;
    io.use_scaling = 0;
    io.scaled_width = io.width;
    io.scaled_height = io.height;

    io.mb_w = io.width;   // for soundness
    io.mb_h = io.height;  // ditto

    VP8ResetProba(&dec.proba);
    ResetSegmentHeader(&dec.segment_hdr);
  }

  // Check if we have all the partition #0 available, and initialize dec.br
  // to read this partition (and this partition only).
  if (frm_hdr.partition_length > buf_size) {
    return VP8SetError(dec, VP8_STATUS_NOT_ENOUGH_DATA, "bad partition length");
  }

  br = &dec.br;
  VP8InitBitReader(br, buf, frm_hdr.partition_length);
  buf += frm_hdr.partition_length;
  buf_size -= frm_hdr.partition_length;

  if (frm_hdr.key_frame) {
    pic_hdr.colorspace = VP8Get(br, "global-header");
    pic_hdr.clamp_type = VP8Get(br, "global-header");
  }
  if (!ParseSegmentHeader(br, &dec.segment_hdr, &dec.proba)) {
    return VP8SetError(dec, VP8_STATUS_BITSTREAM_ERROR, "cannot parse segment header");
  }
  // Filter specs
  if (!ParseFilterHeader(br, dec)) {
    return VP8SetError(dec, VP8_STATUS_BITSTREAM_ERROR, "cannot parse filter header");
  }
  status = ParsePartitions(dec, buf, buf_size);
  if (status != VP8_STATUS_OK) {
    return VP8SetError(dec, status, "cannot parse partitions");
  }

  // quantizer change
  VP8ParseQuant(dec);

  // Frame buffer marking
  if (!frm_hdr.key_frame) {
    return VP8SetError(dec, VP8_STATUS_UNSUPPORTED_FEATURE, "Not a key frame.");
  }

  VP8Get(br, "global-header");  // ignore the value of 'update_proba'

  VP8ParseProba(br, dec);

  // sanitized state
  dec.ready = 1;
  return 1;
}



// See section 13-2: https://datatracker.ietf.org/doc/html/rfc6386#section-13.2
func GetLargeValue(/* const */ br *VP8BitReader, /* const */ p *uint8) int {
  var v int
  if (!VP8GetBit(br, p[3], "coeffs")) {
    if (!VP8GetBit(br, p[4], "coeffs")) {
      v = 2;
    } else {
      v = 3 + VP8GetBit(br, p[5], "coeffs");
    }
  } else {
    if (!VP8GetBit(br, p[6], "coeffs")) {
      if (!VP8GetBit(br, p[7], "coeffs")) {
        v = 5 + VP8GetBit(br, 159, "coeffs");
      } else {
        v = 7 + 2 * VP8GetBit(br, 165, "coeffs");
        v += VP8GetBit(br, 145, "coeffs");
      }
    } else {
      const tab *uint8;
      bit1 := VP8GetBit(br, p[8], "coeffs");
      bit0 := VP8GetBit(br, p[9 + bit1], "coeffs");
      cat := 2 * bit1 + bit0;
      v = 0;
      for tab = kCat3456[cat]; *tab; tab++ {
        v += v + VP8GetBit(br, *tab, "coeffs");
      }
      v += 3 + (8 << cat);
    }
  }
  return v;
}

// Returns the position of the last non-zero coeff plus one
func GetCoeffsFast(/* const */ br *VP8BitReader, /* const */ prob *VP8BandProbas[], ctx int, dq quant_t , n int, out *int16) int {
  var p *uint8 = prob[n].probas[ctx];
  for ; n < 16; n++ {
    if (!VP8GetBit(br, p[0], "coeffs")) {
      return n;  // previous coeff was last non-zero coeff
    }
    while (!VP8GetBit(br, p[1], "coeffs")) {  // sequence of zero coeffs
		n++
      p = prob[n].probas[0];
      if (n == 16) { return 16; }
    }
    {  // non zero coeff
      var p_ctx *VP8ProbaArray = &prob[n + 1].probas[0];
      var v int
      if (!VP8GetBit(br, p[2], "coeffs")) {
        v = 1;
        p = p_ctx[1];
      } else {
        v = GetLargeValue(br, p);
        p = p_ctx[2];
      }
      out[kZigzag[n]] = VP8GetSigned(br, v, "coeffs") * dq[n > 0];
    }
  }
  return 16;
}

// This version of GetCoeffs() uses VP8GetBitAlt() which is an alternate version
// of VP8GetBitAlt() targeting specific platforms.
func GetCoeffsAlt(/* const */ br *VP8BitReader, /* const */ prob *VP8BandProbas[], ctx int , dq quant_t , n int , out *int16) int {
  var p *uint8 = prob[n].probas[ctx];
  for ; n < 16; n++ {
    if (!VP8GetBitAlt(br, p[0], "coeffs")) {
      return n;  // previous coeff was last non-zero coeff
    }
    while (!VP8GetBitAlt(br, p[1], "coeffs")) {  // sequence of zero coeffs
      n++
		p = prob[n].probas[0];
      if (n == 16) {return 16};
    }
    {  // non zero coeff
      var p_ctx *VP8ProbaArray = &prob[n + 1].probas[0];
      var v int
      if (!VP8GetBitAlt(br, p[2], "coeffs")) {
        v = 1;
        p = p_ctx[1];
      } else {
        v = GetLargeValue(br, p);
        p = p_ctx[2];
      }
      out[kZigzag[n]] = VP8GetSigned(br, v, "coeffs") * dq[n > 0];
    }
  }
  return 16;
}


func init() {
	// func WEBP_DSP_INIT_FUNC(InitGetCoeffs) {
  if (VP8GetCPUInfo != nil && VP8GetCPUInfo(kSlowSSSE3)) {
    GetCoeffs = GetCoeffsAlt;
  } else {
    GetCoeffs = GetCoeffsFast;
  }
}

func NzCodeBits(nz_coeffs uint32, nz, dc_nz int) uint32 {
  nz_coeffs <<= 2;
  nz_coeffs |= (nz > 3) ? 3 : (nz > 1) ? 2 : dc_nz;
  return nz_coeffs;
}

func ParseResiduals(/* const */ dec *VP8Decoder, /*const*/ mb *VP8MB, /*const*/ token_br *VP8BitReader)int {
  const *VP8BandProbas(bands *const)[16 + 1] = dec.proba.bands_ptr;
  const ac_proba *VP8BandProbas *const;
  var block *VP8MBData = dec.mb_data + dec.mb_x;
  var q *VP8QuantMatrix = &dec.dqm[block.segment];
  dst *int16 = block.coeffs;
  var left_mb *VP8MB = dec.mb_info - 1;
  uint8 tnz, lnz;
  non_zero_y := 0;
  non_zero_uv := 0;
  int x, y, ch;
  uint32 out_t_nz, out_l_nz;
  var first int

  stdlib.Memset(dst, 0, 384 * sizeof(*dst));
  if (!block.is_i4x4) {  // parse DC
    int16 dc[16] = {0}
    ctx := mb.nz_dc + left_mb.nz_dc;
    nz := GetCoeffs(token_br, bands[1], ctx, q.y2_mat, 0, dc);
    mb.nz_dc = left_mb.nz_dc = (nz > 0);
    if (nz > 1) {  // more than just the DC . perform the full transform
      VP8TransformWHT(dc, dst);
    } else {  // only DC is non-zero . inlined simplified transform
      var i int
      dc0 := (dc[0] + 3) >> 3;
      for (i = 0; i < 16 * 16; i += 16) dst[i] = dc0;
    }
    first = 1;
    ac_proba = bands[0];
  } else {
    first = 0;
    ac_proba = bands[3];
  }

  tnz = mb.nz & float64(0x0f);
  lnz = left_mb.nz & float64(0x0f);
  for y = 0; y < 4; y++ {
    l := lnz & 1;
    nz_coeffs := 0;
    for x = 0; x < 4; x++ {
      ctx := l + (tnz & 1);
      nz := GetCoeffs(token_br, ac_proba, ctx, q.y1_mat, first, dst);
      l = (nz > first);
      tnz = (tnz >> 1) | (l << 7);
      nz_coeffs = NzCodeBits(nz_coeffs, nz, dst[0] != 0);
      dst += 16;
    }
    tnz >>= 4;
    lnz = (lnz >> 1) | (l << 7);
    non_zero_y = (non_zero_y << 8) | nz_coeffs;
  }
  out_t_nz = tnz;
  out_l_nz = lnz >> 4;

  for ch = 0; ch < 4; ch += 2 {
    nz_coeffs := 0;
    tnz = mb.nz >> (4 + ch);
    lnz = left_mb.nz >> (4 + ch);
    for y = 0; y < 2; y++ {
      l := lnz & 1;
      for x = 0; x < 2; x++ {
        ctx := l + (tnz & 1);
        nz := GetCoeffs(token_br, bands[2], ctx, q.uv_mat, 0, dst);
        l = (nz > 0);
        tnz = (tnz >> 1) | (l << 3);
        nz_coeffs = NzCodeBits(nz_coeffs, nz, dst[0] != 0);
        dst += 16;
      }
      tnz >>= 2;
      lnz = (lnz >> 1) | (l << 5);
    }
    // Note: we don't really need the per-4x4 details for U/V blocks.
    non_zero_uv |= nz_coeffs << (4 * ch);
    out_t_nz |= (tnz << 4) << ch;
    out_l_nz |= (lnz & 0xf0) << ch;
  }
  mb.nz = out_t_nz;
  left_mb.nz = out_l_nz;

  block.non_zero_y = non_zero_y;
  block.non_zero_uv = non_zero_uv;

  // We look at the mode-code of each block and check if some blocks have less
  // than three non-zero coeffs (code < 2). This is to afunc dithering flat and
  // empty blocks.
  block.dither = (non_zero_uv & 0xaaaa) ? 0 : q.dither;

  return !(non_zero_y | non_zero_uv);  // will be used for further optimization
}

// Main loop
// Decode one macroblock. Returns false if there is not enough data.
func VP8DecodeMB(/* const */ dec *VP8Decoder, /* const */ token_br *VP8BitReader) int {
  var left *VP8MB = dec.mb_info - 1;
  var mb *VP8MB = dec.mb_info + dec.mb_x;
  var block *VP8MBData = dec.mb_data + dec.mb_x;
  skip := tenary.If(dec.use_skip_proba, block.skip, 0);

  if (!skip) {
    skip = ParseResiduals(dec, mb, token_br);
  } else {
    left.nz = mb.nz = 0;
    if (!block.is_i4x4) {
      left.nz_dc = mb.nz_dc = 0;
    }
    block.non_zero_y = 0;
    block.non_zero_uv = 0;
    block.dither = 0;
  }

  if (dec.filter_type > 0) {  // store filter info
    var finfo *VP8FInfo = dec.f_info + dec.mb_x;
    *finfo = dec.fstrengths[block.segment][block.is_i4x4];
    finfo.f_inner |= !skip;
  }

  return !token_br.eof;
}

// To be called at the start of a new scanline, to initialize predictors.
func VP8InitScanline(/* const */ dec *VP8Decoder) {
  var left *VP8MB = dec.mb_info - 1;
  left.nz = 0;
  left.nz_dc = 0;
  stdlib.Memset(dec.intra_l, B_DC_PRED, sizeof(dec.intra_l));
  dec.mb_x = 0;
}

func ParseFrame(/* const */ dec *VP8Decoder, io *VP8Io) int {
  for dec.mb_y = 0; dec.mb_y < dec.br_mb_y; ++dec.mb_y {
    // Parse bitstream for this row.
    const token_br *VP8BitReader =
        &dec.parts[dec.mb_y & dec.num_parts_minus_one];
    if (!VP8ParseIntraModeRow(&dec.br, dec)) {
      return VP8SetError(dec, VP8_STATUS_NOT_ENOUGH_DATA, "Premature end-of-partition0 encountered.");
    }
    for ; dec.mb_x < dec.mb_w; ++dec.mb_x {
      if (!VP8DecodeMB(dec, token_br)) {
        return VP8SetError(dec, VP8_STATUS_NOT_ENOUGH_DATA, "Premature end-of-file encountered.");
      }
    }
    VP8InitScanline(dec);  // Prepare for next scanline

    // Reconstruct, filter and emit the row.
    if (!VP8ProcessRow(dec, io)) {
      return VP8SetError(dec, VP8_STATUS_USER_ABORT, "Output aborted.");
    }
  }
  if (dec.mt_method > 0) {
    if (!WebPGetWorkerInterface().Sync(&dec.worker)) { return 0; }
  }

  return 1;
}

// Main entry point
// Decode a picture. Will call VP8GetHeaders() if it wasn't done already.
// Returns false in case of error.
func VP8Decode(/* const */ dec *VP8Decoder, /* const */ io *VP8Io) int {
  ok := 0;
  if (dec == nil) {
    return 0;
  }
  if (io == nil) {
    return VP8SetError(dec, VP8_STATUS_INVALID_PARAM, "nil VP8Io parameter in VP8Decode().");
  }

  if (!dec.ready) {
    if (!VP8GetHeaders(dec, io)) {
      return 0;
    }
  }
  assert.Assert(dec.ready);

  // Finish setting up the decoding parameter. Will call io.setup().
  ok = (VP8EnterCritical(dec, io) == VP8_STATUS_OK);
  if (ok) {  // good to go.
    // Will allocate memory and prepare everything.
    if (ok) ok = VP8InitFrame(dec, io);

    // Main decoding loop
    if (ok) ok = ParseFrame(dec, io);

    // Exit.
    ok &= VP8ExitCritical(dec, io);
  }

  if (!ok) {
    VP8Clear(dec);
    return 0;
  }

  dec.ready = 0;
  return ok;
}

// Resets the decoder in its initial state, reclaiming memory.
// Not a mandatory call between calls to VP8Decode().
func VP8Clear(/* const */ dec *VP8Decoder) {
  if (dec == nil) {
    return;
  }
  WebPGetWorkerInterface().End(&dec.worker);
  dec.mem = nil;
  dec.mem_size = 0;
  stdlib.Memset(&dec.br, 0, sizeof(dec.br));
  dec.ready = 0;
}
