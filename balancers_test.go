package main

import (
	"fmt"
	"testing"
)

func TestHostIDBalancer(t *testing.T) {
	for numBuckets := 1; numBuckets < 10; numBuckets++ {

		hostBuckets := make(map[string]int)

	hosts:
		for hostID := 0; hostID < 10; hostID++ {

			host := fmt.Sprintf("host_%06d", hostID)

			bucket, err := HostIDBalancer(host, numBuckets)
			if err != nil {
				t.Fatalf("host %s: %s", host, err)
			}

			t.Logf("[%d buckets] %s assigned to bucket %d", numBuckets, host, bucket)

			expect, ok := hostBuckets[host]
			if !ok {
				hostBuckets[host] = bucket
				continue hosts
			}

			if bucket != expect {
				t.Errorf("host %s was assigned bucket %d, then subsequently assigned %d", host, expect, bucket)
			}
		}
	}
}
