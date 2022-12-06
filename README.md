# ts-query-workers

Command line application for measuring query performance against a device CPU dataset.

## Installation

### Source

```bash
go install github.com/sbward/ts-query-workers
```

### Docker

```bash
docker pull sbward/ts-query-workers:latest
```

## Options

```bash
$ ts-query-workers [-c N] [-db CONNECTION_STRING] [CSV_FILENAME]
```

| Option                  | Usage                                                                                                     |
| ----------------------- | --------------------------------------------------------------------------------------------------------- |
| `-c N`                  | Set number of concurrent workers. Defaults to 10.                                                         |
| `-db CONNECTION_STRING` | Database connection string. Can also be set via the `DB` environment variable.                            |
| `CSV_FILENAME`          | Filename of a CSV file containing query specifications. Can be omitted if a CSV file is piped to `stdin`. |

| Env Var | Usage                                                              |
| ------- | ------------------------------------------------------------------ |
| `DB`    | Database connection string. If set, the `-db` flag can be omitted. |

## Example Usage

```bash
export DB=postgres://USER:PASSWORD@HOST:PORT/DB

ts-query-workers -c 10 datafiles/query_params.csv
```

### Pipe input to Docker

```bash
export DB=postgres://USER:PASSWORD@HOST:PORT/DB

cat datafiles/query_params.csv | docker run -e DB -i sbward/ts-query-workers
```

### Run with custom concurrency

```bash
export DB=postgres://USER:PASSWORD@HOST:PORT/DB

cat datafiles/query_params.csv | docker run -e DB -i sbward/ts-query-workers -c 10
```

### Pass DB option as a flag

Native

```bash
ts-query-workers -db "postgres://USER:PASSWORD@HOST:PORT/DB" datafiles/query_params.csv
```

Docker

```bash
cat datafiles/query_params.csv | docker run -e DB -i sbward/ts-query-workers -db "postgres://USER:PASSWORD@HOST:PORT/DB"
```
