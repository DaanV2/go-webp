package huffman

// Struct for holding the tree header in coded form.
type HuffmanTreeToken struct {
	code       uint8 // value (0..15) or escape code (16,17,18)
	extra_bits uint8 // extra bits for escape codes
}

// Struct to represent the tree codes (depth and bits array).
type HuffmanTreeCode struct {
	num_symbols int // Number of symbols.
	// Code lengths of the symbols.
	code_lengths *uint8
	// Symbol Codes.
	codes *uint16
}

// Struct to represent the Huffman tree.
type HuffmanTree struct {
	total_count      uint32 // Symbol frequency.
	value            int    // Symbol value.
	pool_index_left  int    // Index for the left sub-tree.
	pool_index_right int    // Index for the right sub-tree.
}

// Huffman lookup table entry
type HuffmanCode struct {
	bits  uint8  // number of bits used for this symbol
	value uint16 // symbol value or table offset
}

// long version for holding 32b values
type HuffmanCode32 struct {
	bits  int    // number of bits used for this symbol, // or an impossible value if not a literal code.
	value uint32 // 32b packed ARGB value if literal, // or non-literal symbol otherwise
}
