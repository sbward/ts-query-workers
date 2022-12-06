package genheap

import (
	"container/heap"
)

type Comparable interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Heap is a generic heap that implements the "container/heap" Interface.
type Heap[T Comparable] []T

var _ heap.Interface = (*Heap[int])(nil)

func (h Heap[T]) Len() int {
	return len(h)
}

func (h Heap[T]) Less(i, j int) bool {
	return h[i] < h[j]
}

func (h Heap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Heap[T]) Push(x any) {
	*h = append(*h, x.(T))
}

func (h *Heap[T]) Pop() any {
	n := len(*h)
	old := *h
	*h = old[0 : n-1]
	return old[n-1]
}
