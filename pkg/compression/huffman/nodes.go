package huffman

type Node[T comparable] struct {
	LeftIndex  int  // Going to the left child node index in the Nodes slice, means 0
	RightIndex int  // Going to the right child node index in the Nodes slice, means 1
	Leaf       bool // True if this node is a leaf node, false otherwise
	Value      T    // The value of the leaf node, only valid if Leaf is true
}
