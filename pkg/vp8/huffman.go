// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

// Builds Huffman lookup table assuming code lengths are in symbol order.
// The 'code_lengths' is pre-allocated temporary memory buffer used for creating
// the huffman table.
// Returns built table size or 0 in case of error (invalid tree or
// memory error).
func VP8LBuildHuffmanTable(root_table *HuffmanTables, root_bits int, code_lengths []int, code_lengths_size int) int {
  total_size := BuildHuffmanTable(nil, root_bits, code_lengths, code_lengths_size, nil);
  assert.Assert(code_lengths_size <= MAX_CODE_LENGTHS_SIZE);
  if (total_size == 0 || root_table == nil) {return total_size;}

  if (root_table.curr_segment.curr_table + total_size >=
      root_table.curr_segment.start + root_table.curr_segment.size) {
    // If 'root_table' does not have enough memory, allocate a new segment.
    // The available part of root_table.curr_segment is left unused because we
    // need a contiguous buffer.
    segment_size := root_table.curr_segment.size;
    next *HuffmanTablesSegment = (*HuffmanTablesSegment)WebPSafeMalloc(1, sizeof(*next));
    if (next == nil) return 0;
    // Fill the new segment.
    // We need at least 'total_size' but if that value is small, it is better to
    // allocate a big chunk to prevent more allocations later. 'segment_size' is
    // therefore chosen (any other arbitrary value could be chosen).
    {
      next_size := tenary.If(total_size > segment_size, total_size, segment_size);
      var next_start *HuffmanCode =
          (*HuffmanCode)WebPSafeMalloc(next_size, sizeof(*next_start));
      if (next_start == nil) {
        WebPSafeFree(next);
        return 0;
      }
      next.size = next_size;
      next.start = next_start;
    }
    next.curr_table = next.start;
    next.next = nil;
    // Point to the new segment.
    root_table.curr_segment.next = next;
    root_table.curr_segment = next;
  }
  if (code_lengths_size <= SORTED_SIZE_CUTOFF) {
    // use local stack-allocated array.
    var sorted [SORTED_SIZE_CUTOFF]uint16
	// root_table.curr_segment.curr_table bidi index -> total_size * sizeof(*root_table.curr_segment.curr_table)
    BuildHuffmanTable(root_table.curr_segment.curr_table, root_bits, code_lengths, code_lengths_size, sorted);
  } else {  // rare case. Use heap allocation.
    const sorted *uint16 =
        (*uint16)WebPSafeMalloc(code_lengths_size, sizeof(*sorted));
    if (sorted == nil) return 0;
	// root_table.curr_segment.curr_table bidi index -> total_size * sizeof(*root_table.curr_segment.curr_table)
    // sorted bidi index -> (uint64)code_lengths_size * sizeof(*sorted)
	BuildHuffmanTable(root_table.curr_segment.curr_table, root_bits, code_lengths, code_lengths_size, sorted);
    WebPSafeFree(sorted);
  }
  return total_size;
}

// Allocates a HuffmanTables with 'size' contiguous HuffmanCodes. Returns 0 on
// memory allocation error, 1 otherwise.
func VP8LHuffmanTablesAllocate(size int, huffman_tables *HuffmanTables) int {
  // Have 'segment' point to the first segment for now, 'root'.
  var root *HuffmanTablesSegment = &huffman_tables.root;
  huffman_tables.curr_segment = root;
  root.next = nil;
  // Allocate root.
  {
    const start *HuffmanCode =
        (*HuffmanCode)WebPSafeMalloc(size, sizeof(*root.start));
    if (start == nil) {
      root.start = nil;
      root.size = 0;
      return 0;
    }
    root.size = size;
    root.start = start;
  }
  root.curr_table = root.start;
  return 1;
}

func VP8LHuffmanTablesDeallocate(/* const */ huffman_tables *HuffmanTables) {
  var current, next *HuffmanTablesSegment;
  if (huffman_tables == nil) return;
  // Free the root node.
  current = &huffman_tables.root;
  next = current.next;
  WebPSafeFree(current.start);
  current.start = nil;
  current.size = 0;
  current.next = nil;
  current = next;
  // Free the following nodes.
  while (current != nil) {
    next = current.next;
    WebPSafeFree(current.start);
    WebPSafeFree(current);
    current = next;
  }
}