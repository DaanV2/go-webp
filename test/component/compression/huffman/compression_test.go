package huffman_test

import (
	"testing"

	"github.com/daanv2/go-webp/pkg/compression/huffman"
)

func TestHuffmanCompression(t *testing.T) {

}

func testHuffmanCompression[T comparable](t *testing.T, input []T) {
	builder := huffman.NewTreeBuilder[T]()

	builder.Analyse(input)
}
