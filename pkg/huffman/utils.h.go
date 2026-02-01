package huffman

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Utilities for building and looking up Huffman trees.
//
// Author: Urvang Joshi (urvang@google.com)


import "github.com/daanv2/go-webp/pkg/assert"

import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

// Contiguous memory segment of HuffmanCodes.
type HuffmanTablesSegment struct {
   start  *HuffmanCode(size)
  // Pointer to where we are writing into the segment. Starts at 'start' and
  // cannot go beyond 'start' + 'size'.
  WEBP_UNSAFE_INDEXABLE curr_table *HuffmanCode;
  // Pointer to the next segment in the chain.
  struct next *HuffmanTablesSegment;
  size int;
} HuffmanTablesSegment;

// Chained memory segments of HuffmanCodes.
typedef type HuffmanTables struct {
  HuffmanTablesSegment root;
  // Currently processed segment. At first, this is 'root'.
  curr_segment *HuffmanTablesSegment;
} HuffmanTables;

// Allocates a HuffmanTables with 'size' contiguous HuffmanCodes. Returns 0 on
// memory allocation error, 1 otherwise.
 int VP8LHuffmanTablesAllocate(size int, huffman_tables *HuffmanTables);
func VP8LHuffmanTablesDeallocate(const huffman_tables *HuffmanTables);

const HUFFMAN_PACKED_BITS =6
const HUFFMAN_PACKED_TABLE_SIZE =(uint(1) << HUFFMAN_PACKED_BITS)

// Huffman table group.
// Includes special handling for the following cases:
//  - is_trivial_literal: one common literal base for RED/BLUE/ALPHA (not GREEN)
//  - is_trivial_code: only 1 code (no bit is read from bitstream)
//  - use_packed_table: few enough literal symbols, so all the bit codes
//    can fit into a small look-up table packed_table[]
// The common literal base, if applicable, is stored in 'literal_arb'.
type HTreeGroup struct {
	htrees [HUFFMAN_CODES_PER_META_CODE]*HuffmanCode
	// True, if huffman trees for Red, Blue & Alpha
	// Symbols are trivial (have a single code).
	is_trivial_literal int  
	// If is_trivial_literal is true, this is the
	// ARGB value of the pixel, with Green channel
	// being set to zero.
	literal_arb uint32     
	is_trivial_code int     // true if is_trivial_literal with only one code
	// use packed table below for short literal code
	// table mapping input bits to a packed values, or escape case to literal code
	use_packed_table int    
	packed_table [HUFFMAN_PACKED_TABLE_SIZE]HuffmanCode32
}

// Creates the instance of HTreeGroup with specified number of tree-groups.
func HTreeGroup8LHtreeGroupsNew(num_htree_groups int) *VP;

// Releases the memory allocated for HTreeGroup.
func VP8LHtreeGroupsFree(/* const */ htree_groups *HTreeGroup);

// Builds Huffman lookup table assuming code lengths are in symbol order.
// The 'code_lengths' is pre-allocated temporary memory buffer used for creating
// the huffman table.
// Returns built table size or 0 in case of error (invalid tree or
// memory error).
func VP8LBuildHuffmanTable(/* const */ root_table *HuffmanTables, root_bits int, /* const */   code_lengths []int, code_lengths_size int) int;



#endif  // WEBP_UTILS_HUFFMAN_UTILS_H_
