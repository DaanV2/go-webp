package webp

import "github.com/daanv2/go-webp/pkg/constants"

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

// Create fourcc of the chunk from the chunk tag characters.
func MKFOURCC(a, b, c, d uint32) uint32 {
	return ((a) | (b)<<8 | (c)<<16 | (uint32)(d)<<24)
}

const (
	//go:fix inline
	VP8_SIGNATURE = constants.VP8_SIGNATURE
	//go:fix inline
	VP8_MAX_PARTITION0_SIZE = constants.VP8_MAX_PARTITION0_SIZE
	//go:fix inline
	VP8_MAX_PARTITION_SIZE = constants.VP8_MAX_PARTITION_SIZE
	//go:fix inline
	VP8_FRAME_HEADER_SIZE = constants.VP8_FRAME_HEADER_SIZE
	//go:fix inline
	VP8L_SIGNATURE_SIZE = constants.VP8L_SIGNATURE_SIZE
	//go:fix inline
	VP8L_MAGIC_BYTE = constants.VP8L_MAGIC_BYTE
	//go:fix inline
	VP8L_IMAGE_SIZE_BITS = constants.VP8L_IMAGE_SIZE_BITS
	//go:fix inline
	VP8L_VERSION_BITS = constants.VP8L_VERSION_BITS
	//go:fix inline
	VP8L_VERSION = constants.VP8L_VERSION
	//go:fix inline
	VP8L_FRAME_HEADER_SIZE = constants.VP8L_FRAME_HEADER_SIZE
	//go:fix inline
	MAX_PALETTE_SIZE = constants.MAX_PALETTE_SIZE
	//go:fix inline
	MAX_CACHE_BITS = constants.MAX_CACHE_BITS
	//go:fix inline
	HUFFMAN_CODES_PER_META_CODE = constants.HUFFMAN_CODES_PER_META_CODE
	//go:fix inline
	ARGB_BLACK = constants.ARGB_BLACK
	//go:fix inline
	DEFAULT_CODE_LENGTH = constants.DEFAULT_CODE_LENGTH
	//go:fix inline
	MAX_ALLOWED_CODE_LENGTH = constants.MAX_ALLOWED_CODE_LENGTH
	//go:fix inline
	NUM_LITERAL_CODES = constants.NUM_LITERAL_CODES
	//go:fix inline
	NUM_LENGTH_CODES = constants.NUM_LENGTH_CODES
	//go:fix inline
	NUM_DISTANCE_CODES = constants.NUM_DISTANCE_CODES
	//go:fix inline
	CODE_LENGTH_CODES = constants.CODE_LENGTH_CODES
	//go:fix inline
	MIN_HUFFMAN_BITS = constants.MIN_HUFFMAN_BITS
	//go:fix inline
	NUM_HUFFMAN_BITS = constants.NUM_HUFFMAN_BITS
	//go:fix inline
	MIN_TRANSFORM_BITS = constants.MIN_TRANSFORM_BITS
	//go:fix inline
	NUM_TRANSFORM_BITS = constants.NUM_TRANSFORM_BITS
	//go:fix inline
	TRANSFORM_PRESENT = constants.TRANSFORM_PRESENT
	//go:fix inline
	NUM_TRANSFORMS = constants.NUM_TRANSFORMS
	//go:fix inline
	ALPHA_HEADER_LEN = constants.ALPHA_HEADER_LEN
	//go:fix inline
	ALPHA_NO_COMPRESSION = constants.ALPHA_NO_COMPRESSION
	//go:fix inline
	ALPHA_LOSSLESS_COMPRESSION = constants.ALPHA_LOSSLESS_COMPRESSION
	//go:fix inline
	ALPHA_PREPROCESSED_LEVELS = constants.ALPHA_PREPROCESSED_LEVELS
	//go:fix inline
	TAG_SIZE = constants.TAG_SIZE
	//go:fix inline
	CHUNK_SIZE_BYTES = constants.CHUNK_SIZE_BYTES
	//go:fix inline
	CHUNK_HEADER_SIZE = constants.CHUNK_HEADER_SIZE
	//go:fix inline
	RIFF_HEADER_SIZE = constants.RIFF_HEADER_SIZE
	//go:fix inline
	ANMF_CHUNK_SIZE = constants.ANMF_CHUNK_SIZE
	//go:fix inline
	ANIM_CHUNK_SIZE = constants.ANIM_CHUNK_SIZE
	//go:fix inline
	VP8X_CHUNK_SIZE = constants.VP8X_CHUNK_SIZE
	//go:fix inline
	MAX_CANVAS_SIZE = constants.MAX_CANVAS_SIZE
	//go:fix inline
	MAX_IMAGE_AREA = constants.MAX_IMAGE_AREA
	//go:fix inline
	MAX_LOOP_COUNT = constants.MAX_LOOP_COUNT
	//go:fix inline
	MAX_DURATION = constants.MAX_DURATION
	//go:fix inline
	MAX_POSITION_OFFSET = constants.MAX_POSITION_OFFSET
	//go:fix inline
	MAX_CHUNK_PAYLOAD = constants.MAX_CHUNK_PAYLOAD
)
