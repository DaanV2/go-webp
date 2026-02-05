// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import (
	"github.com/daanv2/go-webp/pkg/libwebp/endian"
)

func WSWAP(x uint32) uint32 {
	return endian.HToLE32(x)
}
