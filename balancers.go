package main

import (
	"fmt"
	"hash"
	"math/rand"

	"github.com/sbward/ts-query-workers/device"
)

// Balancer accepts a value and bucket count, and returns a bucket index
// representing which bucket the value should be assigned to.
type Balancer func(value any, buckets int) (int, error)

var _ Balancer = RandomBalancer

// RandomBalancer selects random buckets, without taking the value into account.
func RandomBalancer(_ any, buckets int) (int, error) {
	return rand.Int() % buckets, nil
}

var _ Balancer = HostIDBalancer

// HostIDBalancer accepts a string value and parses it as a device host string,
// then assigns a bucket based on the host's integer ID.
func HostIDBalancer(value any, buckets int) (int, error) {
	host, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("value must be string but got %T", value)
	}
	id, err := device.ParseHostID(host)
	if err != nil {
		return 0, err
	}
	return id % buckets, nil
}

// NewHashBalancer returns a Balancer that hashes the string value with the given 32-bit hash function,
// then takes the modulo of the hash sum to select the bucket.
func NewHashBalancer(h hash.Hash32) Balancer {
	return func(value any, buckets int) (int, error) {
		defer h.Reset()
		str, ok := value.(string)
		if !ok {
			return 0, fmt.Errorf("value must be a string but got %T", value)
		}
		h.Write([]byte(str))
		return int(h.Sum32()) % buckets, nil
	}
}

// NewQueryHostnameBalancer returns a Balancer that accepts *device.MinMaxCPUQuery values
// and uses the Hostname of the query as the input value into another Balancer.
func NewQueryHostnameBalancer(next Balancer) Balancer {
	return func(value any, buckets int) (int, error) {
		query, ok := value.(*device.MinMaxCPUQuery)
		if !ok {
			return 0, fmt.Errorf("value must be *device.MinMaxCPUQuery but got %T", value)
		}
		return next(query.Hostname, buckets)
	}
}
