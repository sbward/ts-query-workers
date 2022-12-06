package device

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseHostID returns the host identifier from a string with format "host_000002".
func ParseHostID(host string) (int, error) {
	id, err := strconv.Atoi(strings.TrimPrefix(host, "host_"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse host id: %w", err)
	}
	return id, nil
}
