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
)

type VP8BitReader struct {
	// boolean decoder  (keep the field ordering as is!)
	value  bit_t   // current value
	vrange range_t // current range minus 1.0 In [127, 254] interval.
	// number of valid bits left
	// read buffer
	bits int

	// next byte to be read
	buf []uint8 /* (buf_end) */
	// end of read buffer
	// max packed-read position on buffer
	buf_end *uint8
	buf_max *uint8
	eof     bool // true if input is exhausted
}

// Sets the working read buffer.
func VP8BitReaderSetBuffer( /* const */ br *VP8BitReader /*const*/, start []uint8, size uint64) {
	assert.Assert(start != nil)
	br.buf = start
	// C: br.buf_end = start + size
	// C: br.buf_max = tenary.If(size >= sizeof(lbit_t), start+size-sizeof(lbit_t)+1, start)
}

// Initialize the bit reader and the boolean decoder.
func VP8InitBitReader( /* const */ br *VP8BitReader /*const*/, start []uint8, size uint64) {
	assert.Assert(br != nil)
	assert.Assert(start != nil)
	assert.Assert(size < (uint64(1) << 31)) // limit ensured by format and upstream checks
	br.vrange = 255 - 1
	br.value = 0
	br.bits = -8 // to load the very first 8bits
	br.eof = false
	VP8BitReaderSetBuffer(br, start, size)
	VP8LoadNewBytes(br)
}

// Update internal pointers to displace the byte buffer by the
// relative offset 'offset'.
func VP8RemapBitReader( /* const */ br *VP8BitReader, offset ptrdiff_t) {
	if br.buf != nil {
		// C: br.buf += offset
		// C: br.buf_end += offset
		// C: br.buf_max += offset
	}
}

// special case for the tail byte-reading
func VP8LoadFinalBytes( /* const */ br *VP8BitReader) {
	assert.Assert(br != nil && br.buf != nil)
	// Only read 8bits at a time
	// C: if br.buf < br.buf_end {
	//   br.bits += 8
	//   // C: br.value = bit_t(*br.buf++) | (br.value << 8)
	//   // WEBP_SELF_ASSIGN(br.buf_end)
	// } else if !br.eof {
	//   br.value <<= 8
	//   br.bits += 8
	//   br.eof = true
	// } else {
	//   br.bits = 0  // This is to avoid undefined behaviour with shifts.
	// }
}

// return the next value made of 'num_bits' bits
func VP8GetValue( /* const */ br *VP8BitReader, bits int /*const*/, label []byte) uint32 {
	v := uint32(0)
	for bits > 0 {
		bits--
		v |= VP8GetBit(br, 0x80, label) << bits
	}
	return v
}

// return the next value with sign-extension.
func VP8GetSignedValue( /* const */ br *VP8BitReader, bits int /*const*/, label []byte) int {
	value := VP8GetValue(br, bits, label)
	// C: return VP8Get(br, label) ? -value : value
	if VP8Get(br, label) != 0 {
		return -value
	}
	return value
}

// makes sure br.value has at least BITS bits worth of data
func VP8LoadNewBytes( /* const */ br *VP8BitReader) {
	assert.Assert(br != nil && br.buf != nil)
	// Read 'BITS' bits at a time if possible.
	if br.buf < br.buf_max {
		// convert memory type to register type (with some zero'ing!)
		var bits bit_t
		var in_bits lbit_t
		stdlib.MemCpy(&in_bits, br.buf, sizeof(in_bits))

		br.buf += BITS >> 3
		// WEBP_SELF_ASSIGN(br.buf_end)

		if BITS > 32 {
			bits = BSwap64(in_bits)
			bits >>= 64 - BITS
		} else if BITS >= 24 {
			bits = BSwap32(in_bits)
			bits >>= (32 - BITS)
		} else if BITS == 16 {
			bits = BSwap16(in_bits)
		} else { // BITS == 8
			bits = bit_t(in_bits)
		}
		br.value = bits | (br.value << BITS)
		br.bits += BITS
	} else {
		VP8LoadFinalBytes(br) // no need to be inlined
	}
}

// Read a bit with proba 'prob'. Speed-critical function!
func VP8GetBit( /* const */ br *VP8BitReader, prob int /*const*/, label []byte) uint32 {
	// Don't move this declaration! It makes a big speed difference to store
	// 'range' VP *calling *before8LoadNewBytes(), even if this function doesn't
	// alter br.range value.
	var vrange range_t = br.vrange
	if br.bits < 0 {
		VP8LoadNewBytes(br)
	}

	pos := br.bits
	var split range_t = (vrange * range_t(prob)) >> 8
	var value range_t = range_t(br.value >> pos)
	bit := (value > split)
	if bit {
		vrange -= split
		br.value -= bit_t(split+1) << pos
	} else {
		vrange = split + 1
	}
	{
		shift := 7 ^ BitsLog2Floor(vrange)
		vrange <<= shift
		br.bits -= shift
	}
	br.vrange = vrange - 1
	BT_TRACK(br)
	return bit
}

// simplified version of VP8GetBit() for prob=0x80 (note shift is always 1 here)
func VP8GetSigned( /* const */ br *VP8BitReader, v int /*const*/, label []byte) int {
	if br.bits < 0 {
		VP8LoadNewBytes(br)
	}
	{
		pos := br.bits
		var split range_t = br.vrange >> 1
		var value range_t = range_t(br.value >> pos)
		mask := int32(split-value) >> 31 // -1 or 0
		br.bits -= 1
		br.vrange += range_t(mask)
		br.vrange |= 1
		br.value -= bit_t((split+1)&uint32(mask)) << pos
		BT_TRACK(br)
		return (v ^ mask) - mask
	}
}

func VP8GetBitAlt( /* const  */ br *VP8BitReader, prob int /*const*/, label []byte) int {
	// Don't move this declaration! It makes a big speed difference to store
	// 'range' VP *calling *before8LoadNewBytes(), even if this function doesn't
	// alter br.range value.
	var vrange range_t = br.vrange
	if br.bits < 0 {
		VP8LoadNewBytes(br)
	}
	{
		pos := br.bits
		var split range_t = (vrange * prob) >> 8
		var value range_t = range_t(br.value >> pos)
		var bit int // Don't use 'bit := (value > split);", it's slower.
		if value > split {
			vrange -= split + 1
			br.value -= bit_t((split + 1) << pos)
			bit = 1
		} else {
			vrange = split
			bit = 0
		}
		if vrange <= range_t(0x7e) {
			shift := kVP8Log2Range[vrange]
			vrange = kVP8NewRange[vrange]
			br.bits -= shift
		}
		br.vrange = vrange
		BT_TRACK(br)
		return bit
	}
}
