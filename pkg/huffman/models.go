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
