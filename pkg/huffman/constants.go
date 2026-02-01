package huffman

import "github.com/daanv2/go-webp/pkg/constants"

const (
	HUFFMAN_TABLE_BITS = 8
	HUFFMAN_TABLE_MASK = ((1 << HUFFMAN_TABLE_BITS) - 1)

	LENGTHS_TABLE_BITS = 7
	LENGTHS_TABLE_MASK = ((1 << LENGTHS_TABLE_BITS) - 1)

	HUFFMAN_PACKED_BITS       = 6
	HUFFMAN_PACKED_TABLE_SIZE = (uint(1) << HUFFMAN_PACKED_BITS)

	// Huffman data read via DecodeImageStream is represented in two (red and green)
	// bytes.
	MAX_HTREE_GROUPS = 0x10000

	// Maximum code_lengths_size is 2328 (reached for 11-bit color_cache_bits).
	// More commonly, the value is around ~280.
	MAX_CODE_LENGTHS_SIZE = ((1 << constants.MAX_CACHE_BITS) + constants.NUM_LITERAL_CODES + constants.NUM_LENGTH_CODES)
	// Cut-off value for switching between heap and stack allocation.
	SORTED_SIZE_CUTOFF = 512
)
