package stats

import (
	"testing"

	"golang.org/x/exp/slices"
)

type step struct {
	Push   int
	Expect []int
}

var steps = []step{
	{-1, []int{-1}},
	{9, []int{-1, 9}},
	{2, []int{2}},
	{8, []int{2, 8}},
	{3, []int{3}},
	{7, []int{3, 7}},
}

func TestMedian(t *testing.T) {
	m := NewMedian[int]()

	for i, op := range steps {
		m.Push(op.Push)

		t.Log(m.low, m.high)

		if median := m.MedianRaw(); slices.Compare(median, op.Expect) != 0 {
			t.Fatalf("step %d failed: expected %v but got %v", i, op.Expect, median)
		}
	}
}
