package device

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMinMaxCPUQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("failed to init sqlmock:", err)
	}

	q := &MinMaxCPUQuery{
		BucketSize: "1m",
		Hostname:   "host-001",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Hour),
	}

	expect := []MinMaxBucket{
		{
			Bucket: time.Now().Add(-3 * time.Hour),
			Min:    0.1,
			Max:    0.9,
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(minMaxCPUQuerySQL)).
		WithArgs("1m", q.Hostname, q.StartTime, q.EndTime).
		WillReturnRows(
			sqlmock.NewRows([]string{"time", "min_usage", "max_usage"}).
				AddRow(expect[0].Bucket, expect[0].Min, expect[0].Max),
		)

	results, _, err := q.ExecCtx(context.Background(), db)
	if err != nil {
		t.Fatal("failed to execute query:", err)
	}

	if n := len(results); n != 1 {
		t.Errorf("expected 1 result but got %d", n)
	}
	if actualBucket := results[0].Bucket; !actualBucket.Equal(expect[0].Bucket) {
		t.Errorf("expected result bucket %s but got %s", expect[0].Bucket, actualBucket)
	}
	if min := results[0].Min; min != expect[0].Min {
		t.Errorf("expected min %0.1f but got %0.1f", expect[0].Min, min)
	}
	if max := results[0].Max; max != expect[0].Max {
		t.Errorf("expected max %0.1f but got %0.1f", expect[0].Max, max)
	}
}
