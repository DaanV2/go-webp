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
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

type VP8BitWriter struct {
	vrange  int32 // range-1
	value   int32
	run     int // number of outstanding bits
	nb_bits int // number of pending bits
	// internal buffer. Re-allocated regularly. Not owned.
	buf     []uint8 /* (max_pos) */
	pos     uint64
	max_pos uint64
	error   int // true in case of error
}

func VP8PutBit( /* const */ bw *VP8BitWriter, bit int, prob int) int {
	split := (bw.vrange * prob) >> 8
	if bit != 0 {
		bw.value += split + 1
		bw.vrange -= split + 1
	} else {
		bw.vrange = split
	}
	if bw.vrange < 127 { // emit 'shift' bits out and renormalize
		shift := kNorm[bw.vrange]
		bw.vrange = kNewRange[bw.vrange]
		bw.value <<= shift
		bw.nb_bits += shift
		if bw.nb_bits > 0 {
			Flush(bw)
		}
	}
	return bit
}

func VP8PutBitUniform( /* const */ bw *VP8BitWriter, bit int) int {
	split := bw.vrange >> 1
	if bit {
		bw.value += split + 1
		bw.v -= split + 1
	} else {
		bw.vrange = split
	}
	if bw.vrange < 127 {
		bw.vrange = kNewRange[bw.vrange]
		bw.value <<= 1
		bw.nb_bits += 1
		if bw.nb_bits > 0 {
			Flush(bw)
		}
	}
	return bit
}

func BitWriterResize( /* const */ bw *VP8BitWriter, extra_size uint64) int {
	var new_buf []uint8
	var new_size uint64
	needed_size_64b := uint64(bw.pos + extra_size)
	needed_size := uint64(needed_size_64b)
	if needed_size_64b != needed_size {
		bw.error = 1
		return 0
	}
	if needed_size <= bw.max_pos {
		return 1
	}
	// If the following line wraps over 32bit, the test just after will catch it.
	new_size = 2 * bw.max_pos
	if new_size < needed_size {
		new_size = needed_size
	}
	if new_size < 1024 {
		new_size = 1024
	}
	//   new_buf = (*uint8)WebPSafeMalloc(uint64(1), new_size);
	//   if (new_buf == nil) {
	//     bw.error = 1;
	//     return 0;
	//   }
	new_buf = make([]uint8, new_size)

	if bw.pos > 0 {
		assert.Assert(bw.buf != nil)
		stdlib.MemCpy(new_buf, bw.buf, bw.pos)
	}

	bw.buf = new_buf // bidi index -> new_size
	bw.max_pos = new_size
	return 1
}

func Flush( /* const */ bw *VP8BitWriter) {
	s := 8 + bw.nb_bits
	bits := bw.value >> s
	assert.Assert(bw.nb_bits >= 0)
	bw.value -= bits << s
	bw.nb_bits -= 8
	if bits & 0xff != 0xff {
		pos := bw.pos
		if !BitWriterResize(bw, bw.run+1) {
			return
		}
		if bits & 0x100 { // overflow . propagate carry over pending 0xff's
			if pos > 0 {
				bw.buf[pos-1]++
			}
		}
		if bw.run > 0 {
			value := tenary.If(bits&0x100, 0x00, 0xff)
			for ; bw.run > 0; bw.run-- {
				bw.buf[pos] = value
				pos++
			}
		}
		bw.buf[pos] = bits & 0xff
		pos++
		bw.pos = pos
	} else {
		bw.run++ // delay writing of bytes 0xff, pending eventual carry.
	}
}

func VP8PutBits( /* const */ bw *VP8BitWriter, value uint32, nb_bits int) {
	var mask uint32
	assert.Assert(nb_bits > 0 && nb_bits < 32)
	for mask = uint(1) << (nb_bits - 1); mask; mask >>= 1 {
		VP8PutBitUniform(bw, value&mask)
	}
}

func VP8PutSignedBits( /* const */ bw *VP8BitWriter, value int, nb_bits int) {
	if !VP8PutBitUniform(bw, value != 0) {
		return
	}
	if value < 0 {
		VP8PutBits(bw, ((-value)<<1)|1, nb_bits+1)
	} else {
		VP8PutBits(bw, value<<1, nb_bits+1)
	}
}

// Initialize the object. Allocates some initial memory based on expected_size.
func VP8BitWriterInit( /* const */ bw *VP8BitWriter, expected_size uint64) int {
	bw.vrange = 255 - 1
	bw.value = 0
	bw.run = 0
	bw.nb_bits = -8
	bw.pos = 0
	bw.max_pos = 0
	bw.error = 0
	bw.buf = nil

	return tenary.If(expected_size > 0, BitWriterResize(bw, expected_size), 1)
}

// Finalize the bitstream coding. Returns a pointer to the internal buffer.
func VP8BitWriterFinish( /* const */ bw *VP8BitWriter) []uint8 {
	VP8PutBits(bw, 0, 9-bw.nb_bits)
	bw.nb_bits = 0 // pad with zeroes
	Flush(bw)
	return bw.buf
}

// Appends some bytes to the internal buffer. Data is copied.
func VP8BitWriterAppend( /* const */ bw *VP8BitWriter /*const*/, data *uint8, size uint64) int {
	assert.Assert(data != nil)
	if bw.nb_bits != -8 {
		return 0 // Flush() must have been called
	}
	if !BitWriterResize(bw, size) {
		return 0
	}
	stdlib.MemCpy(bw.buf+bw.pos, data, size)
	bw.pos += size
	return 1
}

// return approximate write position (in bits)
func VP8BitWriterPos( /* const */ bw *VP8BitWriter) uint64 {
	nb_bits := 8 + bw.nb_bits // bw.nb_bits is <= 0, note
	return (bw.pos+bw.run)*8 + nb_bits
}

// Returns a pointer to the internal buffer.
func VP8BitWriterBuf( /* const */ bw *VP8BitWriter) *uint8 {
	return bw.buf
}

// Returns the size of the internal buffer.
func VP8BitWriterSize( /* const */ bw *VP8BitWriter) uint64 {
	return bw.pos
}
