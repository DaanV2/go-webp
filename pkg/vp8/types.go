// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

type lbit_t = uint64

// #if (BITS > 24)
// ------------------------------------------------------------------------------
// Derived types and constants:
//
//	bit_t = natural register type for storing 'value' (which is BITS+8 bits)
//	range_t = register for 'range' (which is 8bits only)
type bit_t uint64

// #else
// typedef uint32 bit_t
// #endif

type range_t uint32
type vp8l_val_t uint64   // right now, this bit-reader can only use 64bit.
type vp8l_atype_t uint64 // accumulator type
type vp8l_wtype_t uint32 // writing type

// type used for scores, rate, distortion
// Note that MAX_COST is not the maximum allowed by sizeof(score_t),
// in order to allow overflowing computations.
type score_t int64  