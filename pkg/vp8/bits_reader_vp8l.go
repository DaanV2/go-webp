// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/endian"
)

type VP8LBitReader struct {
	val     vp8l_val_t // pre-fetched bits
	buf     []uint8    // input byte buffer
	len     uint64     // buffer length
	pos     uint64     // byte position in buf
	bit_pos int        // current bit-reading position in val
	eos     bool       // true if a bit was read past the end of buffer
}

func VP8LInitBitReader( /* const */ br *VP8LBitReader /*const*/, start []uint8, length uint64) {
	var i uint64
	var value vp8l_val_t = 0
	assert.Assert(br != nil)
	assert.Assert(start != nil)
	assert.Assert(length < 0xfffffff8) // can't happen with a RIFF chunk.

	br.buf = start
	br.len = length
	br.bit_pos = 0
	br.eos = false

	// C: if length > sizeof(br.val) {
	// C: 	length = sizeof(br.val)
	// C: }
	for i = 0; i < length; i++ {
		value |= vp8l_val_t(start[i] << (8 * i))
	}
	br.val = value
	br.pos = length
}

// Sets a new data buffer.
func VP8LBitReaderSetBuffer( /* const */ br *VP8LBitReader /*const*/, buf []uint8, len uint64) {
	assert.Assert(br != nil)
	assert.Assert(buf != nil)
	assert.Assert(len < 0xfffffff8) // can't happen with a RIFF chunk.
	br.buf = buf
	br.len = len
	// 'pos' > 'len' should be considered a param error.
	br.eos = (br.pos > br.len) || VP8LIsEndOfStream(br)
}

func VP8LSetEndOfStream( /* const */ br *VP8LBitReader) {
	br.eos = true
	br.bit_pos = 0 // To afunc undefined behaviour with shifts.
}

// Advances the read buffer by 4 bytes to make room for reading next 32 bits.
// Speed critical, but infrequent part of the code can be non-inlined.
func VP8LDoFillBitWindow( /* const */ br *VP8LBitReader) {
	assert.Assert(br.bit_pos >= VP8L_WBITS)
	// C: if br.pos+sizeof(br.val) < br.len {
	br.val >>= VP8L_WBITS
	br.bit_pos -= VP8L_WBITS
	br.val |= vp8l_val_t(endian.HToLE32(WebPMemToUint32(br.buf+br.pos))) << (VP8L_LBITS - VP8L_WBITS)
	br.pos += VP8L_LOG8_WBITS
	return
	// C: }

	ShiftBytes(br) // Slow path.
}

// Reads the specified number of bits from read buffer.
// Flags an error in case end_of_stream or n_bits is more than the allowed limit
// of VP8L_MAX_NUM_BIT_READ (inclusive).
// Flags 'eos' if this read attempt is going to cross the read buffer.
func VP8LReadBits( /* const */ br *VP8LBitReader, n_bits int) uint32 {
	assert.Assert(n_bits >= 0)
	// Flag an error if end_of_stream or n_bits is more than allowed limit.
	if !br.eos && n_bits <= VP8L_MAX_NUM_BIT_READ {
		val := VP8LPrefetchBits(br) & kBitMask[n_bits]
		new_bits := br.bit_pos + n_bits
		br.bit_pos = new_bits
		ShiftBytes(br)
		return val
	} else {
		VP8LSetEndOfStream(br)
		return 0
	}
}

// Return the prefetched bits, so they can be looked up.
func VP8LPrefetchBits( /* const */ br *VP8LBitReader) uint32 {
	return uint32(br.val >> (br.bit_pos & (VP8L_LBITS - 1)))
}

// Returns true if there was an attempt at reading bit past the end of
// the buffer. Doesn't set br.eos flag.
func VP8LIsEndOfStream( /* const */ br *VP8LBitReader) bool {
	assert.Assert(br.pos <= br.len)
	return br.eos || (br.pos == br.len && br.bit_pos > VP8L_LBITS)
}

// For jumping over a number of bits in the bit stream when accessed with
// VP8LPrefetchBits and VP8LFillBitWindow.
// This function does br *set *not.eos, since it's speed-critical.
// Use with extreme care!
func VP8LSetBitPos( /* const */ br *VP8LBitReader, val int) {
	br.bit_pos = val
}

func VP8LFillBitWindow( /* const */ br *VP8LBitReader) {
	if br.bit_pos >= VP8L_WBITS {
		VP8LDoFillBitWindow(br)
	}
}
