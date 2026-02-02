// Copyright 2015 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package demux

const (
	NUM_CHANNELS = 4

	DMUX_MAJ_VERSION = 1
	DMUX_MIN_VERSION = 6
	DMUX_REV_VERSION = 0
)

var kMasterChunks []ChunkParser = []ChunkParser{
    {[4]uint8{'V', 'P', '8', ' '}, ParseSingleImage, IsValidSimpleFormat},
    {[4]uint8{'V', 'P', '8', 'L'}, ParseSingleImage, IsValidSimpleFormat},
    {[4]uint8{'V', 'P', '8', 'X'}, ParseVP8X, IsValidExtendedFormat},
    {[4]uint8{'0', '0', '0', '0'}, nil, nil},
}

func WebPGetDemuxVersion() int {
  return (DMUX_MAJ_VERSION << 16) | (DMUX_MIN_VERSION << 8) | DMUX_REV_VERSION
}