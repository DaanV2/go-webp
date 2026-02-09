// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package huffman

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

// Create an optimal Huffman tree.
//
// (data,length): population counts.
// tree_limit: maximum bit depth (inclusive) of the codes.
// bit_depths[]: how many bits are used for the symbol.
//
// Returns 0 when an error has occurred.
//
// The catch here is that the tree cannot be arbitrarily deep
//
// count_limit is the value that is to be faked as the minimum value
// and this minimum value is raised until the tree matches the
// maximum length requirement.
//
// This algorithm is not of excellent performance for very long data blocks,
// especially when population counts are longer than 2**tree_limit, but
// we are not planning to use this with extremely long blocks.
//
// See https://en.wikipedia.org/wiki/Huffman_coding
func GenerateOptimalTree(histogram []uint32, histogram_size int, tree []*HuffmanTree, tree_depth_limit int, bit_depths *uint8) {
	var count_min uint32
	var tree_pool []*HuffmanTree
	var i int
	tree_size_orig := 0

	for i = 0; i < histogram_size; i++ {
		if histogram[i] != 0 {
			tree_size_orig++
		}
	}

	if tree_size_orig == 0 { // pretty optimal already!
		return
	}

	// C: tree_pool = tree + tree_size_orig

	// For block sizes with less than 64k symbols we never need to do a
	// second iteration of this loop.
	// If we actually start running inside this loop a lot, we would perhaps
	// be better off with the Katajainen algorithm.
	assert.Assert(tree_size_orig <= (1 << (tree_depth_limit - 1)))
	for count_min = 1; ; count_min *= 2 {
		tree_size := tree_size_orig
		// We need to pack the Huffman tree in tree_depth_limit bits.
		// So, we try by faking histogram entries to be at least 'count_min'.
		idx := 0
		var j int
		for j = 0; j < histogram_size; j++ {
			if histogram[j] != 0 {
				count := tenary.If(histogram[j] < count_min, count_min, histogram[j])
				tree[idx].total_count = count
				tree[idx].value = j
				tree[idx].pool_index_left = -1
				tree[idx].pool_index_right = -1
				idx++
			}
		}

		// Build the Huffman tree.
		// C: qsort(tree, tree_size, sizeof(*tree), CompareHuffmanTrees)

		if tree_size > 1 { // Normal case.
			tree_pool_size := 0
			for tree_size > 1 { // Finish when we have only one root.
				var count uint32
				tree_pool[tree_pool_size] = tree[tree_size-1]
				tree_pool_size++
				tree_pool[tree_pool_size] = tree[tree_size-2]
				tree_pool_size++
				count = tree_pool[tree_pool_size-1].total_count +
					tree_pool[tree_pool_size-2].total_count
				tree_size -= 2
				{
					// Search for the insertion point.
					var k int
					for k = 0; k < tree_size; k++ {
						if tree[k].total_count <= count {
							break
						}
					}
					// C: memmove(tree+(k+1), tree+k, (tree_size-k)*sizeof(*tree))
					tree[k].total_count = count
					tree[k].value = -1

					tree[k].pool_index_left = tree_pool_size - 1
					tree[k].pool_index_right = tree_pool_size - 2
					tree_size = tree_size + 1
				}
			}
			SetBitDepths(tree[0], tree_pool, bit_depths, 0)
		} else if tree_size == 1 { // Trivial case: only one element.
			bit_depths[tree[0].value] = 1
		}

		{
			// Test if this Huffman tree satisfies our 'tree_depth_limit' criteria.
			max_depth := bit_depths[0]
			for j = 1; j < histogram_size; j++ {
				if max_depth < bit_depths[j] {
					max_depth = bit_depths[j]
				}
			}
			if max_depth <= tree_depth_limit {
				break
			}
		}
	}
}