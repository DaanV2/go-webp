package huffman

import (
	"container/heap"
	"iter"
)

type TreeBuilder[T comparable] struct {
	nodes map[T]treeBuilderNode[T]
}

func NewTreeBuilder[T comparable]() *TreeBuilder[T] {
	return &TreeBuilder[T]{
		nodes: make(map[T]treeBuilderNode[T]),
	}
}

type treeBuilderNode[T comparable] struct {
	frequency int
	node      Node[T]
}

// Analyse takes an input slice of values and counts the frequency of each value, storing the results in the nodes map.
// This method assumes that all inputs values haven't been analysed before, as it will compound on the frequency of the values
func (builder *TreeBuilder[T]) Analyse(input []T) {
	for _, value := range input {
		node, ok := builder.nodes[value]
		if !ok {
			node = treeBuilderNode[T]{
				frequency: 0,
				node: Node[T]{
					LeftIndex:  -1,
					RightIndex: -1,
					Leaf:       true,
					Value:      value,
				},
			}
		}
		node.frequency++
		builder.nodes[value] = node
	}
}

// AnalyseSeq takes an input iterator of values and counts the frequency of each value, storing the results in the nodes map.
// This method assumes that all inputs values haven't been analysed before, as it will compound on the frequency of the values
func (builder *TreeBuilder[T]) AnalyseSeq(input iter.Seq[T]) {
	for value := range input {
		node, ok := builder.nodes[value]
		if !ok {
			node = treeBuilderNode[T]{
				frequency: 0,
				node: Node[T]{
					LeftIndex:  -1,
					RightIndex: -1,
					Leaf:       true,
					Value:      value,
				},
			}
		}
		node.frequency++
		builder.nodes[value] = node
	}
}

// BuildTree builds a Huffman tree from the analysed frequencies using the standard
// greedy merge: repeatedly combine the two lowest-frequency nodes into a new parent
// until a single node remains. The root is reserved at index 0 and filled in once
// the final merge is known, so it never has to move after being written.
func (builder *TreeBuilder[T]) BuildTree() *Tree[T] {
	var zero T

	result := &Tree[T]{
		Nodes: make([]Node[T], 0, 2*len(builder.nodes)),
	}
	if len(builder.nodes) == 0 {
		return result
	}

	// Reserve index 0 for the root node; its children are filled in once the last merge is known.
	result.Nodes = append(result.Nodes, Node[T]{
		LeftIndex:  -1,
		RightIndex: -1,
		Leaf:       false,
		Value:      zero,
	})

	pq := make(priorityQueue, 0, len(builder.nodes))
	for _, leaf := range builder.nodes {
		index := len(result.Nodes)
		result.Nodes = append(result.Nodes, leaf.node)
		pq = append(pq, pqItem{frequency: leaf.frequency, index: index})
	}
	heap.Init(&pq)

	// A single symbol still needs a root pointing at it.
	if pq.Len() == 1 {
		only := heap.Pop(&pq).(pqItem)
		result.Nodes[0].LeftIndex = only.index

		return result
	}

	for pq.Len() > 1 {
		first := heap.Pop(&pq).(pqItem)
		second := heap.Pop(&pq).(pqItem)

		if pq.Len() == 0 {
			result.Nodes[0].LeftIndex = first.index
			result.Nodes[0].RightIndex = second.index

			break
		}

		index := len(result.Nodes)
		result.Nodes = append(result.Nodes, Node[T]{
			LeftIndex:  first.index,
			RightIndex: second.index,
			Leaf:       false,
			Value:      zero,
		})
		heap.Push(&pq, pqItem{frequency: first.frequency + second.frequency, index: index})
	}

	return result
}
