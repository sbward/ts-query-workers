package stats

// Aggregator continuously tracks the count, total, minimum value, maximum value, average value
// and median value of a set of values as they are passed to the aggregator over time.
type Aggregator[T Divisible] struct {
	Count int
	Total T
	Min   T
	Max   T
	Avg   float64
	Med   float64

	median *Median[T]
}

func NewAggregator[T Divisible]() *Aggregator[T] {
	return &Aggregator[T]{
		median: NewMedian[T](),
	}
}

func (a *Aggregator[T]) Push(x T) {
	a.Total += x

	if a.Min > x || a.Count == 0 {
		a.Min = x
	}

	if a.Max < x || a.Count == 0 {
		a.Max = x
	}

	a.Count++

	a.Avg = float64(a.Total) / float64(a.Count)

	a.median.Push(x)

	a.Med = a.median.Median()
}
