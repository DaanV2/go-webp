package constants

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
	// VP8 related constants.
	VP8_SIGNATURE           = 0x9d012a  // Signature in VP8 data.
	VP8_MAX_PARTITION0_SIZE = (1 << 19) // max size of mode partition
	VP8_MAX_PARTITION_SIZE  = (1 << 24) // max size for token partition
	VP8_FRAME_HEADER_SIZE   = 10        // Size of the frame header within VP8 data.

	// VP8L related constants.
	VP8L_SIGNATURE_SIZE    = 1    // VP8L signature size.
	VP8L_MAGIC_BYTE        = 0x2f // VP8L signature byte.
	VP8L_IMAGE_SIZE_BITS   = 14   // Number of bits used to store width and height.
	VP8L_VERSION_BITS      = 3    // 3 bits reserved for version.
	VP8L_VERSION           = 0    // version 0
	VP8L_FRAME_HEADER_SIZE = 5    // Size of the VP8L frame header.

	MAX_PALETTE_SIZE            = 256
	MAX_CACHE_BITS              = 11
	HUFFMAN_CODES_PER_META_CODE = 5
	ARGB_BLACK                  = 0xff000000

	DEFAULT_CODE_LENGTH     = 8
	MAX_ALLOWED_CODE_LENGTH = 15

	NUM_LITERAL_CODES  = 256
	NUM_LENGTH_CODES   = 24
	NUM_DISTANCE_CODES = 40
	CODE_LENGTH_CODES  = 19

	MIN_HUFFMAN_BITS = 2 // min number of Huffman bits
	NUM_HUFFMAN_BITS = 3

	// the maximum number of bits defining a transform is
	// MIN_TRANSFORM_BITS + (1 << NUM_TRANSFORM_BITS) - 1
	MIN_TRANSFORM_BITS = 2
	NUM_TRANSFORM_BITS = 3

	TRANSFORM_PRESENT = 1 // The bit to be written when next data to be read is a transform.
	NUM_TRANSFORMS    = 4 // Maximum number of allowed transform in a bitstream.

	// Alpha related constants.
	ALPHA_HEADER_LEN           = 1
	ALPHA_NO_COMPRESSION       = 0
	ALPHA_LOSSLESS_COMPRESSION = 1
	ALPHA_PREPROCESSED_LEVELS  = 1

	// Mux related constants.
	TAG_SIZE          = 4  // Size of a chunk tag (e.g. "VP8L").
	CHUNK_SIZE_BYTES  = 4  // Size needed to store chunk's size.
	CHUNK_HEADER_SIZE = 8  // Size of a chunk header.
	RIFF_HEADER_SIZE  = 12 // Size of the RIFF header ("RIFFnnnnWEBP").
	ANMF_CHUNK_SIZE   = 16 // Size of an ANMF chunk.
	ANIM_CHUNK_SIZE   = 6  // Size of an ANIM chunk.
	VP8X_CHUNK_SIZE   = 10 // Size of a VP8X chunk.

	MAX_CANVAS_SIZE     = (1 << 24)         // 24-bit max for VP8X width/height.
	MAX_IMAGE_AREA      = (uint64(1) << 32) // 32-bit max for width x height.
	MAX_LOOP_COUNT      = (1 << 16)         // maximum value for loop-count
	MAX_DURATION        = (1 << 24)         // maximum duration
	MAX_POSITION_OFFSET = (1 << 24)         // maximum frame x/y offset

	// Maximum chunk payload is such that adding the header and padding won't
	// overflow a uint32.
	MAX_CHUNK_PAYLOAD = (^uint(0) - CHUNK_HEADER_SIZE - 1)
)
