WITH local AS (
  -- timestamptz AT TIME ZONE 'пояс' → timestamp «как на часах в Москве»
  SELECT created_at AT TIME ZONE 'Europe/Moscow' AS ts
  FROM tickets
),
march AS (
  SELECT extract(hour FROM ts)   AS h,
         extract(isodow FROM ts) AS dow
  FROM local
  WHERE ts >= '2024-03-01' AND ts < '2024-04-01'     -- границы месяца — тоже по Москве
),
s AS (
  SELECT CASE WHEN h < 8 THEN 1 WHEN h < 16 THEN 2 ELSE 3 END AS k, dow
  FROM march
)
SELECT CASE k WHEN 1 THEN 'ночь' WHEN 2 THEN 'день' ELSE 'вечер' END AS shift,
       count(*) FILTER (WHERE dow < 6)  AS weekdays,
       count(*) FILTER (WHERE dow >= 6) AS weekends
FROM s
GROUP BY k
ORDER BY k;
