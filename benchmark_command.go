package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"os"
	"sync"
	"time"

	"github.com/sbward/ts-query-workers/device"
	"github.com/sbward/ts-query-workers/stats"
)

// BenchmarkCommand reads a CSV file of query specifications and executes the queries across a concurrent worker pool.
// After execution completes, a report is printed to stdout with execution statistics.
// All configuration options are required.
type BenchmarkCommand struct {
	// CSV is the source file of query specifications to benchmark.
	CSV *os.File

	// DB is the SQL database to execute queries against.
	DB *sql.DB

	// Concurrency is the number of workers to start.
	Concurrency int

	workers *sync.WaitGroup
}

func (c *BenchmarkCommand) Exec(ctx context.Context) error {
	queries, err := queriesFromCSV(
		csv.NewReader(c.CSV),
		func(q *device.MinMaxCPUQuery) { q.BucketSize = "1m" },
	)
	c.CSV.Close()
	if err != nil {
		return err
	}

	// Assign queries to workers by hashing the query hostname.

	balancer := NewQueryHostnameBalancer(NewHashBalancer(fnv.New32()))

	// Group queries into N buckets, where N is the concurrency.

	buckets, err := Buckets(queries, *concurrency, balancer)
	if err != nil {
		return fmt.Errorf("failed to assign queries to buckets: %w", err)
	}

	fmt.Printf("Benchmarking %d queries across %d workers...\n", len(queries), *concurrency)

	// Launch one worker per bucket. Fan-in results to a single channel.

	results := make(chan *QueryExecutionResult)

	c.workers = &sync.WaitGroup{}

	for bucket, queries := range buckets {
		c.workers.Add(1)
		go c.queryWorker(ctx, bucket, queries, results)
	}

	// Wait for all workers to complete, then close the results channel.

	go func() {
		c.workers.Wait()
		close(results)
	}()

	// Aggregate stats received on the results channel then print a report.

	stats := aggregateStats(*concurrency, results)

	fmt.Println()

	fmt.Println("Execution time:")
	fmt.Println()
	fmt.Println(stats.ExecutionTimeTable())

	fmt.Println()

	fmt.Println("Execution cost:")
	fmt.Println()
	fmt.Println(stats.CostTable())

	return nil
}

// QueryWorker executes a series of queries and sends the results to a result channel.
func (b *BenchmarkCommand) queryWorker(ctx context.Context, bucket int, queries []*device.MinMaxCPUQuery, results chan<- *QueryExecutionResult) {
	defer b.workers.Done()

	for _, query := range queries {
		stats, err := query.ExplainAnalyze(ctx, b.DB)

		result := &QueryExecutionResult{
			Query:  query,
			Stats:  stats,
			Error:  err,
			Worker: bucket,
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case results <- result:
		}
	}
}

// QueryExecutionResult is a report of the result of executing a MinMaxCPUQuery by a Worker.
type QueryExecutionResult struct {
	Query   *device.MinMaxCPUQuery
	Results []device.MinMaxBucket
	Stats   *device.QueryStats
	Error   error
	Worker  int
}

func (q *QueryExecutionResult) String() string {
	start := q.Query.StartTime.Format(csvTimeFormat)
	end := q.Query.EndTime.Format(csvTimeFormat)
	if q.Error != nil {
		return fmt.Sprintf("❌ %s, %s, %s: %s", q.Query.Hostname, start, end, q.Error)
	}
	return fmt.Sprintf("✅ %s, %s, %s -> %s, worker %d", q.Query.Hostname, start, end, q.Stats.ExecutionTime.Round(time.Microsecond), q.Worker)
}

type BenchmarkStats struct {
	ExecTimeGlobal   *stats.Aggregator[time.Duration]
	ExecTimeByWorker []*stats.Aggregator[time.Duration]
	CostGlobal       *stats.Aggregator[float32]
	CostByWorker     []*stats.Aggregator[float32]
}

// AggregateStats aggregates query statistics from the results channel until it closes, then returns the final BenchmarkStats.
func aggregateStats(numWorkers int, results <-chan *QueryExecutionResult) BenchmarkStats {
	// Collects stats for all queries across all workers.
	execTimeGlobal := stats.NewAggregator[time.Duration]()
	costGlobal := stats.NewAggregator[float32]()

	// Collects stats for each worker.
	execTimeWorkers := make([]*stats.Aggregator[time.Duration], numWorkers)
	costWorkers := make([]*stats.Aggregator[float32], numWorkers)

	for result := range results {
		fmt.Println(result)

		execTimeGlobal.Push(result.Stats.ExecutionTime)
		costGlobal.Push(result.Stats.Cost)

		// Execution time per worker

		workerTime := execTimeWorkers[result.Worker]
		if workerTime == nil {
			workerTime = stats.NewAggregator[time.Duration]()
			execTimeWorkers[result.Worker] = workerTime
		}
		workerTime.Push(result.Stats.ExecutionTime)

		// Cost per worker

		workerCost := costWorkers[result.Worker]
		if workerCost == nil {
			workerCost = stats.NewAggregator[float32]()
			costWorkers[result.Worker] = workerCost
		}
		workerCost.Push(result.Stats.Cost)
	}

	return BenchmarkStats{execTimeGlobal, execTimeWorkers, costGlobal, costWorkers}
}

// TableString returns a human-readable table of the benchmark results.
func (b BenchmarkStats) ExecutionTimeTable() string {
	table := "| Worker | Queries | Total | Minimum | Maximum | Average |  Median |\n"
	table += "|--------|---------|-------|---------|---------|---------|---------|\n"
	table += b.tableLineExecTime("ALL", b.ExecTimeGlobal)

	for worker, wstats := range b.ExecTimeByWorker {
		if wstats == nil {
			wstats = &stats.Aggregator[time.Duration]{}
		}
		table += b.tableLineExecTime(fmt.Sprint(worker), wstats)
	}

	return table
}

func (b BenchmarkStats) CostTable() string {
	table := "| Worker | Queries | Total | Minimum | Maximum | Average |  Median |\n"
	table += "|--------|---------|-------|---------|---------|---------|---------|\n"
	table += b.tableLineCost("ALL", b.CostGlobal)

	for worker, wstats := range b.CostByWorker {
		if wstats == nil {
			wstats = &stats.Aggregator[float32]{}
		}
		table += b.tableLineCost(fmt.Sprint(worker), wstats)
	}

	return table
}

func (b BenchmarkStats) tableLineExecTime(worker string, agg *stats.Aggregator[time.Duration]) string {
	return fmt.Sprintf(
		"| %6s | %7d | %5s | %7s | %7s | %7s | %7s |\n",
		worker,
		agg.Count,
		agg.Total.Round(time.Millisecond),
		agg.Min.Round(time.Microsecond),
		agg.Max.Round(time.Microsecond),
		time.Duration(agg.Avg).Round(time.Microsecond),
		time.Duration(agg.Med).Round(time.Microsecond),
	)
}

func (b BenchmarkStats) tableLineCost(worker string, agg *stats.Aggregator[float32]) string {
	return fmt.Sprintf(
		"| %6s | %7d | %5d | %7d | %7d | %7d | %7d |\n",
		worker,
		agg.Count,
		int(agg.Total),
		int(agg.Min),
		int(agg.Max),
		int(agg.Avg),
		int(agg.Med),
	)
}
