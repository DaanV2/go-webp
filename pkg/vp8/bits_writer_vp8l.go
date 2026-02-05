// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

type VP8LBitWriter struct {
	bits vp8l_atype_t // bit accumulator
	used int          // number of bits used in accumulator
	buf  []uint8      /* (end)  */ // start of buffer
	cur  *uint8       // current write position
	end  *uint8       // end of buffer

	// After all bits are written (VP8LBitWriterFinish()), the caller must observe
	// the state of 'error'. A value of 1 indicates that a memory allocation
	// failure has happened during bit writing. A value of 0 indicates successful
	// writing of bits.
	error int
}

func VP8LBitWriterNumBytes( /* const */ bw *VP8LBitWriter) uint64 {
	return (bw.cur - bw.buf) + ((bw.used + 7) >> 3)
}

// Returns 1 on success.
func VP8LBitWriterResize(/* const */ bw *VP8LBitWriter, extra_size uint64 ) int {
  var allocated_buf *uint8;
  var allocated_size uint64 ;
  max_bytes := bw.end - bw.buf;
  current_size := bw.cur - bw.buf;
  size_required_64b := uint64(current_size + extra_size);
  size_required := uint64(size_required_64b);
  if (size_required != size_required_64b) {
    bw.error = 1;
    return 0;
  }
  if (max_bytes > 0 && size_required <= max_bytes) { return 1; }
  allocated_size = (3 * max_bytes) >> 1;
  if (allocated_size < size_required) {allocated_size = size_required;}
  // make allocated size multiple of 1k
  allocated_size = (((allocated_size >> 10) + 1) << 10);

//   allocated_buf = (*uint8)WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
//       *void, WebPSafeMalloc(uint64(1), allocated_size), allocated_size);
  allocated_buf := make([]uint8, allocated_size);

// 	  if (allocated_buf == nil) {
//     bw.error = 1;
//     return 0;
//   }
  if (current_size > 0) {
    stdlib.MemCpy(allocated_buf, bw.buf, current_size);
  }
  
  bw.buf = allocated_buf;
  bw.end = allocated_buf + allocated_size;
  bw.cur = allocated_buf + current_size;
  return 1;
}

// Returns false in case of memory allocation error.
func VP8LBitWriterInit(/* const */ bw *VP8LBitWriter, expected_size uint64) int {
  stdlib.Memset(bw, 0, sizeof(*bw));
  return VP8LBitWriterResize(bw, expected_size);
}

// Returns false in case of memory allocation error.
func VP8LBitWriterClone(/* const */ src *VP8LBitWriter, /*const*/ dst *VP8LBitWriter) int {
  current_size := src.cur - src.buf;
  assert.Assert(src.cur >= src.buf && src.cur <= src.end);
  if (!VP8LBitWriterResize(dst, current_size)) { return 0; }
  stdlib.MemCpy(dst.buf, src.buf, current_size);
  dst.bits = src.bits;
  dst.used = src.used;
  dst.error = src.error;
  dst.cur = dst.buf + current_size;
  return 1;
}

// Resets the cursor of the BitWriter bw to when it was like in bw_init.
func VP8LBitWriterReset(/* const */ bw_init *VP8LBitWriter, /*const*/ bw *VP8LBitWriter) {
  bw.bits = bw_init.bits;
  bw.used = bw_init.used;
  bw.cur = bw.buf + (bw_init.cur - bw_init.buf);
  assert.Assert(bw.cur <= bw.end);
  bw.error = bw_init.error;
}

// Swaps the memory held by two BitWriters.
func VP8LBitWriterSwap(/* const */ src *VP8LBitWriter, /*const*/ dst *VP8LBitWriter) {
  var tmp VP8LBitWriter = *src;
  *src = *dst;
  *dst = tmp;
}

// Internal function for VP8LPutBits flushing VP8L_WRITER_BITS bits from the
// written state.
func VP8LPutBitsFlushBits(/* const */ bw *VP8LBitWriter, used *int, vp8l_atype_t* bits) {
  // If needed, make some room by flushing some bits out.
  if (bw.cur + VP8L_WRITER_BYTES > bw.end) {
    extra_size := (bw.end - bw.buf) + MIN_EXTRA_SIZE;
    if (!CheckSizeOverflow(extra_size) || !VP8LBitWriterResize(bw, uint64(extra_size))) {
      bw.cur = bw.buf;
      bw.error = 1;
      return;
    }
  }
  *(vp8l_wtype_t*)bw.cur = (vp8l_wtype_t)WSWAP((vp8l_wtype_t)*bits);
  bw.cur += VP8L_WRITER_BYTES;
  *bits >>= VP8L_WRITER_BITS;
  *used -= VP8L_WRITER_BITS;
}

// PutBits internal function used in the 16 bit vp8l_wtype_t case.
func VP8LPutBitsInternal(/* const */ bw *VP8LBitWriter, bits uint32, n_bits int) {
  vp8l_atype_t lbits = bw.bits;
  used := bw.used;
  assert.Assert(n_bits <= VP8L_WRITER_MAX_BITS);
  if (n_bits == 0) {return;}
  // Special case of overflow handling for 32bit accumulator (2-steps flush).
  if (used + n_bits >= VP8L_WRITER_MAX_BITS) {
    // Fill up all the VP8L_WRITER_MAX_BITS so it can be flushed out below.
    shift := VP8L_WRITER_MAX_BITS - used;
    lbits |= (vp8l_atype_t)bits << used;
    used = VP8L_WRITER_MAX_BITS;
    n_bits -= shift;
    if (shift >= (int)sizeof(bits) * 8) {
      // Undefined behavior.
      assert.Assert(shift == (int)sizeof(bits) * 8);
      bits = 0;
    } else {
      bits >>= shift;
    }
    assert.Assert(n_bits <= VP8L_WRITER_MAX_BITS);
  }
  // If needed, make some room by flushing some bits out.
  for (used >= VP8L_WRITER_BITS) {
    VP8LPutBitsFlushBits(bw, &used, &lbits);
  }
  bw.bits = lbits | ((vp8l_atype_t)bits << used);
  bw.used = used + n_bits;
}

// Finalize the bitstream coding. Returns a pointer to the internal buffer.
func VP8LBitWriterFinish(/* const */ bw *VP8LBitWriter) *uint8 {
  // flush leftover bits
  if (VP8LBitWriterResize(bw, (bw.used + 7) >> 3)) {
    for bw.used > 0 {
      *bw.cur++ = (uint8)bw.bits;
      bw.bits >>= 8;
      bw.used -= 8;
    }
    bw.used = 0;
  }
  return bw.buf;
}

// This function writes bits into bytes in increasing addresses (little endian),
// and within a byte least-significant-bit first.
// This function can write up to VP8L_WRITER_MAX_BITS bits in one go, but
// VP8LBitReader can only read 24 bits max (VP8L_MAX_NUM_BIT_READ).
// VP8LBitWriter's 'error' flag is set in case of memory allocation error.
func VP8LPutBits(/* const */ bw *VP8LBitWriter, bits uint32, n_bits int) {
	if VP8L_WRITER_BYTES == 4{
	if (n_bits == 0) {return;}
	if (bw.used >= VP8L_WRITER_BITS) {
		VP8LPutBitsFlushBits(bw, &bw.used, &bw.bits);
	}
	bw.bits |= vp8l_atype_t(bits << bw.used);
	bw.used += n_bits;
	}else{
	VP8LPutBitsInternal(bw, bits, n_bits);
	}
}