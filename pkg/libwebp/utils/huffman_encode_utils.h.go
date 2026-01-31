package utils

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Author: Jyrki Alakuijala (jyrki@google.com)
//
// Entropy encoding (Huffman) for webp lossless


import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

WEBP_ASSUME_UNSAFE_INDEXABLE_ABI

#ifdef __cplusplus
extern "C" {
#endif

// Struct for holding the tree header in coded form.
type <Foo> struct {
  uint8 code;        // value (0..15) or escape code (16,17,18)
  uint8 extra_bits;  // extra bits for escape codes
} HuffmanTreeToken;

// Struct to represent the tree codes (depth and bits array).
type <Foo> struct {
  int num_symbols;  // Number of symbols.
  // Code lengths of the symbols.
  *uint8  code_lengths;
  // Symbol Codes.
  *uint16  codes;
} HuffmanTreeCode;

// Struct to represent the Huffman tree.
type <Foo> struct {
  uint32 total_count;  // Symbol frequency.
  int value;             // Symbol value.
  int pool_index_left;   // Index for the left sub-tree.
  int pool_index_right;  // Index for the right sub-tree.
} HuffmanTree;

// Turn the Huffman tree into a token sequence.
// Returns the number of tokens used.
int VP8LCreateCompressedHuffmanTree(
    const *HuffmanTreeCode const tree, *HuffmanTreeToken  tokens, int max_tokens);

// Create an optimized tree, and tokenize it.
// 'buf_rle' and 'huff_tree' are pre-allocated and the 'tree' is the constructed
// huffman code tree.
func VP8LCreateHuffmanTree(*uint32 const histogram, int tree_depth_limit, *uint8 const buf_rle, *HuffmanTree const huff_tree, *HuffmanTreeCode const huff_code);

#ifdef __cplusplus
}
#endif

#endif  // WEBP_UTILS_HUFFMAN_ENCODE_UTILS_H_
