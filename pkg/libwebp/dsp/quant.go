// Copyright 2018 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)



func IsFlat(/* const */ levels []int16, num_blocks int, thresh int) int {
  score := 0
  for ;num_blocks > 0; num_blocks-- {  // TODO(skal): refine positional scoring?
    var i int
    for i = 1; i < 16; i++ {  // omit DC, we're only interested in AC
      score += tenary.If(levels[i] != 0, 1, 0)
      if score > thresh { return 0  }
    }
    levels = levels[16:]
  }
  return 1
}

func IsFlatSource16(/* const */ src []uint8) int {
  v := uint(src[0]) * uint(0x01010101)
  var i int
  for i = 0; i < 16; i++ {
    if (stdlib.MemCmp(src + 0, &v, 4) || stdlib.MemCmp(src + 4, &v, 4) ||
        stdlib.MemCmp(src + 8, &v, 4) || stdlib.MemCmp(src + 12, &v, 4)) {
      return 0
    }
    src += BPS
  }
  return 1
}

