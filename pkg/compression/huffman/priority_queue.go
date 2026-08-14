package huffman

// pqItem references a node already placed in Tree.Nodes, ordered by its frequency for the merge step.
type pqItem struct {
	frequency int
	index     int
}

// priorityQueue is a min-heap of pqItem ordered by frequency, used to repeatedly find the two lowest-frequency nodes.
type priorityQueue []pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].frequency < pq[j].frequency }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x any)        { *pq = append(*pq, x.(pqItem)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]

	return item
}
