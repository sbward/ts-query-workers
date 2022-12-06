package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	concurrency = flag.Int("c", 5, "number of concurrent workers (defaults to 5)")
	connStr     = flag.String("db", os.Getenv("DB"), "database connection string (defaults to DB environment variable)")
)

func main() {
	cmd, err := NewCommandFromCLI(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = cmd.Exec(ctx); err != nil {
		log.Fatal(err)
	}
}

// NewCommandFromCLI reads input and configuration for a BenchmarkCommand from flags and stdin.
func NewCommandFromCLI(args []string) (*BenchmarkCommand, error) {
	flag.Parse()

	f, err := getInputFile()
	if err != nil {
		return nil, err
	}

	connStr, err := getConnStr()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	cmd := &BenchmarkCommand{
		CSV:         f,
		DB:          db,
		Concurrency: *concurrency,
	}

	return cmd, nil
}

// GetInputFile first checks if a file is being provided on Stdin,
// then falls back to expecting a filename argument.
func getInputFile() (*os.File, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Mode()&fs.ModeNamedPipe != 0 || stat.Size() > 0 {
		return os.Stdin, nil
	}

	csvFileName := flag.Arg(0)
	if csvFileName == "" {
		return nil, errors.New("must provide filename argument or stdin")
	}

	return os.Open(csvFileName)
}

// GetConnStr retrieves the DB connection string argument, which must be provided by a flag or env var.
func getConnStr() (string, error) {
	if *connStr == "" {
		return "", errors.New("must provide a database connection string through the -db flag or DB environment variable")
	}
	return *connStr, nil
}
