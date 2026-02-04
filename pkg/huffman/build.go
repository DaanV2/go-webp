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
)

// sorted[code_lengths_size] is a pre-allocated array for sorting symbols
// by code length.
func BuildHuffmanTable( /* const */ root_table []*HuffmanCode, root_bits int /* const */, code_lengths []int, code_lengths_size int, sorted []uint16) int {
	// next available space in table
	var table []*HuffmanCode = root_table
	total_size := 1 << root_bits // total size root table + 2nd level table
	var len int                  // current code length
	var symbol int               // symbol index in original or sorted table
	// number of codes of each length:
	var count [MAX_ALLOWED_CODE_LENGTH + 1]int = [MAX_ALLOWED_CODE_LENGTH + 1]int{0}
	// offsets in sorted table for each length:
	var offset [MAX_ALLOWED_CODE_LENGTH + 1]int

	assert.Assert(code_lengths_size != 0)
	assert.Assert(code_lengths != nil)
	assert.Assert((root_table != nil && sorted != nil) || (root_table == nil && sorted == nil))
	assert.Assert(root_bits > 0)

	// Build histogram of code lengths.
	for symbol = 0; symbol < code_lengths_size; symbol++ {
		if code_lengths[symbol] > MAX_ALLOWED_CODE_LENGTH {
			return 0
		}
		count[code_lengths[symbol]] = count[code_lengths[symbol]] + 1
	}

	// Error, all code lengths are zeros.
	if count[0] == code_lengths_size {
		return 0
	}

	// Generate offsets into sorted symbol table by code length.
	offset[1] = 0
	for len = 1; len < MAX_ALLOWED_CODE_LENGTH; len++ {
		if count[len] > (1 << len) {
			return 0
		}
		offset[len+1] = offset[len] + count[len]
	}

	// Sort symbols by length, by symbol order within each length.
	for symbol = 0; symbol < code_lengths_size; symbol++ {
		symbol_code_length := code_lengths[symbol]
		if code_lengths[symbol] > 0 {
			if sorted != nil {
				assert.Assert(offset[symbol_code_length] < code_lengths_size)
				// The following check is not redundant with the assert. It prevents a
				// potential buffer overflow that the optimizer might not be able to
				// rule out on its own.
				if offset[symbol_code_length] >= code_lengths_size {
					return 0
				}
				sorted[offset[symbol_code_length]] = uint16(symbol)
				offset[symbol_code_length] = offset[symbol_code_length] + 1
			} else {
				offset[symbol_code_length]++
			}
		}
	}

	// Special case code with only one value.
	if offset[MAX_ALLOWED_CODE_LENGTH] == 1 {
		if sorted != nil {
			code := HuffmanCode{
				bits:  0,
				value: sorted[0],
			}

			ReplicateValue(table, 1, total_size, code)
		}
		return total_size
	}

	{
		var step int                  // step size to replicate values in current table
		low := uint32(0xffffffff)     // low bits for current root entry
		mask := total_size - 1        // mask for low bits
		key := 0                      // reversed prefix code
		num_nodes := 1                // number of Huffman tree nodes
		num_open := 1                 // number of open branches in current tree level
		table_bits := root_bits       // key length of current table
		table_size := 1 << table_bits // size of current table
		symbol = 0
		// Fill in root table.
		len = 1
		step = 2
		for len <= root_bits {
			num_open <<= 1
			num_nodes += num_open
			num_open -= count[len]
			if num_open < 0 {
				return 0
			}
			if root_table != nil {
				for ; count[len] > 0; count[len]-- {
					code := HuffmanCode{
						bits:  uint8(len),
						value: sorted[symbol],
					}

					symbol++
					// NOTE: old was &table[key], suspecting it was grabbing the address of the pointer at key as start
					ReplicateValue(table[key:], step, table_size, code)
					key = int(GetNextKey(uint32(key), len)) //NOTE: key used to be uint32
				}
			}

			len++
			step <<= 1
		}

		// Fill in 2nd level tables and add pointers to root table.
		len = root_bits + 1
		step = 2
		for len <= MAX_ALLOWED_CODE_LENGTH {
			num_open <<= 1
			num_nodes += num_open
			num_open -= count[len]
			if num_open < 0 {
				return 0
			}
			for ; count[len] > 0; count[len]-- {
				var code HuffmanCode
				if uint32(key & mask) != low {
					if root_table != nil {
						// This was moving the pointer
						table = table[table_size:]
					}
					table_bits = NextTableBitSize(count, len, root_bits)
					table_size = 1 << table_bits
					total_size += table_size
					low = key & mask
					if root_table != nil {
						root_table[low].bits = (uint8)(table_bits + root_bits)
						root_table[low].value = (uint16)((table - root_table) - low)
					}
				}
				if root_table != nil {
					code.bits = (len - root_bits)
					code.value = sorted[symbol]
					symbol++
					ReplicateValue(&table[key>>root_bits], step, table_size, code)
				}
				key = int(GetNextKey(uint32(key), len))
			}

			len++
			step <<= 1
		}

		// Check if tree is full.
		if num_nodes != 2*offset[MAX_ALLOWED_CODE_LENGTH]-1 {
			return 0
		}
	}

	return total_size
}
