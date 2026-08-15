package huffman

import "errors"

var (
	ErrNeedNodes = errors.New("tree needs atleast a root node and a leaf node")
)

type Decoder[T comparable] struct {
	tree *Tree[T]
}

func NewDecoder[T comparable](tree *Tree[T]) (*Decoder[T], error) {
	if tree == nil {
		return nil, errors.New("tree cannot be nil")
	}
	// Needs atleast the root node + a value node
	if len(tree.Nodes) < TREE_MINIMUM_NODES {
		return nil, ErrNeedNodes
	}

	return &Decoder[T]{
		tree: tree,
	}, nil
}

func (dec *Decoder[T]) DecodeNext(reader BitReader) (T, error) {
	var result T
	if len(dec.tree.Nodes) < TREE_MINIMUM_NODES {
		return result, ErrNeedNodes
	}

	node := dec.tree.Nodes[0]
	for {
		if node.Leaf {
			return node.Value, nil
		}
		bit, err := reader.ReadBit()
		if err != nil {
			return result, err
		}
		// if true, then
		if bit {
			node = dec.tree.Nodes[node.RightIndex]
		} else {
			node = dec.tree.Nodes[node.LeftIndex]
		}
	}

}

func (dec *Decoder[T]) DecodeN(reader BitReader, n int) ([]T, error) {
	result := make([]T, 0, n)
	for range n {
		item, err := dec.DecodeNext(reader)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}
