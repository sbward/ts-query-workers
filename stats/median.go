package stats

import (
	"container/heap"

	"github.com/sbward/ts-query-workers/genheap"
)

type Divisible interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Median tracks the median in a set of values continuously during aggregation.
// Operations to retrieve the median value take constant time.
type Median[T Divisible] struct {
	// Low is a max-heap implemented using negative numbers in a min-heap.
	low *genheap.Heap[T]

	// High is a normal min-heap.
	high *genheap.Heap[T]
}

func NewMedian[T Divisible]() *Median[T] {
	return &Median[T]{
		low:  &genheap.Heap[T]{},
		high: &genheap.Heap[T]{},
	}
}

// Push adds a value to the set.
func (m *Median[T]) Push(x T) {
	// Add the value to the low heap.
	heap.Push(m.low, -x)

	// Pop the max value from the low heap and push it onto the high heap.
	heap.Push(m.high, -heap.Pop(m.low).(T))

	// If the high heap is larger, pop the min value and push it onto the low heap.
	if m.high.Len() > m.low.Len() {
		heap.Push(m.low, -heap.Pop(m.high).(T))
	}
}

// MedianRaw returns the median, or two medians if the size of the set is even.
// If the set is empty, nil is returned.
func (m *Median[T]) MedianRaw() []T {
	if m.low.Len() == 0 && m.high.Len() == 0 {
		return nil
	}
	if m.low.Len() == m.high.Len() {
		return []T{-(*m.low)[0], (*m.high)[0]}
	}
	return []T{-(*m.low)[0]}
}

// Median returns the median value of the set.
func (m *Median[T]) Median() float64 {
	v := m.MedianRaw()
	switch len(v) {
	case 0:
		return 0
	case 1:
		return float64(v[0])
	}
	return (float64(v[0]) + float64(v[1])) / 2
}
