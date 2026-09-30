WITH hours AS (
  SELECT generate_series(TIMESTAMP '2024-02-01 00:00',
                         TIMESTAMP '2024-02-01 05:00',
                         INTERVAL '1 hour') AS hour
),
last_in_hour AS (
  SELECT DISTINCT ON (device, date_trunc('hour', ts))
         device, date_trunc('hour', ts) AS hour, value
  FROM meter
  WHERE value IS NOT NULL
  ORDER BY device, date_trunc('hour', ts), ts DESC
),
grid AS (
  SELECT d.id AS device, h.hour, l.value,
         count(l.value) OVER (PARTITION BY d.id ORDER BY h.hour
                              ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS grp
  FROM devices d
  CROSS JOIN hours h
  LEFT JOIN last_in_hour l ON l.device = d.id AND l.hour = h.hour
)
SELECT device,
       to_char(hour, 'HH24:MI') AS hour,
       max(value) OVER (PARTITION BY device, grp) AS value
FROM grid
ORDER BY device, grid.hour;
