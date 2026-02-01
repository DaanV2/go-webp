package huffman

import "github.com/daanv2/go-webp/pkg/constants"

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

// Contiguous memory segment of HuffmanCodes.
type HuffmanTablesSegment struct {
	start *HuffmanCode //(size)
	// Pointer to where we are writing into the segment. Starts at 'start' and
	// cannot go beyond 'start' + 'size'.
	curr_table *HuffmanCode
	// Pointer to the next segment in the chain.
	next *HuffmanTablesSegment
	size int
}

// Chained memory segments of HuffmanCodes.
type HuffmanTables struct {
	root HuffmanTablesSegment
	// Currently processed segment. At first, this is 'root'.
	curr_segment *HuffmanTablesSegment
}

// Huffman table group.
// Includes special handling for the following cases:
//  - is_trivial_literal: one common literal base for RED/BLUE/ALPHA (not GREEN)
//  - is_trivial_code: only 1 code (no bit is read from bitstream)
//  - use_packed_table: few enough literal symbols, so all the bit codes
//    can fit into a small look-up table packed_table[]
// The common literal base, if applicable, is stored in 'literal_arb'.
type HTreeGroup struct {
	htrees [constants.HUFFMAN_CODES_PER_META_CODE]*HuffmanCode
	// True, if huffman trees for Red, Blue & Alpha
	// Symbols are trivial (have a single code).
	is_trivial_literal int
	// If is_trivial_literal is true, this is the
	// ARGB value of the pixel, with Green channel
	// being set to zero.
	literal_arb uint32
	// true if is_trivial_literal with only one code
	// use packed table below for short literal code
	// table mapping input bits to a packed values, or escape case to literal code
	is_trivial_code  int
	use_packed_table int
	packed_table     [HUFFMAN_PACKED_TABLE_SIZE]HuffmanCode32
}
