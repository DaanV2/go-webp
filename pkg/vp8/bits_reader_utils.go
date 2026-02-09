// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

// If not at EOS, reload up to VP8L_LBITS byte-by-byte
func ShiftBytes(/* const */ br *VP8LBitReader) {
  for br.bit_pos >= 8 && br.pos < br.len {
    br.val >>= 8
    br.val |= (vp8l_val_t(br.buf[br.pos])) << (VP8L_LBITS - 8)
    br.pos++
    br.bit_pos -= 8
  }
  if VP8LIsEndOfStream(br) {
    VP8LSetEndOfStream(br)
  }
}

