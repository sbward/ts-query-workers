SELECT
  time_bucket($1, ts) AS "time",
  min(usage) AS "min_usage",
  max(usage) AS "max_usage"
FROM cpu_usage
WHERE host = $2 AND ts BETWEEN $3 AND $4
GROUP BY time
ORDER BY time