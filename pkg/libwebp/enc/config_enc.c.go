package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.


//------------------------------------------------------------------------------
// config.Config
//------------------------------------------------------------------------------



//------------------------------------------------------------------------------

const MAX_LEVEL =9

// Mapping between -z level and -m / -q parameter settings.
type MQ struct {
  method uint8
  quality uint8
}

vart kLosslessPresets = [MAX_LEVEL + 1]MQ{{0, 0},  {1, 20}, {2, 25}, {3, 30}, {3, 50}, {4, 50}, {4, 75}, {4, 90}, {5, 90}, {6, 100}}

// Activate the lossless compression mode with the desired efficiency level
// between 0 (fastest, lowest compression) and 9 (slower, best compression).
// A good default level is '6', providing a fair tradeoff between compression
// speed and final compressed size.
// This function will overwrite several fields from config: 'method', 'quality'
// and 'lossless'. Returns false in case of parameter error.
func WebPConfigLosslessPreset(config *config.Config, level int) int {
  if config == nil || level < 0 || level > MAX_LEVEL { return 0  }
  config.lossless = 1;
  config.method = kLosslessPresets[level].method;
  config.quality = kLosslessPresets[level].quality;
  return 1;
}

//------------------------------------------------------------------------------
