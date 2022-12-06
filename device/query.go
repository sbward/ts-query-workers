package device

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed min_max_query.sql
var minMaxCPUQuerySQL string

// QuerierCtx is an interface matching several SQL types that provide querying.
type QuerierCtx interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// MinMaxCPUQuery retrieves the min and max CPU usage for a host for every minute in a time range.
type MinMaxCPUQuery struct {
	BucketSize string // Postgres INTERVAL literal, e.g. "1m"
	Hostname   string
	StartTime  time.Time
	EndTime    time.Time
}

// MinMaxBucket is a record of min and max CPU usage during a time bucket.
type MinMaxBucket struct {
	Bucket time.Time
	Min    float32
	Max    float32
}

// QueryStats is a record of the amount of time a query took to run.
type QueryStats struct {
	ExecutionTime time.Duration
	Cost          float32
}

// ExecCtx executes the query with the provided QuerierCtx.
// Unless StatsOnly is true, returns a time series of one-minute MinMaxBuckets for every minute in the time range.
// Also returns QueryStats for the performance of the query.
func (q MinMaxCPUQuery) ExecCtx(ctx context.Context, tx QuerierCtx) ([]MinMaxBucket, QueryStats, error) {
	// Execute the SQL query while measuring execution time.
	start := time.Now()
	rows, err := tx.QueryContext(ctx, minMaxCPUQuerySQL, q.BucketSize, q.Hostname, q.StartTime, q.EndTime)
	stats := QueryStats{
		ExecutionTime: time.Since(start),
	}
	if err != nil {
		return nil, stats, err
	}
	defer rows.Close()

	// Allocate a result slice with capacity for a full result set.
	result := make([]MinMaxBucket, 0, q.maxBuckets())

	// Scan each row into a MinMaxBucket.
	for rows.Next() {
		var bucket MinMaxBucket
		if err := rows.Scan(&bucket.Bucket, &bucket.Min, &bucket.Max); err != nil {
			return nil, stats, err
		}
		result = append(result, bucket)
	}
	return result, stats, nil
}

// MaxBuckets returns the maximum number of 1 minute buckets this query could potentially return.
func (q MinMaxCPUQuery) maxBuckets() int {
	return int(q.EndTime.Sub(q.StartTime) / time.Minute)
}

// pqPlanResult is the result of an EXPLAIN ANALYZE operation with JSON formatting.
type pqPlanResult struct {
	Plan pqPlanNode
}

type pqPlanNode struct {
	NodeType          string  `json:"Node Type"`
	ParallelAware     bool    `json:"Parallel Aware"`
	RelationName      string  `json:"Relation Name"`
	Alias             string  `json:"Alias"`
	StartupCost       float32 `json:"Startup Cost"`
	TotalCost         float32 `json:"Total Cost"`
	PlanRows          int     `json:"Plan Rows"`
	PlanWidth         int     `json:"Plan Width"`
	ActualStartupTime float32 `json:"Actual Startup Time"`
	ActualTotalTime   float32 `json:"Actual Total Time"`
	ActualRows        int     `json:"Actual Rows"`
	ActualLoops       int     `json:"Actual Loops"`
	Plans             []pqPlanNode
}

func (q MinMaxCPUQuery) ExplainAnalyze(ctx context.Context, tx QuerierCtx) (*QueryStats, error) {
	query := "EXPLAIN (ANALYZE, FORMAT JSON) " + minMaxCPUQuerySQL
	rows, err := tx.QueryContext(ctx, query, q.BucketSize, q.Hostname, q.StartTime, q.EndTime)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	result := []pqPlanResult{}
	for rows.Next() {
		var rawPlanJSON string
		if err := rows.Scan(&rawPlanJSON); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if err := json.Unmarshal([]byte(rawPlanJSON), &result); err != nil {
			return nil, fmt.Errorf("parse plan json: %w", err)
		}
	}
	if len(result) != 1 {
		return nil, fmt.Errorf("expected 1 plan result but got %d", len(result))
	}
	stats := &QueryStats{
		ExecutionTime: time.Duration(result[0].Plan.ActualTotalTime * float32(time.Millisecond)),
		Cost:          result[0].Plan.TotalCost,
	}
	return stats, nil
}

func (q MinMaxCPUQuery) String() string {
	return fmt.Sprintf("%s %s %s", q.Hostname, q.StartTime, q.EndTime)
}
