package vp8

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

// maximum value of 'transform_bits' in VP8LEncoder.
const MAX_TRANSFORM_BITS = (MIN_TRANSFORM_BITS + (1 << NUM_TRANSFORM_BITS) - 1)


// WebPGetDecoderVersion Return the decoder's version number, packed in hexadecimal using 8bits for
// each of major/minor/revision. E.g: v2.5.7 is 0x020507.
func WebPGetDecoderVersion() int {
  return (DEC_MAJ_VERSION << 16) | (DEC_MIN_VERSION << 8) | DEC_REV_VERSION;
}