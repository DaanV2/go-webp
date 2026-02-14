package utils

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Alpha plane quantization utility
//
// Author:  Vikas Arora (vikasa@google.com)


import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"



// Replace the input 'data' of size 'width'x'height' with 'num-levels'
// quantized values. If not nil, 'sse' will contain the sum of squared error.
// Valid range for 'num_levels' is [2, 256].
// Returns false in case of error (data is nil, or parameters are invalid).
int QuantizeLevels(/* const */  *uint8((uint64)height *width) data, width, height int, num_levels int, /*const*/ sse *uint64)



#endif  // WEBP_UTILS_QUANT_LEVELS_UTILS_H_
