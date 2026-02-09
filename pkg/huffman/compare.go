package huffman

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

// A comparer function for two Huffman trees: sorts first by 'total count'
// (more comes first), and then by 'value' (more comes first).
func CompareHuffmanTrees(t1 , t2 *HuffmanTree) int {
  if t1.total_count > t2.total_count {
    return -1
  } else if t1.total_count < t2.total_count {
    return 1
  } else {
    assert.Assert(t1.value != t2.value)
    return tenary.If(t1.value < t2.value,  -1, 1)
  }
}