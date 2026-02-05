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

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

// Creates the instance of HTreeGroup with specified number of tree-groups.
func HTreeGroup8LHtreeGroupsNew(num_htree_groups int) []HTreeGroup {
	//   var htree_groups *HTreeGroup = (*HTreeGroup)WebPSafeMalloc(num_htree_groups, sizeof(*htree_groups));
	htree_groups := make([]HTreeGroup, num_htree_groups)
	if htree_groups == nil {
		return nil
	}
	assert.Assert(num_htree_groups <= MAX_HTREE_GROUPS)
	return htree_groups
}

// Returns reverse(reverse(key, len) + 1, len), where reverse(key, len) is the
// bit-wise reversal of the len least significant bits of key.
func GetNextKey(key uint32, len int) uint32 {
	step := uint32(1 << uint32(len-1))
	for (key & step) != 0 {
		step >>= 1
	}

	return tenary.If(step != 0, (key&(step-1))+step, key)
}

// Stores code in table[0], table[step], table[2*step], ..., table[end-step].
// Assumes that end is an integer multiple of step.
func ReplicateValue(table []*HuffmanCode /*(end - step +1)*/, step int, end int, code *HuffmanCode) {
	current_end := end
	assert.Assert(current_end%step == 0)
	for {
		current_end -= step
		table[current_end] = code
		if current_end > 0 {
			continue
		} else {
			break
		}
	}
}

// Returns the table width of the next 2nd level table. count is the histogram
// of bit lengths for the remaining symbols, len is the code length of the next
// processed symbol
func NextTableBitSize(count []int /* (MAX_ALLOWED_CODE_LENGTH + 1) */, len int, root_bits int) int {
	left := 1 << (len - root_bits)
	for len < MAX_ALLOWED_CODE_LENGTH {
		left -= count[len]
		if left <= 0 {
			break
		}
		len++
		left <<= 1
	}
	return len - root_bits
}
