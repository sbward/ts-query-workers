package genheap

import (
	"container/heap"
	"sort"
	"testing"
)

func TestSwap(t *testing.T) {
	h := Heap[int]{1, 2}

	h.Swap(0, 1)

	if h[0] != 2 {
		t.Error("item was not swapped")
	}
}

func TestPush(t *testing.T) {
	h := &Heap[int]{4, 5}

	h.Push(6)

	if (*h)[2] != 6 {
		t.Error("item was not pushed")
	}

	if h.Len() != 3 {
		t.Errorf("length should be 3, but got %d", h.Len())
	}
}

func TestPop(t *testing.T) {
	h := &Heap[float32]{1.5, 2.5}

	item := h.Pop()

	if len(*h) != 1 || (*h)[0] != 1.5 {
		t.Error("item was not popped")
	}
	if item.(float32) != 2.5 {
		t.Errorf("wrong item popped: %0.1f", item)
	}
}

func TestHeapPop(t *testing.T) {
	h := &Heap[int]{5, 4, 3, 2, 1}

	heap.Init(h)

	if min := (*h)[0]; min != 1 {
		t.Fatalf("minimum item should be 1, but got %d", min)
	}

	popped := []int{}

	for h.Len() > 0 {
		popped = append(popped, heap.Pop(h).(int))
	}

	if !sort.IntsAreSorted(popped) {
		t.Error("items were popped out of order")
	}
}

func TestHeapRemove(t *testing.T) {
	h := &Heap[float32]{3.5, 2.5, 1.5}

	heap.Init(h)

	heap.Remove(h, 1)

	if min := (*h)[0]; min != 1.5 {
		t.Errorf("minimum item should be 1.5, but got %0.1f", min)
	}

	if last := (*h)[1]; last != 3.5 {
		t.Errorf("remaining item should be 3.5, but got %0.1f", last)
	}

	if h.Len() != 2 {
		t.Errorf("length should be 2, but got %d", h.Len())
	}
}
