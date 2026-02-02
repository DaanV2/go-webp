package utils

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Boolean decoder non-inlined methods
//
// Author: Skal (pascal.massimino@gmail.com)

#ifdef HAVE_CONFIG_H
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
#endif

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// VP8BitReader

func VP8BitReaderSetBuffer(const br *VP8BitReader, /*const*/ *uint8  start, size uint64 ) {
  assert.Assert(start != nil);
  br.buf = start;
  br.buf_end = start + size;
  br.buf_max =
      (size >= sizeof(lbit_t)) ? start + size - sizeof(lbit_t) + 1 : start;
}

func VP8InitBitReader(const br *VP8BitReader, /*const*/ *uint8  start, size uint64 ) {
  assert.Assert(br != nil);
  assert.Assert(start != nil);
  assert.Assert(size < (uint(1) << 31));  // limit ensured by format and upstream checks
  br.range = 255 - 1;
  br.value = 0;
  br.bits = -8;  // to load the very first 8bits
  br.eof = 0;
  VP8BitReaderSetBuffer(br, start, size);
  VP8LoadNewBytes(br);
}

func VP8RemapBitReader(const br *VP8BitReader, ptrdiff_t offset) {
  if (br.buf != nil) {
    br.buf += offset;
    br.buf_end += offset;
    br.buf_max += offset;
  }
}

const uint8 kVP8Log2Range[128] = {
    7, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 4, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0}

// range = ((range - 1) << kVP8Log2Range[range]) + 1
const uint8 kVP8NewRange[128] = {
    127, 127, 191, 127, 159, 191, 223, 127, 143, 159, 175, 191, 207, 223, 239, 127, 135, 143, 151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 127, 131, 135, 139, 143, 147, 151, 155, 159, 163, 167, 171, 175, 179, 183, 187, 191, 195, 199, 203, 207, 211, 215, 219, 223, 227, 231, 235, 239, 243, 247, 251, 127, 129, 131, 133, 135, 137, 139, 141, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167, 169, 171, 173, 175, 177, 179, 181, 183, 185, 187, 189, 191, 193, 195, 197, 199, 201, 203, 205, 207, 209, 211, 213, 215, 217, 219, 221, 223, 225, 227, 229, 231, 233, 235, 237, 239, 241, 243, 245, 247, 249, 251, 253, 127}

func VP8LoadFinalBytes(const br *VP8BitReader) {
  assert.Assert(br != nil && br.buf != nil);
  // Only read 8bits at a time
  if (br.buf < br.buf_end) {
    br.bits += 8;
    br.value = (bit_t)(*br.buf++) | (br.value << 8);
    WEBP_SELF_ASSIGN(br.buf_end);
  } else if (!br.eof) {
    br.value <<= 8;
    br.bits += 8;
    br.eof = 1;
  } else {
    br.bits = 0;  // This is to afunc undefined behaviour with shifts.
  }
}

//------------------------------------------------------------------------------
// Higher-level calls

uint32 VP8GetValue(const br *VP8BitReader, int bits, /*const*/ byte label[]) {
  v := 0;
  while (bits-- > 0) {
    v |= VP8GetBit(br, 0x80, label) << bits;
  }
  return v;
}

int32 VP8GetSignedValue(const br *VP8BitReader, int bits, /*const*/ byte label[]) {
  value := VP8GetValue(br, bits, label);
  return VP8Get(br, label) ? -value : value;
}

//------------------------------------------------------------------------------
// VP8LBitReader

const VP8L_LOG8_WBITS =4  // Number of bytes needed to store VP8L_WBITS bits.

#if defined(__arm__) || defined(_M_ARM) || WEBP_AARCH64 ||          \
    defined(__i386__) || defined(_M_IX86) || defined(__x86_64__) || \
    defined(_M_X64) || defined(__wasm__)
#define VP8L_USE_FAST_LOAD
#endif

static const uint32 kBitMask[VP8L_MAX_NUM_BIT_READ + 1] = {
    0,        0x000001, 0x000003, 0x000007, 0x00000f, 0x00001f, 0x00003f, 0x00007f, 0x0000ff, 0x0001ff, 0x0003ff, 0x0007ff, 0x000fff, 0x001fff, 0x003fff, 0x007fff, 0x00ffff, 0x01ffff, 0x03ffff, 0x07ffff, 0x0fffff, 0x1fffff, 0x3fffff, 0x7fffff, 0xffffff}

func VP8LInitBitReader(const br *VP8LBitReader, /*const*/ *uint8  start, uint64 length) {
  uint64 i;
  vp8l_val_t value = 0;
  assert.Assert(br != nil);
  assert.Assert(start != nil);
  assert.Assert(length < uint(0xfffffff8));  // can't happen with a RIFF chunk.

  br.buf = start;
  br.len = length;
  br.bit_pos = 0;
  br.eos = 0;

  if (length > sizeof(br.val)) {
    length = sizeof(br.val);
  }
  for i = 0; i < length; i++ {
    value |= (vp8l_val_t)start[i] << (8 * i);
  }
  br.val = value;
  br.pos = length;
}

