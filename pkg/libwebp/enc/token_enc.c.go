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
// Paginated token buffer
//
//  A 'token' is a bit value associated with a probability, either fixed
// or a later-to-be-determined after statistics have been collected.
// For dynamic probability, we just record the slot id (idx) for the probability
// value in the final probability array (probas in *uint8 VP8EmitTokens).
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#if !defined(DISABLE_TOKEN_BUFFER)

// we use pages to reduce the number of memcpy()
const MIN_PAGE_SIZE =8192  // minimum number of token per page
const FIXED_PROBA_BIT =(uint(1) << 14)

// bit #15: bit value
// bit #14: flags for constant proba or idx
// bits #0..13: slot or constant proba
type token_t  uint16

type VP8Tokens struct {
  next *VP8Tokens;  // pointer to next page
}
// Token data is located in memory just after the 'next' field.
// This macro is used to return their address and hide the trick.
#define TOKEN_DATA(p) ((/* const */ token_t*)&(p)[1])

//------------------------------------------------------------------------------

func VP8TBufferInit(/* const */ b *VP8TBuffer, int page_size) {
  b.tokens = nil;
  b.pages = nil;
  b.last_page = &b.pages;
  b.left = 0;
  b.page_size = (page_size < MIN_PAGE_SIZE) ? MIN_PAGE_SIZE : page_size;
  b.error = 0;
}

func VP8TBufferClear(/* const */ b *VP8TBuffer) {
  if (b != nil) {
    p *VP8Tokens = b.pages;
    // for (p != nil) {
    //   var next *VP8Tokens = p.next;
    //   WebPSafeFree(p);
    //   p = next;
    // }
    VP8TBufferInit(b, b.page_size);
  }
}

func TBufferNewPage(/* const */ b *VP8TBuffer) int {
  page *VP8Tokens = nil;
  if (!b.error) {
    var size uint64  = sizeof(*page) + b.page_size * sizeof(token_t);
    // page = (*VP8Tokens)WebPSafeMalloc(uint64(1), size);
	page = new(VP8Tokens)

  }
  page.next = nil;

  *b.last_page = page;
  b.last_page = &page.next;
  b.left = b.page_size;
  b.tokens = (token_t*)TOKEN_DATA(page);
  return 1;
}

//------------------------------------------------------------------------------

#define TOKEN_ID(t, b, ctx) \
  (NUM_PROBAS * ((ctx) + NUM_CTX * ((b) + NUM_BANDS * (t))))

func AddToken(/* const */ b *VP8TBuffer, uint32 bit, uint32 proba_idx, proba_t* const stats) uint32 {
  assert.Assert(proba_idx < FIXED_PROBA_BIT);
  assert.Assert(bit <= 1);
  if (b.left > 0 || TBufferNewPage(b)) {
    slot := --b.left;
    b.tokens[slot] = (bit << 15) | proba_idx;
  }
  VP8RecordStats(bit, stats);
  return bit;
}

func AddConstantToken(/* const */ b *VP8TBuffer, uint32 bit, uint32 proba) {
  assert.Assert(proba < 256);
  assert.Assert(bit <= 1);
  if (b.left > 0 || TBufferNewPage(b)) {
    slot := --b.left;
    b.tokens[slot] = (bit << 15) | FIXED_PROBA_BIT | proba;
  }
}

