package main

import "fmt"

// Buckets divides a slice of values into N buckets with the provided Balancer.
func Buckets[T any](values []T, n int, balance Balancer) ([][]T, error) {
	buckets := make([][]T, n)

	for _, value := range values {
		i, err := balance(value, n)
		if err != nil {
			return nil, fmt.Errorf("balancer failed to assign value '%v': %w", value, err)
		}
		if buckets[i] == nil {
			buckets[i] = []T{}
		}
		buckets[i] = append(buckets[i], value)
	}

	return buckets, nil
}
