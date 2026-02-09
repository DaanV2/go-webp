// Entropy encoding (Huffman) for webp lossless.
package huffman

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/constants"
	"github.com/daanv2/go-webp/pkg/stdlib"
)

// -----------------------------------------------------------------------------
// Util function to optimize the symbol map for RLE coding

// Heuristics for selecting the stride ranges to collapse.
func ValuesShouldBeCollapsedToStrideAverage(a, b int) bool {
	return stdlib.Abs(a-b) < 4
}
func CodeRepeatedValues(repetitions int, tokens *HuffmanTreeToken, value int, prev_value int) *HuffmanTreeToken {
	assert.Assert(value <= MAX_ALLOWED_CODE_LENGTH)
	if value != prev_value {
		tokens.code = value
		tokens.extra_bits = 0
		// C: tokens++
		repetitions--
	}
	for repetitions >= 1 {
		if repetitions < 3 {
			var i int
			for i = 0; i < repetitions; i++ {
				tokens.code = value
				tokens.extra_bits = 0
				// C: tokens++
			}
			break
		} else if repetitions < 7 {
			tokens.code = 16
			tokens.extra_bits = repetitions - 3
			// C: tokens++
			break
		} else {
			tokens.code = 16
			tokens.extra_bits = 3
			// C: tokens++
			repetitions -= 6
		}
	}
	return tokens
}

func CodeRepeatedZeros(repetitions int, tokens []HuffmanTreeToken) []HuffmanTreeToken {
	token_index := 0

	for repetitions >= 1 {
		if repetitions < 3 {
			var i int
			for i = 0; i < repetitions; i++ {
				tokens[token_index].code = 0 // 0-value
				tokens[token_index].extra_bits = 0
				token_index++
			}
			break
		} else if repetitions < 11 {
			tokens[token_index].code = 17
			tokens[token_index].extra_bits = uint8(repetitions - 3)
			token_index++
			break
		} else if repetitions < 139 {
			tokens[token_index].code = 18
			tokens[token_index].extra_bits = uint8(repetitions - 11)
			token_index++
			break
		} else {
			tokens[token_index].code = 18
			tokens[token_index].extra_bits = 0x7f // 138 repeated 0s
			token_index++
			repetitions -= 138
		}
	}
	return tokens
}

// Turn the Huffman tree into a token sequence.
// Returns the number of tokens used.
func VP8LCreateCompressedHuffmanTree(tree *HuffmanTreeCode, tokens *HuffmanTreeToken, max_tokens int) int {
	var current_token *HuffmanTreeToken = tokens
	var starting_token *HuffmanTreeToken = tokens
	// C: var ending_token *HuffmanTreeToken = tokens + max_tokens
	depth_size := tree.num_symbols
	prev_value := 8 // 8 is the initial value for rle.
	i := 0
	assert.Assert(tokens != nil)
	for i < depth_size {
		value := tree.code_lengths[i]
		k := i + 1
		var runs int
		for k < depth_size && tree.code_lengths[k] == value {
			k++
		}
		runs = k - i
		if value == 0 {
			current_token = CodeRepeatedZeros(runs, current_token)
		} else {
			current_token =
				CodeRepeatedValues(runs, current_token, value, prev_value)
			prev_value = value
		}
		i += runs
		// C: assert.Assert(current_token <= ending_token)
	}
	//   (void)ending_token;  // suppress 'unused variable' warning
	// C: return (int)(current_token - starting_token)
}

// -----------------------------------------------------------------------------

// Pre-reversed 4-bit values.
var kReversedBits = [16]uint8{0x0, 0x8, 0x4, 0xc, 0x2, 0xa, 0x6, 0xe, 0x1, 0x9, 0x5, 0xd, 0x3, 0xb, 0x7, 0xf}

func ReverseBits(num_bits int, bits uint32) uint32 {
	retval := uint32(0)
	i := 0
	for i < num_bits {
		i += 4
		retval |= uint32(kReversedBits[bits&0xf] << (constants.MAX_ALLOWED_CODE_LENGTH + 1 - i))
		bits >>= 4
	}
	retval >>= (constants.MAX_ALLOWED_CODE_LENGTH + 1 - num_bits)
	return retval
}

// Get the actual bit values for a tree of bit depths.
func ConvertBitDepthsToSymbols( /* const */ tree *HuffmanTreeCode) {
	// 0 bit-depth means that the symbol does not exist.
	var i int
	var len int
	var next_code [constants.MAX_ALLOWED_CODE_LENGTH + 1]uint32
	var depth_count [constants.MAX_ALLOWED_CODE_LENGTH + 1]int = [constants.MAX_ALLOWED_CODE_LENGTH + 1]int{0}

	assert.Assert(tree != nil)
	len = tree.num_symbols
	for i = 0; i < len; i++ {
		code_length := tree.code_lengths[i]
		assert.Assert(code_length <= MAX_ALLOWED_CODE_LENGTH)
		depth_count[code_length] = depth_count[code_length] + 1
	}
	depth_count[0] = 0 // ignore unused symbol
	next_code[0] = 0
	{
		code := 0
		for i = 1; i <= MAX_ALLOWED_CODE_LENGTH; i++ {
			code = (code + depth_count[i-1]) << 1
			next_code[i] = code
		}
	}
	for i = 0; i < len; i++ {
		code_length := tree.code_lengths[i]
		ncode := next_code[code_length]
		next_code[code_length] = next_code[code_length] + 1
		tree.codes[i] = ReverseBits(code_length, ncode)
	}
}

// Main entry point
// Create an optimized tree, and tokenize it.
// 'buf_rle' and 'huff_tree' are pre-allocated and the 'tree' is the constructed
// huffman code tree.
func VP8LCreateHuffmanTree( /* const */ histogram *uint32, tree_depth_limit int /* const */, buf_rle *uint8 /* const */, huff_tree *HuffmanTree /* const */, huff_code *HuffmanTreeCode) {
	num_symbols := huff_code.num_symbols
	var bounded_histogram *uint32 = histogram // bidi index -> (uint64)num_symbols * sizeof(*histogram)
	var bounded_buf_rle *uint8 = buf_rle      // bidi index -> (uint64)num_symbols * sizeof(*buf_rle)

	// C: stdlib.Memset(bounded_buf_rle, 0, num_symbols*sizeof(*buf_rle))

	OptimizeHuffmanForRle(num_symbols, bounded_buf_rle, bounded_histogram)
	// buff_tree bidi index -> 3 * num_symbols * sizeof(*huff_tree)
	GenerateOptimalTree(bounded_histogram, num_symbols, huff_tree, tree_depth_limit, huff_code.code_lengths)
	// Create the actual bit codes for the bit lengths.
	ConvertBitDepthsToSymbols(huff_code)
}
