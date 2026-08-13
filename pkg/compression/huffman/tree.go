package huffman

type Tree[T comparable] struct {
	Nodes []Node[T]
}

// DetermineCodeFor determines the code for a given leaf node in the tree.
// It find all the others and builds a path from the root to the leaf node, where going left is 0 and going right is 1.
func (tree *Tree[T]) DetermineCodeFor(leafIndex int) (Code, bool) {
	leafNode := tree.Nodes[leafIndex]
	if !leafNode.Leaf {
		return Code{}, false
	}

	// Determine the path from the root to the leaf node
	path := []int{}
	current := leafIndex
	for current != 0 {
		parentIndex, found := tree.FindParentIndex(current)
		if !found {
			return Code{}, false
		}

		path = append(path, current)
		current = parentIndex
	}

	// Build the code from the path
	code := Code{}
	for i := len(path) - 1; i >= 0; i-- {
		nodeIndex := path[i]
		parentIndex, _ := tree.FindParentIndex(nodeIndex)
		parentNode := tree.Nodes[parentIndex]

		// If the current node is the right child of its parent, append a 1 to the code; otherwise, append a 0
		code = code.WithAppendedBit(parentNode.RightIndex == nodeIndex)
	}
	return code, true
}

func (tree *Tree[T]) FindParentIndex(childIndex int) (int, bool) {
	for i, node := range tree.Nodes {
		if node.LeftIndex == childIndex || node.RightIndex == childIndex {
			return i, true
		}
	}
	return -1, false
}
