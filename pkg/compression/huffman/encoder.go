package huffman

import (
	"errors"
	"fmt"
	"iter"
)

type Encoder[T comparable] struct {
	mapValueToCode map[T]Code
}

func BuildEncoder[T comparable](tree *Tree[T]) (*Encoder[T], error) {
	enc := &Encoder[T]{
		mapValueToCode: make(map[T]Code),
	}
	err := enc.buildEncoder(tree)

	return enc, err
}

func (enc *Encoder[T]) buildEncoder(tree *Tree[T]) error {
	var err error

	for i, node := range tree.Nodes {
		if node.Leaf {
			cc, ok := tree.DetermineCodeFor(i)
			if ok {
				enc.mapValueToCode[node.Value] = cc
			} else {
				err = errors.Join(fmt.Errorf("failed to determine code for leaf node with value %v at index %d", node.Value, i))
			}
		}
	}

	return err
}

// Encode encodes a value into its corresponding Huffman code.
func (enc *Encoder[T]) Encode(value T) (Code, bool) {
	code, ok := enc.mapValueToCode[value]

	return code, ok
}

// EncodeTo encodes a slice of values into their corresponding Huffman codes and writes them to the provided BitsWriter.
func (enc *Encoder[T]) EncodeTo(values []T, w BitsWriter) error {
	for _, value := range values {
		code, ok := enc.Encode(value)
		if !ok {
			return fmt.Errorf("value %v not found in encoder", value)
		}
		err := w.WriteBits(code.Value, code.Bits)
		if err != nil {
			return fmt.Errorf("failed to write bits for value %v", value)
		}
	}

	return nil
}

// EncodeSeqTo encodes a sequence of values into their corresponding Huffman codes and writes them to the provided BitsWriter.
func (enc *Encoder[T]) EncodeSeqTo(values iter.Seq[T], w BitsWriter) error {
	for v := range values {
		code, ok := enc.Encode(v)
		if !ok {
			return fmt.Errorf("value %v not found in encoder", v)
		}
		err := w.WriteBits(code.Value, code.Bits)
		if err != nil {
			return fmt.Errorf("failed to write bits for value %v", v)
		}
	}

	return nil
}

// EncodeSeq2To encodes a sequence of pairs of values into their corresponding Huffman codes and writes them to the provided BitsWriter.
func (enc *Encoder[T]) EncodeSeq2To(values iter.Seq2[int, T], w BitsWriter) error {
	for _, v := range values {
		code, ok := enc.Encode(v)
		if !ok {
			return fmt.Errorf("value %v not found in encoder", v)
		}
		err := w.WriteBits(code.Value, code.Bits)
		if err != nil {
			return fmt.Errorf("failed to write bits for value %v", v)
		}
	}

	return nil
}
