package webp

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//  Internal header for constants related to WebP file format.
//
// Author: Urvang (urvang@google.com)


// Create fourcc of the chunk from the chunk tag characters.
func MKFOURCC(a, b, c, d uint32) uint32 {
	return ((a) | (b) << 8 | (c) << 16 | (uint32)(d) << 24)
}

// VP8 related constants.
const VP8_SIGNATURE =0x9d012a             // Signature in VP8 data.
const VP8_MAX_PARTITION0_SIZE =(1 << 19)  // max size of mode partition
const VP8_MAX_PARTITION_SIZE =(1 << 24)   // max size for token partition
const VP8_FRAME_HEADER_SIZE =10  // Size of the frame header within VP8 data.

// VP8L related constants.
const VP8L_SIGNATURE_SIZE =1  // VP8L signature size.
const VP8L_MAGIC_BYTE =0x2f   // VP8L signature byte.
const VP8L_IMAGE_SIZE_BITS = 14                         // Number of bits used to store width and height.
const VP8L_VERSION_BITS =3  // 3 bits reserved for version.
const VP8L_VERSION =0       // version 0
const VP8L_FRAME_HEADER_SIZE =5  // Size of the VP8L frame header.

const MAX_PALETTE_SIZE =256
const MAX_CACHE_BITS =11
const HUFFMAN_CODES_PER_META_CODE =5
const ARGB_BLACK =0xff000000

const DEFAULT_CODE_LENGTH =8
const MAX_ALLOWED_CODE_LENGTH =15

const NUM_LITERAL_CODES =256
const NUM_LENGTH_CODES =24
const NUM_DISTANCE_CODES =40
const CODE_LENGTH_CODES =19

const MIN_HUFFMAN_BITS =2  // min number of Huffman bits
const NUM_HUFFMAN_BITS =3

// the maximum number of bits defining a transform is
// MIN_TRANSFORM_BITS + (1 << NUM_TRANSFORM_BITS) - 1
const MIN_TRANSFORM_BITS =2
const NUM_TRANSFORM_BITS =3

const TRANSFORM_PRESENT =\
  1  // The bit to be written when next data to be read is a transform.
const NUM_TRANSFORMS =4  // Maximum number of allowed transform in a bitstream.
type <FOO> int

const (
  PREDICTOR_TRANSFORM = 0, CROSS_COLOR_TRANSFORM = 1, SUBTRACT_GREEN_TRANSFORM = 2, COLOR_INDEXING_TRANSFORM = 3
} VP8LImageTransformType;

// Alpha related constants.
const ALPHA_HEADER_LEN =1
const ALPHA_NO_COMPRESSION =0
const ALPHA_LOSSLESS_COMPRESSION =1
const ALPHA_PREPROCESSED_LEVELS =1

// Mux related constants.
const TAG_SIZE =4           // Size of a chunk tag (e.g. "VP8L").
const CHUNK_SIZE_BYTES =4   // Size needed to store chunk's size.
const CHUNK_HEADER_SIZE =8  // Size of a chunk header.
const RIFF_HEADER_SIZE =12  // Size of the RIFF header ("RIFFnnnnWEBP").
const ANMF_CHUNK_SIZE =16   // Size of an ANMF chunk.
const ANIM_CHUNK_SIZE =6    // Size of an ANIM chunk.
const VP8X_CHUNK_SIZE =10   // Size of a VP8X chunk.

const MAX_CANVAS_SIZE =(1 << 24)      // 24-bit max for VP8X width/height.
const MAX_IMAGE_AREA =(uint64(1) << 32)    // 32-bit max for width x height.
const MAX_LOOP_COUNT =(1 << 16)       // maximum value for loop-count
const MAX_DURATION =(1 << 24)         // maximum duration
const MAX_POSITION_OFFSET =(1 << 24)  // maximum frame x/y offset

// Maximum chunk payload is such that adding the header and padding won't
// overflow a uint32.
const MAX_CHUNK_PAYLOAD =(~uint(0) - CHUNK_HEADER_SIZE - 1)

#endif  // WEBP_WEBP_FORMAT_CONSTANTS_H_
