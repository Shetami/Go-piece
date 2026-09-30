WITH flagged AS (
  SELECT device, ts, status,
         CASE WHEN status IS DISTINCT FROM lag(status) OVER w THEN 1 ELSE 0 END AS is_start
  FROM device_status
  WINDOW w AS (PARTITION BY device ORDER BY ts)
),
grouped AS (
  SELECT device, ts, status,
         sum(is_start) OVER (PARTITION BY device ORDER BY ts
                             ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS grp
  FROM flagged
),
intervals AS (
  SELECT device, status, min(ts) AS started_at
  FROM grouped
  GROUP BY device, grp, status
),
bounded AS (
  SELECT device, status, started_at,
         coalesce(lead(started_at) OVER (PARTITION BY device ORDER BY started_at),
                  TIMESTAMP '2024-07-01 12:00') AS ended_at
  FROM intervals
)
SELECT device, status,
       to_char(started_at, 'HH24:MI') AS started_at,
       to_char(ended_at,   'HH24:MI') AS ended_at,
       (extract(epoch FROM ended_at - started_at) / 60)::int AS minutes
FROM bounded
ORDER BY device, bounded.started_at;
