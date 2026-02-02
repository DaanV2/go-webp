// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package utils

import "github.com/daanv2/go-webp/pkg/assert"

// Read 16, 24 or 32 bits stored in little-endian order.
func GetLE16(/* const */ data []uint8/* (2) */ )  int {
  return int(data[0] << 0) | (data[1] << 8);
}

func GetLE24(/* const */  data []uint8/* (3) */ )  int {
  return GetLE16(data) | (data[2] << 16);
}

func GetLE32(/* const */ data []uint8/* (4) */) uint32 {
  return uint32(GetLE16(data)) | uint32(GetLE16(data[2:]) << 16);
}

// Store 16, 24 or 32 bits in little-endian order.
func PutLE16(/* const */ data []uint8/* (2) */, val int) {
  assert.Assert(val < (1 << 16));
  data[0] = (val >> 0) & 0xff
  data[1] = (val >> 8) & 0xff
}

func PutLE24(/* const */ data []uint8/* (3) */, val int) {
  assert.Assert(val < (1 << 24))
  PutLE16(data, val & 0xffff)
  data[2] = (val >> 16) & 0xff
}

func PutLE32(/* const */ data []uint8/* (4) */, val uint32 ) {
  PutLE16(data, (int)(val & 0xffff))
  PutLE16(data[2:], (int)(val >> 16))
}
