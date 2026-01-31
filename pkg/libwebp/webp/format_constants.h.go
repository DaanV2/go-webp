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
	// Deprecated: use [constants.VP8_SIGNATURE]
	VP8_SIGNATURE = constants.VP8_SIGNATURE
	// Deprecated: use [constants.VP8_MAX_PARTITION0_SIZE]
	VP8_MAX_PARTITION0_SIZE = constants.VP8_MAX_PARTITION0_SIZE
	// Deprecated: use [constants.VP8_MAX_PARTITION_SIZE]
	VP8_MAX_PARTITION_SIZE = constants.VP8_MAX_PARTITION_SIZE
	// Deprecated: use [constants.VP8_FRAME_HEADER_SIZE]
	VP8_FRAME_HEADER_SIZE = constants.VP8_FRAME_HEADER_SIZE
	// Deprecated: use [constants.VP8L_SIGNATURE_SIZE]
	VP8L_SIGNATURE_SIZE = constants.VP8L_SIGNATURE_SIZE
	// Deprecated: use [constants.VP8L_MAGIC_BYTE]
	VP8L_MAGIC_BYTE = constants.VP8L_MAGIC_BYTE
	// Deprecated: use [constants.VP8L_IMAGE_SIZE_BITS]
	VP8L_IMAGE_SIZE_BITS = constants.VP8L_IMAGE_SIZE_BITS
	// Deprecated: use [constants.VP8L_VERSION_BITS]
	VP8L_VERSION_BITS = constants.VP8L_VERSION_BITS
	// Deprecated: use [constants.VP8L_VERSION]
	VP8L_VERSION = constants.VP8L_VERSION
	// Deprecated: use [constants.VP8L_FRAME_HEADER_SIZE]
	VP8L_FRAME_HEADER_SIZE = constants.VP8L_FRAME_HEADER_SIZE
	// Deprecated: use [constants.MAX_PALETTE_SIZE]
	MAX_PALETTE_SIZE = constants.MAX_PALETTE_SIZE
	// Deprecated: use [constants.MAX_CACHE_BITS]
	MAX_CACHE_BITS = constants.MAX_CACHE_BITS
	// Deprecated: use [constants.HUFFMAN_CODES_PER_META_CODE]
	HUFFMAN_CODES_PER_META_CODE = constants.HUFFMAN_CODES_PER_META_CODE
	// Deprecated: use [constants.ARGB_BLACK]
	ARGB_BLACK = constants.ARGB_BLACK
	// Deprecated: use [constants.DEFAULT_CODE_LENGTH]
	DEFAULT_CODE_LENGTH = constants.DEFAULT_CODE_LENGTH
	// Deprecated: use [constants.MAX_ALLOWED_CODE_LENGTH]
	MAX_ALLOWED_CODE_LENGTH = constants.MAX_ALLOWED_CODE_LENGTH
	// Deprecated: use [constants.NUM_LITERAL_CODES]
	NUM_LITERAL_CODES = constants.NUM_LITERAL_CODES
	// Deprecated: use [constants.NUM_LENGTH_CODES]
	NUM_LENGTH_CODES = constants.NUM_LENGTH_CODES
	// Deprecated: use [constants.NUM_DISTANCE_CODES]
	NUM_DISTANCE_CODES = constants.NUM_DISTANCE_CODES
	// Deprecated: use [constants.CODE_LENGTH_CODES]
	CODE_LENGTH_CODES = constants.CODE_LENGTH_CODES
	// Deprecated: use [constants.MIN_HUFFMAN_BITS]
	MIN_HUFFMAN_BITS = constants.MIN_HUFFMAN_BITS
	// Deprecated: use [constants.NUM_HUFFMAN_BITS]
	NUM_HUFFMAN_BITS = constants.NUM_HUFFMAN_BITS
	// Deprecated: use [constants.MIN_TRANSFORM_BITS]
	MIN_TRANSFORM_BITS = constants.MIN_TRANSFORM_BITS
	// Deprecated: use [constants.NUM_TRANSFORM_BITS]
	NUM_TRANSFORM_BITS = constants.NUM_TRANSFORM_BITS
	// Deprecated: use [constants.TRANSFORM_PRESENT]
	TRANSFORM_PRESENT = constants.TRANSFORM_PRESENT
	// Deprecated: use [constants.NUM_TRANSFORMS]
	NUM_TRANSFORMS = constants.NUM_TRANSFORMS
	// Deprecated: use [constants.ALPHA_HEADER_LEN]
	ALPHA_HEADER_LEN = constants.ALPHA_HEADER_LEN
	// Deprecated: use [constants.ALPHA_NO_COMPRESSION]
	ALPHA_NO_COMPRESSION = constants.ALPHA_NO_COMPRESSION
	// Deprecated: use [constants.ALPHA_LOSSLESS_COMPRESSION]
	ALPHA_LOSSLESS_COMPRESSION = constants.ALPHA_LOSSLESS_COMPRESSION
	// Deprecated: use [constants.ALPHA_PREPROCESSED_LEVELS]
	ALPHA_PREPROCESSED_LEVELS = constants.ALPHA_PREPROCESSED_LEVELS
	// Deprecated: use [constants.TAG_SIZE]
	TAG_SIZE = constants.TAG_SIZE
	// Deprecated: use [constants.CHUNK_SIZE_BYTES]
	CHUNK_SIZE_BYTES = constants.CHUNK_SIZE_BYTES
	// Deprecated: use [constants.CHUNK_HEADER_SIZE]
	CHUNK_HEADER_SIZE = constants.CHUNK_HEADER_SIZE
	// Deprecated: use [constants.RIFF_HEADER_SIZE]
	RIFF_HEADER_SIZE = constants.RIFF_HEADER_SIZE
	// Deprecated: use [constants.ANMF_CHUNK_SIZE]
	ANMF_CHUNK_SIZE = constants.ANMF_CHUNK_SIZE
	// Deprecated: use [constants.ANIM_CHUNK_SIZE]
	ANIM_CHUNK_SIZE = constants.ANIM_CHUNK_SIZE
	// Deprecated: use [constants.VP8X_CHUNK_SIZE]
	VP8X_CHUNK_SIZE = constants.VP8X_CHUNK_SIZE
	// Deprecated: use [constants.MAX_CANVAS_SIZE]
	MAX_CANVAS_SIZE = constants.MAX_CANVAS_SIZE
	// Deprecated: use [constants.MAX_IMAGE_AREA]
	MAX_IMAGE_AREA = constants.MAX_IMAGE_AREA
	// Deprecated: use [constants.MAX_LOOP_COUNT]
	MAX_LOOP_COUNT = constants.MAX_LOOP_COUNT
	// Deprecated: use [constants.MAX_DURATION]
	MAX_DURATION = constants.MAX_DURATION
	// Deprecated: use [constants.MAX_POSITION_OFFSET]
	MAX_POSITION_OFFSET = constants.MAX_POSITION_OFFSET
	// Deprecated: use [constants.MAX_CHUNK_PAYLOAD]
	MAX_CHUNK_PAYLOAD = constants.MAX_CHUNK_PAYLOAD
)