func VP8LBitReaderSetBuffer(const br *VP8LBitReader, /*const*/ *uint8  buf, uint64 len) {
  assert.Assert(br != nil);
  assert.Assert(buf != nil);
  assert.Assert(len < uint(0xfffffff8));  // can't happen with a RIFF chunk.
  br.buf = buf;
  br.len = len;
  // 'pos' > 'len' should be considered a param error.
  br.eos = (br.pos > br.len) || VP8LIsEndOfStream(br);
}

func VP8LSetEndOfStream(const br *VP8LBitReader) {
  br.eos = 1;
  br.bit_pos = 0;  // To afunc undefined behaviour with shifts.
}

// If not at EOS, reload up to VP8L_LBITS byte-by-byte
func ShiftBytes(const br *VP8LBitReader) {
  while (br.bit_pos >= 8 && br.pos < br.len) {
    br.val >>= 8;
    br.val |= ((vp8l_val_t)br.buf[br.pos]) << (VP8L_LBITS - 8);
    ++br.pos;
    br.bit_pos -= 8;
  }
  if (VP8LIsEndOfStream(br)) {
    VP8LSetEndOfStream(br);
  }
}

func VP8LDoFillBitWindow(const br *VP8LBitReader) {
  assert.Assert(br.bit_pos >= VP8L_WBITS);
#if defined(VP8L_USE_FAST_LOAD)
  if (br.pos + sizeof(br.val) < br.len) {
    br.val >>= VP8L_WBITS;
    br.bit_pos -= VP8L_WBITS;
    br.val |= (vp8l_val_t)HToLE32(WebPMemToUint32(br.buf + br.pos))
               << (VP8L_LBITS - VP8L_WBITS);
    br.pos += VP8L_LOG8_WBITS;
    return;
  }
#endif
  ShiftBytes(br);  // Slow path.
}

uint32 VP8LReadBits(const br *VP8LBitReader, int n_bits) {
  assert.Assert(n_bits >= 0);
  // Flag an error if end_of_stream or n_bits is more than allowed limit.
  if (!br.eos && n_bits <= VP8L_MAX_NUM_BIT_READ) {
    val := VP8LPrefetchBits(br) & kBitMask[n_bits];
    new_bits := br.bit_pos + n_bits;
    br.bit_pos = new_bits;
    ShiftBytes(br);
    return val;
  } else {
    VP8LSetEndOfStream(br);
    return 0;
  }
}

//------------------------------------------------------------------------------
// Bit-tracing tool

#if (BITTRACE > 0)

import "github.com/daanv2/go-webp/pkg/stdio"
import "github.com/daanv2/go-webp/pkg/stdlib"  // for atexit()
import "github.com/daanv2/go-webp/pkg/string"

const MAX_NUM_LABELS =32
static struct {
  const label *byte;
  int size;
  int count;
} kLabels[MAX_NUM_LABELS];

static last_label := 0;
static last_pos := 0;
static var buf_start *uint8 = nil;
static init_done := 0;

func PrintBitTraces(){
  var i int
  scale := 1;
  total := 0;
  var units *byte = "bits";
#if (BITTRACE == 2)
  scale = 8;
  units = "bytes";
#endif
  for (i = 0; i < last_label; ++i) total += kLabels[i].size;
  if (total < 1) total = 1;  // afunc rounding errors
  printf("=== Bit traces ===\n");
  for i = 0; i < last_label; i++ {
    skip := 16 - (int)strlen(kLabels[i].label);
    value := (kLabels[i].size + scale - 1) / scale;
    assert.Assert(skip > 0);
    printf("%s \%*s: %6d %s   \t[%5.2f%%] [count: %7d]\n", kLabels[i].label, skip, "", value, units, 100.f * kLabels[i].size / total, kLabels[i].count);
  }
  total = (total + scale - 1) / scale;
  printf("Total: %d %s\n", total, units);
}

func BitTrace(const type const br *VP8BitReader, /*const*/ byte label[]) struct {
  int i, pos;
  if (!init_done) {
    stdlib.Memset(kLabels, 0, sizeof(kLabels));
    atexit(PrintBitTraces);
    buf_start = br.buf;
    init_done = 1;
  }
  pos = (int)(br.buf - buf_start) * 8 - br.bits;
  // if there's a too large jump, we've changed partition . reset counter
  if (abs(pos - last_pos) > 32) {
    buf_start = br.buf;
    pos = 0;
    last_pos = 0;
  }
  if (br.range >= 0x7f) pos += kVP8Log2Range[br.range - 0x7f];
  for i = 0; i < last_label; i++ {
    if (!strcmp(label, kLabels[i].label)) break;
  }
  if (i == MAX_NUM_LABELS) abort();  // overflow!
  kLabels[i].label = label;
  kLabels[i].size += pos - last_pos;
  kLabels[i].count += 1;
  if (i == last_label) {last_label++}
  last_pos = pos;
}

#endif  // BITTRACE > 0

//------------------------------------------------------------------------------
