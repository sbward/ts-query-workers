package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/sbward/ts-query-workers/device"
)

const csvTimeFormat = "2006-01-02 15:04:05"

type queryOption func(*device.MinMaxCPUQuery)

// QueriesFromCSV parses MinMaxCPUQueries from a CSV file.
// Options can be provided to modify each query as they are read.
func queriesFromCSV(r *csv.Reader, opts ...queryOption) ([]*device.MinMaxCPUQuery, error) {
	out := make([]*device.MinMaxCPUQuery, 0)

	for {
		query, err := nextQueryFromCSV(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, fmt.Errorf("CSV parsing failed: %w", err)
		}
		for _, option := range opts {
			option(query)
		}
		out = append(out, query)
	}

	return out, nil
}

// NextQuery parses a device.MinMaxCPUQuery from the next CSV record in r.
// When the end of file is reached io.EOF will be returned.
func nextQueryFromCSV(r *csv.Reader) (*device.MinMaxCPUQuery, error) {
	record, err := r.Read()
	if err != nil {
		return nil, err
	}

	// If the record is a header row, skip it and return the first data row.
	if record[0] == "hostname" {
		return nextQueryFromCSV(r)
	}

	// Parse StartTime.
	start, err := time.Parse(csvTimeFormat, record[1])
	if err != nil {
		line, col := r.FieldPos(1)
		return nil, fmt.Errorf("failed to parse StartTime (line %d, column %d): %w", line, col, err)
	}

	// Parse EndTime.
	end, err := time.Parse(csvTimeFormat, record[2])
	if err != nil {
		line, col := r.FieldPos(2)
		return nil, fmt.Errorf("failed to parse EndTime (line %d, column %d): %w", line, col, err)
	}

	query := &device.MinMaxCPUQuery{
		Hostname:  record[0],
		StartTime: start,
		EndTime:   end,
	}

	return query, nil
}
