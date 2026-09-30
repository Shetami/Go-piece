WITH events AS (
  SELECT stream_id, started_at AS ts, 1 AS delta FROM views
  UNION ALL
  SELECT v.stream_id, coalesce(v.ended_at, s.ended_at), -1
  FROM views v
  JOIN streams s ON s.id = v.stream_id
),
running AS (
  SELECT stream_id, ts,
         sum(delta) OVER (PARTITION BY stream_id ORDER BY ts, delta
                          ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS online
  FROM events
),
peaks AS (
  SELECT stream_id, online, ts,
         row_number() OVER (PARTITION BY stream_id ORDER BY online DESC, ts) AS rn
  FROM running
)
SELECT s.title,
       coalesce(p.online, 0)      AS peak,
       to_char(p.ts, 'HH24:MI')   AS peak_at
FROM streams s
LEFT JOIN peaks p ON p.stream_id = s.id AND p.rn = 1
ORDER BY s.title;
