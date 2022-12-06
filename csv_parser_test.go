package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"testing"

	"github.com/sbward/ts-query-workers/device"
)

//go:embed datafiles/query_params.csv
var csvSampleData string

// The number of records in the sample data file to expect
const csvSampleDataRecords = 200

func TestCSVParser(t *testing.T) {
	queries, err := queriesFromCSV(csv.NewReader(bytes.NewBufferString(csvSampleData)))
	if err != nil {
		t.Fatal(err)
	}
	if len(queries) != csvSampleDataRecords {
		t.Errorf("expected %d records to be parsed, but got %d", csvSampleDataRecords, len(queries))
	}
	for i, query := range queries {
		_, err = device.ParseHostID(query.Hostname)
		if err != nil {
			t.Errorf("query %d failed hostname check: %s", i, err)
		}
		t.Log(query.Hostname, query.StartTime, query.EndTime)
	}
}
