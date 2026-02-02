// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package demux

import "github.com/daanv2/go-webp/pkg/constants"

// Channel extraction from a uint32 representation of a uint8 RGBA/BGRA
// buffer.
func CHANNEL_SHIFT(i int) int {
	if constants.WORDS_BIGENDIAN {
		return (24 - (i) * 8)
	}

	return((i) * 8)
}