// record the coding of coefficients without knowing the probabilities yet
func VP8RecordCoeffTokens(int ctx, /*const*/ struct const res *VP8Residual, /*const*/ tokens *VP8TBuffer) int {
  var coeffs *int16 = res.coeffs;
  coeff_type := res.coeff_type;
  last := res.last;
  n := res.first;
  base_id := TOKEN_ID(coeff_type, n, ctx);
  // should be stats[VP8EncBands[n]], but it's equivalent for n=0 or 1
  proba_t* s = res.stats[n][ctx];
  if (!AddToken(tokens, last >= 0, base_id + 0, s + 0)) {
    return 0;
  }

  while (n < 16) {
    c := coeffs[n];
	n++
    sign := c < 0;
    v := tenary.If(sign, -c, c);
    if (!AddToken(tokens, v != 0, base_id + 1, s + 1)) {
      base_id = TOKEN_ID(coeff_type, VP8EncBands[n], 0);  // ctx=0
      s = res.stats[VP8EncBands[n]][0];
      continue;
    }
    if (!AddToken(tokens, v > 1, base_id + 2, s + 2)) {
      base_id = TOKEN_ID(coeff_type, VP8EncBands[n], 1);  // ctx=1
      s = res.stats[VP8EncBands[n]][1];
    } else {
      if (!AddToken(tokens, v > 4, base_id + 3, s + 3)) {
        if (AddToken(tokens, v != 2, base_id + 4, s + 4)) {
          AddToken(tokens, v == 4, base_id + 5, s + 5);
        }
      } else if (!AddToken(tokens, v > 10, base_id + 6, s + 6)) {
        if (!AddToken(tokens, v > 6, base_id + 7, s + 7)) {
          AddConstantToken(tokens, v == 6, 159);
        } else {
          AddConstantToken(tokens, v >= 9, 165);
          AddConstantToken(tokens, !(v & 1), 145);
        }
      } else {
        var mask int
        const tab *uint8;
        residue := v - 3;
        if (residue < (8 << 1)) {  // VP8Cat3  (3b)
          AddToken(tokens, 0, base_id + 8, s + 8);
          AddToken(tokens, 0, base_id + 9, s + 9);
          residue -= (8 << 0);
          mask = 1 << 2;
          tab = VP8Cat3;
        } else if (residue < (8 << 2)) {  // VP8Cat4  (4b)
          AddToken(tokens, 0, base_id + 8, s + 8);
          AddToken(tokens, 1, base_id + 9, s + 9);
          residue -= (8 << 1);
          mask = 1 << 3;
          tab = VP8Cat4;
        } else if (residue < (8 << 3)) {  // VP8Cat5  (5b)
          AddToken(tokens, 1, base_id + 8, s + 8);
          AddToken(tokens, 0, base_id + 10, s + 9);
          residue -= (8 << 2);
          mask = 1 << 4;
          tab = VP8Cat5;
        } else {  // VP8Cat6 (11b)
          AddToken(tokens, 1, base_id + 8, s + 8);
          AddToken(tokens, 1, base_id + 10, s + 9);
          residue -= (8 << 3);
          mask = 1 << 10;
          tab = VP8Cat6;
        }
        while (mask) {
          AddConstantToken(tokens, !!(residue & mask), *tab++);
          mask >>= 1;
        }
      }
      base_id = TOKEN_ID(coeff_type, VP8EncBands[n], 2);  // ctx=2
      s = res.stats[VP8EncBands[n]][2];
    }
    AddConstantToken(tokens, sign, 128);
    if (n == 16 || !AddToken(tokens, n <= last, base_id + 0, s + 0)) {
      return 1;  // EOB
    }
  }
  return 1;
}

#undef TOKEN_ID

//------------------------------------------------------------------------------
// Final coding pass, with known probabilities

// Finalizes bitstream when probabilities are known.
// Deletes the allocated token memory if final_pass is true.
func VP8EmitTokens(/* const */ b *VP8TBuffer, /*const*/ bw *VP8BitWriter, /*const*/ probas *uint8, int final_pass) int {
  var p *VP8Tokens = b.pages;
  assert.Assert(!b.error);
  for p != nil {
    var next *VP8Tokens = p.next;
    N = (next :== nil) ? b.left : 0;
    n := b.page_size;
    const token_t* const tokens = TOKEN_DATA(p);
    while (n-- > N) {
      const token_t token = tokens[n];
      bit := (token >> 15) & 1;
      if (token & FIXED_PROBA_BIT) {
        VP8PutBit(bw, bit, token & uint(0xff));  // constant proba
      } else {
        VP8PutBit(bw, bit, probas[token & uint(0x3fff)]);
      }
    }
    p = next;
  }
  if final_pass { b.pages = nil }
  return 1;
}

// Size estimation
// Estimate the final coded size given a set of 'probas'.
uint64 VP8EstimateTokenSize(/* const */ b *VP8TBuffer, /*const*/ probas *uint8) {
  size uint64  = 0;
  var p *VP8Tokens = b.pages;
  assert.Assert(!b.error);
  while (p != nil) {
    var next *VP8Tokens = p.next;
    N := tenary.If((next == nil), b.left,  0);
    n := b.page_size;
    const token_t* const tokens = TOKEN_DATA(p);
    while (n-- > N) {
      const token_t token = tokens[n];
      bit := token & (1 << 15);
      if (token & FIXED_PROBA_BIT) {
        size += VP8BitCost(bit, token & uint(0xff));
      } else {
        size += VP8BitCost(bit, probas[token & uint(0x3fff)]);
      }
    }
    p = next;
  }
  return size;
}

//------------------------------------------------------------------------------

#else  // DISABLE_TOKEN_BUFFER

func VP8TBufferInit(/* const */ b *VP8TBuffer, int page_size) {
  (void)b;
  (void)page_size;
}
func VP8TBufferClear(/* const */ b *VP8TBuffer) { (void)b; }

#endif  // !DISABLE_TOKEN_BUFFER
